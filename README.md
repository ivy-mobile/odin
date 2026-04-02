# Odin

Odin 是一个以 Go 编写的基础能力库，当前版本聚焦于数据编解码、协议封装、消息打包、RocketMQ 封装以及一组通用工具包，适合作为服务端基础设施或公共组件库使用。

## 当前模块概览

| 模块 | 说明 | 文档 |
| --- | --- | --- |
| `encoding` | 统一的编解码接口，内置 JSON、MessagePack、Proto、TOML、XML、YAML 实现 | [encoding/README.md](encoding/README.md) |
| `envelope` | 基于 Protobuf 的统一消息信封定义，描述输入输出消息结构 | [envelope/README.md](envelope/README.md) |
| `packet` | 二进制消息打包、解包与心跳处理 | [packet/README.md](packet/README.md) |
| `rmq` | RocketMQ 5.x 生产者与简单消费者封装 | [rmq/README.md](rmq/README.md) |
| `xutil` | 通用工具集合，覆盖配置、日志、ID、网络、缓冲区、任务池等能力 | [xutil/README.md](xutil/README.md) |

## 目录结构

```text
.
├── encoding
├── envelope
├── packet
├── rmq
└── xutil
```

## 变更说明

当前仓库已经移除了早期的游戏框架相关目录，例如 `eventbus`、`game`、`gate`、`locator`、`player`、`registry`、`room` 等。若你依赖这些历史包，请查看对应的历史 tag 或旧分支，不要直接参考当前主线文档。

## 快速开始

```bash
go get github.com/ivy-mobile/odin
```

## 示例

下面的示例展示了如何通过统一编解码接口完成 JSON 编解码：

```go
package main

import (
	"fmt"

	"github.com/ivy-mobile/odin/encoding"
	"github.com/ivy-mobile/odin/encoding/json"
)

type Person struct {
	Name string `json:"name"`
	Age  int    `json:"age"`
}

func main() {
	input := Person{Name: "Alice", Age: 30}

	data, err := encoding.Invoke(json.Name).Marshal(input)
	if err != nil {
		fmt.Println("编码失败:", err)
		return
	}

	var output Person
	if err := encoding.Invoke(json.Name).Unmarshal(data, &output); err != nil {
		fmt.Println("解码失败:", err)
		return
	}

	fmt.Println(string(data))
	fmt.Printf("%+v\n", output)
}
```

如果你需要更细的用法，可以从各模块 README 继续向下查看。
