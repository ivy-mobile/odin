package nats_test

import (
	"context"
	"log"
	"testing"
	"time"

	"github.com/ivy-mobile/odin/eventbus/nats"
)

const (
	loginTopic = "login"
	paidTopic  = "paid"
)

func loginEventHandler(data []byte) {
	log.Printf("%+v\n", string(data))
}

func paidEventHandler(data []byte) {
	log.Printf("%+v\n", string(data))
}

func TestEventbus_Client1_Subscribe(t *testing.T) {
	var (
		err error
		eb  = nats.NewEventbus()
		ctx = context.Background()
	)

	defer eb.Close()

	err = eb.Subscribe(ctx, loginTopic, loginEventHandler)
	if err != nil {
		t.Fatal(err)
	}

	err = eb.Subscribe(ctx, paidTopic, paidEventHandler)
	if err != nil {
		t.Fatal(err)
	}

	t.Log("subscribe success")

	time.Sleep(30 * time.Second)
}

func TestEventbus_Client2_Subscribe(t *testing.T) {
	var (
		err error
		eb  = nats.NewEventbus()
		ctx = context.Background()
	)

	defer eb.Close()

	err = eb.Subscribe(ctx, loginTopic, loginEventHandler)
	if err != nil {
		t.Fatal(err)
	}

	err = eb.Subscribe(ctx, paidTopic, paidEventHandler)
	if err != nil {
		t.Fatal(err)
	}

	err = eb.Unsubscribe(context.Background(), loginTopic)
	if err != nil {
		t.Fatal(err)
	}

	t.Log("subscribe success")

	time.Sleep(30 * time.Second)
}

func TestEventbus_Publish(t *testing.T) {
	var (
		err error
		eb  = nats.NewEventbus()
		ctx = context.Background()
	)

	defer eb.Close()

	err = eb.Publish(ctx, loginTopic, []byte("login"))
	if err != nil {
		t.Fatal(err)
	}

	err = eb.Publish(ctx, paidTopic, []byte("paid"))
	if err != nil {
		t.Fatal(err)
	}

	err = eb.Publish(ctx, loginTopic, []byte("login"))
	if err != nil {
		t.Fatal(err)
	}

	t.Log("publish success")
}
