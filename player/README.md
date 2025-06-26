# 玩家模块
该模块实现了游戏中玩家相关的基础功能，包含玩家接口定义、基础玩家结构、玩家管理器以及消息处理器等。

## 模块结构
| 文件名称 | 功能描述 |
| --- | --- |
| `player.go` | 定义玩家统一接口 `Player` |
| `base.go` | 实现基础玩家结构 `Base`，包含对象池管理 |
| `manager.go` | 实现玩家管理器 `Manager`，用于管理玩家集合 |
| `handler.go` | 定义玩家消息发送器接口 `MsgHandler` |
| `player_test.go` | 包含玩家模块的测试代码 |

## 核心组件
### `Player` 接口
统一的玩家接口，定义了玩家的基本操作，如获取 ID、发送消息、设置离线状态等。

### `Base` 结构体
基础玩家结构，实现了 `Player` 接口。使用对象池进行性能优化，支持玩家操作的协程执行和超时处理。

#### 主要方法
- `GetBase()`: 从对象池获取一个 `Base` 实例
- `PutBase()`: 将 `Base` 实例放回对象池
- `Go()`: 玩家协程执行操作，支持超时处理
- `Close()`: 关闭玩家并释放资源

### `Manager` 结构体
玩家管理器，使用 `sync.Map` 来管理玩家集合，保证并发安全。

#### 主要方法
- `NewManager()`: 新建玩家管理器
- `Add()`: 添加玩家到管理器
- `Remove()`: 从管理器中删除玩家
- `Get()`: 根据 ID 获取玩家
- `Range()`: 遍历管理器中的所有玩家

### `MsgHandler` 接口
玩家消息发送器接口，定义了消息发送方法。

## 使用示例
### 创建玩家
```go
package main

import (
    "time"
    "github.com/ivy-mobile/odin/player"
)

func main() {
    // 实现 MsgHandler 接口
    type MyMsgHandler struct{}
    func (h *MyMsgHandler) SendMessage(seq uint64, uid int64, route, version string, msgID uint64, payload any) error {
        return nil
    }

    handler := &MyMsgHandler{}
    p := player.GetBase(handler, 123, "player1", "avatar1", 3*time.Second)
    defer player.PutBase(p)
}
```

### 使用玩家管理器

```go
package main

import "github.com/ivy-mobile/odin/player"

func main() {
    manager := player.NewManager()
    // 假设 p 是一个 Player 实例
    manager.Add(p)
    if p, ok := manager.Get(123); ok {
        // 使用 player
    }
}