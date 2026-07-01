package userflow

import (
	"context"
	"errors"
	"fmt"
	"sync"
	"sync/atomic"
	"time"

	"github.com/ivy-mobile/odin/envelope"
)

// 预定义错误
var (
	ErrRateLimited = errors.New("rate limit exceeded")
	ErrQueueFull   = errors.New("queue full")
	ErrInvalidUser = errors.New("invalid user id")
	ErrClosed      = errors.New("flow is closed")
)

// EventHandler 事件处理函数
type EventHandler func(ctx context.Context, msg *envelope.InputMessage)

// Flow 用户流量管理器
// 为每个用户维护一个独立的协程和请求队列，保证用户维度请求处理的顺序性
type Flow struct {
	opts    options
	handler EventHandler

	workers   sync.Map // key: int64(userID), value: *worker
	userCount atomic.Int64

	ctx    context.Context
	cancel context.CancelFunc
	wg     sync.WaitGroup

	// 指标
	metrics *Metrics
}

// New 创建用户流量管理器
func New(handler EventHandler, opts ...Option) (*Flow, error) {
	if handler == nil {
		return nil, fmt.Errorf("handler cannot be nil")
	}

	// 应用默认配置
	cfg := defaultOptions

	// 应用用户配置
	for _, opt := range opts {
		opt(&cfg)
	}

	// 验证配置
	if err := cfg.validate(); err != nil {
		return nil, err
	}

	ctx, cancel := context.WithCancel(context.Background())

	f := &Flow{
		opts:    cfg,
		handler: handler,
		ctx:     ctx,
		cancel:  cancel,
	}

	if cfg.enableMetrics {
		f.metrics = NewMetrics()
	}

	return f, nil
}

// Submit 提交消息到用户队列
// 如果用户首次提交，会自动创建对应的处理协程
// 返回 error 表示提交失败（限流或队列满或管理器已关闭）
func (f *Flow) Submit(msg *envelope.InputMessage) error {
	// 检查管理器是否已关闭
	select {
	case <-f.ctx.Done():
		return ErrClosed
	default:
	}

	// 获取用户ID
	userID := msg.GetHeader().GetUid()
	if userID == 0 {
		return ErrInvalidUser
	}

	// 获取或创建工作者
	w := f.getOrCreateWorker(userID)

	// 限流检查（仅在启用时）
	if f.opts.enableRateLimit && !w.limiter.Allow() {
		if f.metrics != nil {
			f.metrics.IncRateLimited(userID)
		}
		return ErrRateLimited
	}

	// 非阻塞入队
	select {
	case w.queue <- msg:
		if f.metrics != nil {
			f.metrics.IncEnqueued(userID)
		}
		return nil
	default:
		if f.metrics != nil {
			f.metrics.IncQueueFull(userID)
		}
		return ErrQueueFull
	}
}

// getOrCreateWorker 获取或创建用户工作者
func (f *Flow) getOrCreateWorker(userID int64) *worker {
	// 快速路径：用户已存在
	if val, ok := f.workers.Load(userID); ok {
		return val.(*worker)
	}

	// 慢速路径：创建新的工作者
	w := newWorker(f.ctx, userID, f.opts.queueSize, f.opts.rateLimit, f.opts.rateBurst)

	// 原子操作：加载或存储
	actual, loaded := f.workers.LoadOrStore(userID, w)
	if loaded {
		// 其他协程已经创建了，关闭我们创建的
		w.close()
		return actual.(*worker)
	}

	// 我们是第一个创建的，启动处理协程
	f.userCount.Add(1)
	f.wg.Add(1)
	go f.processUserEvents(w)

	if f.metrics != nil {
		f.metrics.SetActiveUsers(f.userCount.Load())
	}

	return w
}

// processUserEvents 处理单个用户的事件队列
func (f *Flow) processUserEvents(w *worker) {
	defer f.wg.Done()
	defer f.cleanupWorker(w)

	for {
		select {
		case <-w.ctx.Done():
			// 用户被踢出或管理器关闭，先排空队列中的剩余消息
			f.drainQueue(w)
			return
		case msg, ok := <-w.queue:
			if !ok {
				// 队列已关闭
				return
			}
			// 处理事件
			f.handleEvent(w, msg)
		}
	}
}

// handleEvent 处理单个事件
func (f *Flow) handleEvent(w *worker, msg *envelope.InputMessage) {
	start := time.Now()

	// 使用 recover 捕获 panic
	defer func() {
		if r := recover(); r != nil {
			// panic 被捕获，记录失败指标
			if f.metrics != nil {
				f.metrics.IncFailed(w.userID)
			}
			// TODO: 添加日志记录，输出 panic 信息和堆栈
			// 例如: log.Printf("userflow: panic handling message for user %d: %v\n%s", w.userID, r, debug.Stack())
		}
	}()

	f.handler(w.ctx, msg)

	if f.metrics != nil {
		f.metrics.IncProcessed(w.userID)
		f.metrics.ObserveLatency(w.userID, time.Since(start))
	}
}

// drainQueue 排空队列中的剩余消息
func (f *Flow) drainQueue(w *worker) {
	for {
		select {
		case msg, ok := <-w.queue:
			if !ok {
				return
			}
			f.handleEvent(w, msg)
		default:
			// 队列已空
			return
		}
	}
}

// cleanupWorker 清理工作者
func (f *Flow) cleanupWorker(w *worker) {
	// 关闭队列
	close(w.queue)

	// 从 workers map 中删除
	f.workers.Delete(w.userID)

	// 更新活跃用户计数
	f.userCount.Add(-1)
	if f.metrics != nil {
		f.metrics.SetActiveUsers(f.userCount.Load())
		// 清理用户的 metrics 数据，防止内存泄漏
		f.metrics.DeleteUserMetrics(w.userID)
	}
}

// KickUser 踢出用户，释放对应资源
// 用户长时间离线时应调用此方法
func (f *Flow) KickUser(userID int64) {
	if val, ok := f.workers.Load(userID); ok {
		w := val.(*worker)
		w.cancel()
	}
}

// Close 关闭管理器，等待所有用户的事件处理完成
// 如果超过配置的超时时间，会强制关闭
func (f *Flow) Close() error {
	// 取消所有用户的上下文
	f.cancel()

	// 等待所有协程退出，带超时
	done := make(chan struct{})
	go func() {
		f.wg.Wait()
		close(done)
	}()

	select {
	case <-done:
		return nil
	case <-time.After(f.opts.shutdownTimeout):
		return fmt.Errorf("shutdown timeout after %v", f.opts.shutdownTimeout)
	}
}

// GetMetrics 获取指标（如果启用）
func (f *Flow) GetMetrics() *Metrics {
	return f.metrics
}

// ActiveUserCount 返回当前活跃用户数
func (f *Flow) ActiveUserCount() int64 {
	return f.userCount.Load()
}
