# Room 模块文档

## 模块概述
Room 模块提供了游戏房间的基础实现，包括房间管理、房间状态控制、玩家管理等功能。适用于实现单个房间最多 20 人的游戏场景。

## 模块结构
| 文件名称 | 功能描述 |
| --- | --- |
| `base.go` | 定义基础房间 `Base` 结构体，提供房间的基础能力，如房间 ID、房间名、状态管理、玩家管理等 |
| `base_action.go` | 定义房间操作 `Action` 和操作结果 `ActResult` 结构体 |
| `base_option.go` | 定义房间配置选项和默认值，以及配置选项的设置函数 |
| `manager.go` | 定义房间管理器 `Manager`，用于管理多个房间 |
| `room.go` | 定义房间接口 `Room`，规定了房间的基本方法 |
| `room_test.go` | 包含房间模块的测试代码，模拟了 UNO 牌桌的测试用例 |
| `state.go` | 定义房间状态统一接口 `RoomState` |
| `examples/uno/app.go` | 提供 UNO 游戏房间的模拟实现 |
| `examples/uno/state/gaming.go` | 定义 UNO 游戏的游戏中状态 `Gaming` |

## 核心组件
### `Base` 结构体
基础房间实现，适用于单个房间 20 人以下的场景，提供房间的基础能力，包括：
- 房间 ID、房间名、房间状态管理
- 房间内玩家管理
- 房间内消息广播
- 房间内玩家操作处理

### `Manager` 结构体
房间管理器，用于管理多个房间，提供以下方法：
- `NewManager()`: 创建房间管理器
- `Add(room Room)`: 添加房间
- `Get(id int)`: 获取房间
- `Remove(id int)`: 删除房间
- `Range(fn func(id int, room Room) bool)`: 遍历房间

### `Room` 接口
定义了房间的基本方法，实现该接口的结构体可作为房间使用。

### `RoomState` 接口
定义了房间状态的统一接口，包含 `ID()`、`Name()` 和 `Timeout()` 方法。

## 使用示例
### 创建基础房间
```go
package main

import (
    "github.com/ivy-mobile/odin/room"
    "time"
)

// 假设已经实现了 RoomState 接口的状态结构体
var idleState room.RoomState

func main() {
    // 创建基础房间
    baseRoom, err := room.NewBaseRoom(
        room.With(1, "test-room"),
        room.WithMaxPlayerCount(20),
        room.WithIdleState(idleState),
        room.WithStateTimeoutHandler(func() (uint16, error) {
            // 状态超时处理逻辑
            return 0, nil
        }),
    )
    if err != nil {
        panic(err)
    }
    
    // 启动房间服务
    baseRoom.Serve()
}