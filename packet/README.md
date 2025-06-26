# Packet 模块文档

## 概述
`packet` 模块提供了消息打包和解包的功能，支持心跳包的处理，并且可以处理不同类型的读取器。该模块主要用于网络通信中消息的序列化和反序列化操作，目前仅支持二进制数据格式，使用 `encoding/binary` 包进行数据的读写操作，在消息打包、解包以及心跳包处理等核心功能中均采用二进制方式处理数据。

## 项目结构
```
├── errors.go         # 定义模块相关的错误信息
├── message.go        # 定义消息结构体
├── options.go        # 定义打包器的配置选项
├── packer.go         # 实现打包器的核心逻辑
├── packet.go         # 提供全局打包器的操作接口
└── packet_test.go    # 模块的测试文件
```

## 核心结构体和接口
### `Message` 结构体
<mcfile name="message.go" path="/root/codes/github.com/ivy-mobile/odin/packet/message.go"></mcfile>
```go
type Message struct {
    Seq    int32  // 序列号
    Route  int32  // 路由ID
    Buffer []byte // 消息内容
}
```

### `Packer` 接口
<mcfile name="packer.go" path="/root/codes/github.com/ivy-mobile/odin/packet/packer.go"></mcfile>
```go
type Packer interface {
    // ReadMessage 读取消息
    ReadMessage(reader any) ([]byte, error)
    // PackBuffer 打包消息
    PackBuffer(message *Message) (xbuffer.Buffer, error)
    // PackMessage 打包消息
    PackMessage(message *Message) ([]byte, error)
    // UnpackMessage 解包消息
    UnpackMessage(data []byte) (*Message, error)
    // PackHeartbeat 打包心跳
    PackHeartbeat() ([]byte, error)
    // CheckHeartbeat 检测心跳包
    CheckHeartbeat(data []byte) (bool, error)
}
```

## 核心函数
### 全局打包器初始化
<mcfile name="packet.go" path="/root/codes/github.com/ivy-mobile/odin/packet/packet.go"></mcfile>
```go
func init() {
    globalPacker = NewPacker()
}
```

### 消息打包和解包函数
- `PackMessage(message *Message) ([]byte, error)`: 打包消息。
- `UnpackMessage(data []byte) (*Message, error)`: 解包消息。
- `PackHeartbeat() ([]byte, error)`: 打包心跳包。
- `CheckHeartbeat(data []byte) (bool, error)`: 检测心跳包。

## 错误处理
<mcfile name="errors.go" path="/root/codes/github.com/ivy-mobile/odin/packet/errors.go"></mcfile>
```go
var (
    ErrInvalidReader   = errors.New("invalid reader")
    ErrSeqOverflow     = errors.New("seq overflow")
    ErrRouteOverflow   = errors.New("route overflow")
    ErrMessageTooLarge = errors.New("message too large")
    ErrInvalidMessage  = errors.New("invalid message")
)
```

## 使用示例
### 打包和解包消息
```go
package main

import (
    "fmt"
    "github.com/ivy-mobile/odin/packet"
)

func main() {
    message := &packet.Message{
        Seq:    1,
        Route:  1,
        Buffer: []byte("hello world"),
    }

    // 打包消息
    data, err := packet.PackMessage(message)
    if err != nil {
        fmt.Println("打包消息失败:", err)
        return
    }

    // 解包消息
    unpackedMessage, err := packet.UnpackMessage(data)
    if err != nil {
        fmt.Println("解包消息失败:", err)
        return
    }

    fmt.Printf("解包结果: Seq=%d, Route=%d, Buffer=%s\n", unpackedMessage.Seq, unpackedMessage.Route, string(unpackedMessage.Buffer))
}
```

### 处理心跳包
```go
package main

import (
    "fmt"
    "github.com/ivy-mobile/odin/packet"
)

func main() {
    // 打包心跳包
    heartbeatData, err := packet.PackHeartbeat()
    if err != nil {
        fmt.Println("打包心跳包失败:", err)
        return
    }

    // 检测心跳包
    isHeartbeat, err := packet.CheckHeartbeat(heartbeatData)
    if err != nil {
        fmt.Println("检测心跳包失败:", err)
        return
    }

    fmt.Printf("是否为心跳包: %v\n", isHeartbeat)
}
```

## 测试
模块包含了一系列的基准测试和单元测试，可以在 `packet_test.go` 文件中查看。运行测试命令：
```sh
go test -v ./packet
```