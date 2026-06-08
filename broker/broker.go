package broker

// Broker 消息代理接口
type Broker interface {
	// SendMessage 发送消息, 返回消息ID
	SendMessage(uid int64, gameName, node string, payload []byte) (string, error)
	// ReceiveMessage 监听节点消息
	// fn: 回调函数, uid=用户ID, msgId=消息ID, timestamp=毫秒时间戳, msg=消息体
	ReceiveMessage(gameName, node string, fn func(uid int64, msgId string, timestamp int64, msg []byte)) error
	// Close 释放资源
	Close() error
}
