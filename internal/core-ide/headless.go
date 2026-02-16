package main

import (
	"context"
	"log"
	"os"
	"os/signal"
	"path/filepath"
	"runtime"
	"strings"
	"syscall"
	"time"

	"forge.lthn.ai/core/cli/pkg/agentci"
	"forge.lthn.ai/core/cli/pkg/cli"
	"forge.lthn.ai/core/cli/pkg/config"
	"forge.lthn.ai/core/cli/pkg/forge"
	"forge.lthn.ai/core/cli/pkg/jobrunner"
	forgejosource "forge.lthn.ai/core/cli/pkg/jobrunner/forgejo"
	"forge.lthn.ai/core/cli/pkg/jobrunner/handlers"
)

// hasDisplay returns true if a graphical display is available.
func hasDisplay() bool {
	if runtime.GOOS == "windows" {
		return true
	}
	return os.Getenv("DISPLAY") != "" || os.Getenv("WAYLAND_DISPLAY") != ""
}

// startHeadless runs the job runner in daemon mode without GUI.
func startHeadless() {
	log.Println("Starting Core IDE in headless mode...")

	// Signal handling
	ctx, cancel := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer cancel()

	// Journal
	journalDir := filepath.Join(os.Getenv("HOME"), ".core", "journal")
	journal, err := jobrunner.NewJournal(journalDir)
	if err != nil {
		log.Fatalf("Failed to create journal: %v", err)
	}

	// Forge client
	forgeURL, forgeToken, _ := forge.ResolveConfig("", "")
	forgeClient, err := forge.New(forgeURL, forgeToken)
	if err != nil {
		log.Fatalf("Failed to create forge client: %v", err)
	}

	// Forgejo source — repos from CORE_REPOS env var or default
	repos := parseRepoList(os.Getenv("CORE_REPOS"))
	if len(repos) == 0 {
		repos = []string{"host-uk/core", "host-uk/core-php", "host-uk/core-tenant", "host-uk/core-admin"}
	}

	source := forgejosource.New(forgejosource.Config{
		Repos: repos,
	}, forgeClient)

	// Handlers (order matters — first match wins)
	publishDraft := handlers.NewPublishDraftHandler(forgeClient)
	sendFix := handlers.NewSendFixCommandHandler(forgeClient)
	dismissReviews := handlers.NewDismissReviewsHandler(forgeClient)
	enableAutoMerge := handlers.NewEnableAutoMergeHandler(forgeClient)
	tickParent := handlers.NewTickParentHandler(forgeClient)

	// Agent dispatch — Clotho integration
	cfg, cfgErr := config.New()
	var agentTargets map[string]agentci.AgentConfig
	var clothoCfg agentci.ClothoConfig

	if cfgErr == nil {
		agentTargets, _ = agentci.LoadActiveAgents(cfg)
		clothoCfg, _ = agentci.LoadClothoConfig(cfg)
	}
	if agentTargets == nil {
		agentTargets = map[string]agentci.AgentConfig{}
	}

	spinner := agentci.NewSpinner(clothoCfg, agentTargets)
	log.Printf("Loaded %d agent targets. Strategy: %s", len(agentTargets), clothoCfg.Strategy)

	dispatch := handlers.NewDispatchHandler(forgeClient, forgeURL, forgeToken, spinner)

	// Build poller
	poller := jobrunner.NewPoller(jobrunner.PollerConfig{
		Sources: []jobrunner.JobSource{source},
		Handlers: []jobrunner.JobHandler{
			publishDraft,
			sendFix,
			dismissReviews,
			enableAutoMerge,
			tickParent,
			dispatch, // Last — only matches NeedsCoding signals
		},
		Journal:      journal,
		PollInterval: 60 * time.Second,
		DryRun:       isDryRun(),
	})

	// Daemon with PID file and health check
	daemon := cli.NewDaemon(cli.DaemonOptions{
		PIDFile:    filepath.Join(os.Getenv("HOME"), ".core", "core-ide.pid"),
		HealthAddr: "127.0.0.1:9878",
	})

	if err := daemon.Start(); err != nil {
		log.Fatalf("Failed to start daemon: %v", err)
	}
	daemon.SetReady(true)

	// Start MCP bridge in headless mode too (port 9877)
	go startHeadlessMCP(poller)

	log.Printf("Polling %d repos every %s (dry-run: %v)", len(repos), "60s", poller.DryRun())

	// Run poller in goroutine, block on context
	go func() {
		if err := poller.Run(ctx); err != nil && err != context.Canceled {
			log.Printf("Poller error: %v", err)
		}
	}()

	// Block until signal
	<-ctx.Done()
	log.Println("Shutting down...")
	_ = daemon.Stop()
}

// parseRepoList splits a comma-separated repo list.
func parseRepoList(s string) []string {
	if s == "" {
		return nil
	}
	var repos []string
	for _, r := range strings.Split(s, ",") {
		r = strings.TrimSpace(r)
		if r != "" {
			repos = append(repos, r)
		}
	}
	return repos
}

// isDryRun checks if --dry-run flag was passed.
func isDryRun() bool {
	for _, arg := range os.Args[1:] {
		if arg == "--dry-run" {
			return true
		}
	}
	return false
}
