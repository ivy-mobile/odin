package engine

import (
	"fmt"
	"time"

	"github.com/apache/rocketmq-clients/golang/v5/credentials"

	"github.com/ivy-mobile/odin/conf"
	"github.com/ivy-mobile/odin/rmq"
)

// NewRMQProducer 创建 MQ 生产者
func NewRMQProducer() (*rmq.Producer, error) {
	eCfg := conf.MQ()

	p, err := rmq.NewProducer(eCfg.Endpoint, eCfg.Namespace, eCfg.Group, &credentials.SessionCredentials{
		AccessKey:     eCfg.AccessKey,
		AccessSecret:  eCfg.SecretKey,
		SecurityToken: eCfg.SecurityToken,
	})
	if err != nil {
		return nil, fmt.Errorf("rmq.NewProducer err: %v", err)
	}

	return p, nil
}

// NewRMQConsumer 创建 MQ 消费者
func NewRMQConsumer() (*rmq.Consumer, error) {
	eCfg := conf.MQ()

	c, err := rmq.NewConsumer(
		eCfg.Endpoint,
		eCfg.Namespace,
		eCfg.Group,
		time.Second*5,
		&credentials.SessionCredentials{
			AccessKey:     eCfg.AccessKey,
			AccessSecret:  eCfg.SecretKey,
			SecurityToken: eCfg.SecurityToken,
		},
	)
	if err != nil {
		return nil, fmt.Errorf("rmq.NewConsumer err: %v", err)
	}

	return c, nil
}

// NewRMQTransceiverConsumer 为指定网关节点创建独立消费组用的消费者
func NewRMQTransceiverConsumer(errHandler func(err error)) (*rmq.Consumer, error) {
	eCfg := conf.MQ()

	c, err := rmq.NewConsumerWithOption(
		eCfg.Endpoint, eCfg.Namespace, eCfg.NodeGroup, time.Second*5,
		&credentials.SessionCredentials{
			AccessKey:     eCfg.AccessKey,
			AccessSecret:  eCfg.SecretKey,
			SecurityToken: eCfg.SecurityToken,
		},
		[]rmq.ConsumerOption{
			rmq.WithWatchErrorHandler(errHandler),
		},
	)
	if err != nil {
		return nil, fmt.Errorf("rmq.NewConsumer err: %v", err)
	}

	return c, nil
}
