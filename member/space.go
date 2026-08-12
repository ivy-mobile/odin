package member

import (
	"sync"
	"sync/atomic"
)

// Space 成员管理器
type Space[M Member] struct {
	count atomic.Int32
	data  sync.Map
}

// NewSpace 创建成员管理器
func NewSpace[M Member]() *Space[M] {
	return &Space[M]{}
}

// Get 按用户 ID 获取成员
func (s *Space[M]) Get(id int64) (M, bool) {
	if v, ok := s.data.Load(id); ok {
		return v.(M), true
	}

	var zero M
	return zero, false
}

// Set 保存成员
func (s *Space[M]) Set(m M) {
	_, loaded := s.data.Swap(m.ID(), m)
	if !loaded {
		s.count.Add(1)
	}
}

// Remove 删除成员
func (s *Space[M]) Remove(id int64) {
	if _, loaded := s.data.LoadAndDelete(id); loaded {
		s.count.Add(-1)
	}
}

// All 返回成员列表快照
func (s *Space[M]) All() []M {
	members := make([]M, 0)
	s.data.Range(func(_ interface{}, value interface{}) bool {
		members = append(members, value.(M))
		return true
	})
	return members
}

// Range 遍历成员列表快照
func (s *Space[M]) Range(f func(id int64, m M)) {
	for _, m := range s.All() {
		f(m.ID(), m)
	}
}

// Count 返回当前成员数量
func (s *Space[M]) Count() int {
	return int(s.count.Load())
}
