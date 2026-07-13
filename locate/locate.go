package locate

// Locator 玩家定位器
type Locator interface {
	Name() string
	// BindGateNode 绑定最新网关节点
	BindGateNode(uid int64, node string) error
	// UnBindGateNode 解绑当前网关节点
	UnBindGateNode(uid int64, node string) error
	// GetGateNode 获取玩家当前网关节点
	GetGateNode(uid int64) (string, error)
}
