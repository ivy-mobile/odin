package userflow

import (
	"context"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/ivy-mobile/odin/envelope"
)

// mockMessage 创建测试用消息
func mockMessage(userID int64) *envelope.InputMessage {
	return &envelope.InputMessage{
		Header: &envelope.Header{
			Uid: userID,
		},
		Route: "test.route",
	}
}

// TestNew 测试创建管理器
func TestNew(t *testing.T) {
	t.Run("valid options", func(t *testing.T) {
		handler := func(_ context.Context, _ *envelope.InputMessage) {}
		flow, err := New(handler)
		if err != nil {
			t.Fatalf("expected no error, got %v", err)
		}
		if flow == nil {
			t.Fatal("expected flow, got nil")
		}
		defer flow.Close()
	})

	t.Run("invalid options", func(t *testing.T) {
		handler := func(_ context.Context, _ *envelope.InputMessage) {}
		_, err := New(handler, WithQueueSize(-1))
		if err == nil {
			t.Fatal("expected error for invalid options")
		}
	})

	t.Run("nil handler", func(t *testing.T) {
		_, err := New(nil)
		if err == nil {
			t.Fatal("expected error for nil handler")
		}
	})
}

// TestSubmit 测试提交事件
func TestSubmit(t *testing.T) {
	t.Run("single user single event", func(t *testing.T) {
		var processed atomic.Int32
		handler := func(_ context.Context, _ *envelope.InputMessage) {
			processed.Add(1)
		}

		flow, _ := New(handler)
		defer flow.Close()

		msg := mockMessage(1)
		err := flow.Submit(msg)
		if err != nil {
			t.Fatalf("expected no error, got %v", err)
		}

		// 等待处理完成
		time.Sleep(100 * time.Millisecond)
		if processed.Load() != 1 {
			t.Errorf("expected 1 event processed, got %d", processed.Load())
		}
	})

	t.Run("multiple users", func(t *testing.T) {
		var processed atomic.Int32
		handler := func(_ context.Context, _ *envelope.InputMessage) {
			processed.Add(1)
		}

		flow, _ := New(handler)
		defer flow.Close()

		// 提交 3 个用户的事件
		for i := int64(1); i <= 3; i++ {
			msg := mockMessage(i)
			if err := flow.Submit(msg); err != nil {
				t.Fatalf("expected no error, got %v", err)
			}
		}

		time.Sleep(100 * time.Millisecond)
		if processed.Load() != 3 {
			t.Errorf("expected 3 events processed, got %d", processed.Load())
		}
		if flow.ActiveUserCount() != 3 {
			t.Errorf("expected 3 active users, got %d", flow.ActiveUserCount())
		}
	})

	t.Run("order preservation per user", func(t *testing.T) {
		var mu sync.Mutex
		results := make(map[int64][]string)

		handler := func(_ context.Context, msg *envelope.InputMessage) {
			time.Sleep(10 * time.Millisecond) // 模拟处理时间
			mu.Lock()
			userID := msg.GetHeader().GetUid()
			results[userID] = append(results[userID], msg.GetRoute())
			mu.Unlock()
		}

		flow, _ := New(handler)
		defer flow.Close()

		// 用户1提交3个事件
		routes := []string{"route.A", "route.B", "route.C"}
		for _, route := range routes {
			msg := mockMessage(1)
			msg.Route = route
			_ = flow.Submit(msg)
		}

		time.Sleep(200 * time.Millisecond)

		mu.Lock()
		defer mu.Unlock()
		if len(results[1]) != 3 {
			t.Fatalf("expected 3 events, got %d", len(results[1]))
		}
		// 验证顺序
		for i, route := range results[1] {
			if route != routes[i] {
				t.Errorf("expected %s at position %d, got %s", routes[i], i, route)
			}
		}
	})

	t.Run("invalid user id", func(t *testing.T) {
		handler := func(_ context.Context, _ *envelope.InputMessage) {}
		flow, _ := New(handler)
		defer flow.Close()

		msg := mockMessage(0) // userID = 0
		err := flow.Submit(msg)
		if err == nil {
			t.Error("expected error for invalid user id")
		}
	})
}

