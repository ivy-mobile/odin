package logging

import (
	"context"
	"errors"
	"testing"

	"dubbo.apache.org/dubbo-go/v3/common"
	"dubbo.apache.org/dubbo-go/v3/protocol"
	"dubbo.apache.org/dubbo-go/v3/protocol/invocation"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/ivy-mobile/odin/dbo/header"
	xlogv2 "github.com/ivy-mobile/odin/xutil/xlog/v2"
)

type captureInvoker struct {
	ctx    context.Context
	result protocol.Result
}

func (i *captureInvoker) GetURL() *common.URL {
	return &common.URL{Path: "test.Service"}
}

func (i *captureInvoker) IsAvailable() bool {
	return true
}

func (i *captureInvoker) Destroy() {}

func (i *captureInvoker) Invoke(ctx context.Context, _ protocol.Invocation) protocol.Result {
	i.ctx = ctx
	if i.result == nil {
		return &protocol.RPCResult{}
	}
	return i.result
}

func TestLogFilterAddsMsgIDWhenMissing(t *testing.T) {
	invoker := &captureInvoker{}
	filter := NewLogFilter(xlogv2.New())()
	ctx := header.With(context.Background(), header.Header{header.NodeID: "node-1"})
	invocation := invocation.NewRPCInvocation("Hello", nil, nil)

	result := filter.Invoke(ctx, invoker, invocation)

	require.NotNil(t, result)
	h := header.From(invoker.ctx)
	assert.Equal(t, "node-1", h.NodeID())
	assert.NotEmpty(t, h.MsgID())
}

func TestLogFilterKeepsExistingMsgID(t *testing.T) {
	invoker := &captureInvoker{}
	filter := NewLogFilter(xlogv2.New())()
	ctx := header.With(context.Background(), header.Header{
		header.NodeID: "node-1",
		header.MsgID:  "custom-msg",
	})
	invocation := invocation.NewRPCInvocation("Hello", nil, nil)

	result := filter.Invoke(ctx, invoker, invocation)

	require.NotNil(t, result)
	h := header.From(invoker.ctx)
	assert.Equal(t, "node-1", h.NodeID())
	assert.Equal(t, "custom-msg", h.MsgID())
}

func TestLogFilterLogsSuccessResponse(t *testing.T) {
	type TestResponse struct {
		Code int
		Data string
	}

	invoker := &captureInvoker{
		result: &protocol.RPCResult{
			Rest: &TestResponse{Code: 200, Data: "success"},
		},
	}
	filter := NewLogFilter(xlogv2.New())()
	ctx := header.With(context.Background(), header.Header{header.NodeID: "node-1"})
	invocation := invocation.NewRPCInvocation("TestMethod", nil, nil)

	result := filter.Invoke(ctx, invoker, invocation)

	require.NotNil(t, result)
	assert.NoError(t, result.Error())
	assert.NotNil(t, result.Result())
}

func TestLogFilterLogsErrorResponse(t *testing.T) {
	testErr := errors.New("RPC call failed")
	invoker := &captureInvoker{
		result: &protocol.RPCResult{
			Err: testErr,
		},
	}
	filter := NewLogFilter(xlogv2.New())()
	ctx := header.With(context.Background(), header.Header{header.NodeID: "node-1"})
	invocation := invocation.NewRPCInvocation("TestMethod", nil, nil)

	result := filter.Invoke(ctx, invoker, invocation)

	require.NotNil(t, result)
	assert.Equal(t, testErr, result.Error())
}

func TestGetTypeName(t *testing.T) {
	tests := []struct {
		name     string
		input    any
		expected string
	}{
		{
			name:     "nil value",
			input:    nil,
			expected: "nil",
		},
		{
			name:     "simple type",
			input:    "string",
			expected: "string",
		},
		{
			name:     "pointer type",
			input:    new(int),
			expected: "*int",
		},
		{
			name: "struct type",
			input: struct {
				Name string
			}{Name: "test"},
			expected: "struct { Name string }",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := getTypeName(tt.input)
			assert.Equal(t, tt.expected, result)
		})
	}
}
