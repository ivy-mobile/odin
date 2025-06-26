package rocketmq_test

import (
	"context"
	"fmt"
	"log"
	"testing"
	"time"

	"github.com/ivy-mobile/odin/eventbus/rocketmq"

	"github.com/apache/rocketmq-clients/golang"
	"github.com/apache/rocketmq-clients/golang/credentials"
)

const (
	Topic         = "TestTopic"
	NameSpace     = ""
	Endpoint      = "10.80.40.36:8081"
	ConsumerGroup = "TestGroup"
	AccessKey     = ""
	SecretKey     = ""
)

var (
	// maximum waiting time for receive func
	awaitDuration = time.Second * 5
	// maximum number of messages received at one time
	maxMessageNum int32 = 16
	// invisibleDuration should > 20s
	invisibleDuration = time.Second * 20
	// receive messages in a loop
)

func TestRocketmq(t *testing.T) {

	simpleConsumer, err := golang.NewSimpleConsumer(&golang.Config{
		Endpoint:      Endpoint,
		NameSpace:     NameSpace,
		ConsumerGroup: ConsumerGroup,
		Credentials: &credentials.SessionCredentials{
			AccessKey:    AccessKey,
			AccessSecret: SecretKey,
		},
	},
		golang.WithAwaitDuration(awaitDuration),
		golang.WithSubscriptionExpressions(map[string]*golang.FilterExpression{
			Topic: golang.SUB_ALL,
		}),
	)
	if err != nil {
		log.Fatal(err)
	}
	// start simpleConsumer
	err = simpleConsumer.Start()
	if err != nil {
		log.Fatal(err)
	}
	// gracefule stop simpleConsumer
	defer simpleConsumer.GracefulStop()

	go func() {
		for {
			mvs, err := simpleConsumer.Receive(context.TODO(), maxMessageNum, invisibleDuration)
			if err != nil {
				fmt.Println(err)
			}
			// ack message
			for _, mv := range mvs {
				simpleConsumer.Ack(context.TODO(), mv)
				fmt.Println(mv)
			}
			fmt.Println("wait a moment")
			fmt.Println()
			time.Sleep(time.Second * 3)
		}
	}()
	// run for a while
	time.Sleep(time.Minute)
}

func TestPubSub(t *testing.T) {

	eb, err := rocketmq.NewEventbus(
		rocketmq.WithNameSpace(NameSpace),
		rocketmq.WithEndpoint(Endpoint),
		rocketmq.WithConsumerGroup(ConsumerGroup),
		rocketmq.WithAccessKey(AccessKey),
		rocketmq.WithSecretKey(SecretKey),
	)
	if err != nil {
		t.Fatal(err)
	}

	eb.Subscribe(context.Background(), Topic, func(data []byte) {
		fmt.Println("Topic message: ", string(data))
	})

	go func() {
		// for {
		// 	time.Sleep(time.Second)
		// 	if err := eb.Publish(context.Background(), Topic, []byte("hello")); err != nil {
		// 		t.Log(err)
		// 	}
		// }
	}()

	select {}

}
