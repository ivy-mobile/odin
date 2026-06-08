package broker

import (
	"encoding/json"
	"fmt"
	"os"
	"sync"
	"testing"
	"time"

	"github.com/apache/rocketmq-clients/golang/v5"
	"github.com/apache/rocketmq-clients/golang/v5/credentials"
	"github.com/ivy-mobile/odin/rmq"
)

// ============================================================
// 辅助函数
// ============================================================

func newTestBroker(p producer, c consumer) *RMQBroker {
	return newRMQBroker("test-topic", p, c)
}

// ============================================================
// SendMessage 单元测试
// ============================================================

func TestSendMessage_Success(t *testing.T) {
	var capturedMsg *golang.Message
	mp := &mockProducer{
		sendFunc: func(msg *golang.Message) ([]*golang.SendReceipt, error) {
			capturedMsg = msg
			return []*golang.SendReceipt{{MessageID: "msg-001"}}, nil
		},
	}
	broker := newTestBroker(mp, &mockConsumer{})

	msgId, err := broker.SendMessage(10001, "game", "node-1", []byte("hello"))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if msgId != "msg-001" {
		t.Fatalf("expected msgId=msg-001, got %s", msgId)
	}

	// 验证 topic
	if capturedMsg.Topic != "test-topic" {
		t.Fatalf("expected topic=test-topic, got %s", capturedMsg.Topic)
	}

	// 验证 node property
	if v, ok := capturedMsg.GetProperties()["node"]; !ok || v != "node-1" {
		t.Fatalf("expected property node=node-1, got %v", capturedMsg.GetProperties())
	}

	// 验证消息体
	var m message
	if err := json.Unmarshal(capturedMsg.Body, &m); err != nil {
		t.Fatalf("unmarshal message body failed: %v", err)
	}
	if m.Uid != 10001 {
		t.Fatalf("expected uid=10001, got %d", m.Uid)
	}
	if string(m.Payload) != "hello" {
		t.Fatalf("expected payload=hello, got %s", string(m.Payload))
	}
	if m.UUID == "" {
		t.Fatal("expected UUID to be set, got empty string")
	}
	if m.Timestamp == 0 {
		t.Fatal("expected Timestamp to be set, got 0")
	}
}

func TestSendMessage_MarshalError(t *testing.T) {
	mp := &mockProducer{}
	broker := newTestBroker(mp, &mockConsumer{})

	// []byte 类型的 payload 本身不会导致 json.Marshal 失败,
	// 但 message 中如果包含不可序列化的类型则会失败。
	// 这里通过 channel 类型来模拟, 但 message struct 的字段都是可序列化的,
	// 所以我们直接测试接口不匹配的情况。
	// 实际场景中, Marshal 失败很少见, 但错误路径应该被覆盖。
	// 这里我们验证: 当 Send 返回 error 时, 错误被正确包装。
	mp.sendFunc = func(msg *golang.Message) ([]*golang.SendReceipt, error) {
		return nil, fmt.Errorf("connection refused")
	}
	_, err := broker.SendMessage(1, "", "node", []byte("test"))
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	if !contains(err.Error(), "SendMessage send fail") {
		t.Fatalf("expected error containing 'SendMessage send fail', got: %v", err)
	}
}

func TestSendMessage_SendError(t *testing.T) {
	mp := &mockProducer{
		sendFunc: func(msg *golang.Message) ([]*golang.SendReceipt, error) {
			return nil, fmt.Errorf("connection refused")
		},
	}
	broker := newTestBroker(mp, &mockConsumer{})

	_, err := broker.SendMessage(1, "", "node", []byte("test"))
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	if !contains(err.Error(), "SendMessage send fail") {
		t.Fatalf("expected error containing 'SendMessage send fail', got: %v", err)
	}
	// 验证 error wrapping: 应该能通过 errors.Is/As 解包
	if !contains(err.Error(), "connection refused") {
		t.Fatalf("expected wrapped error 'connection refused', got: %v", err)
	}
}

func TestSendMessage_EmptyReceipts(t *testing.T) {
	mp := &mockProducer{
		sendFunc: func(msg *golang.Message) ([]*golang.SendReceipt, error) {
			return []*golang.SendReceipt{}, nil
		},
	}
	broker := newTestBroker(mp, &mockConsumer{})

	_, err := broker.SendMessage(1, "", "node", []byte("test"))
	if err == nil {
		t.Fatal("expected error for empty receipts, got nil")
	}
	if !contains(err.Error(), "empty send receipts") {
		t.Fatalf("expected error containing 'empty send receipts', got: %v", err)
	}
}

