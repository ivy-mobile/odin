package encoding

import (
	"github.com/ivy-mobile/odin/encoding/json"
	"github.com/ivy-mobile/odin/encoding/msgpack"
	"github.com/ivy-mobile/odin/encoding/proto"
	"github.com/ivy-mobile/odin/encoding/toml"
	"github.com/ivy-mobile/odin/encoding/xml"
	"github.com/ivy-mobile/odin/encoding/yaml"
	"github.com/ivy-mobile/odin/xutil/xlog"
)

var codecs = make(map[string]Codec)

func init() {
	Register(json.DefaultCodec)
	Register(proto.DefaultCodec)
	Register(toml.DefaultCodec)
	Register(xml.DefaultCodec)
	Register(yaml.DefaultCodec)
	Register(msgpack.DefaultCodec)
}

type Codec interface {
	// Name 编解码器类型
	Name() string
	// Marshal 编码
	Marshal(v any) ([]byte, error)
	// Unmarshal 解码
	Unmarshal(data []byte, v any) error
}

// Register 注册编解码器
func Register(codec Codec) {
	if codec == nil {
		xlog.Fatal().Msg("can't register a invalid codec")
	}

	name := codec.Name()

	if name == "" {
		xlog.Fatal().Msg("can't register a codec without name")
	}

	if _, ok := codecs[name]; ok {
		xlog.Warn().Msgf("the old %s codec will be overwritten", name)
	}

	codecs[name] = codec
}

// Invoke 调用编解码器
func Invoke(name string) Codec {
	codec, ok := codecs[name]
	if !ok {
		xlog.Fatal().Msgf("%s codec is not registered", name)
	}
	return codec
}
