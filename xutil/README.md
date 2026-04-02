# xutil 模块文档

## 概述

`xutil` 是 Odin 中的通用工具集合，提供了一批可以独立复用的小包，覆盖缓冲区、配置加载、类型转换、日志、ID 生成、网络抽象、并发和随机数等常见基础能力。

## 子包概览

| 包 | 说明 |
| --- | --- |
| `lang` | 简单语言辅助函数 |
| `queue` | 基于 channel 的轻量队列 |
| `xbuffer` | 二进制读写与无拷贝缓冲区，[详见文档](xbuffer/README.md) |
| `xconf` | 基于 Viper 的配置加载与热更新 |
| `xconv` | 常见类型转换与结构扫描 |
| `xfile` | 文件相关工具函数 |
| `xgo` | 带 panic 恢复的 goroutine 封装 |
| `xid` | UUID、ULID、Snowflake、Sonyflake 等 ID 生成能力 |
| `xlog` | 基于 zerolog 的日志封装 |
| `xlog/v2` | 更新版日志抽象接口 |
| `xnet` | 网络客户端/服务端抽象接口 |
| `xnet/tcp` | TCP 客户端与服务端实现 |
| `xnet/ws` | WebSocket 客户端与服务端实现 |
| `xos` | 系统信号处理辅助 |
| `xpool` | 基于 ants 的任务池，[详见文档](xpool/README.md) |
| `xrand` | 随机数与随机字符串工具 |
| `xreflect` | 反射辅助函数 |
| `xtime` | 时间相关工具函数 |
| `xtype` | 通用空值占位定义 |
| `xvalue` | 任意值包装与读取，[详见文档](xvalue/README.md) |

## 常见使用场景

### 配置加载

使用 `xconf.LoadConfigFromFile` 从 `yaml`、`json`、`toml` 等文件加载配置，并可选择是否监听文件变更。

### 任务调度

使用 `xpool` 管理并发任务；如果任务池不可用，内部会回退到 `xgo.Go` 执行。

### 统一日志

使用 `xlog` 或 `xlog/v2` 输出结构化日志，并支持文件输出与日志切割。

### 网络收发

使用 `xnet` 提供的抽象接口，以及 `xnet/tcp`、`xnet/ws` 的具体实现，快速搭建 TCP 或 WebSocket 通信。

### ID 生成

使用 `xid` 统一生成业务主键或消息标识，例如 UUID、ULID、Snowflake 和 Sonyflake。
