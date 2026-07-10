# Timer 计时器组件

基于 Redis ZSet 实现的分布式计时器组件。

## 功能特性

- ✅ 基于 Redis ZSet 实现，支持分布式部署
- ✅ 自动扫描到期的计时器
- ✅ 支持启动、停止计时器
- ✅ 可配置的扫描间隔和批处理大小
- ✅ 完整的单元测试覆盖
- ✅ 接口化设计，易于扩展和测试

## 使用方法

### 1. 创建计时器实例

```go
import (
    "toka/pkg/timer"
    "github.com/redis/go-redis/v9"
)

// 配置计时器
config := timer.Config{
    RedisClient:  redisClient,      // Redis 客户端
    KeyFormat:    "game:%d:timers", // Redis key 格式
    GameId:       1,                // 游戏ID
    ScanInterval: 1 * time.Second,  // 扫描间隔
    BatchSize:    100,              // 每次处理的最大数量
}

// 创建计时器实例
rt, err := timer.NewRedisTimer(config)
if err != nil {
    // 处理错误
}
```

### 2. 启动计时器

```go
// 启动一个 5 秒后到期的计时器
err := rt.Start(100, "timer1", 5*time.Second)
if err != nil {
    // 处理错误
}
```

### 3. 停止计时器

```go
// 停止指定的计时器
err := rt.Stop(100, "timer1")
if err != nil {
    // 处理错误
}
```

### 4. 监听到期的计时器

```go
ctx := context.Background()

// 定义回调函数
callback := func(tableId int, timerId string, startTime time.Time) {
    // 处理计时器到期逻辑
    fmt.Printf("Timer expired: tableId=%d, timerId=%s, startTime=%v \n", tableId, timerId, startTime)
}

// 开始监听（阻塞调用）
err := rt.Listen(ctx, callback)
if err != nil {
    // 处理错误
}
```

### 5. 在后台监听

```go
go func() {
    ctx, cancel := context.WithCancel(context.Background())
    defer cancel()
    
    callback := func(tableId int, timerId string) {
        // 处理逻辑
    }
    
    _ = rt.Listen(ctx, callback)
}()
```

## 接口说明

### Timer 接口

```go
type Timer interface {
    // Start 启动计时器
    Start(tableId int, timerId string, duration time.Duration) error
    
    // Stop 停止指定的计时器
    Stop(tableId int, timerId string) error
    
    // Listen 开始监听到期的计时器
    Listen(ctx context.Context, callback TimerCallback) error
}
```


## 实现原理

1. **存储方式**: 使用 Redis ZSet 存储计时器
   - Score: 到期时间戳（毫秒）
   - Member: JSON 格式的 TimerMessage

2. **扫描机制**: 定时扫描（默认每秒一次）
   - 使用 `ZRANGEBYSCORE` 获取到期的计时器
   - 批量处理，避免一次处理过多

3. **并发安全**: 
   - 使用 Redis 原子操作
   - 支持分布式部署

## 单元测试

运行测试：

```bash
# 需要 Redis 服务运行在 localhost:6379
go test ./pkg/timer -v

# 如果 Redis 不可用，测试会自动跳过
```

测试覆盖：
- ✅ 创建计时器实例
- ✅ 启动计时器
- ✅ 停止计时器
- ✅ 监听到期计时器
- ✅ 集成测试

## 在 svc 层使用

在 `internal/svc` 中已经集成了该组件，使用方式：

```go
// 1. 初始化（在 service.Init 中）
if err = InitRedisTimer(); err != nil {
    return fmt.Errorf("InitRedisTimer err: %v", err)
}

// 2. 启动监听
if err = ListenAllTableStateTimerByRedis(); err != nil {
    return fmt.Errorf("ListenAllTableStateTimerByRedis err: %v", err)
}

// 3. 启动计时器
StartTableStateTimerByRedis(tableId, timerId, duration)

// 4. 停止计时器
StopTableStateTimerByRedis(tableId, timerId)
```

## 注意事项

1. 确保 Redis 服务正常运行
2. 建议在生产环境中调整 `ScanInterval` 和 `BatchSize` 参数
3. 计时器的精度取决于 `ScanInterval` 设置
4. 如果同一个 `tableId` 和 `timerId` 启动新计时器，会自动替换旧的
