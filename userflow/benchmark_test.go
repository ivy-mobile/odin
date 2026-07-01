package userflow

import (
	"context"
	"testing"

	"github.com/ivy-mobile/odin/envelope"
)

// BenchmarkSubmit 基准测试：提交事件
func BenchmarkSubmit(b *testing.B) {
	handler := func(_ context.Context, _ *envelope.InputMessage) {
		// 空处理器
	}

	flow, _ := New(handler,
		WithRateLimit(10000), // 高限流，避免影响基准测试
		WithRateBurst(10000),
	)
	defer flow.Close()

	msg := mockMessage(1)

	b.ResetTimer()
	for range b.N {
		_ = flow.Submit(msg)
	}
}

// BenchmarkSubmitMultipleUsers 基准测试：多用户提交
func BenchmarkSubmitMultipleUsers(b *testing.B) {
	handler := func(_ context.Context, _ *envelope.InputMessage) {
		// 空处理器
	}

	flow, _ := New(handler,
		WithRateLimit(10000),
		WithRateBurst(10000),
	)
	defer flow.Close()

	b.ResetTimer()
	for i := range b.N {
		userID := int64(i % 100) // 100个用户
		msg := mockMessage(userID)
		_ = flow.Submit(msg)
	}
}

// BenchmarkSubmitWithMetrics 基准测试：启用指标的提交
func BenchmarkSubmitWithMetrics(b *testing.B) {
	handler := func(_ context.Context, _ *envelope.InputMessage) {
		// 空处理器
	}

	flow, _ := New(handler,
		WithRateLimit(10000),
		WithRateBurst(10000),
		WithMetrics(),
	)
	defer flow.Close()

	msg := mockMessage(1)

	b.ResetTimer()
	for range b.N {
		_ = flow.Submit(msg)
	}
}

// BenchmarkSubmitParallel 基准测试：并行提交
func BenchmarkSubmitParallel(b *testing.B) {
	handler := func(_ context.Context, _ *envelope.InputMessage) {
		// 空处理器
	}

	flow, _ := New(handler,
		WithRateLimit(10000),
		WithRateBurst(10000),
	)
	defer flow.Close()

	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		userID := int64(0)
		for pb.Next() {
			userID++
			msg := mockMessage(userID % 100)
			_ = flow.Submit(msg)
		}
	})
}

// BenchmarkGetOrCreateWorker 基准测试：获取或创建工作者
func BenchmarkGetOrCreateWorker(b *testing.B) {
	handler := func(_ context.Context, _ *envelope.InputMessage) {}

	flow, _ := New(handler)
	defer flow.Close()

	b.ResetTimer()
	for range b.N {
		flow.getOrCreateWorker(1)
	}
}

// BenchmarkMetricsIncrement 基准测试：指标递增
func BenchmarkMetricsIncrement(b *testing.B) {
	m := NewMetrics()

	b.ResetTimer()
	for range b.N {
		m.IncEnqueued(1)
	}
}

// BenchmarkMetricsIncrementParallel 基准测试：并行指标递增
func BenchmarkMetricsIncrementParallel(b *testing.B) {
	m := NewMetrics()

	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			m.IncEnqueued(1)
		}
	})
}
