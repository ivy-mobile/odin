package broker

import (
	"github.com/apache/rocketmq-clients/golang/v5"
	"github.com/ivy-mobile/odin/rmq"
)

// MessageView 消息视图接口, 用于解耦 golang.MessageView
type MessageView interface {
	GetBody() []byte
	GetMessageId() string
	GetTopic() string
	GetProperties() map[string]string
}

// producer 生产者接口, 用于解耦 rmq.Producer
type producer interface {
	Send(msg *golang.Message) ([]*golang.SendReceipt, error)
}

// consumer 消费者接口, 用于解耦 rmq.Consumer
type consumer interface {
	SubscribeBySQL92(topic, sql92 string, callback func(msg MessageView) error) error
	Close() error
}

// ProducerAdapter 适配 rmq.Producer 到 producer 接口
type ProducerAdapter struct {
	P *rmq.Producer
}

func (a *ProducerAdapter) Send(msg *golang.Message) ([]*golang.SendReceipt, error) {
	return a.P.Send(msg)
}

// ConsumerAdapter 适配 rmq.Consumer 到 consumer 接口
type ConsumerAdapter struct {
	C *rmq.Consumer
}

func (a *ConsumerAdapter) SubscribeBySQL92(topic, sql92 string, callback func(msg MessageView) error) error {
	return a.C.SubscribeBySQL92(topic, sql92, func(mv *golang.MessageView) error {
		return callback(mv)
	})
}

func (a *ConsumerAdapter) Close() error {
	return a.C.Close()
}

// --- Mock 实现, 仅用于测试 ---

type mockProducer struct {
	sendFunc func(msg *golang.Message) ([]*golang.SendReceipt, error)
}

func (m *mockProducer) Send(msg *golang.Message) ([]*golang.SendReceipt, error) {
	if m.sendFunc != nil {
		return m.sendFunc(msg)
	}
	return nil, nil
}

type mockConsumer struct {
	subscribeBySQL92Func func(topic, sql92 string, callback func(msg MessageView) error) error
	closeFunc            func() error
}

func (m *mockConsumer) SubscribeBySQL92(topic, sql92 string, callback func(msg MessageView) error) error {
	if m.subscribeBySQL92Func != nil {
		return m.subscribeBySQL92Func(topic, sql92, callback)
	}
	return nil
}

func (m *mockConsumer) Close() error {
	if m.closeFunc != nil {
		return m.closeFunc()
	}
	return nil
}

// mockMessageView 实现 MessageView 接口, 用于测试回调
type mockMessageView struct {
	body       []byte
	messageId  string
	topic      string
	properties map[string]string
}

func (m *mockMessageView) GetBody() []byte      { return m.body }
func (m *mockMessageView) GetMessageId() string { return m.messageId }
func (m *mockMessageView) GetTopic() string     { return m.topic }
func (m *mockMessageView) GetProperties() map[string]string {
	if m.properties != nil {
		return m.properties
	}
	return map[string]string{}
}
