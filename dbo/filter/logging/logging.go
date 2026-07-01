package logging

import (
	"context"
	"fmt"
	"time"

	"dubbo.apache.org/dubbo-go/v3/filter"
	"dubbo.apache.org/dubbo-go/v3/protocol"

	"github.com/ivy-mobile/odin/dbo/header"
	"github.com/ivy-mobile/odin/xutil/xid"
	xlogv2 "github.com/ivy-mobile/odin/xutil/xlog/v2"
)

// logFilter 统一dubbo日志过滤器
type logFilter struct {
	logger xlogv2.Logger
}

func NewLogFilter(logger xlogv2.Logger) func() filter.Filter {
	return func() filter.Filter {
		return &logFilter{
			logger: logger.With("module", "filter/logging"),
		}
	}
}

func (l *logFilter) Invoke(ctx context.Context, invoker protocol.Invoker, invocation protocol.Invocation) protocol.Result {
	start := time.Now()
	// 调用方已显式传入 msg-id 时不覆盖，便于外部链路追踪
	cp := header.AddIfAbsent(ctx, header.MsgID, xid.Snowflake())

	// 执行RPC调用
	result := invoker.Invoke(cp, invocation)

	// 记录请求和响应信息
	h := header.From(cp)
	logEvent := l.logger.Info().
		Str("DubboService", invoker.GetURL().Service()).
		Str("Method", invocation.MethodName()).
		Str("msg-id", h.MsgID()).
		Str("node-id", h.NodeID())

	// 检查响应状态
	if err := result.Error(); err != nil {
		// 失败时记录详细错误信息
		logEvent.Err(err).Msg("[DubboRequest] failed")
	} else {
		// 成功时记录基本信息和响应元数据
		if respData := result.Result(); respData != nil {
			// 记录响应类型（避免序列化整个响应体）
			logEvent.Str("resp-type", getTypeName(respData))
		}
		// 记录响应的attachments数量（可选）
		if attachments := result.Attachments(); len(attachments) > 0 {
			logEvent.Int("resp-attachments", len(attachments))
		}
		logEvent.Msgf("[DubboRequest] success, cost: %v", time.Since(start))
	}

	return result
}

// getTypeName 获取类型名称，避免完整的类型路径
func getTypeName(v any) string {
	if v == nil {
		return "nil"
	}
	typeName := fmt.Sprintf("%T", v)
	// 简化类型名称，只保留最后一部分
	if idx := lastIndexAny(typeName, "./"); idx >= 0 {
		return typeName[idx+1:]
	}
	return typeName
}

// lastIndexAny 查找任意字符的最后出现位置
func lastIndexAny(s string, chars string) int {
	for i := len(s) - 1; i >= 0; i-- {
		for j := 0; j < len(chars); j++ {
			if s[i] == chars[j] {
				return i
			}
		}
	}
	return -1
}

func (l *logFilter) OnResponse(_ context.Context, result protocol.Result, _ protocol.Invoker, _ protocol.Invocation) protocol.Result {
	return result
}
