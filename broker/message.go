package broker

type message struct {
	UUID      string `json:"uuid"`
	Uid       int64  `json:"uid"`
	Timestamp int64  `json:"timestamp"` // ms
	Payload   []byte `json:"payload"`
}
