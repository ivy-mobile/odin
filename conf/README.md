# pkg/conf - 统一配置加载包

通用配置加载包，所有游戏服务共享，提供统一的配置结构和加载方式。

## 特性

- 系统配置只读，结构统一（Application、Redis、Log 等）
- 业务配置支持本地文件或 Nacos（Nacos 优先，本地兜底）
- 支持 YAML / JSON / TOML 三种格式，按文件扩展名自动识别
- 系统配置支持环境变量占位符，可设置默认值
- 业务配置支持热更新监听

## 快速开始

### 仅加载系统配置

```go
if err := conf.Load("config/config.yaml"); err != nil {
    panic(err)
}

fmt.Println(conf.Application().ID)
fmt.Println(conf.Redis().Addr)
```

### 同时加载业务配置

```go
var business BusinessConfig
closeFn, err := conf.Load(
    "config/config.yaml",
    conf.WithBusiness("config/business.yaml", &business, true),
)
if err != nil {
    panic(err)
}
defer closeFn()
```

## 加载规则

```
系统配置（本地文件）       业务配置
┌─────────────┐        ┌──────────────────────────┐
│ config.yaml │        │ ConfigCenter 有效？       │
│             │        │   ├─ 是 → 从 Nacos 获取   │
│  - 只读     │        │   └─ 否 → 从本地文件获取  │
│  - 支持环境变量        └──────────────────────────┘
│  - 必填     │
└─────────────┘
```

## 环境变量占位符

系统配置文件支持 `${VAR}` 和 `${VAR:-default}` 语法，仅对系统配置生效，业务配置不做替换。

### 语法说明

| 语法 | 说明 | 示例 |
|------|------|------|
| `${VAR}` | 使用环境变量 `VAR` 的值（等同于 `$VAR`） | `${PORT}` |
| `${VAR-$DEFAULT}` | 环境变量 `VAR` 未设置时，使用 `$DEFAULT` 的值 | `${PORT-8088}` |
| `${VAR:-$DEFAULT}` | 环境变量 `VAR` 未设置或为空时，使用 `$DEFAULT` 的值 | `${PORT:-8088}` |
| `${VAR=$DEFAULT}` | 环境变量 `VAR` 未设置时，使用 `$DEFAULT` 的值 | `${PORT=8088}` |
| `${VAR:=$DEFAULT}` | 环境变量 `VAR` 未设置或为空时，使用 `$DEFAULT` 的值 | `${PORT:=8088}` |
| `${VAR+$OTHER}` | 环境变量 `VAR` 已设置时，使用 `$OTHER` 的值，否则为空字符串 | `${PORT+9090}` |
| `${VAR:+$OTHER}` | 环境变量 `VAR` 已设置且非空时，使用 `$OTHER` 的值，否则为空字符串 | `${PORT:+9090}` |
| `$$VAR` | 转义，结果为字面量 `$VAR` | `$$PORT` → `$PORT` |

> **常用推荐：** `${VAR:-default}` — 环境变量存在用环境变量，不存在或为空用默认值。

### 示例

配置文件 `config.yaml`：

```yaml
application:
  id: ${GAME_ID:-105}
  name: ${SERVER_NAME:-sword-ball}
  port: ${PORT:-8088}

redis:
  addr: ${REDIS_ADDR:-127.0.0.1:6379}
  password: ${REDIS_PASSWORD:-admin123}

config_center:
  ip_addr: ${NACOS_IP:-127.0.0.1}
  port: ${NACOS_PORT:-8848}
  namespace: ${NACOS_NAMESPACE:-dev}
```

**效果：**

| 环境变量 | 值 | 配置结果 |
|----------|-----|---------|
| `GAME_ID=200` | 使用环境变量 | `id: 200` |
| `SERVER_NAME` 未设置 | 使用默认值 | `name: sword-ball` |
| `PORT=9090` | 使用环境变量 | `port: 9090` |
| `REDIS_PASSWORD` 未设置 | 使用默认值 | `password: admin123` |

### 不同格式写法

**YAML：**

```yaml
redis:
  addr: ${REDIS_ADDR:-127.0.0.1:6379}
```

**JSON：**

```json
{
  "redis": {
    "addr": "${REDIS_ADDR:-127.0.0.1:6379}"
  }
}
```

**TOML：**

```toml
[redis]
addr = "${REDIS_ADDR:-127.0.0.1:6379}"
```

### 注意事项

- 环境变量占位符仅对**系统配置**生效，Nacos 业务配置不做替换
- JSON / TOML 对类型要求严格，`int` 类型字段不能使用占位符（替换后为字符串），`string` 类型字段可以正常使用

## 文件格式

根据文件扩展名自动识别格式：

| 扩展名 | 格式 |
|--------|------|
| `.yaml` / `.yml` | YAML |
| `.json` | JSON |
| `.toml` | TOML |
| 其他 | 默认 YAML |

## 系统配置结构

```yaml
application:       # 应用信息
  id: 105
  name: sword-ball
  env: dev
  ws_path: /game
  port: 8088
  pprof_port: 6060

config_center:     # 配置中心（Nacos）
  ip_addr: 127.0.0.1
  port: 8848
  data_id: business-config
  group: DEFAULT_GROUP
  namespace: dev

registry:          # 注册中心（Nacos）
  ip_addr: 127.0.0.1
  port: 8848
  namespace: dev

log:               # 日志配置
  level: debug
  mode: console
  file:
    filename: ./logs/app.log
    max_size: 10

redis:             # Redis
  addr: 127.0.0.1:6379
  password: admin123
  db: 0

mq:                # 消息队列
  endpoint: 127.0.0.1:9876
  group: consumer-group

micros:            # 微服务
  game_center:
    filters: logging
  user:
    filters: logging
  room:
    filters: logging
```

## API

### 加载配置

```go
closeFn, err := conf.Load(filename string, opts ...Option)
```

- `filename`：系统配置文件路径，必填
- `opts`：可选参数
- 返回 `closeFn` 用于释放资源（Nacos 客户端等）

### 业务配置选项

```go
conf.WithBusiness(filename string, target any, watch bool) Option
```

- `filename`：本地业务配置文件路径（Nacos 无效时使用）
- `target`：业务配置结构体指针
- `watch`：是否监听配置变化

### 读取系统配置

```go
conf.Application()  // 应用配置
conf.ConfigCenter() // 配置中心
conf.Registry()     // 注册中心
conf.Log()          // 日志配置
conf.Redis()        // Redis 配置
conf.MQ()           // 消息队列配置
conf.Micros()       // 微服务配置
```

## 完整示例

```go
package main

import (
    "fmt"
    "sword-ball/pkg/conf"
)

type BusinessConfig struct {
    Room *struct {
        SettleTime string `yaml:"settle_time"`
    } `yaml:"room"`
}

func main() {
    var business BusinessConfig
    closeFn, err := conf.Load(
        "config/config.yaml",
        conf.WithBusiness("config/business.yaml", &business, true),
    )
    if err != nil {
        panic(err)
    }
    defer closeFn()

    fmt.Printf("Game: %s (ID: %d)\n", conf.Application().Name, conf.Application().ID)
    fmt.Printf("Redis: %s\n", conf.Redis().Addr)
    fmt.Printf("Room settle time: %s\n", business.Room.SettleTime)
}
```

## 示例文件

- [example.yaml](example.yaml)
- [example.json](example.json)
- [example.toml](example.toml)
