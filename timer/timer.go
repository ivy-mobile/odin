package timer

import (
	"context"
	"fmt"
	"time"

	"github.com/redis/go-redis/v9"
)

// Handler 计时器到期时回调：桌 id、计时器 id、登记时间
type Handler func(tableID int, timerID string, startTime time.Time)

// Timer 基于 Redis ZSet 的牌桌阶段计时抽象
type Timer interface {
	// Start 登记或刷新某桌某计时器的到期点
	Start(tableID int, timerID string, duration time.Duration) error

	// Stop 移除某桌某计时器登记项
	Stop(tableID int, timerID string) error

	// Listen 周期性扫描到期项并回调；ctx 结束则退出
	Listen(ctx context.Context, callback Handler) error
}

// Message 有序集合成员里序列化的计时元数据
type Message struct {
	TableID   int       `json:"table_id"`
	TimerID   string    `json:"timer_id"`
	Timestamp time.Time `json:"timestamp"`
}

// Config 基于 Redis 的计时器运行参数
type Config struct {
	RedisClient  redis.UniversalClient // Redis 客户端
	KeyFormat    string                // Redis key 格式，例如: "game:%d:table_timers"
	GameID       int                   // 游戏ID，用于生成 key
	ScanInterval time.Duration         // 扫描间隔，默认 1 秒
	BatchSize    int                   // 每次处理的最大数量，默认 100
}

// RedisTimer 用有序集合存到期时间戳与 JSON 成员
type RedisTimer struct {
	config Config
}

// NewRedisTimer 校验默认参数并构造实例
func NewRedisTimer(config Config) (*RedisTimer, error) {
	if config.RedisClient == nil {
		return nil, fmt.Errorf("redis client is required")
	}
	if config.KeyFormat == "" {
		config.KeyFormat = "game:%d:table_timers"
	}
	if config.ScanInterval <= 0 {
		config.ScanInterval = 1 * time.Second
	}
	if config.BatchSize <= 0 {
		config.BatchSize = 100
	}

	return &RedisTimer{
		config: config,
	}, nil
}

// 本游戏在 Redis 中的计时器 ZSet 键
func (r *RedisTimer) getTimerKey() string {
	return fmt.Sprintf(r.config.KeyFormat, r.config.GameID)
}
