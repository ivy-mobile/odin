package tcp_test

import (
	"net/http"
	_ "net/http/pprof"
	"testing"

	"github.com/ivy-mobile/odin/packet"
	"github.com/ivy-mobile/odin/xutil/xnet"
	"github.com/ivy-mobile/odin/xutil/xnet/tcp"
)

func TestServer_Simple(t *testing.T) {
	server := tcp.NewServer()

	server.OnStart(func() {
		t.Log("server is started")
	})

	server.OnStop(func() {
		t.Log("server is stopped")
	})

	server.OnConnect(func(conn xnet.Conn) {
		t.Logf("connection is opened, connection id: %d", conn.ID())
	})

	server.OnDisconnect(func(conn xnet.Conn) {
		t.Logf("connection is closed, connection id: %d", conn.ID())
	})

	server.OnReceive(func(conn xnet.Conn, msg []byte) {
		message, err := packet.UnpackMessage(msg)
		if err != nil {
			t.Errorf("unpack message failed: %v", err)
			return
		}

		t.Logf("receive message from client, cid: %d, seq: %d, route: %d, msg: %s", conn.ID(), message.Seq, message.Route, string(message.Buffer))

		msg, err = packet.PackMessage(&packet.Message{
			Seq:    1,
			Route:  1,
			Buffer: []byte("I'm fine~~"),
		})
		if err != nil {
			t.Errorf("pack message failed: %v", err)
			return
		}

		if err = conn.Push(msg); err != nil {
			t.Errorf("push message failed: %v", err)
		}
	})

	if err := server.Start(); err != nil {
		t.Fatalf("start server failed: %v", err)
	}

	select {}
}

func TestServer_Benchmark(t *testing.T) {
	server := tcp.NewServer(
		tcp.WithServerHeartbeatInterval(0),
	)

	server.OnStart(func() {
		t.Log("server is started")
	})

	server.OnReceive(func(conn xnet.Conn, msg []byte) {
		message, err := packet.UnpackMessage(msg)
		if err != nil {
			t.Errorf("unpack message failed: %v", err)
			return
		}

		data, err := packet.PackMessage(&packet.Message{
			Seq:    message.Seq,
			Route:  message.Route,
			Buffer: message.Buffer,
		})
		if err != nil {
			t.Errorf("pack message failed: %v", err)
			return
		}

		if err = conn.Push(data); err != nil {
			t.Errorf("push message failed: %v", err)
			return
		}
	})

	if err := server.Start(); err != nil {
		t.Fatalf("start server failed: %v", err)
	}

	go func() {
		err := http.ListenAndServe(":8089", nil)
		if err != nil {
			t.Errorf("pprof server start failed: %v", err)
		}
	}()

	select {}
}
