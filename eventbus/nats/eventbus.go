package nats

import (
	"context"
	"sync"

	"github.com/ivy-mobile/odin/eventbus"

	"github.com/nats-io/nats.go"
)

type Eventbus struct {
	err  error
	opts *options

	rw   sync.RWMutex
	subs map[string]*nats.Subscription
}

func NewEventbus(opts ...Option) *Eventbus {
	o := defaultOptions()
	for _, opt := range opts {
		opt(o)
	}

	eb := &Eventbus{opts: o}
	eb.opts = o
	eb.subs = make(map[string]*nats.Subscription)

	if o.conn == nil {
		o.conn, eb.err = nats.Connect(o.url, nats.Timeout(o.timeout))
	}

	return eb
}

// Publish 发布事件
func (eb *Eventbus) Publish(ctx context.Context, topic string, payload []byte) error {
	if eb.err != nil {
		return eb.err
	}
	return eb.opts.conn.Publish(topic, payload)
}

// Subscribe 订阅事件
func (eb *Eventbus) Subscribe(ctx context.Context, topic string, handler eventbus.EventHandler) error {
	if eb.err != nil {
		return eb.err
	}

	sub, err := eb.opts.conn.Subscribe(topic, func(msg *nats.Msg) {
		handler(msg.Data)
	})
	if err != nil {
		return err
	}
	eb.rw.Lock()
	eb.subs[topic] = sub
	eb.rw.Unlock()

	return nil
}

// Unsubscribe 取消订阅
func (eb *Eventbus) Unsubscribe(ctx context.Context, topic string) error {
	if eb.err != nil {
		return eb.err
	}

	eb.rw.Lock()
	defer eb.rw.Unlock()

	sub, ok := eb.subs[topic]
	if ok {
		err := sub.Unsubscribe()
		if err != nil {
			return err
		}
		delete(eb.subs, topic)
	}

	return nil
}

// Close 停止监听
func (eb *Eventbus) Close() error {
	if eb.err != nil {
		return eb.err
	}

	eb.opts.conn.Close()

	return nil
}
