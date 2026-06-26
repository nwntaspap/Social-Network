package ws

import (
	"net/http"
	"net/http/httptest"
	"net/url"
	"sync"
	"testing"
	"time"

	"github.com/gorilla/websocket"
)

func TestRecoverWS_CatchesPanic(t *testing.T) {
	called := false
	var mu sync.Mutex
	done := make(chan struct{})

	go func() {
		defer func() {
			mu.Lock()
			called = true
			mu.Unlock()
			close(done)
		}()
		defer recoverWS()
		panic("test panic")
	}()

	select {
	case <-done:
	case <-time.After(time.Second):
		t.Fatal("goroutine did not finish after panic")
	}

	mu.Lock()
	if !called {
		t.Error("defer after recoverWS did not run")
	}
	mu.Unlock()
}

func TestReadPump_PanicRecovery_UnregistersClient(t *testing.T) {
	hub := NewHub()
	serverWS, clientWS := wsPair(t)
	defer clientWS.Close()

	client := NewClient("user-1", hub, serverWS)
	hub.Register(client)

	if !hub.IsOnline("user-1") {
		t.Fatal("client should be registered before panic")
	}

	done := make(chan struct{})
	go func() {
		client.ReadPump(func(c *Client, msg []byte) {
			panic("simulated panic in onMessage")
		})
		close(done)
	}()

	clientWS.WriteMessage(websocket.TextMessage, []byte("trigger"))
	_ = clientWS.WriteMessage(websocket.TextMessage, []byte("trigger again"))

	select {
	case <-done:
	case <-time.After(time.Second):
		t.Fatal("ReadPump did not exit after panic")
	}

	if hub.IsOnline("user-1") {
		t.Error("client should be unregistered after ReadPump panic")
	}
}

func TestWritePump_ExitsOnSendChannelClose(t *testing.T) {
	hub := NewHub()
	serverWS, clientWS := wsPair(t)
	defer clientWS.Close()

	client := NewClient("user-1", hub, serverWS)
	hub.Register(client)

	done := make(chan struct{})
	go func() {
		client.WritePump()
		close(done)
	}()

	close(client.send)

	select {
	case <-done:
	case <-time.After(time.Second):
		t.Fatal("WritePump did not exit after send channel close")
	}
}

func TestWritePump_SendsMessage(t *testing.T) {
	hub := NewHub()
	serverWS, clientWS := wsPair(t)
	defer clientWS.Close()

	client := NewClient("user-1", hub, serverWS)

	go client.WritePump()

	client.Send([]byte("hello world"))

	_, msg, err := clientWS.ReadMessage()
	if err != nil {
		t.Fatalf("ReadMessage error: %v", err)
	}
	if string(msg) != "hello world" {
		t.Errorf("got %q, want %q", string(msg), "hello world")
	}

	close(client.send)
}

func wsPair(t *testing.T) (server, client *websocket.Conn) {
	t.Helper()

	upgrader := websocket.Upgrader{CheckOrigin: func(r *http.Request) bool { return true }}
	serverWS := make(chan *websocket.Conn, 1)

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		conn, err := upgrader.Upgrade(w, r, nil)
		if err != nil {
			return
		}
		serverWS <- conn
	}))

	u, _ := url.Parse(srv.URL)
	u.Scheme = "ws"
	clientConn, _, err := websocket.DefaultDialer.Dial(u.String(), nil)
	if err != nil {
		srv.Close()
		t.Fatalf("dial failed: %v", err)
	}

	serverConn := <-serverWS

	t.Cleanup(func() {
		serverConn.Close()
		srv.Close()
	})

	return serverConn, clientConn
}
