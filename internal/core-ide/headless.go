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

	"github.com/host-uk/core/pkg/cli"
	"github.com/host-uk/core/pkg/jobrunner"
	"github.com/host-uk/core/pkg/jobrunner/github"
	"github.com/host-uk/core/pkg/jobrunner/handlers"
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

	// TODO: Updater integration — the internal/cmd/updater package cannot be
	// imported from the core-ide module due to Go's internal package restriction
	// (separate modules). Move updater to pkg/updater or export a public API to
	// enable auto-update in headless mode.

	// Journal
	journalDir := filepath.Join(os.Getenv("HOME"), ".core", "journal")
	journal, err := jobrunner.NewJournal(journalDir)
	if err != nil {
		log.Fatalf("Failed to create journal: %v", err)
	}

	// GitHub source — repos from CORE_REPOS env var or default
	repos := parseRepoList(os.Getenv("CORE_REPOS"))
	if len(repos) == 0 {
		repos = []string{"host-uk/core", "host-uk/core-php", "host-uk/core-tenant", "host-uk/core-admin"}
	}

	ghSource := github.NewGitHubSource(github.Config{
		Repos: repos,
	})

	// Handlers (order matters — first match wins)
	publishDraft := handlers.NewPublishDraftHandler(nil, "")
	sendFix := handlers.NewSendFixCommandHandler(nil, "")
	resolveThreads := handlers.NewResolveThreadsHandler(nil, "")
	enableAutoMerge := handlers.NewEnableAutoMergeHandler()
	tickParent := handlers.NewTickParentHandler()

	// Build poller
	poller := jobrunner.NewPoller(jobrunner.PollerConfig{
		Sources: []jobrunner.JobSource{ghSource},
		Handlers: []jobrunner.JobHandler{
			publishDraft,
			sendFix,
			resolveThreads,
			enableAutoMerge,
			tickParent,
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
