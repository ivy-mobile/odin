package userflow

import (
	"errors"
	"time"
)

// Option 配置选项函数
type Option func(*options)

// options 内部配置
type options struct {
	queueSize       int
	rateLimit       float64
	rateBurst       int
	shutdownTimeout time.Duration
	enableMetrics   bool
	enableRateLimit bool
}

// 默认配置
var defaultOptions = options{
	queueSize:       10,
	rateLimit:       5,
	rateBurst:       10,
	shutdownTimeout: 5 * time.Second,
	enableMetrics:   false,
	enableRateLimit: true, // 默认启用限流
}

// WithQueueSize 设置每个用户的请求队列大小
func WithQueueSize(size int) Option {
	return func(o *options) {
		o.queueSize = size
	}
}

// WithRateLimit 设置每秒允许的请求数
func WithRateLimit(limit float64) Option {
	return func(o *options) {
		o.rateLimit = limit
	}
}

// WithRateBurst 设置突发请求数
func WithRateBurst(burst int) Option {
	return func(o *options) {
		o.rateBurst = burst
	}
}

// WithShutdownTimeout 设置关闭超时时间
func WithShutdownTimeout(timeout time.Duration) Option {
	return func(o *options) {
		o.shutdownTimeout = timeout
	}
}

// WithMetrics 启用指标收集
func WithMetrics() Option {
	return func(o *options) {
		o.enableMetrics = true
	}
}

// WithRateLimitEnabled 启用或禁用限流
// 默认为启用，可以通过 WithRateLimitEnabled(false) 禁用限流
func WithRateLimitEnabled(enabled bool) Option {
	return func(o *options) {
		o.enableRateLimit = enabled
	}
}

// validate 验证配置
func (o *options) validate() error {
	if o.queueSize <= 0 {
		return errors.New("queueSize must be greater than 0")
	}
	// 只有启用限流时才验证限流参数
	if o.enableRateLimit {
		if o.rateLimit <= 0 {
			return errors.New("rateLimit must be greater than 0")
		}
		if o.rateBurst <= 0 {
			return errors.New("rateBurst must be greater than 0")
		}
	}
	if o.shutdownTimeout <= 0 {
		return errors.New("shutdownTimeout must be greater than 0")
	}
	return nil
}
