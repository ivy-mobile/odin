package xconv

import "github.com/ivy-mobile/odin/encoding/json"

// Scan json反序列化
func Scan(data any, target any) error {
	return json.Unmarshal(Bytes(data), target)
}
