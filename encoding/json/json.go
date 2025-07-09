package json

import (
	"github.com/bytedance/sonic"
)

const Name = "json"

var Codec = &codec{}

type codec struct{}

// Name 编解码器名称
func (codec) Name() string {
	return Name
}

// Marshal 编码
func (codec) Marshal(v any) ([]byte, error) {
	return sonic.Marshal(v)
}

// Unmarshal 解码
func (codec) Unmarshal(data []byte, v any) error {
	return sonic.Unmarshal(data, v)
}

// Marshal 编码
func Marshal(v any) ([]byte, error) {
	return Codec.Marshal(v)
}

// Unmarshal 解码
func Unmarshal(data []byte, v any) error {
	return Codec.Unmarshal(data, v)
}
