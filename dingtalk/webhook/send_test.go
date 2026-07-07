package webhook

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/ivy-mobile/odin/dingtalk/internal/signature"
)

func TestSendTextWithSecret(t *testing.T) {
	const (
		secret    = "secret"
		timestamp = int64(1234567890)
	)

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			t.Fatalf("method = %s, want %s", r.Method, http.MethodPost)
		}
		if got := r.URL.Query().Get("access_token"); got != "token" {
			t.Fatalf("access_token = %q, want token", got)
		}
		if got := r.URL.Query().Get("timestamp"); got != "1234567890" {
			t.Fatalf("timestamp = %q, want 1234567890", got)
		}
		if got, want := r.URL.Query().Get("sign"), signature.Sign(timestamp, secret); got != want {
			t.Fatalf("sign = %q, want %q", got, want)
		}
		if got := r.Header.Get("Content-Type"); !strings.HasPrefix(got, "application/json") {
			t.Fatalf("Content-Type = %q, want application/json", got)
		}

		var msg Message
		if err := json.NewDecoder(r.Body).Decode(&msg); err != nil {
			t.Fatal(err)
		}
		if msg.MsgType != MsgTypeText || msg.Text == nil || msg.Text.Content != "告警: 服务启动成功" {
			t.Fatalf("unexpected message: %+v", msg)
		}
		if msg.At == nil || !msg.At.IsAtAll {
			t.Fatalf("unexpected at: %+v", msg.At)
		}

		_, _ = w.Write([]byte(`{"errcode":0,"errmsg":"ok"}`))
	}))
	defer server.Close()

	err := SendText(
		context.Background(),
		server.URL+"/robot/send?access_token=token",
		"告警: 服务启动成功",
		WithSecret(secret),
		withClock(func() time.Time { return time.UnixMilli(timestamp) }),
		AtAll(),
	)
	if err != nil {
		t.Fatal(err)
	}
}

func TestSendAPIErr(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		_, _ = w.Write([]byte(`{"errcode":310000,"errmsg":"keywords not in content"}`))
	}))
	defer server.Close()

	resp, err := Send(context.Background(), server.URL, NewText("hello"))
	if resp == nil || resp.ErrCode != 310000 {
		t.Fatalf("response = %+v, want errcode 310000", resp)
	}
	var apiErr *APIError
	if !errors.As(err, &apiErr) {
		t.Fatalf("error = %v, want APIError", err)
	}
	if apiErr.Code != 310000 {
		t.Fatalf("apiErr.Code = %d, want 310000", apiErr.Code)
	}
}

func TestSendHTTPError(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		http.Error(w, "bad gateway", http.StatusBadGateway)
	}))
	defer server.Close()

	_, err := Send(context.Background(), server.URL, NewText("hello"))
	var httpErr *HTTPError
	if !errors.As(err, &httpErr) {
		t.Fatalf("error = %v, want HTTPError", err)
	}
	if httpErr.StatusCode != http.StatusBadGateway {
		t.Fatalf("status = %d, want %d", httpErr.StatusCode, http.StatusBadGateway)
	}
	if !strings.Contains(string(httpErr.Body), "bad gateway") {
		t.Fatalf("body = %q, want bad gateway", httpErr.Body)
	}
}

func TestSendMarkdownWithAtMobiles(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var msg Message
		if err := json.NewDecoder(r.Body).Decode(&msg); err != nil {
			t.Fatal(err)
		}
		if msg.MsgType != MsgTypeMarkdown || msg.Markdown == nil {
			t.Fatalf("unexpected message: %+v", msg)
		}
		if msg.At == nil || len(msg.At.AtMobiles) != 1 || msg.At.AtMobiles[0] != "13800138000" {
			t.Fatalf("unexpected at: %+v", msg.At)
		}
		_, _ = w.Write([]byte(`{"errcode":0,"errmsg":"ok"}`))
	}))
	defer server.Close()

	err := SendMarkdown(
		context.Background(),
		server.URL,
		"告警通知",
		"### CPU 使用率过高",
		AtMobiles("13800138000"),
	)
	if err != nil {
		t.Fatal(err)
	}
}

func TestSendLink(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		_, _ = w.Write([]byte(`{"errcode":0,"errmsg":"ok"}`))
	}))
	defer server.Close()

	err := SendLink(context.Background(), server.URL, "标题", "正文", "https://example.com", "")
	if err != nil {
		t.Fatal(err)
	}
}

func TestSendInvalidWebhook(t *testing.T) {
	err := SendText(context.Background(), "", "hello")
	if !errors.Is(err, ErrWebhookEmpty) {
		t.Fatalf("error = %v, want ErrWebhookEmpty", err)
	}
}
