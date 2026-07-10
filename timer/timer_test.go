package timer

import (
	"context"
	"encoding/json"
	"sync"
	"testing"
	"time"

	"github.com/redis/go-redis/v9"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// 构造本机 Redis 客户端供集成测试（失败时由调用方跳过）。
func setupTestRedis(t *testing.T) redis.UniversalClient {
	// 可以使用环境变量指定测试 Redis 地址
	// 如果没有设置，可以使用内存 Redis 或者跳过测试
	client := redis.NewClient(&redis.Options{
		Addr:     "127.0.0.1:6379",
		Password: "", // 无密码
		DB:       15, // 使用 DB 15 进行测试
	})

	ctx := context.Background()
	if err := client.Ping(ctx).Err(); err != nil {
		t.Skipf("Redis not available, skipping test: %v", err)
	}

	// 清理测试数据
	client.FlushDB(ctx)

	return client
}

func TestNewRedisTimer(t *testing.T) {
	tests := []struct {
		name    string
		config  Config
		wantErr bool
	}{
		{
			name: "valid config",
			config: Config{
				RedisClient: setupTestRedis(t),
				KeyFormat:   "test:%d:timers",
				GameID:      1,
			},
			wantErr: false,
		},
		{
			name: "missing redis client",
			config: Config{
				RedisClient: nil,
			},
			wantErr: true,
		},
		{
			name: "default values",
			config: Config{
				RedisClient: setupTestRedis(t),
				GameID:      1,
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			rt, err := NewRedisTimer(tt.config)
			if tt.wantErr {
				assert.Error(t, err)
				assert.Nil(t, rt)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, rt)
				if rt != nil {
					// 检查默认值
					if tt.config.ScanInterval <= 0 {
						assert.Equal(t, 1*time.Second, rt.config.ScanInterval)
					}
					if tt.config.BatchSize <= 0 {
						assert.Equal(t, 100, rt.config.BatchSize)
					}
					if tt.config.KeyFormat == "" {
						assert.Equal(t, "game:%d:table_timers", rt.config.KeyFormat)
					}
				}
			}
		})
	}
}

func TestRedisTimer_Start(t *testing.T) {
	client := setupTestRedis(t)
	defer client.Close()

	rt, err := NewRedisTimer(Config{
		RedisClient: client,
		KeyFormat:   "test:%d:timers",
		GameID:      1,
	})
	require.NoError(t, err)

	t.Run("start timer", func(t *testing.T) {
		err := rt.Start(100, "timer1", 5*time.Second)
		assert.NoError(t, err)

		// 验证计时器已添加到 Redis
		ctx := context.Background()
		key := rt.getTimerKey()
		count, err := client.ZCard(ctx, key).Result()
		assert.NoError(t, err)
		assert.Equal(t, int64(1), count)
	})

	t.Run("start timer with zero duration", func(t *testing.T) {
		err := rt.Start(100, "timer2", 0)
		assert.NoError(t, err)

		// 验证没有添加计时器
		ctx := context.Background()
		key := rt.getTimerKey()
		count, err := client.ZCard(ctx, key).Result()
		assert.NoError(t, err)
		assert.Equal(t, int64(1), count) // 仍然是之前的 1 个
	})

	t.Run("replace existing timer", func(t *testing.T) {
		// 先添加一个计时器
		err := rt.Start(200, "timer3", 10*time.Second)
		assert.NoError(t, err)

		// 再次添加相同 tableId 和 timerId 的计时器
		err = rt.Start(200, "timer3", 20*time.Second)
		assert.NoError(t, err)

		// 验证只有一个计时器（旧的被替换）
		ctx := context.Background()
		key := rt.getTimerKey()
		count, err := client.ZCard(ctx, key).Result()
		assert.NoError(t, err)
		assert.Equal(t, int64(2), count) // 100:timer1 和 200:timer3
	})
}