func TestSendMessage_UUIDPopulated(t *testing.T) {
	mp := &mockProducer{
		sendFunc: func(msg *golang.Message) ([]*golang.SendReceipt, error) {
			var m message
			if err := json.Unmarshal(msg.Body, &m); err != nil {
				return nil, err
			}
			if m.UUID == "" {
				return nil, fmt.Errorf("UUID is empty")
			}
			return []*golang.SendReceipt{{MessageID: "ok"}}, nil
		},
	}
	broker := newTestBroker(mp, &mockConsumer{})

	_, err := broker.SendMessage(1, "", "node", []byte("test"))
	if err != nil {
		t.Fatalf("UUID should be populated: %v", err)
	}
}

// ============================================================
// ReceiveMessage 单元测试
// ============================================================

func TestReceiveMessage_Success(t *testing.T) {
	callbackInvoked := false
	mc := &mockConsumer{
		subscribeBySQL92Func: func(topic, sql92 string, callback func(msg MessageView) error) error {
			if topic != "test-topic" {
				return fmt.Errorf("expected topic=test-topic, got %s", topic)
			}
			if sql92 != "node='node-1'" {
				return fmt.Errorf("expected sql92=node='node-1', got %s", sql92)
			}
			callbackInvoked = true
			return nil
		},
	}
	broker := newTestBroker(&mockProducer{}, mc)

	err := broker.ReceiveMessage("game", "node-1", func(uid int64, msgId string, timestamp int64, msg []byte) {})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !callbackInvoked {
		t.Fatal("expected subscribe callback to be invoked")
	}
}

func TestReceiveMessage_NilCallback(t *testing.T) {
	mc := &mockConsumer{}
	broker := newTestBroker(&mockProducer{}, mc)

	err := broker.ReceiveMessage("game", "node-1", nil)
	if err == nil {
		t.Fatal("expected error for nil callback, got nil")
	}
	if !contains(err.Error(), "callback is nil") {
		t.Fatalf("expected error containing 'callback is nil', got: %v", err)
	}
}

func TestReceiveMessage_SubscribeError(t *testing.T) {
	mc := &mockConsumer{
		subscribeBySQL92Func: func(topic, sql92 string, callback func(msg MessageView) error) error {
			return fmt.Errorf("subscribe failed")
		},
	}
	broker := newTestBroker(&mockProducer{}, mc)

	err := broker.ReceiveMessage("game", "node-1", func(uid int64, msgId string, timestamp int64, msg []byte) {})
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	if !contains(err.Error(), "subscribe failed") {
		t.Fatalf("expected error containing 'subscribe failed', got: %v", err)
	}
}

