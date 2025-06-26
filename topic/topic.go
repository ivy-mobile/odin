package topic

import "fmt"

const (
	Topic_AdminCmd = "admin.%s" // 后台指令
)

// Gate2GameTopic 网关->游戏 消息topic
// gate: 网关服务名 如: game-gateway
// game: 游戏服务名 如: hamster-battle
// TODO v2版本gate和Game都是多节点，topic需要处理指定节点 如:game-gateway-2.hamster-battle-1
func Gate2GameTopic(gate, game string) string {
	return fmt.Sprintf("%s.%s", gate, game)
}

// Game2GateTopic 游戏->网关 消息topic
// gate: 网关服务名 如: game-gateway
// game: 游戏服务名 如: hamster-battle
// TODO v2版本gate和Game都是多节点，topic需要处理指定节点 如:game-gateway-2.hamster-battle-1
func Game2GateTopic(gate, game string) string {
	return fmt.Sprintf("%s.%s", game, gate)
}

// Admin2GameTopic 管理后台->游戏 消息topic
func Admin2GameTopic(game string) string {
	return fmt.Sprintf(Topic_AdminCmd, game)
}
