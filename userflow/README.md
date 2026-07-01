# UserFlow

一个高性能的用户流量管理器，为每个用户维护独立的请求队列和处理协程，保证用户维度请求处理的顺序性。

## 特性

- ✅ **顺序保证**：每个用户的请求在独立协程中顺序处理
- ✅ **限流控制**：基于 Token Bucket 算法的用户级限流
- ✅ **队列管理**：可配置的请求队列大小
- ✅ **关闭控制**：`Close` 会取消上下文并等待 worker 退出，超时返回错误
- ✅ **错误返回**：Submit 返回标准错误，便于上层处理
- ✅ **指标收集**：可选的性能指标收集
- ✅ **并发提交**：支持多协程提交消息
- ✅ **资源管理**：支持通过 `KickUser` 或 `Close` 释放用户资源
- ✅ **Panic 恢复**：自动恢复事件处理中的 panic
- ✅ **上下文传播**：支持上下文取消和超时控制

## 快速开始

```go
package main

import (
    "context"
    "fmt"
    "sword-ball/pkg/userflow"
    "github.com/ivy-mobile/odin/envelope"
)

func main() {
    // 创建事件处理器
    handler := func(ctx context.Context, msg *envelope.InputMessage) {
        userID := msg.GetHeader().GetUid()
        route := msg.GetRoute()
        fmt.Printf("Processing: user=%d, route=%s\n", userID, route)
    }

    // 创建流量管理器（使用默认配置）
    flow, err := userflow.New(handler)
    if err != nil {
        panic(err)
    }
    defer flow.Close()

    // 提交消息
    msg := &envelope.InputMessage{
        Header: &envelope.Header{Uid: 1001},
        Route:  "game.login",
    }
    
    if err := flow.Submit(msg); err != nil {
        fmt.Printf("Submit failed: %v\n", err)
    }
}
```

## 配置（Options 模式）

```go
// 使用默认配置
flow, _ := userflow.New(handler)

// 自定义配置
flow, _ := userflow.New(handler,
    userflow.WithQueueSize(20),              // 队列大小
    userflow.WithRateLimit(10),              // 每秒10个请求
    userflow.WithRateBurst(20),              // 突发20个
    userflow.WithShutdownTimeout(10*time.Second), // 关闭超时
    userflow.WithMetrics(),                  // 启用指标
)

// 禁用限流（适用于内部服务或测试环境）
flow, _ := userflow.New(handler,
    userflow.WithQueueSize(20),
    userflow.WithRateLimitEnabled(false),    // 禁用限流
)
```

### 配置选项

- **WithQueueSize(size int)**: 每个用户的请求队列大小，当队列满时新请求会被拒绝
- **WithRateLimit(limit float64)**: 基于 Token Bucket 算法的限流速率（每秒请求数）
- **WithRateBurst(burst int)**: 允许的突发请求数
- **WithRateLimitEnabled(enabled bool)**: 启用或禁用限流（默认启用），禁用后不进行限流检查
- **WithShutdownTimeout(timeout time.Duration)**: `Close` 等待 worker 退出的最大时间
- **WithMetrics()**: 启用性能指标收集

### 默认值

```go
queueSize:       10
rateLimit:       5 (每秒5个请求)
rateBurst:       10
shutdownTimeout: 5 * time.Second
enableMetrics:   false
enableRateLimit: true (默认启用限流)
```

### 配置建议

根据不同的业务场景，推荐的配置参数：

| 场景 | QueueSize | RateLimit | RateBurst | 说明 |
|------|-----------|-----------|-----------|------|
| 低负载游戏 | 10 | 5 | 10 | 休闲游戏、回合制游戏 |
| 中负载游戏 | 20 | 10 | 20 | 一般在线游戏 |
| 高负载游戏 | 50 | 20 | 50 | 大型多人游戏 |
| 实时对战 | 100 | 50 | 100 | 竞技类、FPS 游戏 |

## 错误处理

`Submit` 方法返回标准错误，使用 `errors.Is` 进行判断：

