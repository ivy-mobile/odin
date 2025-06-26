package eventbus

import (
	"context"
)

type EventHandler func(data []byte)

// Eventbus 事件总线统一接口
// 内置内存、redis、nats、kafka实现,可自定义实现
type Eventbus interface {
	// Publish 发布
	Publish(ctx context.Context, topic string, data []byte) error
	// Subscribe 订阅
	Subscribe(ctx context.Context, topic string, handler EventHandler) error
	// Unsubscribe 取消订阅
	Unsubscribe(ctx context.Context, topic string) error
	// Close 关闭事件总线
	Close() error
}
