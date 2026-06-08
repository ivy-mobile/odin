package broker

import (
	"context"
	"fmt"
	"strconv"
	"sync"
	"testing"
	"time"

	"github.com/apache/rocketmq-clients/golang/v5"
	"github.com/apache/rocketmq-clients/golang/v5/credentials"
	"github.com/ivy-mobile/odin/rmq"
)

func newBroker(Topic, Endpoint, Namespace, Group string) (*RMQBroker, error) {

	// 生产者
	rmqp, err := rmq.NewProducer(Endpoint, Namespace, Group, &credentials.SessionCredentials{})
	if err != nil {
		return nil, err
	}
	rmqc, err := rmq.NewConsumer(Endpoint, Namespace, Group, time.Second*5, &credentials.SessionCredentials{})
	if err != nil {
		return nil, err
	}
	return NewRMQBroker(Topic, rmqp, rmqc), nil
}

func TestRMQBroker(t *testing.T) {
	var (
		Endpoint  = "10.80.1.64:19081"
		Namespace = ""
		Group1    = "TestGroup1"
		Group2    = "TestGroup2"
		Group3    = "TestGroup3"
		Topic     = "ggggggg"
		GameName  = "pelican-prank"
		Node1     = "ABCDEffffdasdefs1"
		Node2     = "ABCDEffffdasdefs2"
		Node3     = "ABCDEffffdasdefs3"
	)

	mp1 := &sync.Map{}
	mp2 := &sync.Map{}

	tc1, _ := newBroker(Topic, Endpoint, Namespace, Group1)
	tc2, _ := newBroker(Topic, Endpoint, Namespace, Group2)
	tc3, _ := newBroker(Topic, Endpoint, Namespace, Group3)

	if err := tc1.ReceiveMessage(GameName, Node1, func(uid int64, msgId string, timestamp int64, msg []byte) {
		mp1.Store(msgId, struct{}{})
		//t.Logf("time: %v, get message, msgId: %v, cost: %v", time.Now(), msgId, time.Since(time.UnixMilli(timestamp)))
	}); err != nil {
		t.Fatal(err)
	}

	if err := tc2.ReceiveMessage(GameName, Node2, func(uid int64, msgId string, timestamp int64, msg []byte) {
		//mp1.Store(msgId, struct{}{})
		t.Logf("[222] time: %v, get message, msgId: %v, cost: %v", time.Now(), msgId, time.Since(time.UnixMilli(timestamp)))
	}); err != nil {
		t.Fatal(err)
	}

	if err := tc3.ReceiveMessage(GameName, Node3, func(uid int64, msgId string, timestamp int64, msg []byte) {
		//mp1.Store(msgId, struct{}{})
		t.Logf("[333] time: %v, get message, msgId: %v, cost: %v", time.Now(), msgId, time.Since(time.UnixMilli(timestamp)))
	}); err != nil {
		t.Fatal(err)
	}

	wg := &sync.WaitGroup{}
	wg.Add(1)
	go func() {
		defer wg.Done()
		var count int64
		for count < 1000 {
			//time.Sleep(time.Millisecond * 20)
			count++
			msgId, err := tc1.SendMessage(10000, GameName, Node1, []byte(strconv.FormatInt(count, 10)))
			if err != nil {
				t.Error(err)
				return
			}
			mp2.Store(msgId, struct{}{})
			//t.Logf("time: %v, send message, msgId: %v", time.Now(), msgId)
		}
	}()
	wg.Wait()
	time.Sleep(time.Second * 2)

	var n1, n2 int
	mp1.Range(func(key, value interface{}) bool {
		n1++
		return true
	})
	mp2.Range(func(key, value interface{}) bool {
		n2++
		return true
	})
	t.Logf("n1:%d, n2:%d", n1, n2)
}

func TestRocketMQ(t *testing.T) {
	var (
		Endpoint  = "10.80.1.64:19081"
		Namespace = ""
		Group     = "TestGroup1"
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
		golang.WithAwaitDuration(time.Second*5),
		golang.WithSubscriptionExpressions(map[string]*golang.FilterExpression{
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