// TestRateLimit 测试限流
func TestRateLimit(t *testing.T) {
	handler := func(_ context.Context, _ *envelope.InputMessage) {}

	flow, _ := New(handler,
		WithRateLimit(2), // 每秒2个
		WithRateBurst(2), // 突发2个
	)
	defer flow.Close()

	var rateLimitErrors int
	// 快速提交5个事件，前2个应该成功，后3个应该被限流
	for i := 0; i < 5; i++ {
		msg := mockMessage(1)
		err := flow.Submit(msg)
		if err == ErrRateLimited {
			rateLimitErrors++
		}
	}

	if rateLimitErrors != 3 {
		t.Errorf("expected 3 rate limit errors, got %d", rateLimitErrors)
	}
}

// TestQueueFull 测试队列满
func TestQueueFull(t *testing.T) {
	var mu sync.Mutex
	var processing bool

	handler := func(_ context.Context, _ *envelope.InputMessage) {
		mu.Lock()
		for processing {
			mu.Unlock()
			time.Sleep(10 * time.Millisecond)
			mu.Lock()
		}
		mu.Unlock()
	}

	flow, _ := New(handler,
		WithQueueSize(2),
		WithRateLimit(1000), // 高限流，避免干扰
	)
	defer flow.Close()

	// 阻塞处理
	mu.Lock()
	processing = true
	mu.Unlock()

	var queueFullErrors int
	// 提交超过队列大小的事件
	for i := 0; i < 5; i++ {
		msg := mockMessage(1)
		err := flow.Submit(msg)
		if err == ErrQueueFull {
			queueFullErrors++
		}
	}

	// 解除阻塞
	mu.Lock()
	processing = false
	mu.Unlock()

	time.Sleep(100 * time.Millisecond)
	if queueFullErrors == 0 {
		t.Error("expected queue full errors")
	}
}

// TestKickUser 测试踢出用户
func TestKickUser(t *testing.T) {
	var processed atomic.Int32
	handler := func(_ context.Context, _ *envelope.InputMessage) {
		processed.Add(1)
	}

	flow, _ := New(handler)
	defer flow.Close()

	// 提交事件
	msg := mockMessage(1)
	_ = flow.Submit(msg)

	time.Sleep(50 * time.Millisecond)
	if flow.ActiveUserCount() != 1 {
		t.Errorf("expected 1 active user, got %d", flow.ActiveUserCount())
	}

	// 踢出用户
	flow.KickUser(1)

	time.Sleep(100 * time.Millisecond)
	if flow.ActiveUserCount() != 0 {
		t.Errorf("expected 0 active users, got %d", flow.ActiveUserCount())
	}
}

// TestClose 测试关闭
func TestClose(t *testing.T) {
	t.Run("graceful shutdown", func(t *testing.T) {
		var processed atomic.Int32
		handler := func(_ context.Context, _ *envelope.InputMessage) {
			time.Sleep(50 * time.Millisecond)
			processed.Add(1)
		}

		flow, _ := New(handler,
			WithRateLimit(100), // 高限流，避免事件被拒绝
			WithRateBurst(100),
		)

		// 提交事件
		for i := 1; i <= 3; i++ {
			msg := mockMessage(int64(i))
			if err := flow.Submit(msg); err != nil {
				t.Fatalf("failed to submit event: %v", err)
			}
		}

		// 等待一小段时间，确保事件开始处理
		time.Sleep(10 * time.Millisecond)

		// 关闭并等待
		err := flow.Close()
		if err != nil {
			t.Fatalf("expected no error, got %v", err)
		}

		if processed.Load() != 3 {
			t.Errorf("expected 3 events processed, got %d", processed.Load())
		}
	})

	t.Run("submit after close", func(t *testing.T) {
		handler := func(_ context.Context, _ *envelope.InputMessage) {}
		flow, _ := New(handler)
		flow.Close()

		msg := mockMessage(1)
		err := flow.Submit(msg)
		if err == nil {
			t.Error("expected error when submitting after close")
		}
	})
}

// TestPanicRecovery 测试 panic 恢复
func TestPanicRecovery(t *testing.T) {
	handler := func(_ context.Context, _ *envelope.InputMessage) {
		panic("test panic")
	}

	flow, _ := New(handler, WithMetrics())
	defer flow.Close()

	msg := mockMessage(1)
	_ = flow.Submit(msg)

	time.Sleep(100 * time.Millisecond)

	// 检查失败指标
	metrics := flow.GetMetrics()
	userMetrics := metrics.GetUserMetrics(1)
	if userMetrics == nil {
		t.Fatal("expected user metrics")
	}
	if userMetrics.Failed.Load() == 0 {
		t.Error("expected panic to be recorded as failure")
	}
}

