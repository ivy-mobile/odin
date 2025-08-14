package rmq

import (
	"context"
	"fmt"

	"github.com/apache/rocketmq-clients/golang/v5"
	"github.com/apache/rocketmq-clients/golang/v5/credentials"
)

type Producer struct {
	ctx context.Context
	p   golang.Producer
}

// NewProducer 创建生产者
func NewProducer(endpoint, namespace, group string, credentials *credentials.SessionCredentials, opts ...golang.ProducerOption) (*Producer, error) {
	p, err := golang.NewProducer(&golang.Config{
		Endpoint:      endpoint,
		NameSpace:     namespace,
		ConsumerGroup: group,
		Credentials:   credentials,
	}, opts...)
	if err != nil {
		return nil, fmt.Errorf("new producer err: %v", err)
	}
	if err = p.Start(); err != nil {
		_ = p.GracefulStop()
		return nil, fmt.Errorf("start producer err: %v", err)
	}
	return &Producer{
		ctx: context.Background(),
		p:   p,
	}, nil
}

// Send 发送消息
func (pc *Producer) Send(msg *golang.Message) ([]*golang.SendReceipt, error) {
	return pc.p.Send(pc.ctx, msg)
}

// SendWithContext 发送消息-自定义context
func (pc *Producer) SendWithContext(ctx context.Context, msg *golang.Message) ([]*golang.SendReceipt, error) {
	return pc.p.Send(ctx, msg)
}

// SendAsync 异步发送消息
func (pc *Producer) SendAsync(msg *golang.Message, callback func(context.Context, []*golang.SendReceipt, error)) {
	pc.p.SendAsync(pc.ctx, msg, callback)
}

// SendAsyncWithContext 异步发送消息-自定义context
func (pc *Producer) SendAsyncWithContext(ctx context.Context, msg *golang.Message, callback func(context.Context, []*golang.SendReceipt, error)) {
	pc.p.SendAsync(ctx, msg, callback)
}

// SendWithTransaction 发送消息-事务
func (pc *Producer) SendWithTransaction(msg *golang.Message, t golang.Transaction) ([]*golang.SendReceipt, error) {
	return pc.p.SendWithTransaction(pc.ctx, msg, t)
}

// SendWithTransactionContext 发送消息-事务-自定义context
func (pc *Producer) SendWithTransactionContext(ctx context.Context, msg *golang.Message, t golang.Transaction) ([]*golang.SendReceipt, error) {
	return pc.p.SendWithTransaction(ctx, msg, t)
}

// Close 关闭
func (pc *Producer) Close() error {
	return pc.p.GracefulStop()
}
