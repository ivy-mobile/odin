package webhook

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"
	"time"

	"github.com/stretchr/testify/require"

	"github.com/ivy-mobile/odin/dingtalk/internal/signature"
)

func TestSendTextWithSecret(t *testing.T) {
	const (
		secret    = "secret"
		timestamp = int64(1234567890)
	)

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		require.Equal(t, http.MethodPost, r.Method)
		require.Equal(t, "token", r.URL.Query().Get("access_token"))
		require.Equal(t, "1234567890", r.URL.Query().Get("timestamp"))
		require.Equal(t, signature.Sign(timestamp, secret), r.URL.Query().Get("sign"))
		require.Contains(t, r.Header.Get("Content-Type"), "application/json")

		var msg Message
		require.NoError(t, json.NewDecoder(r.Body).Decode(&msg))
		require.Equal(t, MsgTypeText, msg.MsgType)
		require.NotNil(t, msg.Text)
		require.Equal(t, "告警: 服务启动成功", msg.Text.Content)
		require.NotNil(t, msg.At)
		require.True(t, msg.At.IsAtAll)

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
	require.NoError(t, err)
}

func TestSendAPIErr(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		_, _ = w.Write([]byte(`{"errcode":310000,"errmsg":"keywords not in content"}`))
	}))
	defer server.Close()

	resp, err := Send(context.Background(), server.URL, NewText("hello"))
	require.NotNil(t, resp)
	require.Equal(t, 310000, resp.ErrCode)

	var apiErr *APIError
	require.ErrorAs(t, err, &apiErr)
	require.Equal(t, 310000, apiErr.Code)
}

func TestErrorStrings(t *testing.T) {
	require.Contains(t, (&APIError{Code: 1, Message: "bad"}).Error(), "code=1")
	require.Contains(t, (&HTTPError{StatusCode: http.StatusBadGateway}).Error(), "502")
}

func TestSendHTTPError(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		http.Error(w, "bad gateway", http.StatusBadGateway)
	}))
	defer server.Close()

	_, err := Send(context.Background(), server.URL, NewText("hello"))
	var httpErr *HTTPError
	require.ErrorAs(t, err, &httpErr)
	require.Equal(t, http.StatusBadGateway, httpErr.StatusCode)
	require.Contains(t, string(httpErr.Body), "bad gateway")
}

func TestSendDecodeResponseError(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		_, _ = w.Write([]byte(`not-json`))
	}))
	defer server.Close()

	_, err := Send(context.Background(), server.URL, NewText("hello"))
	require.Error(t, err)
	require.Contains(t, err.Error(), "decode response")
}

func TestSendReadResponseError(t *testing.T) {
	httpClient := &http.Client{Transport: roundTripFunc(func(*http.Request) (*http.Response, error) {
		return &http.Response{
			StatusCode: http.StatusOK,
			Body:       errReadCloser{},
			Header:     make(http.Header),
		}, nil
	})}

	_, err := Send(context.Background(), "https://example.com/robot/send", NewText("hello"), WithHTTPClient(httpClient))
	require.Error(t, err)
	require.Contains(t, err.Error(), "read response")
}

func TestSendHTTPClientError(t *testing.T) {
	httpClient := &http.Client{Transport: roundTripFunc(func(*http.Request) (*http.Response, error) {
		return nil, errors.New("network down")
	})}

	_, err := Send(context.Background(), "https://example.com/robot/send", NewText("hello"), WithHTTPClient(httpClient))
	require.Error(t, err)
	require.Contains(t, err.Error(), "network down")
}

func TestSanitizeRequestErrorPlainError(t *testing.T) {
	err := sanitizeRequestError(errors.New("plain"))
	require.Error(t, err)
	require.Contains(t, err.Error(), "plain")
}

func TestSendMarkdownWithAtMobiles(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var msg Message
		require.NoError(t, json.NewDecoder(r.Body).Decode(&msg))
		require.Equal(t, MsgTypeMarkdown, msg.MsgType)
		require.NotNil(t, msg.Markdown)
		require.NotNil(t, msg.At)
		require.Len(t, msg.At.AtMobiles, 1)
		require.Equal(t, "13800138000", msg.At.AtMobiles[0])
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
	require.NoError(t, err)
}

func TestSendWithNilContextAndTimeout(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		_, _ = w.Write([]byte(`{"errcode":0,"errmsg":"ok"}`))
	}))
	defer server.Close()

	resp, err := Send(nil, server.URL, NewText("hello"), WithTimeout(time.Second))
	require.NoError(t, err)
	require.Equal(t, 0, resp.ErrCode)
}

func TestSendLink(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		_, _ = w.Write([]byte(`{"errcode":0,"errmsg":"ok"}`))
	}))
	defer server.Close()

	err := SendLink(context.Background(), server.URL, "标题", "正文", "https://example.com", "")
	require.NoError(t, err)
}

func TestSendActionMessages(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		_, _ = w.Write([]byte(`{"errcode":0,"errmsg":"ok"}`))
	}))
	defer server.Close()

	err := SendSingleActionCard(context.Background(), server.URL, "标题", "正文", "按钮", "https://example.com", BtnHorizontal)
	require.NoError(t, err)

	err = SendActionCard(context.Background(), server.URL, "标题", "正文", []ActionCardButton{
		{Title: "按钮", ActionURL: "https://example.com"},
	}, BtnVertical)
	require.NoError(t, err)

	err = SendFeedCard(context.Background(), server.URL, []FeedCardLink{
		{Title: "标题", MessageURL: "https://example.com", PicURL: "https://example.com/pic.png"},
	})
	require.NoError(t, err)
}

func TestSendInvalidWebhook(t *testing.T) {
	err := SendText(context.Background(), "", "hello")
	require.ErrorIs(t, err, ErrWebhookEmpty)
}

func TestSendInvalidWebhookFormat(t *testing.T) {
	tests := []string{
		"://bad",
		"ftp://example.com/robot/send",
	}
	for _, webhook := range tests {
		t.Run(webhook, func(t *testing.T) {
			err := SendText(context.Background(), webhook, "hello")
			require.ErrorIs(t, err, ErrWebhookInvalid)
		})
	}
}

func TestSendInvalidMessage(t *testing.T) {
	_, err := Send(context.Background(), "https://example.com/robot/send", nil)
	require.ErrorIs(t, err, ErrMessageNil)
}

func TestRequestURLWithoutSecret(t *testing.T) {
	u, err := url.Parse("https://example.com/robot/send?access_token=token")
	require.NoError(t, err)

	got := requestURL(u, applyOptions())
	require.Equal(t, "https://example.com/robot/send?access_token=token", got)
}

type roundTripFunc func(*http.Request) (*http.Response, error)

func (f roundTripFunc) RoundTrip(req *http.Request) (*http.Response, error) {
	return f(req)
}

type errReadCloser struct{}

func (errReadCloser) Read([]byte) (int, error) {
	return 0, errors.New("read failed")
}

func (errReadCloser) Close() error {
	return nil
}
