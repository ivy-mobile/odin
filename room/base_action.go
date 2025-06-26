package room

// Action 动作
type Action struct {
	fn     func() (uint16, error) // 动作核心逻辑
	result chan ActResult         // 动作结果
}

// ActResult 动作结果
type ActResult struct {
	code uint16 // 状态码,取值范围:0-65535
	err  error
}

// OK 是否成功

func (r *ActResult) OK() bool {
	return r.err == nil
}

// Err 错误信息
func (r *ActResult) Err() error {
	return r.err
}
