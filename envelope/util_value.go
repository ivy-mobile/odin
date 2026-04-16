package envelope

// 提供了对 envelope.Value 的一些常用操作工具方法

// Int32V int32 转 Value(pb)
func Int32V(v int32) *Value {
	return &Value{Value: &Value_I32{I32: v}}
}

// Int64V int64 转 Value(pb)
func Int64V(v int64) *Value {
	return &Value{Value: &Value_I64{I64: v}}
}

// Float32V float32 转 Value(pb)
func Float32V(v float32) *Value {
	return &Value{Value: &Value_F32{F32: v}}
}

// Float64V float64 转 Value(pb)
func Float64V(v float64) *Value {
	return &Value{Value: &Value_F64{F64: v}}
}

// StrV string 转 Value(pb)
func StrV(v string) *Value {
	return &Value{Value: &Value_Str{Str: v}}
}

// BoolV bool 转 Value(pb)
func BoolV(v bool) *Value {
	return &Value{Value: &Value_Bool{Bool: v}}
}

// BytesV []byte 转 Value(pb)
func BytesV(v []byte) *Value {
	return &Value{Value: &Value_Bytes{Bytes: v}}
}

// UInt32V uint32 转 Value(pb)
func UInt32V(v uint32) *Value {
	return &Value{Value: &Value_U32{U32: v}}
}

// UInt64V uint64 转 Value(pb)
func UInt64V(v uint64) *Value {
	return &Value{Value: &Value_U64{U64: v}}
}
