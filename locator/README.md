# locator 组件说明

> ⚠️ **注意：Redis 版本需大于 6.0。**

`locator` 是一个基于Redis实现的用于分布式游戏服务中用户节点定位的组件，支持将用户与网关节点（Gate）和游戏节点（Game）进行绑定、解绑及查询。其核心目标是高效、可靠地管理用户与后端节点的映射关系，支持多实例部署和高并发场景。

## 主要功能

- **绑定节点**  
  支持将用户与指定的 Gate 节点或 Game 节点进行绑定，并同步到 Redis 和本地缓存。

- **解绑节点**  
  支持解绑用户与 Gate 节点或 Game 节点的关系，确保数据一致性。

- **节点查询**  
  支持根据用户 ID 查询其当前绑定的 Gate 节点或 Game 节点，优先从本地缓存获取，缓存未命中时自动回源 Redis。

- **缓存一致性**  
  采用本地 LRU 缓存 + Redis 的双层存储结构，结合 singleflight 防止缓存击穿，提升高并发下的性能和一致性。

## 依赖说明

- 依赖 Redis 作为后端存储，支持单机、集群等多种部署方式。
- 本地缓存采用 [hashicorp/golang-lru](https://github.com/hashicorp/golang-lru)。
- 防缓存击穿采用 [golang.org/x/sync/singleflight](https://pkg.go.dev/golang.org/x/sync/singleflight)。

## 主要接口

```go
type Locator interface {
    BindGate(ctx context.Context, uid int64, gateID string) error
    BindGame(ctx context.Context, uid int64, gameName, nodeID string) error
    UnbindGate(ctx context.Context, uid int64, gateID string) error
    UnbindGame(ctx context.Context, uid int64, gameName, nodeID string) error
    GetGateNode(ctx context.Context, uid int64) (string, error)
    GetGameNode(ctx context.Context, uid int64, gameName string) (string, error)
}
```

## 配置项说明

通过可选参数 Option 进行灵活配置：

- `WithAddrs(addrs ...string)`：设置 Redis 地址，默认 `127.0.0.1:6379`
- `WithDB(db int)`：设置 Redis 数据库编号，默认 0
- `WithUsername(username string)`：设置 Redis 用户名
- `WithPassword(password string)`：设置 Redis 密码
- `WithMaxRetries(maxRetries int)`：设置最大重试次数，默认 3
- `WithPrefix(prefix string)`：设置 Redis key 前缀，默认 `ivy`
- `WithMaxCacheSize(size int)`：设置本地 LRU 缓存大小，默认 1000
- `WithClient(client redis.UniversalClient)`：自定义 Redis 客户端
- `WithContext(ctx context.Context)`：自定义上下文

## 初始化示例

```go
import (
    "github.com/redis/go-redis/v9"
    "your_project/pkg/locator"
)

loc := locator.New(
    locator.WithAddrs("127.0.0.1:6379"),
    locator.WithPassword("yourpassword"),
    locator.WithPrefix("mygame"),
    locator.WithMaxCacheSize(2000),
)
```

## 典型使用场景

- 用户登录时，绑定其所在的网关节点（Gate）。
- 用户进入某个游戏房间时，绑定其所在的游戏节点（Game）。
- 用户断开连接或退出游戏时，解绑对应节点。
- 业务服务需要快速定位用户当前所在的节点，实现消息路由或状态同步。

## 数据结构说明

- Gate 节点绑定：`{prefix}:locator:user:{uid}:gate`，类型为 string
- Game 节点绑定：`{prefix}:locator:user:{uid}:game`，类型为 hash，field 为 gameName，value 为 nodeID

## 事件广播

目前绑定/解绑事件预留了广播接口（TODO），可根据业务需求扩展。

---

如需详细用法和二次开发，请参考源码及注释。