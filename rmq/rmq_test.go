package rmq

import (
	"fmt"
	"testing"
	"time"

	"github.com/apache/rocketmq-clients/golang/v5"
	"github.com/apache/rocketmq-clients/golang/v5/credentials"
	"github.com/stretchr/testify/assert"
)

func TestRMQ_TAG(t *testing.T) {

	var (
		Endpoint  = "10.80.1.64:19081"
		Namespace = ""
		TAG       = "test_tag"
		Topic     = "GameTest"
		Group     = "TestGroup"
	)
	// 生产者
	p, err := NewProducer(Endpoint, Namespace, Group, &credentials.SessionCredentials{})
	assert.Equal(t, err, nil)

	for i := range 10 {
		msg := &golang.Message{
			Topic: Topic,
			Body:  []byte(fmt.Sprintf("test-2-%d", i)),
		}
		msg.SetTag(TAG)
		//msg.SetMessageGroup(golang.SPAN_ATTRIBUTE_VALUE_ROCKETMQ_NORMAL_MESSAGE)      // 简单消息 - 默认
		//msg.SetMessageGroup(golang.SPAN_ATTRIBUTE_VALUE_ROCKETMQ_FIFO_MESSAGE)        // 顺序消息
		//msg.SetMessageGroup(golang.SPAN_ATTRIBUTE_VALUE_ROCKETMQ_DELAY_MESSAGE)       // 延迟消息
		//msg.SetMessageGroup(golang.SPAN_ATTRIBUTE_VALUE_ROCKETMQ_TRANSACTION_MESSAGE) // 事务消息
		_, err = p.Send(msg)
		assert.Equal(t, err, nil)
	}

	// 消费者
	c, err := NewConsumer(Endpoint, Namespace, Group, time.Second*5, &credentials.SessionCredentials{})
	assert.Equal(t, err, nil)

	err = c.SubscribeByTag(Topic, TAG, func(msg *golang.MessageView) error {
		fmt.Println("msg:", string(msg.GetBody()))
		return nil
	})
	if err != nil {
		t.Errorf("Subscribe()1 error = %v", err)
		return
	}
	time.Sleep(time.Second * 5)

	_ = c.Close()
	_ = p.Close()
}

func TestRMQ_ALL(t *testing.T) {

	var (
		Endpoint  = "10.80.1.64:19081"
		Namespace = ""
		TAG       = "test_tag"
		Topic     = "GameTest"
		Group     = "TestGroup"
	)
	// 生产者
	p, err := NewProducer(Endpoint, Namespace, Group, &credentials.SessionCredentials{})
	assert.Equal(t, err, nil)

	for i := range 10 {
		msg := &golang.Message{
			Topic: Topic,
			Body:  []byte(fmt.Sprintf("test-2-%d", i)),
		}
		msg.SetTag(TAG)
		//msg.SetMessageGroup(golang.SPAN_ATTRIBUTE_VALUE_ROCKETMQ_NORMAL_MESSAGE)      // 简单消息 - 默认
		//msg.SetMessageGroup(golang.SPAN_ATTRIBUTE_VALUE_ROCKETMQ_FIFO_MESSAGE)        // 顺序消息
		//msg.SetMessageGroup(golang.SPAN_ATTRIBUTE_VALUE_ROCKETMQ_DELAY_MESSAGE)       // 延迟消息
		//msg.SetMessageGroup(golang.SPAN_ATTRIBUTE_VALUE_ROCKETMQ_TRANSACTION_MESSAGE) // 事务消息
		_, err = p.Send(msg)
		assert.Equal(t, err, nil)
	}

	// 消费者
	c, err := NewConsumer(Endpoint, Namespace, Group, time.Second*5, &credentials.SessionCredentials{})
	assert.Equal(t, err, nil)

	err = c.Subscribe(Topic, func(msg *golang.MessageView) error {
		fmt.Println("msg:", string(msg.GetBody()))
		return nil
	})
	if err != nil {
		t.Errorf("Subscribe()1 error = %v", err)
		return
	}
	time.Sleep(time.Second * 5)
	_ = c.Close()
	_ = p.Close()
}

func TestRMQ_SQL92(t *testing.T) {

	var (
		Endpoint  = "10.80.1.64:19081"
		Namespace = ""
		TAG       = "test_tag"
		Topic     = "GameTest"
		Group     = "TestGroup"
	)
	// 生产者
	p, err := NewProducer(Endpoint, Namespace, Group, &credentials.SessionCredentials{})
	assert.Equal(t, err, nil)

	for i := range 10 {
		msg := &golang.Message{
			Topic: Topic,
			Body:  []byte(fmt.Sprintf("test-2-%d", i)),
		}
		msg.SetTag(TAG)
		msg.AddProperty("p1", "v1")
		//msg.SetMessageGroup(golang.SPAN_ATTRIBUTE_VALUE_ROCKETMQ_NORMAL_MESSAGE)      // 简单消息 - 默认
		//msg.SetMessageGroup(golang.SPAN_ATTRIBUTE_VALUE_ROCKETMQ_FIFO_MESSAGE)        // 顺序消息
		//msg.SetMessageGroup(golang.SPAN_ATTRIBUTE_VALUE_ROCKETMQ_DELAY_MESSAGE)       // 延迟消息
		//msg.SetMessageGroup(golang.SPAN_ATTRIBUTE_VALUE_ROCKETMQ_TRANSACTION_MESSAGE) // 事务消息
		_, err = p.Send(msg)
		assert.Equal(t, err, nil)
	}

	// 消费者
	c, err := NewConsumer(Endpoint, Namespace, Group, time.Second*5, &credentials.SessionCredentials{})
	assert.Equal(t, err, nil)

	err = c.SubscribeBySQL92(Topic, "p1='v1'", func(msg *golang.MessageView) error {
		fmt.Println("msg:", string(msg.GetBody()))
		return nil
	})
	if err != nil {
		t.Errorf("Subscribe()1 error = %v", err)
		return
	}
	time.Sleep(time.Second * 5)
	_ = c.Close()
	_ = p.Close()
}
