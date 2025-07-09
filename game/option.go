package game

import (
	"github.com/ivy-mobile/odin/encoding"
	"github.com/ivy-mobile/odin/encoding/proto"
	"github.com/ivy-mobile/odin/enum"
	"github.com/ivy-mobile/odin/eventbus"
	"github.com/ivy-mobile/odin/generator"
	"github.com/ivy-mobile/odin/registry"
)

type Option func(*options)

type options struct {
	id string
	// 网关服统一的服务名, 如:Gate,与网关设置的GameServiceName一致, 将根据配置的此服务名进行发现所有的游戏服务
	gateServiceName string
	// 游戏服统一的服务名, 如:Game,与网关设置的GameServiceName一致, 将根据配置的此服务名进行发现所有的游戏服务
	serviceName string
	// 游戏名,如：hamster-battle,区别各个游戏
	name string
	// 编解码器
	codec encoding.Codec
	// 服务注册与发现中心
	registry registry.Registry
	// 房间ID生成器
	roomIdGenerator generator.RoomIdGenerator
	// eventbus 事件总线
	eventbus eventbus.Eventbus
	// 后台指令消息处理器
	adminCmdHandler CmdMessageHandler
}

func defaultOptions() *options {
	return &options{
		codec:           proto.DefaultCodec, // 默认使用proto编解码
		serviceName:     enum.DefaultGameServiceName,
		gateServiceName: enum.DefaultGateServiceName,
	}
}

// WithID 游戏ID
func WithID(id string) Option {
	return func(o *options) {
		o.id = id
	}
}

// WithGateServiceName 网关服务名称 - 默认Gate
// 推荐使用默认设置，不调用此方法
func WithGateServiceName(name string) Option {
	return func(o *options) {
		o.gateServiceName = name
	}
}

// WithName 游戏名  - 默认为空
// name: 游戏服务唯一标识 如: hamster-battle
func WithName(name string) Option {
	return func(o *options) {
		o.name = name
	}
}

// WithServiceName 游戏公共服务名  - 默认为gate-service
// 推荐使用默认设置，不调用此方法
func WithServiceName(serviceName string) Option {
	return func(o *options) {
		o.serviceName = serviceName
	}
}

// WithEventbus 事件总线
func WithEventbus(eb eventbus.Eventbus) Option {
	return func(o *options) {
		o.eventbus = eb
	}
}

// WithRoomIdGenerator 房间ID生成器
func WithRoomIdGenerator(roomIdGenerator generator.RoomIdGenerator) Option {
	return func(o *options) {
		o.roomIdGenerator = roomIdGenerator
	}
}

// WithCodec 编解码器 - 默认使用 proto
func WithCodec(codec encoding.Codec) Option {
	return func(o *options) {
		o.codec = codec
	}
}

// WithRegistry 服务注册与发现中心
func WithRegistry(registry registry.Registry) Option {
	return func(o *options) {
		o.registry = registry
	}
}
