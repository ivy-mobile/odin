package envelope

// 从 map 中获取值

func GetMapValue(m map[string]*Value, key string) *Value {
	return m[key]
}

func GetMapInt32(m map[string]*Value, key string) int32 {
	return m[key].GetI32()
}

func GetMapInt64(m map[string]*Value, key string) int64 {
	return m[key].GetI64()
}

func GetMapUint32(m map[string]*Value, key string) uint32 {
	return m[key].GetU32()
}

func GetMapUint64(m map[string]*Value, key string) uint64 {
	return m[key].GetU64()
}

func GetMapFloat32(m map[string]*Value, key string) float32 {
	return m[key].GetF32()
}

func GetMapFloat64(m map[string]*Value, key string) float64 {
	return m[key].GetF64()
}

func GetMapStr(m map[string]*Value, key string) string {
	return m[key].GetStr()
}

func GetMapBool(m map[string]*Value, key string) bool {
	return m[key].GetBool()
}

func GetMapBytes(m map[string]*Value, key string) []byte {
	return m[key].GetBytes()
}

// 向 map 中添加值

func PutMapInt32(m map[string]*Value, key string, value int32) {
	m[key] = Int32V(value)
}

func PutMapInt64(m map[string]*Value, key string, value int64) {
	m[key] = Int64V(value)
}

func PutMapUint32(m map[string]*Value, key string, value uint32) {
	m[key] = UInt32V(value)
}

func PutMapUint64(m map[string]*Value, key string, value uint64) {
	m[key] = UInt64V(value)
}

func PutMapFloat32(m map[string]*Value, key string, value float32) {
	m[key] = Float32V(value)
}

func PutMapFloat64(m map[string]*Value, key string, value float64) {
	m[key] = Float64V(value)
}

func PutMapStr(m map[string]*Value, key string, value string) {
	m[key] = StrV(value)
}

func PutMapBool(m map[string]*Value, key string, value bool) {
	m[key] = BoolV(value)
}

func PutMapBytes(m map[string]*Value, key string, value []byte) {
	m[key] = BytesV(value)
}
