# Nacos Registry

`pkg/registry/nacos` 是 `sword-ball/pkg/registry` 的 Nacos 适配实现，用于把服务实例注册到 Nacos，并通过 Nacos 查询或监听服务实例变更。

## 功能

- 实现 `registry.Registry` 接口。
- 支持服务注册和注销。
- 支持按服务名查询健康实例。
- 支持通过 Watch 监听服务实例变更。
- 支持 Nacos `group`、`cluster`、`weight` 和默认协议类型配置。
- 从 endpoint 协议中写入实例 metadata 的 `kind`，从 `ServiceInstance.Version` 写入 `version`。

## 服务名约定

当前实现会使用 `ServiceInstance.Name` 作为 Nacos 的 `ServiceName` 注册和注销，不会自动追加 endpoint 协议后缀。

例如：

```go
svc := &registry.ServiceInstance{
    Name:      "room.grpc",
    Version:   "v1.0.0",
    Endpoints: []string{"grpc://127.0.0.1:9000"},
}
```

如果希望按 `room.grpc` 查询或监听，注册时也应把 `Name` 设置为 `room.grpc`。

## 配置项

`New` 默认配置：

| 配置 | 默认值             | 说明 |
| --- |-----------------| --- |
| `Group` | `DEFAULT_GROUP` | Nacos group |
| `Cluster` | `DEFAULT`       | Nacos cluster |
| `Weight` | `100`           | 默认实例权重 |
| `Kind` | `ws`            | 查询或监听结果没有 `kind` metadata 时使用的协议 |

## 使用示例

```go
package main

import (
    "context"

    "sword-ball/pkg/registry"
    nregistry "sword-ball/pkg/registry/nacos"

    "github.com/nacos-group/nacos-sdk-go/v2/clients"
    "github.com/nacos-group/nacos-sdk-go/v2/common/constant"
    "github.com/nacos-group/nacos-sdk-go/v2/vo"
)

func main() {
    client, err := clients.NewNamingClient(vo.NacosClientParam{
        ClientConfig: &constant.ClientConfig{
            NamespaceId:         "public",
            TimeoutMs:           5000,
            NotLoadCacheAtStart: true,
            LogDir:              "/tmp/nacos/log",
            CacheDir:            "/tmp/nacos/cache",
            LogLevel:            "info",
        },
        ServerConfigs: []constant.ServerConfig{
            *constant.NewServerConfig("127.0.0.1", 8848),
        },
    })
    if err != nil {
        panic(err)
    }

    reg := nregistry.New(
        client,
        nregistry.Group("DEFAULT_GROUP"),
        nregistry.Cluster("DEFAULT"),
        nregistry.Weight(100),
    )

    svc := &registry.ServiceInstance{
        ID:        "node-1",
        Name:      "room.grpc",
        Version:   "v1.0.0",
        Metadata:  map[string]string{"idc": "shanghai"},
        Endpoints: []string{"grpc://127.0.0.1:9000"},
    }

    ctx := context.Background()
    if err := reg.Register(ctx, svc); err != nil {
        panic(err)
    }
    defer reg.Deregister(ctx, svc)

    instances, err := reg.GetService(ctx, "room.grpc")
    if err != nil {
        panic(err)
    }
    _ = instances
}
```

## Watch 示例

```go
watcher, err := reg.Watch(context.Background(), "room.grpc")
if err != nil {
    panic(err)
}
defer watcher.Stop()

instances, err := watcher.Next()
if err != nil {
    panic(err)
}
_ = instances
```

`Watch` 创建后会先触发一次查询，后续由 Nacos subscribe callback 驱动再次查询。

## Metadata 和权重

注册时：

- 如果 `ServiceInstance.Metadata` 为空，会写入 `kind` 和 `version`。
- 如果 `ServiceInstance.Metadata` 不为空，会复制原 metadata，并覆盖写入 `kind`、`version`、`id`。
- 如果 metadata 中包含 `weight`，且能解析为 float，则优先使用该权重；否则使用 `Weight` 的默认权重。

查询时：

- endpoint 的协议优先使用实例 metadata 中的 `kind`。
- 如果没有 `kind`，使用 `Kind`。
- 返回 metadata 会补充 `weight`，其值为 Nacos 实例权重向上取整后的字符串。

## 测试

推荐优先运行不依赖真实 Nacos 的单元测试：

```powershell
go test ./pkg/registry/nacos -run "TestRegistry_(RegisterBuildsNacosParams|RegisterUsesServiceNameAsNacosServiceName|DeregisterBuildsNacosParams|GetServiceMapsInstances|WatchMapsServiceAndUnsubscribes)$" -count=1
```

完整包测试中包含连接真实 Nacos 的集成测试，依赖 `registry_test.go` 中配置的 Nacos 地址和 namespace：

```powershell
go test ./pkg/registry/nacos -count=1
```

如果本地无法访问对应 Nacos 服务，集成测试可能失败。

## 注意事项

- `ServiceInstance.Name` 不能为空，否则返回 `ErrServiceInstanceNameEmpty`。
- endpoint 必须是合法 URL，并且 host 中必须包含可解析端口，例如 `grpc://127.0.0.1:9000`。
- `Register` 和 `Deregister` 都会遍历 `ServiceInstance.Endpoints`，每个 endpoint 对应一个 Nacos 实例。
- `GetService` 只查询健康实例，内部使用 `SelectInstances` 且 `HealthyOnly=true`。
