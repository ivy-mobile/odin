package member

import (
	"time"

	"github.com/ivy-mobile/odin/envelope"
)

// Member 定义成员管理器依赖的最小公共能力
type Member interface {
	ID() int64
	Snapshot() *envelope.Member
}

// Base 成员（用户），作为所有游戏共用的房间成员模型
type Base struct {
	id             int64     // 玩家 ID
	isRobot        bool      // 是否机器人
	nickname       string    // 昵称
	avatar         string    // 头像
	gender         string    // 性别
	isReady        bool      // 是否已准备
	lastActiveTime time.Time // 最近活跃时间
	offline        bool      // 是否离线
	seatID         int       // 座位 ID
	roomID         int       // 房间 ID
}

// NewBase 创建成员基础模型
func NewBase(
	id int64,
	isRobot bool,
	nickname string,
	avatar string,
	gender string,
	isReady bool,
	lastActiveTime time.Time,
	offline bool,
	seatID int,
	roomID int,
) *Base {
	return &Base{
		id:             id,
		isRobot:        isRobot,
		nickname:       nickname,
		avatar:         avatar,
		gender:         gender,
		isReady:        isReady,
		lastActiveTime: lastActiveTime,
		offline:        offline,
		seatID:         seatID,
		roomID:         roomID,
	}
}

// ID 返回玩家 ID
func (m *Base) ID() int64 {
	return m.id
}

// IsRobot 返回成员是否为机器人
func (m *Base) IsRobot() bool {
	return m.isRobot
}

// SetRobot 设置成员是否为机器人
func (m *Base) SetRobot(isRobot bool) {
	m.isRobot = isRobot
}

// Nickname 返回成员昵称
func (m *Base) Nickname() string {
	return m.nickname
}

// SetNickname 设置成员昵称
func (m *Base) SetNickname(nickname string) {
	m.nickname = nickname
}

// Avatar 返回成员头像
func (m *Base) Avatar() string {
	return m.avatar
}

// SetAvatar 设置成员头像
func (m *Base) SetAvatar(avatar string) {
	m.avatar = avatar
}

// Gender 返回成员性别
func (m *Base) Gender() string {
	return m.gender
}

// SetGender 设置成员性别
func (m *Base) SetGender(gender string) {
	m.gender = gender
}

// IsReady 返回成员是否已准备
func (m *Base) IsReady() bool {
	return m.isReady
}

// SetReady 设置成员准备状态
func (m *Base) SetReady(isReady bool) {
	m.isReady = isReady
}

// LastActiveTime 返回成员最近活跃时间
func (m *Base) LastActiveTime() time.Time {
	return m.lastActiveTime
}

// SetLastActiveTime 设置成员最近活跃时间
func (m *Base) SetLastActiveTime(lastActiveTime time.Time) {
	m.lastActiveTime = lastActiveTime
}

// Offline 返回成员是否离线
func (m *Base) Offline() bool {
	return m.offline
}

// SetOffline 设置成员离线状态
func (m *Base) SetOffline(offline bool) {
	m.offline = offline
}

// SeatID 返回成员座位 ID
func (m *Base) SeatID() int {
	return m.seatID
}

// SetSeatID 设置成员座位 ID
func (m *Base) SetSeatID(seatID int) {
	m.seatID = seatID
}

// RoomID 返回成员房间 ID
func (m *Base) RoomID() int {
	return m.roomID
}

// SetRoomID 设置成员房间 ID
func (m *Base) SetRoomID(roomID int) {
	m.roomID = roomID
}

// Snapshot 构造成员协议快照
func (m *Base) Snapshot() *envelope.Member {
	return &envelope.Member{
		Uid:      m.ID(),
		Nickname: m.Nickname(),
		Avatar:   m.Avatar(),
		Gender:   m.Gender(),
		SeatId:   int32(m.SeatID()),
		IsReady:  m.IsReady(),
		Meta:     make(map[string]*envelope.Value),
	}
}
