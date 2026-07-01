package userflow

import (
	"testing"
	"time"
)

// TestMetricsBasic 测试基本指标功能
func TestMetricsBasic(t *testing.T) {
	m := NewMetrics()

	// 测试活跃用户数
	m.SetActiveUsers(5)
	if m.GetActiveUsers() != 5 {
		t.Errorf("expected 5 active users, got %d", m.GetActiveUsers())
	}

	// 测试用户指标计数
	m.IncEnqueued(1)
	m.IncEnqueued(1)
	m.IncProcessed(1)
	m.IncFailed(1)
	m.IncRateLimited(1)
	m.IncQueueFull(1)

	um := m.GetUserMetrics(1)
	if um == nil {
		t.Fatal("expected user metrics")
	}

	if um.Enqueued.Load() != 2 {
		t.Errorf("expected 2 enqueued, got %d", um.Enqueued.Load())
	}
	if um.Processed.Load() != 1 {
		t.Errorf("expected 1 processed, got %d", um.Processed.Load())
	}
	if um.Failed.Load() != 1 {
		t.Errorf("expected 1 failed, got %d", um.Failed.Load())
	}
	if um.RateLimited.Load() != 1 {
		t.Errorf("expected 1 rate limited, got %d", um.RateLimited.Load())
	}
	if um.QueueFull.Load() != 1 {
		t.Errorf("expected 1 queue full, got %d", um.QueueFull.Load())
	}
}

// TestMetricsLatency 测试延迟统计
func TestMetricsLatency(t *testing.T) {
	m := NewMetrics()

	// 记录延迟
	m.ObserveLatency(1, 100*time.Millisecond)
	m.ObserveLatency(1, 200*time.Millisecond)
	m.ObserveLatency(1, 300*time.Millisecond)

	um := m.GetUserMetrics(1)
	if um == nil {
		t.Fatal("expected user metrics")
	}

	avgLatency := um.AverageLatency()
	expected := 200 * time.Millisecond
	// 允许小的误差
	diff := avgLatency - expected
	if diff < 0 {
		diff = -diff
	}
	if diff > time.Millisecond {
		t.Errorf("expected average latency ~%v, got %v", expected, avgLatency)
	}
}

// TestMetricsSnapshot 测试快照
func TestMetricsSnapshot(t *testing.T) {
	m := NewMetrics()

	m.IncEnqueued(1)
	m.IncProcessed(1)
	m.IncFailed(1)
	m.ObserveLatency(1, 100*time.Millisecond)

	um := m.GetUserMetrics(1)
	snapshot := um.GetSnapshot()

	if snapshot.Enqueued != 1 {
		t.Errorf("expected 1 enqueued, got %d", snapshot.Enqueued)
	}
	if snapshot.Processed != 1 {
		t.Errorf("expected 1 processed, got %d", snapshot.Processed)
	}
	if snapshot.Failed != 1 {
		t.Errorf("expected 1 failed, got %d", snapshot.Failed)
	}
	if snapshot.AverageLatency != 100*time.Millisecond {
		t.Errorf("expected 100ms latency, got %v", snapshot.AverageLatency)
	}
}

// TestMetricsGetAllUserMetrics 测试获取所有用户指标
func TestMetricsGetAllUserMetrics(t *testing.T) {
	m := NewMetrics()

	// 为3个用户记录指标
	for i := int64(1); i <= 3; i++ {
		m.IncEnqueued(i)
		m.IncProcessed(i)
	}

	allMetrics := m.GetAllUserMetrics()
	if len(allMetrics) != 3 {
		t.Errorf("expected 3 users, got %d", len(allMetrics))
	}

	for i := int64(1); i <= 3; i++ {
		um, ok := allMetrics[i]
		if !ok {
			t.Errorf("expected metrics for user %d", i)
			continue
		}
		if um.Enqueued.Load() != 1 {
			t.Errorf("expected 1 enqueued for user %d, got %d", i, um.Enqueued.Load())
		}
	}
}

// TestMetricsNonExistentUser 测试不存在的用户
func TestMetricsNonExistentUser(t *testing.T) {
	m := NewMetrics()

	um := m.GetUserMetrics(999)
	if um != nil {
		t.Error("expected nil for non-existent user")
	}
}

// TestMetricsZeroLatency 测试零延迟
func TestMetricsZeroLatency(t *testing.T) {
	um := &UserMetrics{}

	avgLatency := um.AverageLatency()
	if avgLatency != 0 {
		t.Errorf("expected 0 latency for no samples, got %v", avgLatency)
	}
}

// TestMetricsConcurrent 测试并发安全
func TestMetricsConcurrent(t *testing.T) {
	m := NewMetrics()

	const goroutines = 100
	const iterations = 100

	done := make(chan bool)

	for i := 0; i < goroutines; i++ {
		go func() {
			for j := 0; j < iterations; j++ {
				m.IncEnqueued(1)
				m.IncProcessed(1)
				m.ObserveLatency(1, time.Millisecond)
			}
			done <- true
		}()
	}

	for i := 0; i < goroutines; i++ {
		<-done
	}

	um := m.GetUserMetrics(1)
	expectedCount := int64(goroutines * iterations)

	if um.Enqueued.Load() != expectedCount {
		t.Errorf("expected %d enqueued, got %d", expectedCount, um.Enqueued.Load())
	}
	if um.Processed.Load() != expectedCount {
		t.Errorf("expected %d processed, got %d", expectedCount, um.Processed.Load())
	}
	if um.latencyCount.Load() != expectedCount {
		t.Errorf("expected %d latency samples, got %d", expectedCount, um.latencyCount.Load())
	}
}