// TestContextCancellation 测试上下文取消
func TestContextCancellation(t *testing.T) {
	var ctxCancelled atomic.Bool
	handler := func(ctx context.Context, _ *envelope.InputMessage) {
		<-ctx.Done()
		ctxCancelled.Store(true)
	}

	flow, _ := New(handler)
	defer flow.Close()

	msg := mockMessage(1)
	_ = flow.Submit(msg)

	time.Sleep(50 * time.Millisecond)
	flow.KickUser(1)

	time.Sleep(100 * time.Millisecond)
	if !ctxCancelled.Load() {
		t.Error("expected context to be canceled")
	}
}

// TestConcurrentSubmit 测试并发提交
func TestConcurrentSubmit(t *testing.T) {
	var processed atomic.Int32
	handler := func(_ context.Context, _ *envelope.InputMessage) {
		processed.Add(1)
	}

	flow, _ := New(handler)
	defer flow.Close()

	const goroutines = 10
	const eventsPerGoroutine = 100

	var wg sync.WaitGroup
	wg.Add(goroutines)

	for i := 0; i < goroutines; i++ {
		go func(userID int64) {
			defer wg.Done()
			for j := 0; j < eventsPerGoroutine; j++ {
				msg := mockMessage(userID)
				_ = flow.Submit(msg)
			}
		}(int64(i))
	}

	wg.Wait()
	time.Sleep(500 * time.Millisecond)

	// 由于限流，不是所有事件都会被处理
	// 但至少应该处理了一部分
	if processed.Load() == 0 {
		t.Error("expected some events to be processed")
	}
}

// TestMetrics 测试指标收集
func TestMetrics(t *testing.T) {
	handler := func(_ context.Context, _ *envelope.InputMessage) {}

	flow, _ := New(handler,
		WithMetrics(),
		WithRateLimit(2),
		WithRateBurst(2),
	)
	defer flow.Close()

	// 提交事件，触发限流
	for i := 0; i < 5; i++ {
		msg := mockMessage(1)
		_ = flow.Submit(msg)
	}

	time.Sleep(200 * time.Millisecond)

	metrics := flow.GetMetrics()
	if metrics == nil {
		t.Fatal("expected metrics to be enabled")
	}

	userMetrics := metrics.GetUserMetrics(1)
	if userMetrics == nil {
		t.Fatal("expected user metrics")
	}

	snapshot := userMetrics.GetSnapshot()
	if snapshot.Enqueued != 2 {
		t.Errorf("expected 2 enqueued, got %d", snapshot.Enqueued)
	}
	if snapshot.RateLimited != 3 {
		t.Errorf("expected 3 rate limited, got %d", snapshot.RateLimited)
	}
	if snapshot.Processed != 2 {
		t.Errorf("expected 2 processed, got %d", snapshot.Processed)
	}
}

// TestRateLimitDisabled 测试禁用限流
func TestRateLimitDisabled(t *testing.T) {
	var count int32
	handler := func(_ context.Context, _ *envelope.InputMessage) {
		atomic.AddInt32(&count, 1)
	}

	// 创建流量管理器，禁用限流
	flow, _ := New(handler,
		WithRateLimit(1),            // 设置一个很低的限流
		WithRateBurst(1),            // 突发也很低
		WithRateLimitEnabled(false), // 但是禁用限流
	)
	defer flow.Close()

	// 快速提交多个事件，应该都能成功（不会被限流）
	const eventCount = 10
	for i := 0; i < eventCount; i++ {
		msg := mockMessage(1)
		err := flow.Submit(msg)
		if err != nil {
			t.Errorf("submit failed: %v (should not be rate limited)", err)
		}
	}

	// 等待处理完成
	time.Sleep(100 * time.Millisecond)

	// 验证所有事件都被处理了
	processed := atomic.LoadInt32(&count)
	if processed != eventCount {
		t.Errorf("expected %d events processed, got %d", eventCount, processed)
	}
}