func TestRedisTimer_Stop(t *testing.T) {
	client := setupTestRedis(t)
	defer client.Close()

	rt, err := NewRedisTimer(Config{
		RedisClient: client,
		KeyFormat:   "test:%d:timers",
		GameID:      1,
	})
	require.NoError(t, err)

	// 先添加一些计时器
	err = rt.Start(100, "timer1", 5*time.Second)
	require.NoError(t, err)
	err = rt.Start(200, "timer2", 10*time.Second)
	require.NoError(t, err)

	t.Run("stop existing timer", func(t *testing.T) {
		err := rt.Stop(100, "timer1")
		assert.NoError(t, err)

		// 验证计时器已删除
		ctx := context.Background()
		key := rt.getTimerKey()
		count, err := client.ZCard(ctx, key).Result()
		assert.NoError(t, err)
		assert.Equal(t, int64(1), count) // 只剩下 timer2
	})

	t.Run("stop non-existing timer", func(t *testing.T) {
		err := rt.Stop(999, "not-exist")
		assert.NoError(t, err) // 应该不报错
	})
}

func TestRedisTimer_Listen(t *testing.T) {
	client := setupTestRedis(t)
	defer client.Close()

	rt, err := NewRedisTimer(Config{
		RedisClient:  client,
		KeyFormat:    "test:%d:timers",
		GameID:       1,
		ScanInterval: 100 * time.Millisecond, // 快速扫描用于测试
		BatchSize:    10,
	})
	require.NoError(t, err)

	t.Run("listen expired timers", func(t *testing.T) {
		var wg sync.WaitGroup
		var mu sync.Mutex
		expiredTimers := make([]Message, 0)

		callback := func(tableID int, timerID string, _ time.Time) {
			mu.Lock()
			expiredTimers = append(expiredTimers, Message{
				TableID: tableID,
				TimerID: timerID,
			})
			mu.Unlock()
			wg.Done()
		}

		ctx, cancel := context.WithCancel(context.Background())
		defer cancel()

		// 启动监听
		go func() {
			_ = rt.Listen(ctx, callback)
		}()

		// 添加一个即将到期的计时器
		wg.Add(1)
		err := rt.Start(100, "timer1", 200*time.Millisecond)
		require.NoError(t, err)

		// 添加另一个计时器
		wg.Add(1)
		err = rt.Start(200, "timer2", 300*time.Millisecond)
		require.NoError(t, err)

		// 等待计时器到期
		done := make(chan struct{})
		go func() {
			wg.Wait()
			close(done)
		}()

		select {
		case <-done:
			// 验证回调被调用
			mu.Lock()
			assert.GreaterOrEqual(t, len(expiredTimers), 2)
			// 验证包含我们添加的计时器
			found1, found2 := false, false
			for _, tm := range expiredTimers {
				if tm.TableID == 100 && tm.TimerID == "timer1" {
					found1 = true
				}
				if tm.TableID == 200 && tm.TimerID == "timer2" {
					found2 = true
				}
			}
			assert.True(t, found1, "timer1 should be expired")
			assert.True(t, found2, "timer2 should be expired")
			mu.Unlock()
		case <-time.After(2 * time.Second):
			t.Fatal("timers did not expire in time")
		}
	})

	t.Run("listen with context cancellation", func(t *testing.T) {
		ctx, cancel := context.WithCancel(context.Background())

		done := make(chan struct{})
		go func() {
			err := rt.Listen(ctx, nil)
			assert.NoError(t, err)
			close(done)
		}()

		// 取消上下文
		cancel()

		select {
		case <-done:
			// 成功退出
		case <-time.After(1 * time.Second):
			t.Fatal("Listen did not stop after context cancellation")
		}
	})
}

