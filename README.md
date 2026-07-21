# Odin

Odin 是一个以 Go 编写的基础能力库，当前版本聚焦于数据编解码、协议封装、消息打包、RocketMQ 封装以及一组通用工具包，适合作为服务端基础设施或公共组件库使用。

## 当前模块概览

| 模块 | 说明 | 文档 |
| --- | --- | --- |
| `encoding` | 统一的编解码接口，内置 JSON、MessagePack、Proto、TOML、XML、YAML 实现 | [encoding/README.md](encoding/README.md) |
| `envelope` | 基于 Protobuf 的统一消息信封定义，描述输入输出消息结构 | [envelope/README.md](envelope/README.md) |
| `packet` | 二进制消息打包、解包与心跳处理 | [packet/README.md](packet/README.md) |
| `rmq` | RocketMQ 5.x 生产者与简单消费者封装 | [rmq/README.md](rmq/README.md) |
| `dingtalk` | 钉钉开放能力集成，当前支持群自定义机器人 webhook | [dingtalk/README.md](dingtalk/README.md) |
| `xutil` | 通用工具集合，覆盖配置、日志、ID、网络、缓冲区、任务池等能力 | [xutil/README.md](xutil/README.md) |

## 目录结构

```text
.
├── encoding
├── envelope
├── dingtalk
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

## Odin CLI

安装命令行工具：

```bash
go install github.com/ivy-mobile/odin/cmd/odin@latest
```

`cmd/odin` 是独立 Go module。本地开发时可进入该目录执行 `go install .`；远程发布使用 `cmd/odin/vX.Y.Z` 形式的子模块标签，例如 `cmd/odin/v0.1.0`。

使用默认的 `game-skeleton` 模板创建项目：

```bash
odin new uno --app-id 107
odin new ab-cd --app-id 108
```

项目名只能包含小写英文字母，多个单词使用单个短横线分隔。项目会创建在当前目录下，短横线形式用于目录名和 Go module，下划线形式用于 Go 标识符。

可以通过参数指定其他 Git 模板仓库和分支：

```bash
odin new uno --app-id 107 -r https://example.com/team/game-layout.git
odin new uno --app-id 107 -r git@example.com:team/game-layout.git -b develop
```

也可以通过环境变量设置模板仓库：

```bash
ODIN_LAYOUT_REPO=https://example.com/team/game-layout.git odin new uno --app-id 107
```

模板仓库的选择优先级为 `--repo`、`ODIN_LAYOUT_REPO`、默认模板仓库。未指定 `--branch` 时使用远端默认分支。`--app-id` 必填且必须为正整数；WebSocket 路径按项目名推导，例如 `mono-pink` 对应 `/party-pop/game/mono/pink`。运行环境需要安装 Git，并提前配置好私有仓库所需的 SSH 或 HTTPS 凭据。

所有模板仓库都必须在根目录提供 `.odin-template.yaml`。首版清单版本为 `1`，支持按点路径修改 YAML 标量字段，以及对指定 UTF-8 文件执行带匹配数量校验的精确文本替换。生成结果不会包含该清单。

清单可以使用以下三个内置变量：

- `{{ .Project }}`：命令行传入的项目名，例如 `mono-pink`。
- `{{ .ProjectRoute }}`：将项目名中的短横线替换为 `/`，例如 `mono/pink`。
- `{{ .AppID }}`：`--app-id` 的正整数值。

清单结构示例：

```yaml
version: 1

yaml:
  - file: config/application.yaml
    set:
      - path: application.id
        value: "{{ .AppID }}"
        type: int
      - path: application.ws_path
        value: "/party-pop/game/{{ .ProjectRoute }}"

text:
  - file: api/provider_config_test.go
    replacements:
      - old: '"name: todo-rpc",'
        new: '"name: {{ .Project }}-rpc",'
        count: 1
```

`yaml[].set[].path` 使用点分隔的 mapping key，只能修改已经存在的标量；`type` 支持 `string`（默认）和 `int`。`text` 规则只接受模板根目录内的普通 UTF-8 文件，每条规则的实际匹配数必须与正整数 `count` 完全一致。绝对路径、`..`、`.git`、符号链接和二进制文件都会被拒绝；重复 YAML key、字段不存在或类型不匹配也会使生成失败。清单不能修改 `api/todo.proto`、`api/todo.pb.go` 和 `api/todo.triple.go`。

CLI 生成项目文件后不会自动执行 `go mod tidy` 或 `git init`。

发布 CLI 时使用与独立 module 对应的子目录标签：

```bash
git tag cmd/odin/v0.1.0
git push origin cmd/odin/v0.1.0
```

标签推送后可通过 `go install github.com/ivy-mobile/odin/cmd/odin@latest` 安装该版本。

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
