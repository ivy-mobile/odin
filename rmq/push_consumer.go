package rmq

import (
	"context"
	"fmt"
	"time"

	"github.com/apache/rocketmq-clients/golang/v5"
	"github.com/apache/rocketmq-clients/golang/v5/credentials"
)

type PushConsumer struct {
	ctx               context.Context
	cancel            context.CancelFunc
	c                 golang.PushConsumer
	maxMessageNum     int32
	invisibleDuration time.Duration
	watchErrorHandler func(err error)
}

// NewPushConsumerWithOption 创建消费者, 支持 odin 自定义选项
func NewPushConsumerWithOption(
	endpoint,
	namespace,
	group string,
	awaitDuration time.Duration,
	credentials *credentials.SessionCredentials,
	opts ...golang.PushConsumerOption) (*PushConsumer, error) {

	opts = append(opts, golang.WithPushAwaitDuration(awaitDuration))
	pc, err := golang.NewPushConsumer(
		&golang.Config{
			Endpoint:      endpoint,
			NameSpace:     namespace,
			ConsumerGroup: group,
			Credentials:   credentials,
		},
		opts...,
	)
	if err != nil {
		return nil, fmt.Errorf("new push consumer error: %v", err)
	}
	if err := pc.Start(); err != nil {
		return nil, fmt.Errorf("start consumer error: %v", err)
	}
	ctx, cancel := context.WithCancel(context.Background())
	c := &PushConsumer{
		ctx:               ctx,
		cancel:            cancel,
		c:                 pc,
		maxMessageNum:     16,
		invisibleDuration: time.Second * 10, // 至少10000
	}
	return c, nil
}

// Subscribe 订阅主题 - 所有消息
func (sc *PushConsumer) Subscribe(topic string) error {
	if topic == "" {
		return fmt.Errorf("topic is empty")
	}
	return sc.c.Subscribe(topic, golang.SUB_ALL)
}

// SubscribeByTag 订阅主题,tag过滤
func (sc *PushConsumer) SubscribeByTag(topic, tag string) error {
	if topic == "" {
		return fmt.Errorf("topic is empty")
	}
	if tag == "" {
		return fmt.Errorf("tag is empty")
	}
	return sc.c.Subscribe(topic, golang.NewFilterExpression(tag))
}

// SubscribeBySQL92 订阅主题, sql92 过滤
func (sc *PushConsumer) SubscribeBySQL92(topic, sql92 string) error {
	if topic == "" {
		return fmt.Errorf("topic is empty")
	}
	if sql92 == "" {
		return fmt.Errorf("sql92 is empty")
	}
	if err := sc.c.Subscribe(topic, golang.NewFilterExpressionWithType(sql92, golang.SQL92)); err != nil {
		return fmt.Errorf("subscribe error: %v", err)
	}
	return nil
}

// Unsubscribe 取消订阅主题
func (sc *PushConsumer) Unsubscribe(topic string) error {
	return sc.c.Unsubscribe(topic)
}

// Close 关闭消费者
func (sc *PushConsumer) Close() error {
	sc.cancel()
	return sc.c.GracefulStop()
}
