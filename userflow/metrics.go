package userflow

import (
	"sync"
	"sync/atomic"
	"time"
)

// Metrics 指标收集器
type Metrics struct {
	// 全局指标
	activeUsers atomic.Int64

	// 每个用户的指标
	userMetrics sync.Map // key: int64(userID), value: *UserMetrics
}

// UserMetrics 单个用户的指标
type UserMetrics struct {
	Enqueued    atomic.Int64 // 入队数量
	Processed   atomic.Int64 // 处理成功数量
	Failed      atomic.Int64 // 处理失败数量
	RateLimited atomic.Int64 // 被限流数量
	QueueFull   atomic.Int64 // 队列满数量

	// 延迟统计
	totalLatency atomic.Int64 // 总延迟（纳秒）
	latencyCount atomic.Int64 // 延迟样本数
}

// NewMetrics 创建指标收集器
func NewMetrics() *Metrics {
	return &Metrics{}
}

// SetActiveUsers 设置活跃用户数
func (m *Metrics) SetActiveUsers(count int64) {
	m.activeUsers.Store(count)
}

// GetActiveUsers 获取活跃用户数
func (m *Metrics) GetActiveUsers() int64 {
	return m.activeUsers.Load()
}

// getUserMetrics 获取或创建用户指标
func (m *Metrics) getUserMetrics(userID int64) *UserMetrics {
	if val, ok := m.userMetrics.Load(userID); ok {
		return val.(*UserMetrics)
	}

	um := &UserMetrics{}
	actual, _ := m.userMetrics.LoadOrStore(userID, um)
	return actual.(*UserMetrics)
}

// IncEnqueued 增加入队计数
func (m *Metrics) IncEnqueued(userID int64) {
	m.getUserMetrics(userID).Enqueued.Add(1)
}

// IncProcessed 增加处理成功计数
func (m *Metrics) IncProcessed(userID int64) {
	m.getUserMetrics(userID).Processed.Add(1)
}

// IncFailed 增加处理失败计数
func (m *Metrics) IncFailed(userID int64) {
	m.getUserMetrics(userID).Failed.Add(1)
}

// IncRateLimited 增加限流计数
func (m *Metrics) IncRateLimited(userID int64) {
	m.getUserMetrics(userID).RateLimited.Add(1)
}

// IncQueueFull 增加队列满计数
func (m *Metrics) IncQueueFull(userID int64) {
	m.getUserMetrics(userID).QueueFull.Add(1)
}

// ObserveLatency 记录延迟
func (m *Metrics) ObserveLatency(userID int64, latency time.Duration) {
	um := m.getUserMetrics(userID)
	um.totalLatency.Add(int64(latency))
	um.latencyCount.Add(1)
}

// GetUserMetrics 获取指定用户的指标
func (m *Metrics) GetUserMetrics(userID int64) *UserMetrics {
	if val, ok := m.userMetrics.Load(userID); ok {
		return val.(*UserMetrics)
	}
	return nil
}

// GetAllUserMetrics 获取所有用户的指标
func (m *Metrics) GetAllUserMetrics() map[int64]*UserMetrics {
	result := make(map[int64]*UserMetrics)
	m.userMetrics.Range(func(key, value interface{}) bool {
		result[key.(int64)] = value.(*UserMetrics)
		return true
	})
	return result
}

// AverageLatency 计算平均延迟
func (um *UserMetrics) AverageLatency() time.Duration {
	count := um.latencyCount.Load()
	if count == 0 {
		return 0
	}
	total := um.totalLatency.Load()
	return time.Duration(total / count)
}

// Snapshot 用户指标快照
type Snapshot struct {
	Enqueued       int64
	Processed      int64
	Failed         int64
	RateLimited    int64
	QueueFull      int64
	AverageLatency time.Duration
}

// GetSnapshot 获取用户指标快照
func (um *UserMetrics) GetSnapshot() Snapshot {
	return Snapshot{
		Enqueued:       um.Enqueued.Load(),
		Processed:      um.Processed.Load(),
		Failed:         um.Failed.Load(),
		RateLimited:    um.RateLimited.Load(),
		QueueFull:      um.QueueFull.Load(),
		AverageLatency: um.AverageLatency(),
	}
}
