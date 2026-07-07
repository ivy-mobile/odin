# DingTalk 模块文档

## 概述

`dingtalk` 模块用于集成钉钉开放能力。当前已支持群自定义机器人 webhook，后续可以在该目录下继续扩展应用机器人、事件回调、审批、通讯录等能力。

## 当前子模块

| 子模块 | 说明 | 文档 |
| --- | --- | --- |
| `webhook` | 群自定义机器人 webhook 消息发送，支持 text、link、markdown、actionCard、feedCard 与加签 | [webhook/README.md](webhook/README.md) |

## 目录规划

```text
dingtalk/
├── internal/     # 钉钉能力内部共享实现
└── webhook/      # 群自定义机器人 webhook
```

## 设计约定

- 根目录 `dingtalk` 不直接暴露具体业务 API，具体能力通过子包提供。
- `internal` 目录只放子包共享实现，避免过早扩大公开 API 面。
- 不在错误信息中暴露完整 webhook，降低 `access_token` 泄露风险。
