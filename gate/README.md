# Gate 模块文档

该模块实现了一个基于 WebSocket 的业务网关，主要负责用户链接管理和消息转发。

## 模块结构
| 文件名称 | 功能描述 |
| --- | --- |
| `gate.go` | 定义 `Gate` 结构体，实现网关的核心功能，包括启动、关闭和验证等操作 |
| `gate_test.go` | 包含 `Gate` 模块的测试代码 |
| `handler.go` | 处理用户连接、断开和消息接收等逻辑 |
| `option.go` | 定义网关的配置选项和默认值 |
| `session.go` | 实现会话管理器，用于管理用户会话 |

## 核心组件
### `Gate` 结构体
业务网关的核心结构体，提供了网关的启动、关闭、验证等功能。

#### 主要方法
- `New()`: 创建一个新的 `Gate` 实例
- `Start()`: 启动网关服务
- `shutdown()`: 关闭网关服务
- `validateOptions()`: 验证配置选项的有效性

### `Sessions` 结构体
会话管理器，用于管理所有用户会话，保证并发安全。

#### 主要方法
- `NewSessions()`: 创建一个新的会话管理器
- `Get()`: 根据用户 ID 获取会话
- `Set()`: 设置用户会话
- `Remove()`: 移除用户会话
- `Send()`: 向指定用户发送二进制消息
- `SendText()`: 向指定用户发送文本消息

### 配置选项
在 `option.go` 中定义了一系列配置选项，用于自定义网关行为。主要配置选项如下：
| 配置项 | 描述 | 默认值 |
| --- | --- | --- |
| `id` | 实例 ID | 无 |
| `name` | 实例名称 | `game-gateway` |
| `port` | 端口号 | 无 |
| `pattern` | 路由匹配模式 | 无 |
| `writeWait` | 写入超时时间 | `10s` |
| `pongWait` | pong 等待时间 | `60s` |
| `pingPeriod` | ping 之间的时间间隔 | `54s` |
| `maxMessageSize` | 消息最大字节数 | `512` |
| `codec` | 编码解码器 | `proto.Codec` |
| `eventbus` | 事件总线 | 无 |

### 消息处理
在 `handler.go` 中实现了用户连接、断开和消息接收等逻辑：
- `handleConnect()`: 处理用户连接，验证用户 ID 并保存会话
- `handleDisconnect()`: 处理用户断开连接，释放资源并移除会话
- `handleMessage()`: 处理文本消息，采用 JSON 协议
- `handleMessageBinary()`: 处理二进制消息，采用 Proto 协议
- `dispatch()`: 分发消息到事件总线

## 使用示例
### 创建并启动网关
```go
package main

import (
    "time"
    "github.com/ivy-mobile/odin/encoding/json"
    "github.com/ivy-mobile/odin/gate"
)

func main() {
    g := gate.New(
        gate.WithID("test"),
        gate.WithName("test"),
        gate.WithPort(":8080"),
        gate.WithCodec(json.Codec),
        gate.WithPattern("/ws"),
        gate.WithWriteWait(10*time.Second),
        gate.WithPongWait(60*time.Second),
        gate.WithPingPeriod(30*time.Second),
        gate.WithMaxMessageSize(1024),
    )
    g.Start()
}