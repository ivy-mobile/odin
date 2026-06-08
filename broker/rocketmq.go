package broker

import (
	"fmt"
	"time"

	"github.com/apache/rocketmq-clients/golang/v5"
	"github.com/apache/rocketmq-clients/golang/v5/credentials"
	"github.com/ivy-mobile/odin/conf"
	"github.com/ivy-mobile/odin/encoding/json"
	"github.com/ivy-mobile/odin/rmq"
)

type RMQBroker struct {
	topic string
	rmqp  *rmq.Producer
	rmqc  *rmq.PushConsumer
}

func NewRMQBroker(topic, gameName, nodeId string, cfg conf.RocketMQConfig, msgHandler func(string, string, string, []byte)) (*RMQBroker, error) {
	// producer
	p, err := rmq.NewProducer(cfg.Endpoint, cfg.Namespace, cfg.Group, &credentials.SessionCredentials{
		AccessKey:     cfg.AccessKey,
		AccessSecret:  cfg.SecretKey,
		SecurityToken: cfg.SecurityToken,
	}, golang.WithTopics(topic))
	if err != nil {
		return nil, fmt.Errorf("broker new producer err: %v", err)
	}
	// consumer
	c, err := rmq.NewPushConsumerWithOption(cfg.Endpoint, cfg.Namespace, cfg.NodeGroup, time.Second*5, &credentials.SessionCredentials{
		AccessKey:     cfg.AccessKey,
		AccessSecret:  cfg.SecretKey,
		SecurityToken: cfg.SecurityToken,
	},
		golang.WithPushSubscriptionExpressions(map[string]*golang.FilterExpression{
			topic: golang.NewFilterExpression(gameName), // 按 gameName tag 过滤
		}),
		golang.WithPushMessageListener(&golang.FuncMessageListener{
			Consume: func(mv *golang.MessageView) golang.ConsumerResult {
				tag := mv.GetTag()
				node := mv.GetProperties()["node"]
				//fmt.Println("|||| ", *tag, node)
				//if *tag == gameName && nodeId == node {
				msgHandler(*tag, node, mv.GetMessageId(), mv.GetBody())
				//}
				return golang.SUCCESS
			},
		}),
		golang.WithPushConsumptionThreadCount(20),
		golang.WithPushMaxCacheMessageCount(1024),
	)

	if err != nil {
		_ = p.Close()
		return nil, fmt.Errorf("broker new consumer err: %v", err)
	}
	//if err := c.SubscribeByTag(topic, gameName); err != nil {
	//	return nil, err
	//}
	return &RMQBroker{
		topic: topic,
		rmqp:  p,
		rmqc:  c,
	}, nil
}

var _ Broker = (*RMQBroker)(nil)

func (r *RMQBroker) SendMessage(uid int64, gameName, node string, payload []byte) (string, error) {
	m := message{
		Uid:       uid,
		Timestamp: time.Now().UnixMilli(),
		Payload:   payload,
	}
	body, err := json.Marshal(m)
	if err != nil {
		return "", fmt.Errorf("SendMessage marshal json fail: %v", err)
	}
	msg := &golang.Message{
		Topic: r.topic,
		Body:  body,
	}
	msg.SetTag(gameName)
	msg.AddProperty("node", node)
	srs, err := r.rmqp.Send(msg)
	if err != nil {
		return "", fmt.Errorf("SendMessage send fail: %v", err)
	}
	return srs[0].MessageID, nil
}

func (r *RMQBroker) Close() error {

	if r.rmqc != nil {
		if err := r.rmqc.Close(); err != nil {
			return err
		}
	}
	if r.rmqp != nil {
		if err := r.rmqp.Close(); err != nil {
			return err
		}
	}
	return nil
}
