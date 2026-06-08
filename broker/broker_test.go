package broker

import (
	"context"
	"fmt"
	"sync"
	"testing"
	"time"

	"github.com/apache/rocketmq-clients/golang/v5"
	"github.com/apache/rocketmq-clients/golang/v5/credentials"
	"github.com/ivy-mobile/odin/conf"
	"github.com/stretchr/testify/require"
)

func TestRocketMQ(t *testing.T) {
	var (
		Endpoint  = "10.80.1.64:19081"
		Namespace = ""
		Group     = "TestGroup"
		Topic     = "ggggggg"
		GameName  = "pelican-prank"
		Node      = "Axxxxx"
	)
	producer, err := golang.NewProducer(&golang.Config{
		Endpoint:      Endpoint,
		NameSpace:     Namespace,
		ConsumerGroup: Group,
		Credentials:   &credentials.SessionCredentials{},
	},
	//golang.WithTopics(Topic),
	)
	if err != nil {
		t.Fatal(err)
	}
	_ = producer.Start()
	defer func() {
		_ = producer.GracefulStop()
	}()
	// new simpleConsumer instance
	simpleConsumer, err := golang.NewSimpleConsumer(&golang.Config{
		Endpoint:      Endpoint,
		NameSpace:     Namespace,
		ConsumerGroup: Group,
		Credentials:   &credentials.SessionCredentials{},
	},
		golang.WithSimpleAwaitDuration(time.Second*5),
		golang.WithSimpleSubscriptionExpressions(map[string]*golang.FilterExpression{
			Topic: golang.NewFilterExpression(GameName + "-" + Node),
		}),
	)

	if err != nil {
		t.Fatal(err)
	}
	// start simpleConsumer
	err = simpleConsumer.Start()
	if err != nil {
		t.Fatal(err)
	}

	//simpleConsumer.Subscribe(Topic, golang.NewFilterExpression(GameName+"-"+Node))

	// gracefule stop simpleConsumer
	defer func() {
		_ = simpleConsumer.GracefulStop()
	}()
	mp := &sync.Map{}
	go func() {

		for {
			mvs, err := simpleConsumer.Receive(context.TODO(), 16, time.Second*20)
			if err != nil {
				fmt.Println(err)
				continue
			}
			// ack message
			for _, mv := range mvs {
				_ = simpleConsumer.Ack(context.TODO(), mv)
				mp.Delete(mv.GetMessageId())
			}
		}
	}()

	go func() {
		var count int
		for count <= 100 {
			time.Sleep(time.Millisecond * 20)
			tag := GameName + "-" + Node
			rs, err := producer.Send(context.Background(), &golang.Message{
				Topic: Topic,
				Body:  []byte("hello world"),
				Tag:   &tag,
			})
			if err != nil {
				continue
			}
			count++
			mp.Store(rs[0].MessageID, struct{}{})
		}
	}()
	// run for a while
	time.Sleep(time.Second * 10)
	t.Logf("====== 111")
	mp.Range(func(key, value interface{}) bool {
		t.Logf(" === %v", key)
		return true
	})
}

func newBroker(Topic, GameName, Node, Endpoint, Namespace, Group string) (*RMQBroker, error) {

	return NewRMQBroker(Topic, GameName, Node, conf.RocketMQConfig{
		Endpoint:  Endpoint,
		Namespace: Namespace,
		NodeGroup: Group,
	}, func(gameName, nodeID, messageID string, msg []byte) {
		fmt.Println(Node+"<-"+nodeID, "===========================", gameName, ".", nodeID, messageID)
	})
}

//  1. 每个服务多个节点使用同一个消费组
//     1.1. 通过Tag、SQL92或业务层过滤当前节点消息: 不是当前节点的消息不ACK，这种方式其实是在利用mq的重新投递，而重新投递的时间最少为10s，业务上无法接受
//     1.2. 广播消息到所有节点: 每个节点拿到消息后业务层判断是否是当前节点的消息，不满足条件的直接丢弃即可，新版rocketmq sdk不支持广播
//  2. 每个服务多个节点使用多个消费组
//     2.1. 不管是使用Tag、SQL92或业务层过滤当前节点消息: 消费组爆炸
func TestBroker(t *testing.T) {
	var (
		Endpoint  = "10.80.1.64:19081"
		Namespace = ""
		Group     = "TestGroup"
		Topic     = "test2-topic"
		GameName  = "pelican-prank"
		Node1     = "A"
		Node2     = "B"
		Node3     = "C"
	)

	tc1, err := newBroker(Topic, GameName, Node1, Endpoint, Namespace, Group)
	require.Nil(t, err)
	_, err = newBroker(Topic, GameName, Node2, Endpoint, Namespace, Group)
	require.Nil(t, err)
	_, err = newBroker(Topic, GameName, Node3, Endpoint, Namespace, Group)
	require.Nil(t, err)

	sendNode := Node2
	go func() {
		var count int
		for count < 10 {
			time.Sleep(time.Second * 1)
			msgId, err := tc1.SendMessage(10000, GameName, sendNode, []byte("hello"))
			if err != nil {
				t.Error(err)
				return
			}
			count++
			t.Logf("[%s -> %s] time: %v, send message, msgId: %v", Node1, sendNode, time.Now(), msgId)
		}
	}()

	time.Sleep(time.Second * 40)

}
