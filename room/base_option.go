package room

import "time"

type Option func(*options)

type options struct {
	id                  int
	name                string
	maxPlayerCount      int                    // 最大玩家数
	actionTimeout       time.Duration          // 单个操作超时时间
	idleState           RoomState              // 创建房间时初始状态
	stateTimeoutHandler func() (uint16, error) // 状态超时执行函数 - 控制房间状态流转
}

func defaultOptions() *options {
	return &options{
		actionTimeout: 3 * time.Second,
	}
}

// With 设置房间ID和名称
func With(id int, name string) Option {
	return func(o *options) {
		o.id = id
		o.name = name
	}
}

// WithMaxPlayerCount 设置创建房间时初始状态，无默认值，必须手动设置
func WithMaxPlayerCount(count int) Option {
	return func(o *options) {
		o.maxPlayerCount = count
	}
}

// WithIdleState 设置创建房间时初始状态, 默认无
func WithIdleState(state RoomState) Option {
	return func(o *options) {
		o.idleState = state
	}
}

// WithActionTimeout 设置单个操作超时时间, 默认3秒
func WithActionTimeout(timeout time.Duration) Option {
	return func(o *options) {
		o.actionTimeout = timeout
	}
}

// WithStateTimeoutHandler 设置状态超时执行函数, 当状态倒计时结束时执行
func WithStateTimeoutHandler(handler func() (uint16, error)) Option {
	return func(o *options) {
		o.stateTimeoutHandler = handler
	}
}
