package main

import (
	"context"
	"log"

	"forge.lthn.ai/core/go/pkg/mcp/ide"
	"forge.lthn.ai/core/go/pkg/ws"
	"github.com/wailsapp/wails/v3/pkg/application"
)

// IDEService provides core IDE bindings for the frontend.
type IDEService struct {
	app    *application.App
	ideSub *ide.Subsystem
	hub    *ws.Hub
}

// NewIDEService creates a new IDEService.
func NewIDEService(ideSub *ide.Subsystem, hub *ws.Hub) *IDEService {
	return &IDEService{ideSub: ideSub, hub: hub}
}

// ServiceName returns the service name for Wails.
func (s *IDEService) ServiceName() string { return "IDEService" }

// ServiceStartup is called when the Wails application starts.
func (s *IDEService) ServiceStartup(_ context.Context, _ application.ServiceOptions) error {
	log.Println("IDEService started")
	return nil
}

// ServiceShutdown is called when the Wails application shuts down.
func (s *IDEService) ServiceShutdown() error {
	log.Println("IDEService shutdown")
	return nil
}

// ConnectionStatus represents the IDE bridge connection state.
type ConnectionStatus struct {
	BridgeConnected bool   `json:"bridgeConnected"`
	LaravelURL      string `json:"laravelUrl"`
	WSClients       int    `json:"wsClients"`
	WSChannels      int    `json:"wsChannels"`
}

// GetConnectionStatus returns the current bridge and WebSocket status.
func (s *IDEService) GetConnectionStatus() ConnectionStatus {
	connected := false
	if s.ideSub.Bridge() != nil {
		connected = s.ideSub.Bridge().Connected()
	}

	stats := s.hub.Stats()
	return ConnectionStatus{
		BridgeConnected: connected,
		WSClients:       stats.Clients,
		WSChannels:      stats.Channels,
	}
}

// DashboardData aggregates data for the dashboard view.
type DashboardData struct {
	Connection ConnectionStatus `json:"connection"`
}

// GetDashboard returns aggregated dashboard data.
func (s *IDEService) GetDashboard() DashboardData {
	return DashboardData{
		Connection: s.GetConnectionStatus(),
	}
}

// ShowWindow shows a named window.
func (s *IDEService) ShowWindow(name string) {
	if s.app == nil {
		return
	}
	if w, ok := s.app.Window.Get(name); ok {
		w.Show()
		w.Focus()
	}
}
