// Package userflow 提供了一个高性能的用户流量管理器。
//
// 该包为每个用户维护独立的请求队列和处理协程，保证用户维度请求处理的顺序性，
// 同时支持限流、指标收集、关闭控制等功能。
//
// # 核心特性
//
//   - 顺序保证：每个用户的请求在独立协程中顺序处理
//   - 限流控制：基于 Token Bucket 算法的用户级限流
//   - 队列管理：可配置的请求队列大小
//   - 关闭控制：Close 会取消上下文并等待 worker 退出，超时返回错误
//   - 错误返回：Submit 返回错误，便于上层处理
//   - 指标收集：可选的性能指标收集
//   - 并发提交：支持多协程提交消息
//
// # 基本使用
//
// 创建流量管理器并提交消息：
//
//	handler := func(ctx context.Context, msg *envelope.InputMessage) {
//	    userID := msg.GetHeader().GetUid()
//	    route := msg.GetRoute()
//	    // 处理消息逻辑
//	}
//
//	flow, err := userflow.New(handler)
//	if err != nil {
//	    panic(err)
//	}
//	defer flow.Close()
//
//	msg := &envelope.InputMessage{
//	    Header: &envelope.Header{Uid: 1001},
//	    Route:  "game.login",
//	}
//
//	if err := flow.Submit(msg); err != nil {
//	    log.Printf("Submit failed: %v", err)
//	}
//
// # 自定义配置
//
// 使用 Options 模式自定义配置：
//
//	flow, _ := userflow.New(handler,
//	    userflow.WithQueueSize(20),              // 队列大小
//	    userflow.WithRateLimit(10),              // 每秒10个请求
//	    userflow.WithRateBurst(20),              // 突发20个
//	    userflow.WithShutdownTimeout(10*time.Second), // 关闭超时
//	    userflow.WithMetrics(),                  // 启用指标
//	    userflow.WithRateLimitEnabled(false),    // 禁用限流（可选）
//	)
//
// # 错误处理
//
// Submit 返回错误，可以使用 errors.Is 判断预定义错误：
//
//	err := flow.Submit(msg)
//	if err != nil {
//	    switch {
//	    case errors.Is(err, userflow.ErrRateLimited):
//	        log.Println("request rate limited")
//	    case errors.Is(err, userflow.ErrQueueFull):
//	        log.Println("user queue full")
//	    case errors.Is(err, userflow.ErrClosed):
//	        log.Println("flow closed")
//	    case errors.Is(err, userflow.ErrInvalidUser):
//	        log.Println("invalid user id")
//	    default:
//	        log.Printf("Submit error: %v", err)
//	    }
//	}
//
// # 指标收集
//
// 启用指标后可以获取详细的性能数据：
//
//	flow, _ := userflow.New(handler, userflow.WithMetrics())
//
//	metrics := flow.GetMetrics()
//	userMetrics := metrics.GetUserMetrics(userID)
//	snapshot := userMetrics.GetSnapshot()
//	// 访问 snapshot.Enqueued, snapshot.Processed 等
//
// # 用户管理
//
// 踢出用户释放资源：
//
//	flow.KickUser(userID)
//
// 获取活跃用户数：
//
//	count := flow.ActiveUserCount()
//
// # 性能考虑
//
//   - 每个活跃用户占用一个 goroutine 和一个队列（约几 KB）
//   - 不同用户的请求可以并发处理，同一用户的请求顺序处理
//   - 合理设置限流参数，防止单个用户占用过多资源
//   - 队列大小应根据业务特点调整
//
// 更多信息请参阅 README.md
package userflow
