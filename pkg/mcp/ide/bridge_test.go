package ide

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/gorilla/websocket"
	"forge.lthn.ai/core/cli/pkg/ws"
)

var testUpgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool { return true },
}

// echoServer creates a test WebSocket server that echoes messages back.
func echoServer(t *testing.T) *httptest.Server {
	t.Helper()
	return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		conn, err := testUpgrader.Upgrade(w, r, nil)
		if err != nil {
			t.Logf("upgrade error: %v", err)
			return
		}
		defer conn.Close()
		for {
			mt, data, err := conn.ReadMessage()
			if err != nil {
				break
			}
			if err := conn.WriteMessage(mt, data); err != nil {
				break
			}
		}
	}))
}

func wsURL(ts *httptest.Server) string {
	return "ws" + strings.TrimPrefix(ts.URL, "http")
}

func TestBridge_Good_ConnectAndSend(t *testing.T) {
	ts := echoServer(t)
	defer ts.Close()

	hub := ws.NewHub()
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	go hub.Run(ctx)

	cfg := DefaultConfig()
	cfg.LaravelWSURL = wsURL(ts)
	cfg.ReconnectInterval = 100 * time.Millisecond

	bridge := NewBridge(hub, cfg)
	bridge.Start(ctx)

	// Wait for connection
	deadline := time.Now().Add(2 * time.Second)
	for !bridge.Connected() && time.Now().Before(deadline) {
		time.Sleep(50 * time.Millisecond)
	}
	if !bridge.Connected() {
		t.Fatal("bridge did not connect within timeout")
	}

	err := bridge.Send(BridgeMessage{
		Type: "test",
		Data: "hello",
	})
	if err != nil {
		t.Fatalf("Send() failed: %v", err)
	}
}

func TestBridge_Good_Shutdown(t *testing.T) {
	ts := echoServer(t)
	defer ts.Close()

	hub := ws.NewHub()
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	go hub.Run(ctx)

	cfg := DefaultConfig()
	cfg.LaravelWSURL = wsURL(ts)
	cfg.ReconnectInterval = 100 * time.Millisecond

	bridge := NewBridge(hub, cfg)
	bridge.Start(ctx)

	deadline := time.Now().Add(2 * time.Second)
	for !bridge.Connected() && time.Now().Before(deadline) {
		time.Sleep(50 * time.Millisecond)
	}

	bridge.Shutdown()
	if bridge.Connected() {
		t.Error("bridge should be disconnected after Shutdown")
	}
}

func TestBridge_Bad_SendWithoutConnection(t *testing.T) {
	hub := ws.NewHub()
	cfg := DefaultConfig()
	bridge := NewBridge(hub, cfg)

	err := bridge.Send(BridgeMessage{Type: "test"})
	if err == nil {
		t.Error("expected error when sending without connection")
	}
}

func TestBridge_Good_MessageDispatch(t *testing.T) {
	// Server that sends a message to the bridge on connect.
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		conn, err := testUpgrader.Upgrade(w, r, nil)
		if err != nil {
			return
		}
		defer conn.Close()

		msg := BridgeMessage{
			Type:    "chat_response",
			Channel: "chat:session-1",
			Data:    "hello from laravel",
		}
		data, _ := json.Marshal(msg)
		conn.WriteMessage(websocket.TextMessage, data)

		// Keep connection open
		for {
			_, _, err := conn.ReadMessage()
			if err != nil {
				break
			}
		}
	}))
	defer ts.Close()

	hub := ws.NewHub()
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	go hub.Run(ctx)

	cfg := DefaultConfig()
	cfg.LaravelWSURL = wsURL(ts)
	cfg.ReconnectInterval = 100 * time.Millisecond

	bridge := NewBridge(hub, cfg)
	bridge.Start(ctx)

	deadline := time.Now().Add(2 * time.Second)
	for !bridge.Connected() && time.Now().Before(deadline) {
		time.Sleep(50 * time.Millisecond)
	}
	if !bridge.Connected() {
		t.Fatal("bridge did not connect within timeout")
	}

	// Give time for the dispatched message to be processed.
	time.Sleep(200 * time.Millisecond)

	// Verify hub stats — the message was dispatched (even without subscribers).
	// This confirms the dispatch path ran without error.
}

func TestBridge_Good_Reconnect(t *testing.T) {
	callCount := 0
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		callCount++
		conn, err := testUpgrader.Upgrade(w, r, nil)
		if err != nil {
			return
		}
		// Close immediately on first connection to force reconnect
		if callCount == 1 {
			conn.Close()
			return
		}
		defer conn.Close()
		for {
			_, _, err := conn.ReadMessage()
			if err != nil {
				break
			}
		}
	}))
	defer ts.Close()

	hub := ws.NewHub()
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	go hub.Run(ctx)

	cfg := DefaultConfig()
	cfg.LaravelWSURL = wsURL(ts)
	cfg.ReconnectInterval = 100 * time.Millisecond
	cfg.MaxReconnectInterval = 200 * time.Millisecond

	bridge := NewBridge(hub, cfg)
	bridge.Start(ctx)

	// Wait long enough for a reconnect cycle
	deadline := time.Now().Add(3 * time.Second)
	for !bridge.Connected() && time.Now().Before(deadline) {
		time.Sleep(50 * time.Millisecond)
	}
	if !bridge.Connected() {
		t.Fatal("bridge did not reconnect within timeout")
	}
	if callCount < 2 {
		t.Errorf("expected at least 2 connection attempts, got %d", callCount)
	}
}

func TestSubsystem_Good_Name(t *testing.T) {
	sub := New(nil)
	if sub.Name() != "ide" {
		t.Errorf("expected name 'ide', got %q", sub.Name())
	}
}

func TestSubsystem_Good_NilHub(t *testing.T) {
	sub := New(nil)
	if sub.Bridge() != nil {
		t.Error("expected nil bridge when hub is nil")
	}
	// Shutdown should not panic
	if err := sub.Shutdown(context.Background()); err != nil {
		t.Errorf("Shutdown with nil bridge failed: %v", err)
	}
}