func TestRedisTimer_Integration(t *testing.T) {
	client := setupTestRedis(t)
	defer client.Close()

	rt, err := NewRedisTimer(Config{
		RedisClient:  client,
		KeyFormat:    "test:%d:timers",
		GameID:       1,
		ScanInterval: 50 * time.Millisecond,
		BatchSize:    10,
	})
	require.NoError(t, err)

	// 集成测试：完整的启动、监听、停止流程
	var wg sync.WaitGroup
	var mu sync.Mutex
	expiredCount := 0

	callback := func(_ int, _ string, _ time.Time) {
		mu.Lock()
		expiredCount++
		mu.Unlock()
		wg.Done()
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	// 启动监听
	go func() {
		_ = rt.Listen(ctx, callback)
	}()

	// 添加多个计时器
	timers := []struct {
		tableID int
		timerID string
		delay   time.Duration
	}{
		{100, "timer1", 200 * time.Millisecond},
		{200, "timer2", 300 * time.Millisecond},
		{300, "timer3", 400 * time.Millisecond},
	}

	for _, tm := range timers {
		wg.Add(1)
		err = rt.Start(tm.tableID, tm.timerID, tm.delay)
		require.NoError(t, err)
	}

	// 在第一个计时器到期前停止其中一个
	time.Sleep(100 * time.Millisecond)
	err = rt.Stop(200, "timer2")
	require.NoError(t, err)
	wg.Done() // 手动减少计数，因为这个计时器被停止了

	// 等待所有计时器到期
	done := make(chan struct{})
	go func() {
		wg.Wait()
		close(done)
	}()

	select {
	case <-done:
		mu.Lock()
		// 应该只有 2 个计时器到期（timer1 和 timer3，timer2 被停止了）
		assert.Equal(t, 2, expiredCount)
		mu.Unlock()
	case <-time.After(2 * time.Second):
		t.Fatal("timers did not expire in time")
	}
}

// 计时消息 JSON 往返。
func TestTimerMessage_Marshal(t *testing.T) {
	tm := Message{
		TableID: 100,
		TimerID: "test-timer",
	}

	data, err := json.Marshal(tm)
	assert.NoError(t, err)
	assert.NotEmpty(t, data)

	var tm2 Message
	err = json.Unmarshal(data, &tm2)
	assert.NoError(t, err)
	assert.Equal(t, tm.TableID, tm2.TableID)
	assert.Equal(t, tm.TimerID, tm2.TimerID)
}

func TestRedisTimer(t *testing.T) {
	client := setupTestRedis(t)
	defer client.Close()

	rt, err := NewRedisTimer(Config{
		RedisClient:  client,
		KeyFormat:    "test:%d:timers",
		GameID:       1,
		ScanInterval: time.Millisecond * 100,
	})
	require.NoError(t, err)

	go func() {
		if listenErr := rt.Listen(context.Background(), func(tableId int, timerId string, sendTime time.Time) {
			t.Logf("【1】 %v timer expired: tableId = %v, timerId = %v, cost: %v", time.Now(), tableId, timerId, time.Since(sendTime))
		}); listenErr != nil {
			t.Errorf("Listen failed: %v", listenErr)
		}
	}()

	go func() {
		if listenErr := rt.Listen(context.Background(), func(tableId int, timerId string, sendTime time.Time) {
			t.Logf("【2】 %v timer expired: tableId = %v, timerId = %v, cost: %v", time.Now(), tableId, timerId, time.Since(sendTime))
		}); listenErr != nil {
			t.Errorf("Listen failed: %v", listenErr)
		}
	}()
	err = rt.Start(100, "timer1s", 1*time.Second)
	require.NoError(t, err)

	err = rt.Start(100, "timer2s", 2*time.Second)
	require.NoError(t, err)

	err = rt.Start(200, "timer5s", 5*time.Second)
	require.NoError(t, err)
	err = rt.Start(200, "timer6s", 6*time.Second)
	require.NoError(t, err)

	err = rt.Start(300, "timer10s", 10*time.Second)
	require.NoError(t, err)
	err = rt.Start(300, "timer11s", 11*time.Second)
	require.NoError(t, err)

	err = rt.Start(400, "timer15s", 15*time.Second)
	require.NoError(t, err)
	err = rt.Start(400, "timer16s", 16*time.Second)
	require.NoError(t, err)

	err = rt.Start(500, "timer20s", 20*time.Second)
	require.NoError(t, err)
	err = rt.Start(500, "timer21s", 21*time.Second)
	require.NoError(t, err)

	err = rt.Start(600, "timer60s", 60*time.Second)
	require.NoError(t, err)
	err = rt.Start(600, "timer61s", 61*time.Second)
	require.NoError(t, err)

	time.Sleep(time.Second * 62)
}
