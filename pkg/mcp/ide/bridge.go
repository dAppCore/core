package ide

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"sync"
	"time"

	"github.com/gorilla/websocket"
	"forge.lthn.ai/core/cli/pkg/ws"
)

// BridgeMessage is the wire format between the IDE and Laravel.
type BridgeMessage struct {
	Type      string    `json:"type"`
	Channel   string    `json:"channel,omitempty"`
	SessionID string    `json:"sessionId,omitempty"`
	Data      any       `json:"data,omitempty"`
	Timestamp time.Time `json:"timestamp"`
}

// Bridge maintains a WebSocket connection to the Laravel core-agentic
// backend and forwards responses to a local ws.Hub.
type Bridge struct {
	cfg  Config
	hub  *ws.Hub
	conn *websocket.Conn

	mu        sync.Mutex
	connected bool
	cancel    context.CancelFunc
}

// NewBridge creates a bridge that will connect to the Laravel backend and
// forward incoming messages to the provided ws.Hub channels.
func NewBridge(hub *ws.Hub, cfg Config) *Bridge {
	return &Bridge{cfg: cfg, hub: hub}
}

// Start begins the connection loop in a background goroutine.
// Call Shutdown to stop it.
func (b *Bridge) Start(ctx context.Context) {
	ctx, b.cancel = context.WithCancel(ctx)
	go b.connectLoop(ctx)
}

// Shutdown cleanly closes the bridge.
func (b *Bridge) Shutdown() {
	if b.cancel != nil {
		b.cancel()
	}
	b.mu.Lock()
	defer b.mu.Unlock()
	if b.conn != nil {
		b.conn.Close()
		b.conn = nil
	}
	b.connected = false
}

// Connected reports whether the bridge has an active connection.
func (b *Bridge) Connected() bool {
	b.mu.Lock()
	defer b.mu.Unlock()
	return b.connected
}

// Send sends a message to the Laravel backend.
func (b *Bridge) Send(msg BridgeMessage) error {
	b.mu.Lock()
	defer b.mu.Unlock()
	if b.conn == nil {
		return fmt.Errorf("bridge: not connected")
	}
	msg.Timestamp = time.Now()
	data, err := json.Marshal(msg)
	if err != nil {
		return fmt.Errorf("bridge: marshal failed: %w", err)
	}
	return b.conn.WriteMessage(websocket.TextMessage, data)
}

// connectLoop reconnects to Laravel with exponential backoff.
func (b *Bridge) connectLoop(ctx context.Context) {
	delay := b.cfg.ReconnectInterval
	for {
		select {
		case <-ctx.Done():
			return
		default:
		}

		if err := b.dial(ctx); err != nil {
			log.Printf("ide bridge: connect failed: %v", err)
			select {
			case <-ctx.Done():
				return
			case <-time.After(delay):
			}
			delay = min(delay*2, b.cfg.MaxReconnectInterval)
			continue
		}

		// Reset backoff on successful connection
		delay = b.cfg.ReconnectInterval
		b.readLoop(ctx)
	}
}

func (b *Bridge) dial(ctx context.Context) error {
	dialer := websocket.Dialer{
		HandshakeTimeout: 10 * time.Second,
	}
	conn, _, err := dialer.DialContext(ctx, b.cfg.LaravelWSURL, nil)
	if err != nil {
		return err
	}

	b.mu.Lock()
	b.conn = conn
	b.connected = true
	b.mu.Unlock()

	log.Printf("ide bridge: connected to %s", b.cfg.LaravelWSURL)
	return nil
}

func (b *Bridge) readLoop(ctx context.Context) {
	defer func() {
		b.mu.Lock()
		if b.conn != nil {
			b.conn.Close()
		}
		b.connected = false
		b.mu.Unlock()
	}()

	for {
		select {
		case <-ctx.Done():
			return
		default:
		}

		_, data, err := b.conn.ReadMessage()
		if err != nil {
			log.Printf("ide bridge: read error: %v", err)
			return
		}

		var msg BridgeMessage
		if err := json.Unmarshal(data, &msg); err != nil {
			log.Printf("ide bridge: unmarshal error: %v", err)
			continue
		}

		b.dispatch(msg)
	}
}

// dispatch routes an incoming message to the appropriate ws.Hub channel.
func (b *Bridge) dispatch(msg BridgeMessage) {
	if b.hub == nil {
		return
	}

	wsMsg := ws.Message{
		Type: ws.TypeEvent,
		Data: msg.Data,
	}

	channel := msg.Channel
	if channel == "" {
		channel = "ide:" + msg.Type
	}

	if err := b.hub.SendToChannel(channel, wsMsg); err != nil {
		log.Printf("ide bridge: dispatch to %s failed: %v", channel, err)
	}
}
