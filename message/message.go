package message

// Message 统一消息数据接口，兼容json 、proto
type Message interface {
	// GetSeq 序列号
	GetSeq() uint64
	// GetUid 用户ID
	GetUid() int64
	// GetRoute 路由ID
	GetRoute() string
	// GetGame 游戏名
	GetGame() string
	// GetMsgId 消息ID
	GetMsgId() uint64
	// GetTimestamp 时间戳
	GetTimestamp() int64
	// GetVersion 版本号
	GetVersion() string
	// GetPayload 具体的业务数据
	GetPayload() []byte
}
