# DingTalk Webhook 模块文档

## 概述

`dingtalk/webhook` 提供钉钉群自定义机器人 webhook 消息发送能力，支持普通 webhook 与加签 webhook。

支持的发送函数：

| 函数 | 说明 |
| --- | --- |
| `SendText` | 发送 text 消息 |
| `SendMarkdown` | 发送 markdown 消息 |
| `SendLink` | 发送 link 消息 |
| `SendSingleActionCard` | 发送单按钮 actionCard 消息 |
| `SendActionCard` | 发送独立跳转按钮 actionCard 消息 |
| `SendFeedCard` | 发送 feedCard 消息 |

## 快速开始

```go
package main

import (
	"context"
	"fmt"

	"github.com/ivy-mobile/odin/dingtalk/webhook"
)

func main() {
	url := "https://oapi.dingtalk.com/robot/send?access_token=xxx"

	if err := webhook.SendText(
		context.Background(),
		url,
		"告警: 服务启动成功",
		webhook.WithSecret("SECxxx"),
		webhook.AtAll(),
	); err != nil {
		fmt.Println("发送失败:", err)
		return
	}
}
```

## Markdown 消息

```go
err := webhook.SendMarkdown(
	context.Background(),
	url,
	"告警通知",
	"### CPU 使用率过高\n> 当前值: 92%",
	webhook.WithSecret("SECxxx"),
	webhook.AtMobiles("13800138000"),
)
```

## Link 消息

```go
err := webhook.SendLink(
	context.Background(),
	url,
	"发布完成",
	"服务已发布完成，点击查看详情",
	"https://example.com/releases/1",
	"https://example.com/release.png",
	webhook.WithSecret("SECxxx"),
)
```

## ActionCard 消息

```go
err := webhook.SendActionCard(
	context.Background(),
	url,
	"发布完成",
	"### 服务发布完成\n点击查看详情",
	[]webhook.ActionCardButton{
		{Title: "查看发布单", ActionURL: "https://example.com/releases/1"},
	},
	webhook.BtnHorizontal,
	webhook.WithSecret("SECxxx"),
)
```

## 通用发送

如果需要先构造消息体，也可以使用 `Send`：

```go
msg := webhook.NewFeedCard(
	webhook.FeedCardLink{
		Title:      "发布完成",
		MessageURL: "https://example.com/releases/1",
		PicURL:     "https://example.com/release.png",
	},
)

resp, err := webhook.Send(context.Background(), url, msg, webhook.WithSecret("SECxxx"))
```

## 可选配置

### 加签

使用 `WithSecret` 后，发送函数会自动追加 `timestamp` 和 `sign` 参数：

```go
err := webhook.SendText(
	context.Background(),
	url,
	"告警: 服务启动成功",
	webhook.WithSecret("SECxxx"),
)
```

### @ 用户

`text` 和 `markdown` 消息支持 @ 用户：

```go
err := webhook.SendMarkdown(
	context.Background(),
	url,
	"告警通知",
	"### CPU 使用率过高",
	webhook.AtMobiles("13800138000"),
	webhook.AtUserIDs("user-id"),
)
```

### 超时时间

```go
err := webhook.SendText(
	context.Background(),
	url,
	"告警: 服务启动成功",
	webhook.WithTimeout(3*time.Second),
)
```

### 自定义 HTTP 客户端

```go
httpClient := &http.Client{Timeout: 3 * time.Second}
err := webhook.SendText(
	context.Background(),
	url,
	"告警: 服务启动成功",
	webhook.WithHTTPClient(httpClient),
)
```

## 错误处理

- `APIError`: 钉钉返回 `errcode != 0`
- `HTTPError`: HTTP 状态码不是 2xx
- 参数错误会返回预定义错误，例如 `ErrWebhookEmpty`、`ErrMessageContentEmpty`
