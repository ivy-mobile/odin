package game

import (
	"github.com/ivy-mobile/odin/encoding"
	"github.com/ivy-mobile/odin/encoding/proto"
	"github.com/ivy-mobile/odin/eventbus"
	"github.com/ivy-mobile/odin/generator"
)

type Option func(*options)

type options struct {
	id   string
	name string
	// 编解码器
	codec encoding.Codec
	// 房间ID生成器
	roomIdGenerator generator.RoomIdGenerator
	// eventbus 事件总线
	eventbus eventbus.Eventbus
	// 后台指令消息处理器
	adminCmdHandler AdminMessageHandler
}

func defaultOptions() *options {
	return &options{
		codec: proto.DefaultCodec, // 默认使用proto编解码
	}
}

// WithID 游戏ID
func WithID(id string) Option {
	return func(o *options) {
		o.id = id
	}
}

// WithName 游戏名
func WithName(name string) Option {
	return func(o *options) {
		o.name = name
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

// WithAdminCmdHandler 后台指令消息处理器，默认为nil,不处理
func WithAdminCmdHandler(adminCmdHandler AdminMessageHandler) Option {
	return func(o *options) {
		o.adminCmdHandler = adminCmdHandler
	}
}

// WithCodec 编解码器 - 默认使用 proto
func WithCodec(codec encoding.Codec) Option {
	return func(o *options) {
		o.codec = codec
	}
}
