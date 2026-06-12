package broker

type message struct {
	UUID string `json:"uuid"`
	//nolint:revive // JSON 字段兼容历史 uid 命名。
	Uid       int64  `json:"uid"`
	Timestamp int64  `json:"timestamp"` // ms
	Payload   []byte `json:"payload"`
}
