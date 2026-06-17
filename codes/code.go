package codes

const (
	OK                  = "A000000"
	InternalServerError = "G100000" // 服务内部异常
	BadRequest          = "G100001" // 请求失败
	Unauthorized        = "G100002" // 未授权
	RequestRPCFail      = "G100003" // 远程调用错误
	TxCommitErr         = "G100004" // 事务提交错误
	OptDatabaseErr      = "G100005" // 数据库操作错误
	RequestErr          = "G100006" // 错误的请求数据
	OptRedisErr         = "G100007" // Redis操作错误
	ProtobufErr         = "G100008" // Protobuf错误
	JSONMarshalErr      = "G100009" // Json序列化错误
	InvalidUserID       = "G100010" // 无效的用户ID
	InvalidTableID      = "G100011" // 无效的牌桌ID
	InvalidGameID       = "G100012" // 无效的游戏id
	InvalidOpt          = "G100013" // 无效的操作
)
