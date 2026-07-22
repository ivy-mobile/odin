package member

import (
	"time"

	"github.com/ivy-mobile/odin/envelope"
)

// Member 成员（用户），作为所有游戏共用的房间成员模型
type Member struct {
	ID             int64     // 玩家 ID
	IsRobot        bool      // 是否机器人
	Nickname       string    // 昵称
	Avatar         string    // 头像
	Gender         string    // 性别
	IsReady        bool      // 是否已准备
	LastActiveTime time.Time // 最近活跃时间
	Offline        bool      // 是否离线

	SeatID int // 座位 ID
	RoomID int // 房间 ID
}

// Snapshot 构造成员协议快照
func (m *Member) Snapshot() *envelope.Member {
	meta := make(map[string]*envelope.Value)

	return &envelope.Member{
		Uid:      m.ID,
		Nickname: m.Nickname,
		Avatar:   m.Avatar,
		Gender:   m.Gender,
		SeatId:   int32(m.SeatID),
		IsReady:  m.IsReady,
		Meta:     meta,
	}
}
