# Nacos Registry

`registry/nacos` 是 `github.com/ivy-mobile/odin/registry` 的 Nacos 适配实现，基于 `github.com/nacos-group/nacos-sdk-go/v2` 的 `naming_client.INamingClient` 完成服务注册、注销、查询和监听。

## 功能

- 实现 `registry.Registry` 和 `registry.Watcher` 接口。
- 支持一个 `ServiceInstance` 按多个 endpoint 注册为多个 Nacos 实例。
- 支持按服务名注销实例。
- 支持按服务名查询健康实例。
- 支持通过 Nacos subscribe 监听服务变更。
- 支持 Nacos `group`、`cluster`、`weight` 和默认协议类型配置。
- 注册和查询时会在 Nacos metadata 与 `registry.ServiceInstance` 之间转换 `kind`、`version`、`id`、`weight` 字段。

## 创建 Registry

```go
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

reg := nacos.New(
    client,
    nacos.Group("DEFAULT_GROUP"),
    nacos.Cluster("DEFAULT"),
    nacos.Weight(100),
    nacos.Kind("grpc"),
)
```

完整 import 示例：

```go
import (
    "github.com/ivy-mobile/odin/registry/nacos"
    "github.com/nacos-group/nacos-sdk-go/v2/clients"
    "github.com/nacos-group/nacos-sdk-go/v2/common/constant"
    "github.com/nacos-group/nacos-sdk-go/v2/vo"
)
```

## 配置项

`New` 使用以下默认配置：

| 配置 | 默认值 | 说明 |
| --- | --- | --- |
| `Group` | `constant.DEFAULT_GROUP` | Nacos group，通常是 `DEFAULT_GROUP` |
| `Cluster` | `DEFAULT` | Nacos cluster |
| `Weight` | `100` | 注册时的默认实例权重，查询映射时也作为兜底权重 |
| `Kind` | `ws` | 查询或监听结果没有 `kind` metadata 时使用的 endpoint 协议 |

## 服务名约定

当前实现直接使用 `ServiceInstance.Name` 作为 Nacos `ServiceName`，不会根据 endpoint 协议自动追加后缀。

例如注册时：

```go
svc := &registry.ServiceInstance{
    ID:        "node-1",
    Name:      "room.grpc",
    Version:   "v1.0.0",
    Endpoints: []string{"grpc://127.0.0.1:9000"},
}
```

后续查询和监听也应使用同一个服务名：

```go
instances, err := reg.GetService(ctx, "room.grpc")
watcher, err := reg.Watch(ctx, "room.grpc")
```

## 注册

`Register` 会遍历 `ServiceInstance.Endpoints`，每个 endpoint 调用一次 Nacos `RegisterInstance`：

```go
svc := &registry.ServiceInstance{
    ID:        "node-1",
    Name:      "room.grpc",
    Version:   "v1.0.0",
    Metadata:  map[string]string{"idc": "shanghai", "weight": "12.5"},
    Endpoints: []string{"grpc://127.0.0.1:9000"},
}

if err := reg.Register(context.Background(), svc); err != nil {
    panic(err)
}
```

注册行为：

- `ServiceInstance.Name` 为空时返回 `ErrServiceInstanceNameEmpty`。
- endpoint 必须是带 host 和 port 的合法 URL，例如 `grpc://127.0.0.1:9000`。
- Nacos `ServiceName` 使用 `ServiceInstance.Name`。
- Nacos `Ip`、`Port` 从 endpoint 的 host 中解析。
- Nacos `ClusterName`、`GroupName` 使用当前配置。
- Nacos 实例固定以 `Enable=true`、`Healthy=true`、`Ephemeral=true` 注册。
- metadata 为空时写入 `kind` 和 `version`。
- metadata 非空时会复制原 metadata，并覆盖写入 `kind`、`version`、`id`，不会修改入参 map。
- metadata 非空且包含可解析的 `weight` 时，优先使用该权重；否则使用 `Weight` 配置值。

## 注销

`Deregister` 同样会遍历 `ServiceInstance.Endpoints`，每个 endpoint 调用一次 Nacos `DeregisterInstance`：

```go
if err := reg.Deregister(context.Background(), svc); err != nil {
    panic(err)
}
```

注销行为：

- Nacos `ServiceName` 使用 `ServiceInstance.Name`。
- Nacos `Ip`、`Port` 从 endpoint 的 host 中解析。
- Nacos `GroupName`、`Cluster` 使用当前配置。
- Nacos 实例按 `Ephemeral=true` 注销。
- endpoint 解析失败、端口解析失败、SDK 返回错误或注销结果为 false 时都会返回错误。

## 查询

`GetService` 使用 Nacos `SelectInstances` 查询健康实例：

```go
instances, err := reg.GetService(context.Background(), "room.grpc")
if err != nil {
    panic(err)
}
```

查询行为：

- `ServiceName` 使用传入的 `serviceName`。
- `GroupName` 使用当前配置。
- `HealthyOnly=true`。
- 当前查询没有传入 cluster 过滤。

Nacos `model.Instance` 会被映射成 `registry.ServiceInstance`：

- `ID` 来自 `InstanceId`。
- `Name` 来自 Nacos 返回的 `ServiceName`。
- `Version` 来自 metadata `version`。
- `Metadata` 会复制 Nacos metadata，避免修改 SDK 返回的原始 map。
- endpoint 形如 `{kind}://{ip}:{port}`。
- endpoint 的 `kind` 优先来自 metadata `kind`，没有时使用 `Kind` 配置。
- metadata `weight` 会被补充为实例权重向上取整后的字符串。
- Nacos 实例权重大于 0 时使用实例权重，否则使用 `Weight` 配置值。

## Watch

`Watch` 会订阅指定服务，并返回 `registry.Watcher`：

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

监听行为：

- 创建 watcher 时会调用 Nacos `Subscribe`。
- 订阅参数包含 `ServiceName`、`GroupName` 和当前配置的单个 `Cluster`。
- 创建后会主动触发一次 `Next`，用于拉取当前服务列表。
- Nacos subscribe callback 触发后，下一次 `Next` 会重新调用 Nacos `GetService`。
- `Next` 在 context 取消或超时时返回对应错误。
- `Stop` 会先调用 Nacos `Unsubscribe`，再取消内部 context。

## 测试

推荐优先运行不依赖真实 Nacos 的单元测试：

```powershell
go test ./registry/nacos -run "TestRegistry_(RegisterBuildsNacosParams|RegisterUsesServiceNameAsNacosServiceName|RegisterReturnsErrorWhenNacosReturnsFalse|DeregisterBuildsNacosParams|DeregisterReturnsErrorWhenNacosReturnsFalse|GetServiceMapsInstances|WatchMapsServiceAndUnsubscribes)$" -count=1
```

完整包测试中包含连接真实 Nacos 的集成测试，依赖 `registry_test.go` 中配置的 Nacos 地址和 namespace：

```powershell
go test ./registry/nacos -count=1
```

如果本地无法访问对应 Nacos 服务，集成测试可能失败。

## 注意事项

- `Register` 和 `Deregister` 当前没有使用传入的 `context.Context`。
- `Register` 会校验 `ServiceInstance.Name` 非空，`Deregister` 不做该校验。
- endpoint 的 host 必须能被 `net.SplitHostPort` 解析，推荐始终写成 `scheme://host:port`。
- `ServiceInstance.Metadata["weight"]` 只有在 metadata 非空时才会参与注册权重解析。
- `GetService` 只查询健康实例；`Watch.Next` 通过 `GetService` 拉取当前订阅服务的实例列表。
