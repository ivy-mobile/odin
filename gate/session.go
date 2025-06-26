package gate

import (
	"sync"

	"github.com/ivy-mobile/odin/xutil/xlog"

	"github.com/olahol/melody"
)

// SessionManager 会话管理器
type Sessions struct {
	mu sync.RWMutex              // 读写锁
	ss map[int64]*melody.Session // 所有会话
}

func NewSessions() *Sessions {
	return &Sessions{
		ss: make(map[int64]*melody.Session),
	}
}

func (ss *Sessions) Get(userId int64) *melody.Session {
	ss.mu.RLock()
	defer ss.mu.RUnlock()
	return ss.ss[userId]
}

// Set 设置会话
func (ss *Sessions) Set(userId int64, session *melody.Session) {
	ss.mu.Lock()
	defer ss.mu.Unlock()
	ss.ss[userId] = session
}

// Remove 移除会话
func (ss *Sessions) Remove(userId int64) {
	ss.mu.Lock()
	defer ss.mu.Unlock()
	delete(ss.ss, userId)
}

// Send 发送消息 - 仅支持二进制消息
func (ss *Sessions) Send(msg []byte, uids ...int64) {
	for _, uid := range uids {
		session := ss.Get(uid)
		if session == nil {
			continue
		}
		if err := session.WriteBinary(msg); err != nil {
			xlog.Error().Msgf("send binary message to userID: %d failed: %s", uid, err.Error())
		}
	}
}

// SendText 发送消息 - 仅支持文本消息
func (ss *Sessions) SendText(msg []byte, uids ...int64) {
	for _, uid := range uids {
		session := ss.Get(uid)
		if session == nil {
			continue
		}
		if err := session.Write(msg); err != nil {
			xlog.Error().Msgf("send text message to userID: %d failed: %s", uid, err.Error())
		}
	}
}