```go
err := flow.Submit(msg)
if err != nil {
    switch {
    case errors.Is(err, userflow.ErrRateLimited):
        log.Println("请求被限流")
        // 通知客户端降低请求频率
    case errors.Is(err, userflow.ErrQueueFull):
        log.Println("队列已满")
        // 通知客户端稍后重试
    case errors.Is(err, userflow.ErrClosed):
        log.Println("管理器已关闭")
    case errors.Is(err, userflow.ErrInvalidUser):
        log.Println("无效的用户ID")
    default:
        log.Printf("未知错误: %v", err)
    }
}
```

### 预定义错误

```go
var (
    ErrRateLimited = errors.New("rate limit exceeded")
    ErrQueueFull   = errors.New("queue full")
    ErrInvalidUser = errors.New("invalid user id")
    ErrClosed      = errors.New("flow is closed")
)
```

## 使用场景

### 1. WebSocket 消息处理

```go
// 消息处理器
handler := func(ctx context.Context, msg *envelope.InputMessage) {
    route := msg.GetRoute()
    switch route {
    case "game.login":
        handleLogin(ctx, msg)
    case "game.move":
        handleMove(ctx, msg)
    default:
        log.Printf("Unknown route: %s", route)
    }
}

// 创建流量管理器
flow, _ := userflow.New(handler,
    userflow.WithQueueSize(20),
    userflow.WithRateLimit(10),
)
defer flow.Close()

// WebSocket 消息回调
ws.HandleMessage(func(s *melody.Session, data []byte) {
    var msg envelope.InputMessage
    if err := proto.Unmarshal(data, &msg); err != nil {
        return
    }
    
    if err := flow.Submit(&msg); err != nil {
        switch {
        case errors.Is(err, userflow.ErrRateLimited):
            s.Write([]byte("请求过于频繁，请稍后再试"))
        case errors.Is(err, userflow.ErrQueueFull):
            s.Write([]byte("服务器繁忙，请稍后再试"))
        }
    }
})
```

### 2. 带指标监控

```go
flow, _ := userflow.New(handler, userflow.WithMetrics())

// 定期输出指标
go func() {
    ticker := time.NewTicker(1 * time.Minute)
    defer ticker.Stop()
    
    for range ticker.C {
        metrics := flow.GetMetrics()
        for userID, um := range metrics.GetAllUserMetrics() {
            snapshot := um.GetSnapshot()
            log.Printf("User %d: processed=%d, failed=%d, latency=%v",
                userID, snapshot.Processed, snapshot.Failed, snapshot.AverageLatency)
        }
    }
}()
```

### 3. 关闭控制

```go
// 捕获退出信号
sigChan := make(chan os.Signal, 1)
signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM)

go func() {
    <-sigChan
    log.Println("Shutting down...")
    
    // 停止接收新请求
    stopAcceptingRequests()
    
    // 关闭管理器：取消上下文并等待 worker 退出
    if err := flow.Close(); err != nil {
        log.Printf("Shutdown error: %v", err)
    } else {
        log.Println("Shutdown completed")
    }
    
    os.Exit(0)
}()
```

## 高级用法

### 指标收集

```go
flow, _ := userflow.New(handler, userflow.WithMetrics())

// 获取指标
metrics := flow.GetMetrics()
userMetrics := metrics.GetUserMetrics(userID)
snapshot := userMetrics.GetSnapshot()

fmt.Printf("Enqueued: %d\n", snapshot.Enqueued)
fmt.Printf("Processed: %d\n", snapshot.Processed)
fmt.Printf("Failed: %d\n", snapshot.Failed)
fmt.Printf("RateLimited: %d\n", snapshot.RateLimited)
fmt.Printf("QueueFull: %d\n", snapshot.QueueFull)
fmt.Printf("Avg Latency: %v\n", snapshot.AverageLatency)
```

### 用户管理

```go
// 踢出用户（释放资源）
flow.KickUser(userID)

// 获取活跃用户数
count := flow.ActiveUserCount()
```

### 上下文控制

