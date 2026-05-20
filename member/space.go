package member

import (
	"sync"
	"sync/atomic"
)

// Space 成员管理器
type Space struct {
	count int32
	data  sync.Map
}

// NewSpace 创建成员管理器
func NewSpace() *Space {
	return &Space{}
}

// Get 按用户 ID 获取成员
func (s *Space) Get(id int64) *Member {
	if v, ok := s.data.Load(id); ok {
		return v.(*Member)
	}
	return nil
}

// Set 保存成员
func (s *Space) Set(id int64, m *Member) {
	atomic.AddInt32(&s.count, 1)
	s.data.Store(id, m)
}

// Remove 删除成员
func (s *Space) Remove(id int64) {
	atomic.AddInt32(&s.count, -1)
	s.data.Delete(id)
}

// All 返回成员列表快照。
func (s *Space) All() []*Member {
	members := make([]*Member, 0)
	s.data.Range(func(key, value interface{}) bool {
		members = append(members, value.(*Member))
		return true
	})
	return members
}

// Range 遍历成员列表快照。
func (s *Space) Range(f func(id int64, m *Member)) {
	for _, m := range s.All() {
		f(m.ID, m)
	}
}

// Count 返回当前成员数量。
func (s *Space) Count() int {
	return int(atomic.LoadInt32(&s.count))
}
