# rmq 模块文档

## 概述

`rmq` 是对 Apache RocketMQ Go 5.x 客户端的一层轻量封装，提供了生产者与简单消费者两个入口，简化了启动、关闭和常用发送、订阅流程。

## 文件说明

```text
consumer.go    # 简单消费者封装
producer.go    # 生产者封装
rmq_test.go    # 基础测试
```

## 核心能力

### `Producer`

- `NewProducer(...)`：创建并启动生产者
- `Send(...)`：同步发送消息
- `SendAsync(...)`：异步发送消息
- `SendWithTransaction(...)`：发送事务消息
- `Close()`：优雅关闭生产者

### `Consumer`

- `NewConsumer(...)`：创建并启动简单消费者
- `Subscribe(...)`：订阅某个 Topic 的全部消息
- `SubscribeByTag(...)`：按 Tag 过滤订阅
- `SubscribeBySQL92(...)`：按 SQL92 表达式过滤订阅
- `Unsubscribe(...)`：取消订阅
- `Close()`：停止接收并优雅关闭消费者

消费者在回调返回 `nil` 时会自动执行 `Ack`；如果返回错误，则该消息不会被确认。

## 使用示例

### 创建生产者

```go
package main

import (
	golang "github.com/apache/rocketmq-clients/golang/v5"
	"github.com/apache/rocketmq-clients/golang/v5/credentials"
	"github.com/ivy-mobile/odin/rmq"
)

func main() {
	cred := &credentials.SessionCredentials{
		AccessKey:    "access-key",
		AccessSecret: "secret-key",
	}

	producer, err := rmq.NewProducer(
		"127.0.0.1:8081",
		"default",
		"demo-producer",
		cred,
	)
	if err != nil {
		panic(err)
	}
	defer producer.Close()

	_, err = producer.Send(&golang.Message{
		Topic: "demo-topic",
		Body:  []byte("hello odin"),
	})
	if err != nil {
		panic(err)
	}
}
```

### 创建消费者

```go
package main

import (
	"time"

	golang "github.com/apache/rocketmq-clients/golang/v5"
	"github.com/apache/rocketmq-clients/golang/v5/credentials"
	"github.com/ivy-mobile/odin/rmq"
)

func main() {
	cred := &credentials.SessionCredentials{
		AccessKey:    "access-key",
		AccessSecret: "secret-key",
	}

	consumer, err := rmq.NewConsumer(
		"127.0.0.1:8081",
		"default",
		"demo-consumer",
		5*time.Second,
		cred,
	)
	if err != nil {
		panic(err)
	}
	defer consumer.Close()

	if err := consumer.Subscribe("demo-topic", func(msg *golang.MessageView) error {
		return nil
	}); err != nil {
		panic(err)
	}

	select {}
}
```
