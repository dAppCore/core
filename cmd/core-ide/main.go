// Package main provides the Core IDE desktop application.
// Core IDE connects to the Laravel core-agentic backend via MCP bridge,
// providing a chat interface for AI agent sessions, build monitoring,
// and a system dashboard.
package main

import (
	"context"
	"embed"
	"io/fs"
	"log"
	"runtime"

	"github.com/host-uk/core/cmd/core-ide/icons"
	"github.com/host-uk/core/pkg/mcp/ide"
	"github.com/host-uk/core/pkg/ws"
	"github.com/wailsapp/wails/v3/pkg/application"
)

//go:embed all:frontend/dist/core-ide/browser
var assets embed.FS

func main() {
	staticAssets, err := fs.Sub(assets, "frontend/dist/core-ide/browser")
	if err != nil {
		log.Fatal(err)
	}

	// Create shared WebSocket hub for real-time streaming
	hub := ws.NewHub()
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	go hub.Run(ctx)

	// Create IDE subsystem (bridge to Laravel core-agentic)
	ideSub := ide.New(hub)
	ideSub.StartBridge(ctx)

	// Create Wails services
	ideService := NewIDEService(ideSub, hub)
	chatService := NewChatService(ideSub)
	buildService := NewBuildService(ideSub)

	app := application.New(application.Options{
		Name:        "Core IDE",
		Description: "Host UK Platform IDE - AI Agent Sessions, Build Monitoring & Dashboard",
		Services: []application.Service{
			application.NewService(ideService),
			application.NewService(chatService),
			application.NewService(buildService),
		},
		Assets: application.AssetOptions{
			Handler: application.AssetFileServerFS(staticAssets),
		},
		Mac: application.MacOptions{
			ActivationPolicy: application.ActivationPolicyAccessory,
		},
	})

	ideService.app = app

	setupSystemTray(app, ideService)

	log.Println("Starting Core IDE...")
	log.Println("  - System tray active")
	log.Println("  - Bridge connecting to Laravel core-agentic...")

	if err := app.Run(); err != nil {
		log.Fatal(err)
	}

	cancel()
}

// setupSystemTray configures the system tray icon, menu, and windows.
func setupSystemTray(app *application.App, ideService *IDEService) {
	systray := app.SystemTray.New()
	systray.SetTooltip("Core IDE")

	if runtime.GOOS == "darwin" {
		systray.SetTemplateIcon(icons.TrayTemplate)
	} else {
		systray.SetDarkModeIcon(icons.TrayDark)
		systray.SetIcon(icons.TrayLight)
	}

	// Tray panel window
	trayWindow := app.Window.NewWithOptions(application.WebviewWindowOptions{
		Name:             "tray-panel",
		Title:            "Core IDE",
		Width:            400,
		Height:           500,
		URL:              "/tray",
		Hidden:           true,
		Frameless:        true,
		BackgroundColour: application.NewRGB(22, 27, 34),
	})
	systray.AttachWindow(trayWindow).WindowOffset(5)

	// Main IDE window
	app.Window.NewWithOptions(application.WebviewWindowOptions{
		Name:             "main",
		Title:            "Core IDE",
		Width:            1400,
		Height:           900,
		URL:              "/main",
		Hidden:           true,
		BackgroundColour: application.NewRGB(22, 27, 34),
	})

	// Settings window
	app.Window.NewWithOptions(application.WebviewWindowOptions{
		Name:             "settings",
		Title:            "Core IDE Settings",
		Width:            600,
		Height:           500,
		URL:              "/settings",
		Hidden:           true,
		BackgroundColour: application.NewRGB(22, 27, 34),
	})

	// Tray menu
	trayMenu := app.Menu.New()

	statusItem := trayMenu.Add("Status: Connecting...")
	statusItem.SetEnabled(false)

	trayMenu.AddSeparator()

	trayMenu.Add("Open IDE").OnClick(func(ctx *application.Context) {
		if w, ok := app.Window.Get("main"); ok {
			w.Show()
			w.Focus()
		}
	})

	trayMenu.Add("Settings...").OnClick(func(ctx *application.Context) {
		if w, ok := app.Window.Get("settings"); ok {
			w.Show()
			w.Focus()
		}
	})

	trayMenu.AddSeparator()

	trayMenu.Add("Quit Core IDE").OnClick(func(ctx *application.Context) {
		app.Quit()
	})

	systray.SetMenu(trayMenu)
}
