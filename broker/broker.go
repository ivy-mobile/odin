package broker

// Broker 消息代理
type Broker interface {
	// SendMessage 发送消息
	SendMessage(uid int64, gameName, node string, payload []byte) (string, error)
	// Close 释放资源
	Close() error
}
