package ide

import (
	"context"

	"github.com/host-uk/core/pkg/ws"
	"github.com/modelcontextprotocol/go-sdk/mcp"
)

// Subsystem implements mcp.Subsystem and mcp.SubsystemWithShutdown for the IDE.
type Subsystem struct {
	cfg    Config
	bridge *Bridge
	hub    *ws.Hub
}

// New creates an IDE subsystem. The ws.Hub is used for real-time forwarding;
// pass nil if headless (tools still work but real-time streaming is disabled).
func New(hub *ws.Hub, opts ...Option) *Subsystem {
	cfg := DefaultConfig()
	for _, opt := range opts {
		opt(&cfg)
	}
	var bridge *Bridge
	if hub != nil {
		bridge = NewBridge(hub, cfg)
	}
	return &Subsystem{cfg: cfg, bridge: bridge, hub: hub}
}

// Name implements mcp.Subsystem.
func (s *Subsystem) Name() string { return "ide" }

// RegisterTools implements mcp.Subsystem.
func (s *Subsystem) RegisterTools(server *mcp.Server) {
	s.registerChatTools(server)
	s.registerBuildTools(server)
	s.registerDashboardTools(server)
}

// Shutdown implements mcp.SubsystemWithShutdown.
func (s *Subsystem) Shutdown(_ context.Context) error {
	if s.bridge != nil {
		s.bridge.Shutdown()
	}
	return nil
}

// Bridge returns the Laravel WebSocket bridge (may be nil in headless mode).
func (s *Subsystem) Bridge() *Bridge { return s.bridge }

// StartBridge begins the background connection to the Laravel backend.
func (s *Subsystem) StartBridge(ctx context.Context) {
	if s.bridge != nil {
		s.bridge.Start(ctx)
	}
}
