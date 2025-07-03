package rocketmq

import (
	"context"
	"sync"
	"time"

	"github.com/ivy-mobile/odin/eventbus"
	"github.com/ivy-mobile/odin/xutil/xgo"
	"github.com/ivy-mobile/odin/xutil/xlog"

	rmq "github.com/apache/rocketmq-clients/golang"
	"github.com/apache/rocketmq-clients/golang/credentials"
)

// Eventbus RocketMQ 事件总线实现
type Eventbus struct {
	ctx    context.Context
	cancel context.CancelFunc
	once   sync.Once

	producer rmq.Producer
	consumer rmq.SimpleConsumer
	handlers map[string]eventbus.EventHandler
}

// NewEventbus 创建新的 RocketMQ 事件总线实例
func NewEventbus(opts ...Option) (*Eventbus, error) {
	options := defaultOptions()
	for _, opt := range opts {
		opt(options)
	}

	// 创建生产者
	producer, err := rmq.NewProducer(&rmq.Config{
		Endpoint:  options.Endpoint,
		NameSpace: options.NameSpace,
		Credentials: &credentials.SessionCredentials{
			AccessKey:    options.AccessKey,
			AccessSecret: options.SecretKey,
		},
	})
	if err != nil {
		return nil, err
	}

	// 启动生产者
	if err = producer.Start(); err != nil {
		return nil, err
	}

	// 创建消费者
	consumer, err := rmq.NewSimpleConsumer(&rmq.Config{
		Endpoint:      options.Endpoint,
		NameSpace:     options.NameSpace,
		ConsumerGroup: options.ConsumerGroup,
		Credentials: &credentials.SessionCredentials{
			AccessKey:    options.AccessKey,
			AccessSecret: options.SecretKey,
		},
	}, rmq.WithAwaitDuration(5*time.Second))
	if err != nil {
		producer.GracefulStop()
		return nil, err
	}

	// 启动消费者
	if err := consumer.Start(); err != nil {
		producer.GracefulStop()
		return nil, err
	}

	ctx, cancel := context.WithCancel(context.Background())
	e := &Eventbus{
		producer: producer,
		consumer: consumer,
		ctx:      ctx,
		cancel:   cancel,
		handlers: make(map[string]eventbus.EventHandler),
	}
	e.watch()
	return e, nil
}

// Publish 发布消息
func (eb *Eventbus) Publish(ctx context.Context, topic string, data []byte) error {
	message := &rmq.Message{
		Topic: topic,
		Body:  data,
	}
	_, err := eb.producer.Send(ctx, message)
	if err != nil {
		return err
	}
	return nil
}

// Subscribe 订阅消息
func (eb *Eventbus) Subscribe(ctx context.Context, topic string, handler eventbus.EventHandler) error {

	// 设置订阅表达式
	if err := eb.consumer.Subscribe(topic, rmq.SUB_ALL); err != nil {
		return err
	}
	// 保存处理器
	eb.handlers[topic] = handler

	// 确保只启动一次
	eb.once.Do(func() {
		eb.watch()
	})
	return nil
}

// Unsubscribe 取消订阅
func (eb *Eventbus) Unsubscribe(ctx context.Context, topic string) error {
	if err := eb.consumer.Unsubscribe(topic); err != nil {
		return err
	}
	delete(eb.handlers, topic)
	return nil
}

func (eb *Eventbus) watch() {

	// 启动消息消费协程
	xgo.Go(func() {
		for {
			select {
			case <-eb.ctx.Done():
				return
			default:
				messages, err := eb.consumer.Receive(eb.ctx, 16, 20*time.Second)
				if err != nil {
					xlog.Debug().Msgf("Failed to receive message: %s", err.Error())
					continue
				}
				for _, msg := range messages {
					// 查找处理器
					handler, ok := eb.handlers[msg.GetTopic()]
					if !ok {
						continue
					}
					// 调用处理器
					handler(msg.GetBody())
					// 确认消息
					if err := eb.consumer.Ack(eb.ctx, msg); err != nil {
						xlog.Error().Msgf("Failed to ack message: %s", err.Error())
					}
				}
			}
		}
	})
}

// Close 关闭事件总线
func (eb *Eventbus) Close() error {
	eb.cancel()
	if err := eb.producer.GracefulStop(); err != nil {
		return err
	}
	return eb.consumer.GracefulStop()
}
