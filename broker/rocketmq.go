package broker

import (
	"fmt"
	"time"

	"github.com/apache/rocketmq-clients/golang/v5"

	"github.com/ivy-mobile/odin/encoding/json"
	"github.com/ivy-mobile/odin/rmq"
)

type RMQBroker struct {
	topic string
	rmqp  *rmq.Producer
	rmqc  *rmq.Consumer
}

func NewRMQBroker(topic string, producer *rmq.Producer, consumer *rmq.Consumer) *RMQBroker {
	return &RMQBroker{
		topic: topic,
		rmqp:  producer,
		rmqc:  consumer,
	}
}

var _ Broker = (*RMQBroker)(nil)

func getNode(gameName, node string) string {
	return gameName + "." + node
}

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
	msg.AddProperty("node", getNode(gameName, node))
	srs, err := r.rmqp.Send(msg)
	if err != nil {
		return "", fmt.Errorf("SendMessage send fail: %v", err)
	}
	return srs[0].MessageID, nil
}

func (r *RMQBroker) ReceiveMessage(gameName, node string, fn func(uid int64, msgId string, timestamp int64, data []byte)) error {
	return r.rmqc.SubscribeBySQL92(r.topic, fmt.Sprintf("node='%s'", getNode(gameName, node)), func(msg *golang.MessageView) error {
		var m message
		if err := json.Unmarshal(msg.GetBody(), &m); err != nil {
			return fmt.Errorf("ReceiveMessage unmarshal fail: %v", err)
		}
		fn(m.Uid, msg.GetMessageId(), m.Timestamp, m.Payload)
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
		if err := r.rmqp.Close(); err != nil {
			return err
		}
	}
	return nil
}