```go
handler := func(ctx context.Context, msg *envelope.InputMessage) {
    // 检查上下文是否已取消
    select {
    case <-ctx.Done():
        return // 用户已被踢出或管理器已关闭
    default:
    }
    
    // 处理逻辑
    // ...
}
```

## 性能优化

### 1. 减少内存分配

```go
// 使用对象池复用消息对象
var msgPool = sync.Pool{
    New: func() interface{} {
        return &envelope.InputMessage{}
    },
}

func submitMessage(flow *userflow.Flow, userID int64, route string) {
    msg := msgPool.Get().(*envelope.InputMessage)
    msg.Header = &envelope.Header{Uid: userID}
    msg.Route = route
    
    flow.Submit(msg)
    
    // 处理完成后回收
    // (需要在 handler 中实现)
}
```

### 2. 批量处理

如果业务允许，可以在 handler 中积累一批消息再处理：

```go
handler := func(ctx context.Context, msg *envelope.InputMessage) {
    batch = append(batch, msg)
    
    if len(batch) >= 10 {
        processBatch(batch)
        batch = batch[:0]
    }
}
```

### 3. 异步处理

对于耗时操作，可以在 handler 中再次异步处理：

```go
handler := func(ctx context.Context, msg *envelope.InputMessage) {
    // 快速验证
    if !validate(msg) {
        return
    }
    
    // 耗时操作异步处理
    go func() {
        heavyProcess(msg)
    }()
}
```

### 4. 合理设置队列大小

- 队列太小：容易触发 QueueFull，丢失消息
- 队列太大：内存占用高，积压消息更多
- 建议：根据消息处理速度和突发量设置，一般 10-50 即可

## 调试技巧

### 1. 启用详细日志

```go
handler := func(ctx context.Context, msg *envelope.InputMessage) {
    start := time.Now()
    userID := msg.GetHeader().GetUid()
    route := msg.GetRoute()
    
    log.Printf("[Start] User %d, route %s", userID, route)
    
    process(msg)
    
    log.Printf("[Done] User %d, route %s, cost %v",
        userID, route, time.Since(start))
}
```

### 2. 监控活跃用户

```go
go func() {
    ticker := time.NewTicker(10 * time.Second)
    defer ticker.Stop()
    
    for range ticker.C {
        count := flow.ActiveUserCount()
        log.Printf("Active users: %d", count)
        
        if count > 1000 {
            log.Printf("WARNING: Too many active users!")
        }
    }
}()
```

### 3. 检测慢请求

```go
handler := func(ctx context.Context, msg *envelope.InputMessage) {
    start := time.Now()
    
    process(msg)
    
    if cost := time.Since(start); cost > 100*time.Millisecond {
        userID := msg.GetHeader().GetUid()
        route := msg.GetRoute()
        log.Printf("[SlowRequest] User %d, route %s, cost %v",
            userID, route, cost)
    }
}
```

### 4. 监控错误率

```go
var (
    submitTotal   atomic.Int64
    rateLimited   atomic.Int64
    queueFull     atomic.Int64
)

go func() {
    ticker := time.NewTicker(1 * time.Minute)
    defer ticker.Stop()
    
    for range ticker.C {
        total := submitTotal.Swap(0)
        limited := rateLimited.Swap(0)
        full := queueFull.Swap(0)
        
        if total > 0 {
            limitedRate := float64(limited) / float64(total) * 100
            fullRate := float64(full) / float64(total) * 100
            
            log.Printf("Submit stats: total=%d, rate_limited=%.2f%%, queue_full=%.2f%%",
                total, limitedRate, fullRate)
        }
    }
}()

// 提交时统计
err := flow.Submit(msg)
submitTotal.Add(1)
if errors.Is(err, userflow.ErrRateLimited) {
    rateLimited.Add(1)
} else if errors.Is(err, userflow.ErrQueueFull) {
    queueFull.Add(1)
}
```

## 性能考虑

