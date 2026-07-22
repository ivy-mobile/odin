# Odin CLI

Odin CLI 用于从版本化 Git 模板创建 Go 项目。当前提供 `new` 命令和版本查询参数。

## 安装

```bash
go install github.com/ivy-mobile/odin/cmd/odin@latest
```

`cmd/odin` 是独立 Go module。本地开发时可在本目录执行：

```bash
go install .
```

查看当前版本：

```bash
odin -v
# 或
odin --version
```

通过版本标签执行 `go install` 时，Odin 会读取构建信息中的 module 版本；本地开发构建显示 `dev`。

## 创建项目

```text
odin new <project> --id <positive-integer> [-r|--repo <git-source>] [-b|--branch <branch>]
```

使用默认 `game-skeleton` 模板：

```bash
odin new uno --id 107
odin new mono-pink --id 108
```

项目名必须匹配 `^[a-z]+(?:-[a-z]+)*$`，即只允许小写英文字母以及分隔单词的单个短横线。项目会创建在当前目录下，短横线形式用于目录名和 Go module，下划线形式用于 Go 标识符。

`--id` 必填且必须为正整数。WebSocket 路径根据项目名推导：

- `uno` → `/party-pop/game/uno`
- `mono-pink` → `/party-pop/game/mono/pink`

指定其他 Git 模板仓库和分支：

```bash
odin new uno --id 107 -r https://example.com/team/game-layout.git
odin new uno --id 107 -r git@example.com:team/game-layout.git -b develop
```

也可以通过环境变量设置模板仓库：

```bash
ODIN_LAYOUT_REPO=https://example.com/team/game-layout.git odin new uno --id 107
```

模板仓库选择优先级为 `--repo`、`ODIN_LAYOUT_REPO`、默认仓库 `https://cnb.cool/ivy-party-pop/backend/go/game-skeleton`。未指定 `--branch` 时使用远端默认分支。运行环境需要安装 Git，并提前配置私有仓库所需的 SSH 或 HTTPS 凭据。

生成过程不会自动执行 `go mod tidy`、`git init`、模板脚本或协议生成命令。

## 模板清单

每个模板仓库必须在根目录提供 `.odin-template.yaml`。当前只支持 `version: 1`，生成结果不会包含该清单。

清单可以使用三个内置变量：

- `{{ .Project }}`：原始项目名，例如 `mono-pink`。
- `{{ .ProjectRoute }}`：将短横线替换为 `/`，例如 `mono/pink`。
- `{{ .AppID }}`：`--id` 的正整数值。

清单示例：

```yaml
version: 1
project_readme: .odin-project-readme.md

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

`yaml[].set[].path` 使用点分隔的 mapping key，只能修改已经存在的标量。`type` 支持 `string`（默认）和 `int`，修改时保留字段顺序与注释。

`text` 规则只接受模板根目录内的普通 UTF-8 文件。每条规则的实际匹配数必须与正整数 `count` 完全一致，避免模板内容漂移后静默生成错误项目。

绝对路径、`..`、`.git`、符号链接和二进制编辑目标都会被拒绝；重复 YAML key、字段不存在或类型不匹配也会使生成失败。清单不能修改 `api/todo.proto`、`api/todo.pb.go` 和 `api/todo.triple.go`。

`project_readme` 可以指定生成项目使用的精简 README 源文件。该文件支持相同的三个内置变量，渲染后写入项目根目录的 `README.md`；源文件和模板仓库原有 README 都不会进入生成结果。

所有转换都在临时目录完成。目标已存在或任意步骤失败时，Odin 不会覆盖目标，也不会留下半成品。

## 本地验证

```bash
go test ./...
go vet ./...
golangci-lint run ./...
```

## 发布

CLI 使用 GoReleaser Community 交叉编译以下平台：

- Windows `amd64`、`arm64`
- Linux `amd64`、`arm64`
- macOS `amd64`、`arm64`

Linux 和 macOS 产物为 `tar.gz`，Windows 产物为 `zip`。每次发布还会生成 `checksums.txt`。

### 本地快照

安装与 CI 相同版本的 GoReleaser：

```bash
go install github.com/goreleaser/goreleaser/v2@v2.17.0
```

在仓库根目录执行配置检查和快照构建：

```powershell
$env:RELEASE_VERSION = "0.1.0-snapshot"
goreleaser check --config cmd/odin/.goreleaser.yaml
goreleaser release --snapshot --clean --config cmd/odin/.goreleaser.yaml
```

产物生成在仓库根目录的 `dist/`。解压当前平台的压缩包后执行：

```powershell
./odin -v
```

预期输出 `odin version v0.1.0-snapshot`。校验下载文件：

```bash
cd dist
sha256sum -c checksums.txt
```

### 正式发布

独立 module 使用稳定 SemVer 格式的 `cmd/odin/vX.Y.Z` 标签：

```bash
git tag cmd/odin/v0.1.0
git push origin cmd/odin/v0.1.0
```

GitHub Actions 只监听 `cmd/odin/v*`，根库的 `v*` 标签不会触发 CLI 发布。工作流会执行测试和静态检查，生成 6 个压缩包与 `checksums.txt`，然后将它们上传到对应标签的 GitHub Release。

标签推送后，可通过以下命令安装：

```bash
go install github.com/ivy-mobile/odin/cmd/odin@latest
```
