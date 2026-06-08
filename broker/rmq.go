package broker

import (
	"errors"
	"fmt"
	"time"

	"github.com/apache/rocketmq-clients/golang/v5"
	"github.com/ivy-mobile/odin/encoding/json"
	"github.com/ivy-mobile/odin/rmq"
	"github.com/ivy-mobile/odin/xutil/xid"
)

type RMQBroker struct {
	topic string
	rmqp  producer
	rmqc  consumer
}

// NewRMQBroker 创建 RMQBroker
func NewRMQBroker(topic string, producer *rmq.Producer, consumer *rmq.Consumer) *RMQBroker {
	return &RMQBroker{
		topic: topic,
		rmqp:  &ProducerAdapter{P: producer},
		rmqc:  &ConsumerAdapter{C: consumer},
	}
}

// newRMQBroker 创建 RMQBroker (支持接口注入, 便于测试)
func newRMQBroker(topic string, p producer, c consumer) *RMQBroker {
	return &RMQBroker{
		topic: topic,
		rmqp:  p,
		rmqc:  c,
	}
}

var _ Broker = (*RMQBroker)(nil)

func (r *RMQBroker) SendMessage(uid int64, _, node string, payload []byte) (string, error) {
	m := message{
		UUID:      xid.UUID(),
		Uid:       uid,
		Timestamp: time.Now().UnixMilli(),
		Payload:   payload,
	}
	body, err := json.Marshal(m)
	if err != nil {
		return "", fmt.Errorf("SendMessage marshal json fail: %w", err)
	}
	msg := &golang.Message{
		Topic: r.topic,
		Body:  body,
	}
	msg.AddProperty("node", node)
	srs, err := r.rmqp.Send(msg)
	if err != nil {
		return "", fmt.Errorf("SendMessage send fail: %w", err)
	}
	if len(srs) == 0 {
		return "", errors.New("SendMessage send fail: empty send receipts")
	}
	return srs[0].MessageID, nil
}

func (r *RMQBroker) ReceiveMessage(_, node string, fn func(uid int64, msgId string, timestamp int64, data []byte)) error {
	if fn == nil {
		return errors.New("ReceiveMessage callback is nil")
	}
	return r.rmqc.SubscribeBySQL92(r.topic, fmt.Sprintf("node='%s'", node), func(mv MessageView) error {
		var m message
		if err := json.Unmarshal(mv.GetBody(), &m); err != nil {
			return fmt.Errorf("ReceiveMessage unmarshal fail: %w", err)
		}
		fn(m.Uid, mv.GetMessageId(), m.Timestamp, m.Payload)
		return nil
	})
}

func (r *RMQBroker) Close() error {
	if r.rmqc != nil {
		if err := r.rmqc.Close(); err != nil {
			return err
		}
	}
	if r.rmqp != nil {
		// producer 没有 Close 方法在接口中, 但实际实现中有
		// 这里通过类型断言来关闭
		if closer, ok := r.rmqp.(interface{ Close() error }); ok {
			if err := closer.Close(); err != nil {
				return err
			}
		}
	}
	return nil
}
