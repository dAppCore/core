package main

import (
	"fmt"
	"testing"
	"time"

	"github.com/host-uk/core/internal/bugseti"
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

	// Fill beyond the cap with fresh entries.
	svc.mu.Lock()
	for i := 0; i < maxWorkspaces+20; i++ {
		svc.workspaces[fmt.Sprintf("ws-%d", i)] = &Workspace{
			CreatedAt: time.Now().Add(-time.Duration(i) * time.Minute),
		}
	}
	svc.cleanup()
	svc.mu.Unlock()

	if got := svc.ActiveWorkspaces(); got != maxWorkspaces {
		t.Errorf("expected %d workspaces after cap cleanup, got %d", maxWorkspaces, got)
	}
}

func TestCleanup_EvictsOldestWhenOverCap(t *testing.T) {
	svc := NewWorkspaceService(bugseti.NewConfigService())

	// Create maxWorkspaces+1 entries; the newest should survive.
	svc.mu.Lock()
	for i := 0; i <= maxWorkspaces; i++ {
		svc.workspaces[fmt.Sprintf("ws-%d", i)] = &Workspace{
			CreatedAt: time.Now().Add(-time.Duration(maxWorkspaces-i) * time.Minute),
		}
	}
	svc.cleanup()
	svc.mu.Unlock()

	// The newest entry (ws-<maxWorkspaces>) should still exist.
	newest := fmt.Sprintf("ws-%d", maxWorkspaces)
	if m := svc.GetMedium(newest); m != nil {
		// GetMedium returns nil for entries with nil Medium, which is expected here.
		// We just want to verify the key still exists.
	}

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
