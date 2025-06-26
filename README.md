# Odin 工具集

Odin（奥丁）是北欧神话中众神之王，他拥有强大的智慧与力量，掌管战争、智慧、魔法等。本项目以 Odin 命名，寓意该工具集如同众神之王一般，为游戏开发提供强大、全面且智能的支持。

Odin 是一个 Go 语言编写的工具集，提供了游戏开发相关的各种模块，方便开发者快速搭建游戏服务。

## 模块概览

| 模块名 | 功能简介 | 文档地址 |
| --- | --- | --- |
| [encoding](#encoding-编码模块) | 提供多种数据编解码器的实现，支持常见数据格式的编码和解码 | **[前往](encoding/README.md)** |
| [eventbus](#eventbus-事件总线模块) | 实现多种消息队列的事件总线，支持 Redis、NATS、RocketMQ 等 | **[前往](eventbus/README.md)** |
| [game](#game-游戏核心模块) | 游戏核心服务逻辑，包含游戏启动、消息处理、玩家和房间管理等 | **[前往](game/README.md)** |
| [gate](#gate-网关模块) | 基于 WebSocket 的业务网关，负责用户连接管理和消息转发 | **[前往](gate/README.md)** |
| [locator](#locator-定位模块) | 基于 Redis 的分布式游戏服务用户节点定位组件 | **[前往](locator/README.md)** |
| [packet](#packet-消息包处理模块) | 消息打包、解包和心跳处理模块 | **[前往](packet/README.md)** |
| [player](#player-玩家模块) | 玩家相关功能模块，包含玩家接口、基础玩家结构和玩家管理器 | **[前往](player/README.md)** |
| [room](#room-房间模块) | 游戏房间实现模块，包含房间管理、状态控制和玩家管理 | **[前往](room/README.md)** |
| [xutil](#xutil-工具模块) | 包含多个实用工具子模块，如缓冲区操作、任务池等 |  |

## 快速开始

### 1. 克隆项目
```bash
go get github.com/ivy-mobile/odin.git
```

### 2. 使用示例
```go
package main

import (
    "fmt"
    "github.com/ivy-mobile/odin/encoding/json"
)

// 定义一个示例结构体
type Person struct {
    Name string `json:"name"`
    Age  int    `json:"age"`
}

func main() {
    // 注册 JSON 编解码器

    // 创建一个 Person 实例
    p := Person{Name: "Alice", Age: 30}

    // 编码为 JSON
    data, err := json.Marshal(p)
    if err != nil {
        fmt.Println("编码失败:", err)
        return
    }
    fmt.Println("编码结果:", string(data))

    // 解码 JSON
    var ps Person
    err = json.Unmarshal(data, &ps)
    if err != nil {
        fmt.Println("解码失败:", err)
        return
    }
    fmt.Println("解码结果:", ps)
}

```