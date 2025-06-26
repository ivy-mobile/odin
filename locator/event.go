package locator

import (
	"encoding/json"
)

type EventType int8

const (
	EventType_BindGate   EventType = iota + 1 // 绑定网关节点
	EventType_UnbindGate                      // 解绑网关节点
	EventType_BindGame                        // 绑定游戏节点
	EventType_UnbindGame                      // 解绑游戏节点
)

type Event struct {
	EventType EventType
	Uid       int64  // 用户ID
	NodeName  string // 节点名
	NodeID    string // 节点ID
}

func (e Event) MarshalBinary() (data []byte, err error) {
	return json.Marshal(e)
}

func (e *Event) UnmarshalBinary(data []byte) error {
	return json.Unmarshal(data, e)
}

type EventChannel int8

const (
	EventChannel_Gate EventChannel = iota + 1 // 网关事件
	EventChannel_Game                         // 游戏事件
)
