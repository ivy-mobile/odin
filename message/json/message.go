package msgjson

import (
	"github.com/ivy-mobile/odin/message"
)

// JsonMessage Json消息统一格式
type JsonMessage struct {
	Seq       uint64 `json:"seq"`       // 序列号
	Uid       int64  `json:"uid"`       // 用户ID
	Route     string `json:"route"`     // 路由ID
	Game      string `json:"game"`      // 游戏服务名
	MsgID     uint64 `json:"msgId"`     // 消息ID
	Timestamp int64  `json:"timestamp"` // 时间戳 - 毫秒
	Version   string `json:"version"`   // 版本号
	Payload   []byte `json:"payload"`   // 具体游戏的业务数据
}

var _ message.Message = (*JsonMessage)(nil)

// GetSeq 获取序列号
func (m *JsonMessage) GetSeq() uint64 {
	return m.Seq
}

// GetUid 获取用户ID
func (m *JsonMessage) GetUid() int64 {
	return m.Uid
}

// GetRoute 获取路由ID
func (m *JsonMessage) GetRoute() string {
	return m.Route
}

// GetGame 获取游戏名
func (m *JsonMessage) GetGame() string {
	return m.Game
}

// GetMsgId 获取消息ID
func (m *JsonMessage) GetMsgId() uint64 {
	return m.MsgID
}

// GetTimestamp 获取时间戳
func (m *JsonMessage) GetTimestamp() int64 {
	return m.Timestamp
}

// GetPayload 获取具体的业务数据
func (m *JsonMessage) GetPayload() []byte {
	return m.Payload
}

// GetVersion 获取版本号
func (m *JsonMessage) GetVersion() string {
	return m.Version
}
