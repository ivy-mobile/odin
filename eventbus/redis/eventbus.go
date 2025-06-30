package redis

import (
	"context"
	"strings"

	"github.com/ivy-mobile/odin/eventbus"
	"github.com/ivy-mobile/odin/xutil/xconv"
	"github.com/ivy-mobile/odin/xutil/xgo"

	"github.com/redis/go-redis/v9"
)

type Eventbus struct {
	ctx    context.Context
	cancel context.CancelFunc
	opts   *options
	sub    *redis.PubSub

	handlers map[string]eventbus.EventHandler
}

func NewEventbus(opts ...Option) *Eventbus {
	o := defaultOptions()
	for _, opt := range opts {
		opt(o)
	}

	if o.prefix == "" {
		o.prefix = defaultPrefix
	}

	if o.client == nil {
		o.client = redis.NewUniversalClient(&redis.UniversalOptions{
			Addrs:      o.addrs,
			DB:         o.db,
			Username:   o.username,
			Password:   o.password,
			MaxRetries: o.maxRetries,
		})
	}

	eb := &Eventbus{
		handlers: make(map[string]eventbus.EventHandler),
	}
	eb.ctx, eb.cancel = context.WithCancel(o.ctx)
	eb.opts = o
	eb.sub = eb.opts.client.Subscribe(eb.ctx)
	eb.watch()
	return eb
}

// Publish 发布事件
func (eb *Eventbus) Publish(ctx context.Context, topic string, payload []byte) error {
	return eb.opts.client.Publish(ctx, eb.buildChannelKey(topic), payload).Err()
}

// Subscribe 订阅事件
func (eb *Eventbus) Subscribe(ctx context.Context, topic string, handler eventbus.EventHandler) error {
	err := eb.sub.Subscribe(ctx, eb.buildChannelKey(topic))
	if err != nil {
		return err
	}
	eb.handlers[topic] = handler
	return nil
}

// Unsubscribe 取消订阅
func (eb *Eventbus) Unsubscribe(ctx context.Context, topic string) error {
	err := eb.sub.Unsubscribe(ctx, eb.buildChannelKey(topic))
	if err != nil {
		return err
	}
	delete(eb.handlers, topic)
	return nil
}

func (eb *Eventbus) watch() {
	xgo.Go(func() {
		for {
			select {
			case <-eb.ctx.Done():
				return
			case msg := <-eb.sub.Channel():
				topic := eb.parseChannelKey(msg.Channel)
				handler, ok := eb.handlers[topic]
				if ok {
					handler(xconv.Bytes(msg.Payload))
				}
			}
		}
	})
}

// Close 停止监听
func (eb *Eventbus) Close() error {
	eb.cancel()
	return eb.sub.Close()
}

// build channel key pass by topic
func (eb *Eventbus) buildChannelKey(topic string) string {
	if eb.opts.prefix == "" {
		return topic
	} else {
		return eb.opts.prefix + ":" + topic
	}
}

// parse to topic from channel key
func (eb *Eventbus) parseChannelKey(channel string) string {
	if eb.opts.prefix == "" {
		return channel
	} else {
		return strings.TrimPrefix(channel, eb.opts.prefix+":")
	}
}
