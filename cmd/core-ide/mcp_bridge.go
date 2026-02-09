package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"sync"

	"github.com/host-uk/core/pkg/ws"
	"github.com/wailsapp/wails/v3/pkg/application"
)

// MCPBridge is the SERVER bridge that exposes MCP tools via HTTP.
// AI agents call these endpoints to control windows, execute JS in webviews,
// access the clipboard, show notifications, and query the app state.
type MCPBridge struct {
	app          *application.App
	hub          *ws.Hub
	claudeBridge *ClaudeBridge
	port         int
	running      bool
	mu           sync.Mutex
}

// NewMCPBridge creates a new MCP bridge server.
func NewMCPBridge(hub *ws.Hub, port int) *MCPBridge {
	cb := NewClaudeBridge("ws://localhost:9876/ws")
	return &MCPBridge{
		hub:          hub,
		claudeBridge: cb,
		port:         port,
	}
}

// ServiceName returns the Wails service name.
func (b *MCPBridge) ServiceName() string { return "MCPBridge" }

// ServiceStartup is called by Wails when the app starts.
func (b *MCPBridge) ServiceStartup(_ context.Context, _ application.ServiceOptions) error {
	b.app = application.Get()
	go b.startHTTPServer()
	log.Printf("MCP Bridge started on port %d", b.port)
	return nil
}

// ServiceShutdown is called when the app shuts down.
func (b *MCPBridge) ServiceShutdown() error {
	b.mu.Lock()
	defer b.mu.Unlock()
	b.running = false
	return nil
}

// startHTTPServer starts the HTTP server for MCP tools and WebSocket.
func (b *MCPBridge) startHTTPServer() {
	b.mu.Lock()
	b.running = true
	b.mu.Unlock()

	// Start the Claude bridge (CLIENT → MCP core on :9876)
	b.claudeBridge.Start()

	mux := http.NewServeMux()

	// WebSocket endpoint for Angular frontend
	mux.HandleFunc("/ws", b.hub.HandleWebSocket)

	// Claude bridge WebSocket relay (GUI clients ↔ MCP core)
	mux.HandleFunc("/claude", b.claudeBridge.HandleWebSocket)

	// MCP server endpoints
	mux.HandleFunc("/mcp", b.handleMCPInfo)
	mux.HandleFunc("/mcp/tools", b.handleMCPTools)
	mux.HandleFunc("/mcp/call", b.handleMCPCall)

	// Health check
	mux.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]any{
			"status":       "ok",
			"mcp":          true,
			"claudeBridge": b.claudeBridge.Connected(),
		})
	})

	addr := fmt.Sprintf("127.0.0.1:%d", b.port)
	log.Printf("MCP HTTP server listening on %s", addr)

	if err := http.ListenAndServe(addr, mux); err != nil {
		log.Printf("MCP HTTP server error: %v", err)
	}
}

// handleMCPInfo returns MCP server information.
func (b *MCPBridge) handleMCPInfo(w http.ResponseWriter, _ *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("Access-Control-Allow-Origin", "*")

	json.NewEncoder(w).Encode(map[string]any{
		"name":    "core-ide",
		"version": "0.1.0",
		"capabilities": map[string]any{
			"webview":       true,
			"windowControl": true,
			"clipboard":     true,
			"notifications": true,
			"websocket":     fmt.Sprintf("ws://localhost:%d/ws", b.port),
			"claude":        fmt.Sprintf("ws://localhost:%d/claude", b.port),
		},
	})
}

// handleMCPTools returns the list of available tools.
func (b *MCPBridge) handleMCPTools(w http.ResponseWriter, _ *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("Access-Control-Allow-Origin", "*")

	tools := []map[string]string{
		// Window management
		{"name": "window_list", "description": "List all windows with positions and sizes"},
		{"name": "window_get", "description": "Get info about a specific window"},
		{"name": "window_position", "description": "Move a window to specific coordinates"},
		{"name": "window_size", "description": "Resize a window"},
		{"name": "window_bounds", "description": "Set position and size in one call"},
		{"name": "window_maximize", "description": "Maximize a window"},
		{"name": "window_minimize", "description": "Minimize a window"},
		{"name": "window_restore", "description": "Restore from maximized/minimized"},
		{"name": "window_focus", "description": "Bring window to front"},
		{"name": "window_visibility", "description": "Show or hide a window"},
		{"name": "window_title", "description": "Change window title"},
		{"name": "window_title_get", "description": "Get current window title"},
		{"name": "window_fullscreen", "description": "Toggle fullscreen mode"},
		{"name": "window_always_on_top", "description": "Pin window above others"},
		{"name": "window_create", "description": "Create a new window at specific position"},
		{"name": "window_close", "description": "Close a window by name"},
		{"name": "window_background_colour", "description": "Set window background colour with alpha"},
		// Webview interaction
		{"name": "webview_eval", "description": "Execute JavaScript in a window's webview"},
		{"name": "webview_navigate", "description": "Navigate window to a URL"},
		{"name": "webview_list", "description": "List windows with webview info"},
		// System integration
		{"name": "clipboard_read", "description": "Read text from system clipboard"},
		{"name": "clipboard_write", "description": "Write text to system clipboard"},
		// System tray
		{"name": "tray_set_tooltip", "description": "Set system tray tooltip"},
		{"name": "tray_set_label", "description": "Set system tray label"},
	}
	json.NewEncoder(w).Encode(map[string]any{"tools": tools})
}

