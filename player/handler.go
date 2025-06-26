package player

// MsgHandler 玩家消息发送器
type MsgHandler interface {
	SendMessage(seq uint64, uid int64, route, version string, msgID uint64, payload any) error
}