1. **内存使用**：每个活跃用户占用一个 goroutine 和一个队列（约几 KB）
2. **并发性**：不同用户的请求可以并发处理，同一用户的请求顺序处理
3. **限流**：合理设置限流参数，防止单个用户占用过多资源
4. **队列大小**：队列大小应根据业务特点调整，避免过大导致内存浪费
5. **关闭超时**：设置合理的关闭超时，避免长时间等待

## 性能基准

```
BenchmarkSubmit-8                     ~107 ns/op    0-1 allocs/op
BenchmarkSubmitMultipleUsers-8        ~156 ns/op    1 allocs/op
BenchmarkSubmitWithMetrics-8          ~106 ns/op    0-1 allocs/op
BenchmarkSubmitParallel-8             ~58 ns/op     1 allocs/op
BenchmarkGetOrCreateWorker-8          ~50 ns/op     0-1 allocs/op
BenchmarkMetricsIncrement-8           ~20 ns/op     0 allocs/op
BenchmarkMetricsIncrementParallel-8   ~10 ns/op     0 allocs/op
```

## 测试

运行所有测试：

```bash
cd pkg/userflow
go test -v
```

查看测试覆盖率：

```bash
go test -cover
```

运行竞态检测：

```bash
go test -race
```

运行基准测试：

```bash
go test -bench=. -benchmem
```

压力测试：

```bash
go test -bench=BenchmarkSubmitParallel -benchtime=10s
```

## 测试覆盖

- **测试范围**: 覆盖创建、提交、限流、队列、关闭、指标等主要路径
- **测试覆盖率**: 94.7%
- **竞态检测**: 通过
- **基准测试**: 7 个

### 测试类别

1. 创建管理器（有效/无效配置、nil 处理器）
2. 提交消息（单用户、多用户、顺序保证、无效用户ID）
3. 限流测试
4. 队列满测试
5. 踢出用户测试
6. 关闭控制测试
7. Panic 恢复测试
8. 上下文取消测试
9. 并发提交测试
10. 指标收集测试

## 与原实现的对比

| 特性 | 原实现 | 新实现 |
|------|--------|--------|
| 类型安全 | ⚠️ 使用 `interface{}` | ✅ 直接使用 `*envelope.InputMessage` |
| 配置 | ❌ 硬编码 | ✅ Options 模式 |
| 错误处理 | ❌ 直接写 session | ✅ 标准错误返回 |
| 测试 | ❌ 难以测试 | ✅ 94.7% 覆盖率 |
| 指标 | ❌ 无 | ✅ 可选指标收集 |
| 文档 | ❌ 少 | ✅ 完整文档 |
| 并发提交 | ⚠️ 不受控 | ✅ 支持多协程提交 |
| 资源清理 | ⚠️ 分散 | ✅ 通过 `KickUser` / `Close` 管理 |

## 最佳实践

1. **合理设置限流**：根据服务器性能和业务特点设置限流参数
2. **启用指标**：在生产环境启用指标收集，便于监控和调优
3. **及时踢出用户**：用户离线后及时调用 `KickUser` 释放资源
4. **错误处理**：使用 `errors.Is` 判断错误类型并进行相应处理
5. **关闭管理器**：应用退出时先停止接收新请求，再调用 `Close()` 等待 worker 退出
6. **监控慢请求**：记录并优化耗时较长的请求
7. **控制活跃用户数**：监控活跃用户数，避免资源耗尽

## 文件结构

```
pkg/userflow/
├── flow.go            # Flow 结构体 + 错误定义
├── worker.go          # Worker 工作者
├── options.go         # Options 模式配置
├── metrics.go         # 指标收集
├── doc.go             # 包文档
├── flow_test.go       # Flow 功能测试
├── options_test.go    # Options 测试
├── metrics_test.go    # 指标测试
├── benchmark_test.go  # 基准测试
└── README.md                  # 本文档
```

## 设计原则

1. **单一职责**：每个文件关注一个功能点
2. **Options 模式**：灵活的配置方式
3. **标准错误**：使用 Go 标准错误处理
4. **并发模型**：不同用户并行处理，同一用户顺序处理
5. **测试驱动**：高测试覆盖率
6. **性能优先**：优化热路径，减少分配

## License

MIT
