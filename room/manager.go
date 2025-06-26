package room

import "sync"

// Manager 房间管理器
type Manager struct {
	rooms sync.Map
}

// NewManager 创建房间管理器
func NewManager() *Manager {
	return &Manager{
		rooms: sync.Map{},
	}
}

// AddRoom 添加房间
func (m *Manager) Add(room Room) {
	m.rooms.Store(room.ID(), room)
}

// GetRoom 获取房间
func (m *Manager) Get(id int) (Room, bool) {
	room, ok := m.rooms.Load(id)
	if !ok {
		return nil, false
	}
	return room.(Room), ok
}

// Remove 删除房间
func (m *Manager) Remove(id int) {
	m.rooms.Delete(id)
}

// Range 遍历房间
// 若fn返回false，则终止循环
func (m *Manager) Range(fn func(id int, room Room) bool) {
	m.rooms.Range(func(key, value any) bool {
		return fn(key.(int), value.(Room))
	})
}
