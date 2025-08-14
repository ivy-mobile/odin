package rmq

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/apache/rocketmq-clients/golang/v5"
	"github.com/apache/rocketmq-clients/golang/v5/credentials"
	"github.com/ivy-mobile/odin/xutil/xgo"
	"github.com/ivy-mobile/odin/xutil/xlog"
)

type Consumer struct {
	ctx               context.Context
	cancel            context.CancelFunc
	wg                sync.WaitGroup
	once              sync.Once
	subs              sync.Map // 已订阅topic
	c                 golang.SimpleConsumer
	maxMessageNum     int32
	invisibleDuration time.Duration
}

// NewConsumer 消费者
// golang.WithAwaitDuration(time.Second*5),
func NewConsumer(endpoint, namespace, group string,
	credentials *credentials.SessionCredentials,
	opts ...golang.SimpleConsumerOption) (*Consumer, error) {

	sc, err := golang.NewSimpleConsumer(
		&golang.Config{
			Endpoint:      endpoint,
			NameSpace:     namespace,
			ConsumerGroup: group,
			Credentials:   credentials,
		},
		opts...,
	)
	if err != nil {
		return nil, fmt.Errorf("new simple consumer error: %v", err)
	}
	if err := sc.Start(); err != nil {
		return nil, fmt.Errorf("start consumer error: %v", err)
	}
	ctx, cancel := context.WithCancel(context.Background())
	return &Consumer{
		ctx:               ctx,
		cancel:            cancel,
		c:                 sc,
		maxMessageNum:     16,
		invisibleDuration: time.Second * 20, // 至少10000
	}, nil
}

// Subscribe 订阅主题 - 所有消息
// callback: 回调函数，当callback返回 error==nil 时, 会ack消息，error!= nil 时，不会ack消息
func (sc *Consumer) Subscribe(topic string, callback func(msg *golang.MessageView) error) error {

	if _, ok := sc.subs.LoadOrStore(topic, callback); ok {
		return fmt.Errorf("topic %s has been subscribed", topic)
	}
	if topic == "" {
		return fmt.Errorf("topic is empty")
	}
	if err := sc.c.Subscribe(topic, golang.SUB_ALL); err != nil {
		return fmt.Errorf("subscribe error: %v", err)
	}
	// 第一次订阅执行即可
	sc.once.Do(sc.watch)
	return nil
}

// SubscribeByTag 订阅主题,tag过滤
// callback: 回调函数，当callback返回 error==nil 时, 会ack消息，error!= nil 时，不会ack消息
func (sc *Consumer) SubscribeByTag(topic, tag string, callback func(msg *golang.MessageView) error) error {

	if _, ok := sc.subs.LoadOrStore(topic, callback); ok {
		return fmt.Errorf("topic %s has been subscribed", topic)
	}
	if topic == "" {
		return fmt.Errorf("topic is empty")
	}
	if tag == "" {
		return fmt.Errorf("tag is empty")
	}
	if err := sc.c.Subscribe(topic, golang.NewFilterExpression(tag)); err != nil {
		return fmt.Errorf("subscribe error: %v", err)
	}
	// 第一次订阅执行即可
	sc.once.Do(sc.watch)
	return nil
}

// SubscribeBySQL92 订阅主题,sql92 过滤
// callback: 回调函数，当callback返回 error==nil 时, 会ack消息，error!= nil 时，不会ack消息
func (sc *Consumer) SubscribeBySQL92(topic, sql92 string, callback func(msg *golang.MessageView) error) error {

	if _, ok := sc.subs.LoadOrStore(topic, callback); ok {
		return fmt.Errorf("topic %s has been subscribed", topic)
	}
	if topic == "" {
		return fmt.Errorf("topic is empty")
	}
	if sql92 == "" {
		return fmt.Errorf("sql92 is empty")
	}
	if err := sc.c.Subscribe(topic, golang.NewFilterExpressionWithType(sql92, golang.SQL92)); err != nil {
		return fmt.Errorf("subscribe error: %v", err)
	}
	// 第一次订阅执行即可
	sc.once.Do(sc.watch)
	return nil
}

// Unsubscribe 取消订阅主题
func (sc *Consumer) Unsubscribe(topic string) error {
	if err := sc.c.Unsubscribe(topic); err != nil {
		return fmt.Errorf("unsubscribe error: %v", err)
	}
	sc.subs.Delete(topic)
	return nil
}

// Close 关闭消费者
func (sc *Consumer) Close() error {
	sc.cancel()
	sc.wg.Wait()
	return sc.c.GracefulStop()
}

// 监听消息
func (sc *Consumer) watch() {
	sc.wg.Add(1)

	xgo.Go(func() {
		defer sc.wg.Done()
		for {
			select {
			case <-sc.ctx.Done():
				return
			default:
				msgs, err := sc.c.Receive(sc.ctx, sc.maxMessageNum, sc.invisibleDuration)
				if err != nil {
					continue
				}
				for _, msg := range msgs {
					cbFunc, ok := sc.subs.Load(msg.GetTopic())
					if !ok || cbFunc == nil {
						continue
					}
					if err = cbFunc.(func(msg *golang.MessageView) error)(msg); err != nil {
						continue
					}
					if err = sc.c.Ack(sc.ctx, msg); err != nil {
						xlog.Error().Msgf("simple consumer ack error: %v, msg: %v", err, msg)
						continue
					}
				}
			}
		}
	})
}
