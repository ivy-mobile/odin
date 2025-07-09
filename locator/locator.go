package locator

import "context"

type NodeType string

const (
	NodeTypeGate NodeType = "gate" // 网关节点
	NodeTypeGame NodeType = "game" // 游戏节点
)

// Locator 定位器
// 用于定位用户所在的 Game节点 或 Gate节点
type Locator interface {

	// BindGate 绑定网关节点
	BindGate(ctx context.Context, uid int64, gateID string) error
	// UnbindGate 解绑网关节点
	UnbindGate(ctx context.Context, uid int64, gateID string) error
	// GetGateNode 获取用户所在的 Gate 节点
	GetGateNode(ctx context.Context, uid int64) (string, error)

	// BindGame 绑定游戏节点
	BindGame(ctx context.Context, uid int64, gameName, gameID string) error
	// UnbindGame 解绑游戏节点
	UnbindGame(ctx context.Context, uid int64, gameName, gameID string) error
	// GetGameNode 获取用户所在的 Game 节点
	GetGameNode(ctx context.Context, uid int64, gameName string) (string, error)

	// WatchChange 监听变化
	WatchChange(ctx context.Context, channels ...EventChannel)
}