// handleMCPCall handles tool calls via HTTP POST.
func (b *MCPBridge) handleMCPCall(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Access-Control-Allow-Methods", "POST, OPTIONS")
	w.Header().Set("Access-Control-Allow-Headers", "Content-Type")

	if r.Method == "OPTIONS" {
		w.WriteHeader(http.StatusOK)
		return
	}
	if r.Method != "POST" {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req struct {
		Tool   string         `json:"tool"`
		Params map[string]any `json:"params"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	var result map[string]any
	if len(req.Tool) > 8 && req.Tool[:8] == "webview_" {
		result = b.executeWebviewTool(req.Tool, req.Params)
	} else {
		result = b.executeWindowTool(req.Tool, req.Params)
	}
	json.NewEncoder(w).Encode(result)
}

// executeWindowTool handles window, clipboard, tray, and notification tools.
func (b *MCPBridge) executeWindowTool(tool string, params map[string]any) map[string]any {
	if b.app == nil {
		return map[string]any{"error": "app not available"}
	}

	switch tool {
	case "window_list":
		return b.windowList()

	case "window_get":
		name := strParam(params, "name")
		return b.windowGet(name)

	case "window_position":
		name := strParam(params, "name")
		x := intParam(params, "x")
		y := intParam(params, "y")
		w, ok := b.app.Window.Get(name)
		if !ok {
			return map[string]any{"error": "window not found", "name": name}
		}
		w.SetPosition(x, y)
		return map[string]any{"success": true, "name": name, "x": x, "y": y}

	case "window_size":
		name := strParam(params, "name")
		width := intParam(params, "width")
		height := intParam(params, "height")
		w, ok := b.app.Window.Get(name)
		if !ok {
			return map[string]any{"error": "window not found", "name": name}
		}
		w.SetSize(width, height)
		return map[string]any{"success": true, "name": name, "width": width, "height": height}

	case "window_bounds":
		name := strParam(params, "name")
		x := intParam(params, "x")
		y := intParam(params, "y")
		width := intParam(params, "width")
		height := intParam(params, "height")
		w, ok := b.app.Window.Get(name)
		if !ok {
			return map[string]any{"error": "window not found", "name": name}
		}
		w.SetPosition(x, y)
		w.SetSize(width, height)
		return map[string]any{"success": true, "name": name, "x": x, "y": y, "width": width, "height": height}

	case "window_maximize":
		name := strParam(params, "name")
		w, ok := b.app.Window.Get(name)
		if !ok {
			return map[string]any{"error": "window not found", "name": name}
		}
		w.Maximise()
		return map[string]any{"success": true, "action": "maximize"}

	case "window_minimize":
		name := strParam(params, "name")
		w, ok := b.app.Window.Get(name)
		if !ok {
			return map[string]any{"error": "window not found", "name": name}
		}
		w.Minimise()
		return map[string]any{"success": true, "action": "minimize"}

	case "window_restore":
		name := strParam(params, "name")
		w, ok := b.app.Window.Get(name)
		if !ok {
			return map[string]any{"error": "window not found", "name": name}
		}
		w.Restore()
		return map[string]any{"success": true, "action": "restore"}

	case "window_focus":
		name := strParam(params, "name")
		w, ok := b.app.Window.Get(name)
		if !ok {
			return map[string]any{"error": "window not found", "name": name}
		}
		w.Show()
		w.Focus()
		return map[string]any{"success": true, "action": "focus"}

	case "window_visibility":
		name := strParam(params, "name")
		visible, _ := params["visible"].(bool)
		w, ok := b.app.Window.Get(name)
		if !ok {
			return map[string]any{"error": "window not found", "name": name}
		}
		if visible {
			w.Show()
		} else {
			w.Hide()
		}
		return map[string]any{"success": true, "visible": visible}

	case "window_title":
		name := strParam(params, "name")
		title := strParam(params, "title")
		w, ok := b.app.Window.Get(name)
		if !ok {
			return map[string]any{"error": "window not found", "name": name}
		}
		w.SetTitle(title)
		return map[string]any{"success": true, "title": title}

	case "window_title_get":
		name := strParam(params, "name")
		_, ok := b.app.Window.Get(name)
		if !ok {
			return map[string]any{"error": "window not found", "name": name}
		}
		// Wails v3 Window interface has SetTitle but no Title getter;
		// return the window name as a fallback identifier.
		return map[string]any{"name": name}

	case "window_fullscreen":
		name := strParam(params, "name")
		fullscreen, _ := params["fullscreen"].(bool)
		w, ok := b.app.Window.Get(name)
		if !ok {
			return map[string]any{"error": "window not found", "name": name}
		}
		if fullscreen {
			w.Fullscreen()
		} else {
			w.UnFullscreen()
		}
		return map[string]any{"success": true, "fullscreen": fullscreen}

	case "window_always_on_top":
		name := strParam(params, "name")
		onTop, _ := params["onTop"].(bool)
		w, ok := b.app.Window.Get(name)
		if !ok {
			return map[string]any{"error": "window not found", "name": name}
		}
		w.SetAlwaysOnTop(onTop)
		return map[string]any{"success": true, "alwaysOnTop": onTop}

	case "window_create":
		name := strParam(params, "name")
		title := strParam(params, "title")
		url := strParam(params, "url")
		x := intParam(params, "x")
		y := intParam(params, "y")
		width := intParam(params, "width")
		height := intParam(params, "height")
		if width == 0 {
			width = 800
		}
		if height == 0 {
			height = 600
		}
		opts := application.WebviewWindowOptions{
			Name:             name,
			Title:            title,
			URL:              url,
			Width:            width,
			Height:           height,
			Hidden:           false,
			BackgroundColour: application.NewRGB(22, 27, 34),
		}
		w := b.app.Window.NewWithOptions(opts)
		if x != 0 || y != 0 {
			w.SetPosition(x, y)
		}
		return map[string]any{"success": true, "name": name}

	case "window_close":
		name := strParam(params, "name")
		w, ok := b.app.Window.Get(name)
		if !ok {
			return map[string]any{"error": "window not found", "name": name}
		}
		w.Close()
		return map[string]any{"success": true, "action": "close"}

	case "window_background_colour":
		name := strParam(params, "name")
		r := uint8(intParam(params, "r"))
		g := uint8(intParam(params, "g"))
		bv := uint8(intParam(params, "b"))
		a := uint8(intParam(params, "a"))
		if a == 0 {
			a = 255
		}
		w, ok := b.app.Window.Get(name)
		if !ok {
			return map[string]any{"error": "window not found", "name": name}
		}
		w.SetBackgroundColour(application.NewRGBA(r, g, bv, a))
		return map[string]any{"success": true}

	case "clipboard_read":
		text, ok := b.app.Clipboard.Text()
		if !ok {
			return map[string]any{"error": "failed to read clipboard"}
		}
		return map[string]any{"text": text}

	case "clipboard_write":
		text, _ := params["text"].(string)
		ok := b.app.Clipboard.SetText(text)
		if !ok {
			return map[string]any{"error": "failed to write clipboard"}
		}
		return map[string]any{"success": true}

	case "tray_set_tooltip":
		// System tray is managed at startup; this is informational
		return map[string]any{"info": "tray tooltip can be set via system tray menu"}

	case "tray_set_label":
		return map[string]any{"info": "tray label can be set via system tray menu"}

	default:
		return map[string]any{"error": "unknown tool", "tool": tool}
	}
}

// executeWebviewTool handles webview/JS tools.
func (b *MCPBridge) executeWebviewTool(tool string, params map[string]any) map[string]any {
	if b.app == nil {
		return map[string]any{"error": "app not available"}
	}

	switch tool {
	case "webview_eval":
		windowName := strParam(params, "window")
		code := strParam(params, "code")
		w, ok := b.app.Window.Get(windowName)
		if !ok {
			return map[string]any{"error": "window not found", "window": windowName}
		}
		w.ExecJS(code)
		return map[string]any{"success": true, "window": windowName}

	case "webview_navigate":
		windowName := strParam(params, "window")
		url := strParam(params, "url")
		w, ok := b.app.Window.Get(windowName)
		if !ok {
			return map[string]any{"error": "window not found", "window": windowName}
		}
		w.SetURL(url)
		return map[string]any{"success": true, "url": url}

	case "webview_list":
		return b.windowList()

	default:
		return map[string]any{"error": "unknown webview tool", "tool": tool}
	}
}

// windowList returns info for all known windows.
func (b *MCPBridge) windowList() map[string]any {
	knownNames := []string{"tray-panel", "main", "settings"}
	var windows []map[string]any
	for _, name := range knownNames {
		w, ok := b.app.Window.Get(name)
		if !ok {
			continue
		}
		x, y := w.Position()
		width, height := w.Size()
		windows = append(windows, map[string]any{
			"name":   name,
			"title":  w.Name(),
			"x":      x,
			"y":      y,
			"width":  width,
			"height": height,
		})
	}
	return map[string]any{"windows": windows}
}

// windowGet returns info for a specific window.
func (b *MCPBridge) windowGet(name string) map[string]any {
	w, ok := b.app.Window.Get(name)
	if !ok {
		return map[string]any{"error": "window not found", "name": name}
	}
	x, y := w.Position()
	width, height := w.Size()
	return map[string]any{
		"window": map[string]any{
			"name":   name,
			"title":  w.Name(),
			"x":      x,
			"y":      y,
			"width":  width,
			"height": height,
		},
	}
}

// Parameter helpers
func strParam(params map[string]any, key string) string {
	if v, ok := params[key].(string); ok {
		return v
	}
	return ""
}

func intParam(params map[string]any, key string) int {
	if v, ok := params[key].(float64); ok {
		return int(v)
	}
	return 0
}
