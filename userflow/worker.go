package userflow

import (
	"context"
	"sync"

	"golang.org/x/time/rate"

	"github.com/ivy-mobile/odin/envelope"
)

// worker 每个用户的工作者，包含独立的队列和限流器
type worker struct {
	userID    int64
	queue     chan *envelope.InputMessage
	limiter   *rate.Limiter
	ctx       context.Context
	cancel    context.CancelFunc
	closeOnce sync.Once
}

// newWorker 创建新的用户工作者
func newWorker(parentCtx context.Context, userID int64, queueSize int, rateLimit float64, rateBurst int) *worker {
	ctx, cancel := context.WithCancel(parentCtx)
	return &worker{
		userID:  userID,
		queue:   make(chan *envelope.InputMessage, queueSize),
		limiter: rate.NewLimiter(rate.Limit(rateLimit), rateBurst),
		ctx:     ctx,
		cancel:  cancel,
	}
}

// close 关闭工作者（只取消上下文，不关闭 queue）
// queue 的关闭由 processUserEvents 的 cleanupWorker 负责
func (w *worker) close() {
	w.closeOnce.Do(func() {
		w.cancel()
	})
}
