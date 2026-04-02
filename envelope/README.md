# envelope 模块文档

## 概述

`envelope` 提供了一套基于 Protobuf 的统一消息结构，用于描述客户端输入消息和服务端输出消息。它适合作为业务协议最外层的信封层，再在 `payload` 或 `data` 字段中承载具体业务数据。

## 文件说明

```text
envelope.proto    # Protobuf 协议定义
envelope.pb.go    # 生成后的 Go 代码
```

## 核心消息

### `Header`

统一消息头，包含序列号、用户 ID、游戏 ID、消息 ID、时间戳和版本号等基础元信息。

### `InputMessage`

客户端到服务端的输入消息：

- `header`：通用头信息
- `route`：业务路由
- `payload`：序列化后的业务请求体

### `OutputMessage`

服务端响应或推送消息：

- `Header`：通用头信息
- `msg_type`：消息类型，区分推送或响应
- `error_code` / `error_msg`：错误信息
- `msg_tag`：消息标签
- `data`：序列化后的业务数据

## 使用示例

```go
package main

import (
	"fmt"
	"time"

	"github.com/ivy-mobile/odin/envelope"
	"google.golang.org/protobuf/proto"
)

func main() {
	msg := &envelope.InputMessage{
		Header: &envelope.Header{
			Seq:       1,
			Uid:       10001,
			GameId:    100,
			MsgId:     "msg-1",
			Timestamp: time.Now().UnixMilli(),
			Version:   "v1",
		},
		Route:   "ping",
		Payload: []byte("hello"),
	}

	data, err := proto.Marshal(msg)
	if err != nil {
		panic(err)
	}

	var decoded envelope.InputMessage
	if err := proto.Unmarshal(data, &decoded); err != nil {
		panic(err)
	}

	fmt.Println(decoded.Route)
}
```
