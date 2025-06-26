# xpool 任务池模块

`xpool` 是 `github.com/ivy-mobile/odin` 项目中的一个任务池模块，借助 `ants` 库实现高效的任务池管理。该模块提供了任务池的初始化、任务添加和资源释放等功能。

## 功能概述
- **任务池初始化**：支持自定义任务池大小、是否非阻塞和是否禁用清除功能。
- **任务添加**：可将任务添加到任务池执行，若添加失败则使用 `xgo.Go` 方法执行任务。
- **资源释放**：提供释放任务池资源的方法。

## 代码结构
### `options.go`
- 定义任务池配置选项和默认值。
- 提供设置任务池大小、是否非阻塞和是否禁用清除功能的选项函数。

### `pool.go`
- 定义任务池接口和默认实现。
- 实现任务池的初始化、任务添加和资源释放方法。
- 提供全局任务池的设置和获取方法。

## 使用示例
```go
package main

import (
    "github.com/ivy-mobile/odin/xpool"
)

func main() {
    // 创建自定义任务池
    pool := xpool.NewPool(
        xpool.WithSize(200000),
        xpool.WithNonblocking(false),
        xpool.WithDisablePurge(false),
    )

    // 设置全局任务池
    xpool.SetPool(pool)

    // 添加任务
    xpool.AddTask(func() {
        // 任务逻辑
    })

    // 释放任务池资源
    xpool.Release()
}
```

## 配置选项
| 选项函数          | 描述                     | 默认值 |
|-------------------|--------------------------|--------|
| `WithSize`        | 设置任务池大小           | 100000  |
| `WithNonblocking` | 设置是否非阻塞           | true   |
| `WithDisablePurge`| 设置是否禁用清除功能     | true   |

## 注意事项
- 若全局任务池未初始化，调用 `AddTask` 方法时会直接使用 `xgo.Go` 方法执行任务。
- 重新设置全局任务池时，会先释放原有任务池的资源。