package player

import "sync"

// Manager 玩家管理器
type Manager struct {
	players sync.Map
}

// NewManager 新建玩家管理器
func NewManager() *Manager {
	return &Manager{
		players: sync.Map{},
	}
}

// Add 添加玩家
func (m *Manager) Add(player Player) {
	m.players.Store(player.ID(), player)
}

// Remove 删除玩家
func (m *Manager) Remove(id int64) {
	m.players.Delete(id)
}

// Get 获取玩家
func (m *Manager) Get(id int64) (Player, bool) {
	p, ok := m.players.Load(id)
	if !ok {
		return nil, false
	}
	return p.(Player), ok
}

// Range 遍历玩家
func (m *Manager) Range(fn func(id int64, player Player) bool) {
	m.players.Range(func(key, value any) bool {
		return fn(key.(int64), value.(Player))
	})
}
