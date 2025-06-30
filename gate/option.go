package gate

import (
	"context"
	"time"

	"github.com/ivy-mobile/odin/encoding"
	"github.com/ivy-mobile/odin/encoding/proto"
	"github.com/ivy-mobile/odin/enum"
	"github.com/ivy-mobile/odin/eventbus"
	"github.com/ivy-mobile/odin/registry"

	"github.com/olahol/melody"
)

type (
	Option func(*options)

	MessageHandler func(s *melody.Session, msg []byte) // 消息处理函数
)

type options struct {
	ctx             context.Context // 上下文
	id              string          // 实例ID
	name            string          // 实例名称
	gameServiceName string          // 游戏服务总名称 如: Game

	codec encoding.Codec // 编码解码器

	// websocket 配置
	port                      string        // 端口
	pattern                   string        // 路由匹配模式
	writeWait                 time.Duration // 写入超时时间
	pongWait                  time.Duration // pong 等待时间
	pingPeriod                time.Duration // ping 之间的时间间隔
	maxMessageSize            int64         // 消息最大字节数
	messageBufferSize         int           // 在会话缓冲区开始丢弃消息之前，该缓冲区中所能容纳的最大消息数量
	concurrentMessageHandling bool          // 并发处理来自会话的消息

	// 事件总线
	eventbus eventbus.Eventbus
	// 服务注册与发现
	registry registry.Registry
}

func defaultOptions() *options {
	return &options{
		ctx:                       context.Background(),
		name:                      enum.DefaultGateServiceName,
		gameServiceName:           enum.DefaultGameServiceName,
		writeWait:                 10 * time.Second,
		pongWait:                  60 * time.Second,
		pingPeriod:                54 * time.Second,
		maxMessageSize:            512,
		messageBufferSize:         1024,
		concurrentMessageHandling: false,
		codec:                     proto.DefaultCodec,
	}
}

func WithContext(ctx context.Context) Option {
	return func(os *options) {
		os.ctx = ctx
	}
}

// WithID 设置实例ID
func WithID(id string) Option {
	return func(os *options) {
		os.id = id
	}
}

// WithName 设置实例名称 默认 gate
func WithName(name string) Option {
	return func(os *options) {
		os.name = name
	}
}

// WithGameServiceName 设置游戏服务名称 默认 game-service
// 所有游戏都将注册名称为 game-service,每个实例的分别通过metadata中数据进行区分
func WithGameServiceName(name string) Option {
	return func(os *options) {
		os.gameServiceName = name
	}
}

// WithPort 设置端口 默认 :8080
func WithPort(port string) Option {
	return func(os *options) {
		os.port = port
	}
}

// WithPattern 设置路由匹配模式 默认 /ws
func WithPattern(pattern string) Option {
	return func(os *options) {
		os.pattern = pattern
	}
}

// WithWriteWait 设置写入超时时间 默认 10s
func WithWriteWait(writeWait time.Duration) Option {
	return func(os *options) {
		os.writeWait = writeWait
	}
}

// WithPongWait 设置 pong 等待时间 默认 60s
func WithPongWait(pongWait time.Duration) Option {
	return func(os *options) {
		os.pongWait = pongWait
	}
}

// WithPingPeriod 设置 ping 之间的时间间隔 默认 54s
func WithPingPeriod(pingPeriod time.Duration) Option {
	return func(os *options) {
		os.pingPeriod = pingPeriod
	}
}

// WithMaxMessageSize 设置消息最大字节数 默认 512
func WithMaxMessageSize(maxMessageSize int64) Option {
	return func(os *options) {
		os.maxMessageSize = maxMessageSize
	}
}

// WithEventbus 设置事件总线
func WithEventbus(eventbus eventbus.Eventbus) Option {
	return func(os *options) {
		os.eventbus = eventbus
	}
}

// WithCodec 设置编码解码器 - 默认使用proto
func WithCodec(codec encoding.Codec) Option {
	return func(os *options) {
		os.codec = codec
	}
}

// WithRegistry 设置服务注册发现 - 必选
func WithRegistry(registry registry.Registry) Option {
	return func(os *options) {
		os.registry = registry
	}
}
