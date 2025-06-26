package ws_test

import (
	"net/http"
	"testing"

	"github.com/ivy-mobile/odin/packet"
	"github.com/ivy-mobile/odin/xutil/xgo"
	"github.com/ivy-mobile/odin/xutil/xnet"
	"github.com/ivy-mobile/odin/xutil/xnet/ws"
)

func TestServer(t *testing.T) {
	server := ws.NewServer()
	server.OnStart(func() {
		t.Logf("server is started")
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
			t.Error(err)
			return
		}

		t.Logf("receive msg from client, connection id: %d, seq: %d, route: %d, msg: %s", conn.ID(), message.Seq, message.Route, string(message.Buffer))

		msg, err = packet.PackMessage(&packet.Message{
			Seq:    1,
			Route:  1,
			Buffer: []byte("I'm fine~~"),
		})
		if err != nil {
			t.Fatal(err)
		}

		if err = conn.Push(msg); err != nil {
			t.Error(err)
		}
	})
	server.OnUpgrade(func(w http.ResponseWriter, r *http.Request) (allowed bool) {
		return true
	})

	if err := server.Start(); err != nil {
		t.Fatal(err)
	}

	xgo.Go(func() {
		err := http.ListenAndServe(":8089", nil)
		if err != nil {
			t.Errorf("pprof server start failed: %v", err)
		}
	})

	select {}
}

func TestServer_Benchmark(t *testing.T) {
	server := ws.NewServer()
	server.OnStart(func() {
		t.Logf("server is started")
	})
	server.OnReceive(func(conn xnet.Conn, msg []byte) {
		_, err := packet.UnpackMessage(msg)
		if err != nil {
			t.Error(err)
			return
		}

		msg, err = packet.PackMessage(&packet.Message{
			Seq:    1,
			Route:  1,
			Buffer: []byte("I'm fine~~"),
		})
		if err != nil {
			t.Fatal(err)
		}

		if err = conn.Push(msg); err != nil {
			t.Error(err)
		}
	})
	server.OnUpgrade(func(w http.ResponseWriter, r *http.Request) (allowed bool) {
		return true
	})

	if err := server.Start(); err != nil {
		t.Fatal(err)
	}

	xgo.Go(func() {
		err := http.ListenAndServe(":8089", nil)
		if err != nil {
			t.Errorf("pprof server start failed: %v", err)
		}
	})

	select {}
}
