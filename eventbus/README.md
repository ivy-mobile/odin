# Ivy Go Kit Eventbus 模块

## 概述
Ivy Go Kit 的 Eventbus 模块提供了一个事件总线的实现，支持多种消息中间件（如 Redis、NATS、RocketMQ 等）作为后端。该模块允许开发者方便地发布和订阅事件，实现组件间的解耦通信。

## 代码结构
```
.eventbus.go         # 事件总线接口
redis/              # Redis 实现
  |-- eventbus.go   # Redis 事件总线实现
  |-- eventbus_test.go # Redis 测试用例
  |-- options.go    # Redis 配置选项
nats/               # NATS 实现
  |-- eventbus.go   # NATS 事件总线实现
  |-- eventbus_test.go # NATS 测试用例
  |-- options.go    # NATS 配置选项
rocketmq/           # RocketMQ 实现
  |-- eventbus.go   # RocketMQ 事件总线实现
  |-- eventbus_test.go # RocketMQ 测试用例
  |-- options.go    # RocketMQ 配置选项
```

## 核心概念
- **Eventbus 接口**：定义了事件总线的基本操作，包括发布事件、订阅事件、取消订阅和关闭事件总线。

## 主要实现
### RocketMQ 事件总线 (`rocketmq.Eventbus`)
基于 RocketMQ 简单消息 实现的事件总线。

### NATS 事件总线 (`nats.Eventbus`)
基于 NATS 消息系统实现的事件总线。

### Redis 事件总线 (`redis.Eventbus`)
基于 Redis 实现的事件总线，支持自定义配置和外部客户端。

## 使用示例
### 订阅事件
```go
import (
    "context"
    "github.com/ivy-mobile/odin/eventbus/redis"
    "log"
)

func loginEventHandler(data []byte) {
    log.Printf("%+v\n", data)
}

func main() {
    eb := redis.NewEventbus()
    ctx := context.Background()

    err := eb.Subscribe(ctx, "login", loginEventHandler)
    if err != nil {
        log.Fatal(err)
    }
}
```

### 发布事件
```go
import (
    "context"
    "github.com/ivy-mobile/odin/eventbus/redis"
)

func main() {
    eb := redis.NewEventbus()
    ctx := context.Background()

    err := eb.Publish(ctx, "login", "login")
    if err != nil {
        panic(err)
    }
}
```

## 自定义事件总线
可根据需求自定义事件总线，实现 `Eventbus` 接口，即实现 `Publish`、`Subscribe`、`Unsubscribe`、`Close` 方法。如: Kafka、ActiveMQ等。

## 配置选项
不同的消息中间件实现提供了不同的配置选项，可以通过 `Option` 函数来配置这些选项。