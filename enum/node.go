package enum

import "fmt"

// 节点类型标识
const (
	NodeType_Gate = "Gate" // 网关节点类
	NodeType_Game = "Game" // 游戏节点类
)

// GameNodeName 游戏节点名称
func GameNodeName(serviceName, serviceId, alias string) string {
	return fmt.Sprintf("%s-%s-%s", serviceName, alias, serviceId)
}

// GateNodeName 网关节点名称
func GateNodeName(serviceName, serviceId string) string {
	return fmt.Sprintf("%s-%s", serviceName, serviceId)
}

// 节点状态
const (
	NodeState_Work     = "Work"     // 工作中
	NodeState_Hang     = "Hang"     // 挂起
	NodeState_Shutdown = "Shutdown" // // 关闭
)