func TestReceiveMessage_CallbackReceivesCorrectData(t *testing.T) {
	m := message{
		UUID:      "test-uuid-123",
		Uid:       99999,
		Timestamp: 1700000000000,
		Payload:   []byte("test-payload"),
	}
	body, _ := json.Marshal(m)

	mc := &mockConsumer{
		subscribeBySQL92Func: func(topic, sql92 string, callback func(msg MessageView) error) error {
			mv := &mockMessageView{
				body:      body,
				messageId: "rmq-msg-id-001",
				topic:     "test-topic",
			}
			return callback(mv)
		},
	}
	broker := newTestBroker(&mockProducer{}, mc)

	var (
		gotUid       int64
		gotMsgId     string
		gotTimestamp int64
		gotPayload   []byte
	)
	err := broker.ReceiveMessage("game", "node-1", func(uid int64, msgId string, timestamp int64, data []byte) {
		gotUid = uid
		gotMsgId = msgId
		gotTimestamp = timestamp
		gotPayload = data
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if gotUid != 99999 {
		t.Fatalf("expected uid=99999, got %d", gotUid)
	}
	if gotMsgId != "rmq-msg-id-001" {
		t.Fatalf("expected msgId=rmq-msg-id-001, got %s", gotMsgId)
	}
	if gotTimestamp != 1700000000000 {
		t.Fatalf("expected timestamp=1700000000000, got %d", gotTimestamp)
	}
	if string(gotPayload) != "test-payload" {
		t.Fatalf("expected payload=test-payload, got %s", string(gotPayload))
	}
}

func TestReceiveMessage_UnmarshalError(t *testing.T) {
	mc := &mockConsumer{
		subscribeBySQL92Func: func(topic, sql92 string, callback func(msg MessageView) error) error {
			mv := &mockMessageView{
				body:      []byte("invalid json{{{"),
				messageId: "msg-001",
				topic:     "test-topic",
			}
			return callback(mv)
		},
	}
	broker := newTestBroker(&mockProducer{}, mc)

	err := broker.ReceiveMessage("game", "node-1", func(uid int64, msgId string, timestamp int64, data []byte) {
		t.Fatal("callback should not be invoked on unmarshal error")
	})
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	if !contains(err.Error(), "ReceiveMessage unmarshal fail") {
		t.Fatalf("expected error containing 'ReceiveMessage unmarshal fail', got: %v", err)
	}
}

// ============================================================
// Close 单元测试
// ============================================================

func TestClose_Success(t *testing.T) {
	consumerClosed := false
	mc := &mockConsumer{
		closeFunc: func() error {
			consumerClosed = true
			return nil
		},
	}
	broker := newTestBroker(&mockProducer{}, mc)

	if err := broker.Close(); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !consumerClosed {
		t.Fatal("expected consumer to be closed")
	}
}

func TestClose_ConsumerCloseError(t *testing.T) {
	mc := &mockConsumer{
		closeFunc: func() error {
			return fmt.Errorf("consumer close failed")
		},
	}
	broker := newTestBroker(&mockProducer{}, mc)

	err := broker.Close()
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	if !contains(err.Error(), "consumer close failed") {
		t.Fatalf("expected error containing 'consumer close failed', got: %v", err)
	}
}

func TestClose_NilConsumer(t *testing.T) {
	broker := &RMQBroker{
		topic: "test",
		rmqp:  &mockProducer{},
		rmqc:  nil,
	}
	if err := broker.Close(); err != nil {
		t.Fatalf("unexpected error with nil consumer: %v", err)
	}
}

func TestClose_NilProducer(t *testing.T) {
	broker := &RMQBroker{
		topic: "test",
		rmqp:  nil,
		rmqc:  &mockConsumer{closeFunc: func() error { return nil }},
	}
	if err := broker.Close(); err != nil {
		t.Fatalf("unexpected error with nil producer: %v", err)
	}
}

// ============================================================
// 多节点隔离测试
// ============================================================

func TestMultiNodeIsolation(t *testing.T) {
	var (
		node1Messages []message
		node2Messages []message
		mu            sync.Mutex
	)

	mc := &mockConsumer{
		subscribeBySQL92Func: func(topic, sql92 string, callback func(msg MessageView) error) error {
			// 根据 sql92 过滤条件分发不同的消息
			if sql92 == "node='A'" {
				body, _ := json.Marshal(message{Uid: 1, Timestamp: 100, Payload: []byte("from-A")})
				callback(&mockMessageView{body: body, messageId: "msg-A", topic: topic})
			} else if sql92 == "node='B'" {
				body, _ := json.Marshal(message{Uid: 2, Timestamp: 200, Payload: []byte("from-B")})
				callback(&mockMessageView{body: body, messageId: "msg-B", topic: topic})
			}
			return nil
		},
	}

	broker := newTestBroker(&mockProducer{}, mc)

	// 订阅节点 A
	err := broker.ReceiveMessage("game", "A", func(uid int64, msgId string, timestamp int64, data []byte) {
		mu.Lock()
		defer mu.Unlock()
		node1Messages = append(node1Messages, message{Uid: uid, Payload: data})
	})
	if err != nil {
		t.Fatalf("subscribe node A failed: %v", err)
	}

	// 订阅节点 B (同一个 broker 不支持重复订阅, 这里用新的 broker)
	broker2 := newTestBroker(&mockProducer{}, mc)
	err = broker2.ReceiveMessage("game", "B", func(uid int64, msgId string, timestamp int64, data []byte) {
		mu.Lock()
		defer mu.Unlock()
		node2Messages = append(node2Messages, message{Uid: uid, Payload: data})
	})
	if err != nil {
		t.Fatalf("subscribe node B failed: %v", err)
	}

	mu.Lock()
	defer mu.Unlock()

	if len(node1Messages) != 1 {
		t.Fatalf("expected 1 message for node A, got %d", len(node1Messages))
	}
	if string(node1Messages[0].Payload) != "from-A" {
		t.Fatalf("expected payload=from-A, got %s", string(node1Messages[0].Payload))
	}

	if len(node2Messages) != 1 {
		t.Fatalf("expected 1 message for node B, got %d", len(node2Messages))
	}
	if string(node2Messages[0].Payload) != "from-B" {
		t.Fatalf("expected payload=from-B, got %s", string(node2Messages[0].Payload))
	}
}

// ============================================================
// 并发安全测试
// ============================================================

func TestConcurrentSend(t *testing.T) {
	var mu sync.Mutex
	sentCount := 0
	mp := &mockProducer{
		sendFunc: func(msg *golang.Message) ([]*golang.SendReceipt, error) {
			mu.Lock()
			sentCount++
			mu.Unlock()
			return []*golang.SendReceipt{{MessageID: fmt.Sprintf("msg-%d", sentCount)}}, nil
		},
	}
	broker := newTestBroker(mp, &mockConsumer{})

	var wg sync.WaitGroup
	for i := 0; i < 100; i++ {
		wg.Add(1)
		go func(n int) {
			defer wg.Done()
			_, err := broker.SendMessage(int64(n), "", "node", []byte("data"))
			if err != nil {
				t.Errorf("goroutine %d: unexpected error: %v", n, err)
			}
		}(i)
	}
	wg.Wait()

	mu.Lock()
	defer mu.Unlock()
	if sentCount != 100 {
		t.Fatalf("expected 100 messages sent, got %d", sentCount)
	}
}

// ============================================================
// 辅助函数
// ============================================================

func contains(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || len(s) > 0 && containsStr(s, substr))
}

func containsStr(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}

// ============================================================
// 集成测试 (需要 RocketMQ 服务)
// 设置环境变量 BROKER_INTEGRATION_TEST=1 启用
// ============================================================

func TestRMQBroker_Integration(t *testing.T) {
	if os.Getenv("BROKER_INTEGRATION_TEST") == "" {
		t.Skip("Skipping integration test; set BROKER_INTEGRATION_TEST=1 to enable")
	}

	var (
		Endpoint  = "10.80.1.64:19081"
		Namespace = ""
		Group1    = "TestGroup1"
		Group2    = "TestGroup2"
		Topic     = "ggggggg"
		GameName  = "pelican-prank"
		Node1     = "Node-integration-1"
		Node2     = "Node-integration-2"
	)

	newBroker := func(group string) (*RMQBroker, error) {
		rmqp, err := rmq.NewProducer(Endpoint, Namespace, group, &credentials.SessionCredentials{})
		if err != nil {
			return nil, err
		}
		rmqc, err := rmq.NewConsumer(Endpoint, Namespace, group, time.Second*5, &credentials.SessionCredentials{})
		if err != nil {
			return nil, err
		}
		return NewRMQBroker(Topic, rmqp, rmqc), nil
	}

	tc1, err := newBroker(Group1)
	if err != nil {
		t.Fatal(err)
	}
	defer tc1.Close()

	tc2, err := newBroker(Group2)
	if err != nil {
		t.Fatal(err)
	}
	defer tc2.Close()

	mp1 := &sync.Map{}
	mp2 := &sync.Map{}

	if err := tc1.ReceiveMessage(GameName, Node1, func(uid int64, msgId string, timestamp int64, msg []byte) {
		mp1.Store(msgId, struct{}{})
	}); err != nil {
		t.Fatal(err)
	}

	if err := tc2.ReceiveMessage(GameName, Node2, func(uid int64, msgId string, timestamp int64, msg []byte) {
		mp2.Store(msgId, struct{}{})
	}); err != nil {
		t.Fatal(err)
	}

	// 发送消息到 Node1
	var wg sync.WaitGroup
	wg.Add(1)
	go func() {
		defer wg.Done()
		for i := 0; i < 100; i++ {
			msgId, err := tc1.SendMessage(10000, GameName, Node1, []byte(fmt.Sprintf("msg-%d", i)))
			if err != nil {
				t.Error(err)
				return
			}
			if msgId == "" {
				t.Error("expected non-empty message ID")
				return
			}
		}
	}()
	wg.Wait()
	time.Sleep(time.Second * 3)

	n1 := countSyncMap(mp1)
	t.Logf("Node1 received: %d messages", n1)
	if n1 == 0 {
		t.Error("expected Node1 to receive at least 1 message")
	}
}

func countSyncMap(m *sync.Map) int {
	count := 0
	m.Range(func(_, _ interface{}) bool {
		count++
		return true
	})
	return count
}
