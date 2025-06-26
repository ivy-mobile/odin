package gate

import (
	"context"

	"github.com/ivy-mobile/odin/encoding/json"
	"github.com/ivy-mobile/odin/encoding/proto"
	"github.com/ivy-mobile/odin/message"
	msgjson "github.com/ivy-mobile/odin/message/json"
	msgproto "github.com/ivy-mobile/odin/message/proto"
	"github.com/ivy-mobile/odin/topic"
	"github.com/ivy-mobile/odin/xutil/xconv"
	"github.com/ivy-mobile/odin/xutil/xlog"

	"github.com/olahol/melody"
)

const (
	UserKey = "userId"
)

func (g *Gate) handleConnect(s *melody.Session) {

	temp := s.Request.FormValue(UserKey)
	if temp == "" {
		xlog.Error().Msgf("[Connect] faild, addr: %s ", s.Request.RemoteAddr)
		s.Write([]byte("userId is empty"))
		s.Close()
		return
	}
	userId := xconv.Int64(temp)
	xlog.Info().Msgf("[Connect] userId: %v, addr: %s ", userId, s.Request.RemoteAddr)
	// 保存到 melody.Session
	s.Set("userId", userId)
	// 保存到 会话管理器
	g.sessions.Set(userId, s)
}

func (g *Gate) handleDisconnect(s *melody.Session) {
	temp, ok := s.Get("userId")
	if !ok {
		xlog.Error().Msgf("[Disconnect] failed: userId not found, addr: %s", s.Request.RemoteAddr)
		return
	}
	// 关闭会话 - 释放资源
	err := s.Close()
	if err != nil {
		xlog.Error().Msgf("[Disconnect] Close melody.Session failed, err: %v", err)
	}
	userId := temp.(int64)
	// 移除会话
	g.sessions.Remove(userId)
	xlog.Info().Msgf("[Disconnect] userId: %v, addr: %s", userId, s.Request.RemoteAddr)
}

// 处理文本消息 - 采用json协议
func (g *Gate) handleMessage(s *melody.Session, msg []byte) {

	if g.opts.codec.Name() != json.Name {
		xlog.Error().Msgf("[HandleMessage] codec not json, codec: %s", g.opts.codec.Name())
		return
	}
	g.dispatch(msg) // 分发
}

// 处理二进制消息 - 采用proto协议
func (g *Gate) handleMessageBinary(s *melody.Session, msg []byte) {

	if g.opts.codec.Name() != proto.Name {
		xlog.Error().Msgf("[HandleMessage] codec not json, codec: %s", g.opts.codec.Name())
		return
	}
	g.dispatch(msg) // 分发
}

// 分发消息
func (g *Gate) dispatch(data []byte) {

	// 1. 判断事件总线是否可用
	eb := g.opts.eventbus
	if eb == nil {
		xlog.Error().Msg("[dispatch] eventbus is nil, can not publish event")
		return
	}

	// 2. proto解析数据
	var msg message.Message
	switch g.opts.codec.Name() {
	case json.Name:
		var jsonMsg msgjson.JsonMessage
		if err := json.Unmarshal(data, &jsonMsg); err != nil {
			xlog.Error().Msgf("[dispatch] failed, unmarshal faild, err: %s", err.Error())
			return
		}
		msg = &jsonMsg
	case proto.Name:
		var protoMsg msgproto.Message
		if err := proto.Unmarshal(data, &protoMsg); err != nil {
			xlog.Error().Msgf("[dispatch] failed, unmarshal faild, err: %s", err.Error())
			return
		}
		msg = &protoMsg
	}

	// 3. 验证用户-链接 是否匹配
	s := g.sessions.Get(msg.GetUid())
	if s == nil {
		xlog.Error().Msgf("[dispatch] failed, user session not found, userId: %v", msg.GetUid())
		return
	}
	if uid, ok := s.Get(UserKey); !ok || uid.(int64) != msg.GetUid() {
		xlog.Error().Msgf("[dispatch] failed, user session not match, userId: %v", msg.GetUid())
		return
	}

	// 4. 构造topic
	// topic = 当前服务名.目标服务名
	// TODO 未来需要处理多节点场景下，精确转发
	tp := topic.Gate2GameTopic(g.Name(), msg.GetGame())

	// 4, 通过事件总线发布
	if err := eb.Publish(context.Background(), tp, data); err != nil {
		xlog.Error().Msgf("[dispatch] Publish failed, err: %v", err)
	}
}
