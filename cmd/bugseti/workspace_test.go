package main

import (
	"fmt"
	"testing"
	"time"

	"forge.lthn.ai/core/cli/internal/bugseti"
)

func TestCleanup_TTL(t *testing.T) {
	svc := NewWorkspaceService(bugseti.NewConfigService())

	// Seed with entries that are older than TTL.
	svc.mu.Lock()
	for i := 0; i < 5; i++ {
		svc.workspaces[fmt.Sprintf("old-%d", i)] = &Workspace{
			CreatedAt: time.Now().Add(-25 * time.Hour),
		}
	}
	// Add one fresh entry.
	svc.workspaces["fresh"] = &Workspace{
		CreatedAt: time.Now(),
	}
	svc.cleanup()
	svc.mu.Unlock()

	if got := svc.ActiveWorkspaces(); got != 1 {
		t.Errorf("expected 1 workspace after TTL cleanup, got %d", got)
	}
}

func TestCleanup_MaxSize(t *testing.T) {
	svc := NewWorkspaceService(bugseti.NewConfigService())

	maxCap := svc.maxCap()

	// Fill beyond the cap with fresh entries.
	svc.mu.Lock()
	for i := 0; i < maxCap+20; i++ {
		svc.workspaces[fmt.Sprintf("ws-%d", i)] = &Workspace{
			CreatedAt: time.Now().Add(-time.Duration(i) * time.Minute),
		}
	}
	svc.cleanup()
	svc.mu.Unlock()

	if got := svc.ActiveWorkspaces(); got != maxCap {
		t.Errorf("expected %d workspaces after cap cleanup, got %d", maxCap, got)
	}
}

func TestCleanup_EvictsOldestWhenOverCap(t *testing.T) {
	svc := NewWorkspaceService(bugseti.NewConfigService())

	maxCap := svc.maxCap()

	// Create maxCap+1 entries; the newest should survive.
	svc.mu.Lock()
	for i := 0; i <= maxCap; i++ {
		svc.workspaces[fmt.Sprintf("ws-%d", i)] = &Workspace{
			CreatedAt: time.Now().Add(-time.Duration(maxCap-i) * time.Minute),
		}
	}
	svc.cleanup()
	svc.mu.Unlock()

	// The newest entry (ws-<maxCap>) should still exist.
	newest := fmt.Sprintf("ws-%d", maxCap)

	svc.mu.RLock()
	_, exists := svc.workspaces[newest]
	svc.mu.RUnlock()
	if !exists {
		t.Error("expected newest workspace to survive eviction")
	}

	// The oldest entry (ws-0) should have been evicted.
	svc.mu.RLock()
	_, exists = svc.workspaces["ws-0"]
	svc.mu.RUnlock()
	if exists {
		t.Error("expected oldest workspace to be evicted")
	}
}

func TestCleanup_ReturnsEvictedCount(t *testing.T) {
	svc := NewWorkspaceService(bugseti.NewConfigService())

	svc.mu.Lock()
	for i := 0; i < 3; i++ {
		svc.workspaces[fmt.Sprintf("old-%d", i)] = &Workspace{
			CreatedAt: time.Now().Add(-25 * time.Hour),
		}
	}
	svc.workspaces["fresh"] = &Workspace{
		CreatedAt: time.Now(),
	}
	evicted := svc.cleanup()
	svc.mu.Unlock()

	if evicted != 3 {
		t.Errorf("expected 3 evicted entries, got %d", evicted)
	}
}

func TestStartStop(t *testing.T) {
	svc := NewWorkspaceService(bugseti.NewConfigService())
	svc.Start()

	// Add a stale entry while the sweeper is running.
	svc.mu.Lock()
	svc.workspaces["stale"] = &Workspace{
		CreatedAt: time.Now().Add(-25 * time.Hour),
	}
	svc.mu.Unlock()

	// Stop should return without hanging.
	svc.Stop()
}

func TestConfigurableTTL(t *testing.T) {
	cfg := bugseti.NewConfigService()
	svc := NewWorkspaceService(cfg)

	// Default TTL should be 24h (1440 minutes).
	if got := svc.ttl(); got != 24*time.Hour {
		t.Errorf("expected default TTL of 24h, got %s", got)
	}

	// Default max cap should be 100.
	if got := svc.maxCap(); got != 100 {
		t.Errorf("expected default max cap of 100, got %d", got)
	}
}

func TestNilConfigFallback(t *testing.T) {
	svc := &WorkspaceService{
		config:     nil,
		workspaces: make(map[string]*Workspace),
		done:       make(chan struct{}),
		stopped:    make(chan struct{}),
	}

	if got := svc.ttl(); got != defaultWorkspaceTTL {
		t.Errorf("expected fallback TTL %s, got %s", defaultWorkspaceTTL, got)
	}
	if got := svc.maxCap(); got != defaultMaxWorkspaces {
		t.Errorf("expected fallback max cap %d, got %d", defaultMaxWorkspaces, got)
	}
}
