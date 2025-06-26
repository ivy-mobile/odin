# github.com/ivy-mobile/odin 编码模块

该模块提供了多种数据编解码器的实现，支持常见的数据格式，方便在不同场景下进行数据的编码和解码操作。

## 支持的编解码器

| 编解码器 | 包名 | 实现文件 | 依赖库 |
| --- | --- | --- | --- |
| JSON | `json` | `json/json.go` | `github.com/bytedance/sonic` |
| MessagePack | `msgpack` | `msgpack/msgpack.go` | `github.com/vmihailenco/msgpack/v5` |
| Protocol Buffers | `proto` | `proto/proto.go` | `google.golang.org/protobuf/proto` |
| TOML | `toml` | `toml/toml.go` | `github.com/BurntSushi/toml` |
| XML | `xml` | `xml/xml.go` | 标准库 `encoding/xml` |
| YAML | `yaml` | `yaml/yaml.go` | `gopkg.in/yaml.v3` |

## 核心接口

### `Codec` 接口
定义了编解码器的基本方法，包括获取编解码器名称、编码和解码操作。
```go
// encoding/codec.go

type Codec interface {
    // Name 编解码器类型
    Name() string
    // Marshal 编码
    Marshal(v any) ([]byte, error)
    // Unmarshal 解码
    Unmarshal(data []byte, v any) error
}
```

## 核心函数

### `Register`
用于注册编解码器到全局注册表中。
```go
// encoding/codec.go

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
```

### `Invoke`
根据编解码器名称调用对应的编解码器。
```go
// encoding/codec.go

func Invoke(name string) Codec {
    codec, ok := codecs[name]
    if !ok {
        xlog.Fatal().Msgf("%s codec is not registered", name)
    }
    return codec
}
```

## 初始化
在 `codec.go` 的 `init` 函数中，已经注册了所有支持的编解码器。
```go
// encoding/codec.go

func init() {
    Register(json.DefaultCodec)
    Register(proto.DefaultCodec)
    Register(toml.DefaultCodec)
    Register(xml.DefaultCodec)
    Register(yaml.DefaultCodec)
    Register(msgpack.DefaultCodec)
}
```

## 使用示例
```go
package main

import (
    "fmt"
    "github.com/ivy-mobile/odin/encoding"
    "github.com/ivy-mobile/odin/encoding/json"
)

func main() {
    // 注册编解码器（已在 init 函数中完成）
    // 调用 JSON 编解码器
    data := map[string]interface{}{"message": "Hello, World!"}
    encoded, err := encoding.Invoke(json.Name).Marshal(data)
    if err != nil {
        fmt.Println("编码失败:", err)
        return
    }
    fmt.Println("编码结果:", string(encoded))

    var decoded map[string]interface{}
    err = encoding.Invoke(json.Name).Unmarshal(encoded, &decoded)
    if err != nil {
        fmt.Println("解码失败:", err)
        return
    }
    fmt.Println("解码结果:", decoded)
}
```