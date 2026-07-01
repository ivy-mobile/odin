package userflow

import (
	"testing"
	"time"
)

// TestDefaultOptions 测试默认选项
func TestDefaultOptions(t *testing.T) {
	opts := defaultOptions

	if opts.queueSize != 10 {
		t.Errorf("expected queueSize 10, got %d", opts.queueSize)
	}
	if opts.rateLimit != 5 {
		t.Errorf("expected rateLimit 5, got %f", opts.rateLimit)
	}
	if opts.rateBurst != 10 {
		t.Errorf("expected rateBurst 10, got %d", opts.rateBurst)
	}
	if opts.shutdownTimeout != 5*time.Second {
		t.Errorf("expected shutdownTimeout 5s, got %v", opts.shutdownTimeout)
	}
	if opts.enableMetrics {
		t.Error("expected enableMetrics false by default")
	}
}

// TestOptionsValidate 测试选项验证
func TestOptionsValidate(t *testing.T) {
	tests := []struct {
		name      string
		opts      options
		expectErr bool
	}{
		{
			name:      "valid options",
			opts:      defaultOptions,
			expectErr: false,
		},
		{
			name: "invalid queueSize",
			opts: options{
				queueSize:       0,
				rateLimit:       5,
				rateBurst:       10,
				shutdownTimeout: 5 * time.Second,
			},
			expectErr: true,
		},
		{
			name: "negative queueSize",
			opts: options{
				queueSize:       -1,
				rateLimit:       5,
				rateBurst:       10,
				shutdownTimeout: 5 * time.Second,
			},
			expectErr: true,
		},
		{
			name: "invalid rateLimit",
			opts: options{
				queueSize:       10,
				rateLimit:       0,
				rateBurst:       10,
				shutdownTimeout: 5 * time.Second,
				enableRateLimit: true, // 启用限流才验证
			},
			expectErr: true,
		},
		{
			name: "negative rateLimit",
			opts: options{
				queueSize:       10,
				rateLimit:       -1,
				rateBurst:       10,
				shutdownTimeout: 5 * time.Second,
				enableRateLimit: true, // 启用限流才验证
			},
			expectErr: true,
		},
		{
			name: "invalid rateBurst",
			opts: options{
				queueSize:       10,
				rateLimit:       5,
				rateBurst:       0,
				shutdownTimeout: 5 * time.Second,
				enableRateLimit: true, // 启用限流才验证
			},
			expectErr: true,
		},
		{
			name: "invalid rateLimit but disabled",
			opts: options{
				queueSize:       10,
				rateLimit:       0, // 无效值
				rateBurst:       0, // 无效值
				shutdownTimeout: 5 * time.Second,
				enableRateLimit: false, // 禁用限流，不验证
			},
			expectErr: false, // 不应该报错
		},
		{
			name: "invalid shutdownTimeout",
			opts: options{
				queueSize:       10,
				rateLimit:       5,
				rateBurst:       10,
				shutdownTimeout: 0,
			},
			expectErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.opts.validate()
			if tt.expectErr && err == nil {
				t.Error("expected error, got nil")
			}
			if !tt.expectErr && err != nil {
				t.Errorf("expected no error, got %v", err)
			}
		})
	}
}

// TestWithOptions 测试选项函数
func TestWithOptions(t *testing.T) {
	opts := defaultOptions

	WithQueueSize(20)(&opts)
	if opts.queueSize != 20 {
		t.Errorf("expected queueSize 20, got %d", opts.queueSize)
	}

	WithRateLimit(10.0)(&opts)
	if opts.rateLimit != 10.0 {
		t.Errorf("expected rateLimit 10.0, got %f", opts.rateLimit)
	}

	WithRateBurst(20)(&opts)
	if opts.rateBurst != 20 {
		t.Errorf("expected rateBurst 20, got %d", opts.rateBurst)
	}

	WithShutdownTimeout(10 * time.Second)(&opts)
	if opts.shutdownTimeout != 10*time.Second {
		t.Errorf("expected shutdownTimeout 10s, got %v", opts.shutdownTimeout)
	}

	WithMetrics()(&opts)
	if !opts.enableMetrics {
		t.Error("expected enableMetrics true")
	}
}
