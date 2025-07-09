# Game 模块文档

该模块实现了游戏服务的核心逻辑，负责管理游戏的启动、消息处理、玩家和房间管理等功能。

## 模块结构
| 文件名称 | 功能描述 |
| --- | --- |
| `context.go` | 定义游戏上下文接口和默认实现，用于处理消息和玩家交互 |
| `game.go` | 定义 `Game` 结构体，实现游戏的核心功能，包括启动、关闭、消息处理等操作 |
| `game_test.go` | 包含 `Game` 模块的测试代码，演示路由注册和消息模拟 |
| `handler.go` | 定义消息处理器类型和通用处理器包装器，用于解析业务数据 |
| `option.go` | 定义游戏的配置选项和默认值 |

## 核心组件
### `Game` 结构体
游戏服务的核心结构体，提供游戏的启动、关闭、消息处理等功能。

#### 主要方法
- `New()`: 创建一个新的 `Game` 实例
- `Start()`: 启动游戏服务
- `shutdown()`: 关闭游戏服务
- `validateOptions()`: 验证配置选项的有效性
- `RegisterRouter()`: 注册消息路由处理器
- `SendMessage()`: 发送消息至网关

### `Context` 接口及 `defaultContext` 结构体
`Context` 接口定义了游戏上下文中的基本操作，`defaultContext` 是其默认实现，用于处理消息和玩家交互。

#### 主要方法
- `Seq()`: 获取消息序列号
- `Uid()`: 获取用户 ID
- `Player()`: 获取玩家信息
- `Resp()`: 发送响应消息
- `Push()`: 推送消息给指定玩家
- `PushToRoom()`: 推送消息给房间内所有玩家

### 消息处理器
#### `GameMessageHandler`
处理游戏消息的函数类型，接收 `Game` 实例和消息对象。

#### `Handler` 函数
通用处理器包装器，使用泛型自动解析业务数据 `payload`，避免重复编码。

## 配置选项
在 `option.go` 中定义了一系列配置选项，用于自定义游戏服务行为。主要配置选项如下：
| 配置项 | 描述 | 默认值 |
| --- | --- | --- |
| `id` | 游戏 ID | 无 |
| `name` | 游戏名称 | 无 |
| `codec` | 编解码器 | `proto.Codec` |
| `roomIdGenerator` | 房间 ID 生成器 | 无 |
| `eventbus` | 事件总线 | 无 |
| `adminCmdHandler` | 后台指令消息处理器 | 无 |

## 消息处理流程
1. **启动游戏**：调用 `Start()` 方法启动游戏服务，验证配置选项并监听外部消息。
2. **消息接收**：通过事件总线订阅网关消息和后台指令消息。
3. **消息处理**：接收到消息后，将其写入网关消息通道，由 `handlerGateMessage()` 方法解析并调用对应的路由处理器。
4. **消息发送**：使用 `SendMessage()` 方法将消息发送至网关。

## 使用示例
### 创建并启动游戏服务
```go
package main

import (
    "github.com/ivy-mobile/odin/encoding/json"
    "github.com/ivy-mobile/odin/eventbus/redis"
    "github.com/ivy-mobile/odin/game"
)

func main() {
    g := game.New(
        game.WithID("1"),
        game.WithName("test"),
        game.WithEventbus(redis.NewEventbus(
            redis.WithAddrs("localhost:6379"),
            redis.WithPassword(""),
        )),
        game.WithCodec(json.Codec),
        game.WithAdminCmdHandler(func(data []byte) {
            // 处理后台指令
        }),
    )

    // 注册路由
    g.RegisterRouter("1.0.0", "Heartbeat", game.Handler(Heartbeat))
    g.RegisterRouter("1.0.0", "Login", game.Handler(Login))

    g.Start()
}

func Login(ctx game.Context, req *struct{ Msg string }) {
    // 处理登录逻辑
}

func Heartbeat(ctx game.Context, req *struct{}) {
    // 处理心跳逻辑
}