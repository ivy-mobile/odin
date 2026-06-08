package broker

// Broker 消息代理
type Broker interface {
	// SendMessage 发送消息
	SendMessage(uid int64, gameName, node string, payload []byte) (string, error)
	// ReceiveMessage 监听节点消息
	ReceiveMessage(gameName, node string, fn func(uid int64, msgId string, timestamp int64, msg []byte)) error
	// Close 释放资源
	Close() error
}
