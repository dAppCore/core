package ratelimit

import (
	"context"
	"path/filepath"
	"testing"
	"time"
)

func TestCanSend_Good(t *testing.T) {
	rl, _ := New()
	rl.filePath = filepath.Join(t.TempDir(), "ratelimits.yaml")

	model := "test-model"
	rl.Quotas[model] = ModelQuota{MaxRPM: 10, MaxTPM: 1000, MaxRPD: 100}

	if !rl.CanSend(model, 100) {
		t.Errorf("Expected CanSend to return true for fresh state")
	}
}

func TestCanSend_RPMExceeded_Bad(t *testing.T) {
	rl, _ := New()
	model := "test-rpm"
	rl.Quotas[model] = ModelQuota{MaxRPM: 2, MaxTPM: 1000000, MaxRPD: 100}

	rl.RecordUsage(model, 10, 10)
	rl.RecordUsage(model, 10, 10)

	if rl.CanSend(model, 10) {
		t.Errorf("Expected CanSend to return false after exceeding RPM")
	}
}

func TestCanSend_TPMExceeded_Bad(t *testing.T) {
	rl, _ := New()
	model := "test-tpm"
	rl.Quotas[model] = ModelQuota{MaxRPM: 10, MaxTPM: 100, MaxRPD: 100}

	rl.RecordUsage(model, 50, 40) // 90 tokens used

	if rl.CanSend(model, 20) { // 90 + 20 = 110 > 100
		t.Errorf("Expected CanSend to return false when estimated tokens exceed TPM")
	}
}

func TestCanSend_RPDExceeded_Bad(t *testing.T) {
	rl, _ := New()
	model := "test-rpd"
	rl.Quotas[model] = ModelQuota{MaxRPM: 10, MaxTPM: 1000000, MaxRPD: 2}

	rl.RecordUsage(model, 10, 10)
	rl.RecordUsage(model, 10, 10)

	if rl.CanSend(model, 10) {
		t.Errorf("Expected CanSend to return false after exceeding RPD")
	}
}

func TestCanSend_UnlimitedModel_Good(t *testing.T) {
	rl, _ := New()
	model := "test-unlimited"
	rl.Quotas[model] = ModelQuota{MaxRPM: 0, MaxTPM: 0, MaxRPD: 0}

	// Should always be allowed
	for i := 0; i < 1000; i++ {
		rl.RecordUsage(model, 100, 100)
	}
	if !rl.CanSend(model, 999999) {
		t.Errorf("Expected unlimited model to always allow sends")
	}
}

func TestRecordUsage_PrunesOldEntries_Good(t *testing.T) {
	rl, _ := New()
	model := "test-prune"
	rl.Quotas[model] = ModelQuota{MaxRPM: 5, MaxTPM: 1000000, MaxRPD: 100}

	// Manually inject old data
	oldTime := time.Now().Add(-2 * time.Minute)
	rl.State[model] = &UsageStats{
		Requests: []time.Time{oldTime, oldTime, oldTime},
		Tokens: []TokenEntry{
			{Time: oldTime, Count: 100},
			{Time: oldTime, Count: 100},
		},
		DayStart: time.Now(),
	}

	// CanSend triggers prune
	if !rl.CanSend(model, 10) {
		t.Errorf("Expected CanSend to return true after pruning old entries")
	}

	stats := rl.State[model]
	if len(stats.Requests) != 0 {
		t.Errorf("Expected 0 requests after pruning old entries, got %d", len(stats.Requests))
	}
}

func TestPersistAndLoad_Good(t *testing.T) {
	tmpDir := t.TempDir()
	path := filepath.Join(tmpDir, "ratelimits.yaml")

	rl1, _ := New()
	rl1.filePath = path
	model := "persist-test"
	rl1.Quotas[model] = ModelQuota{MaxRPM: 50, MaxTPM: 5000, MaxRPD: 500}
	rl1.RecordUsage(model, 100, 100)

	if err := rl1.Persist(); err != nil {
		t.Fatalf("Persist failed: %v", err)
	}

	rl2, _ := New()
	rl2.filePath = path
	if err := rl2.Load(); err != nil {
		t.Fatalf("Load failed: %v", err)
	}

	stats := rl2.Stats(model)
	if stats.RPM != 1 {
		t.Errorf("Expected RPM 1 after load, got %d", stats.RPM)
	}
	if stats.TPM != 200 {
		t.Errorf("Expected TPM 200 after load, got %d", stats.TPM)
	}
}

func TestWaitForCapacity_Ugly(t *testing.T) {
	rl, _ := New()
	model := "wait-test"
	rl.Quotas[model] = ModelQuota{MaxRPM: 1, MaxTPM: 1000000, MaxRPD: 100}

	rl.RecordUsage(model, 10, 10) // Use up the 1 RPM

	ctx, cancel := context.WithTimeout(context.Background(), 100*time.Millisecond)
	defer cancel()

	err := rl.WaitForCapacity(ctx, model, 10)
	if err != context.DeadlineExceeded {
		t.Errorf("Expected DeadlineExceeded, got %v", err)
	}
}

func TestDefaultQuotas_Good(t *testing.T) {
	rl, _ := New()
	expected := []string{
		"gemini-3-pro-preview",
		"gemini-3-flash-preview",
		"gemini-2.0-flash",
	}
	for _, m := range expected {
		if _, ok := rl.Quotas[m]; !ok {
			t.Errorf("Expected default quota for %s", m)
		}
	}
}

func TestAllStats_Good(t *testing.T) {
	rl, _ := New()
	rl.RecordUsage("gemini-3-pro-preview", 1000, 500)

	all := rl.AllStats()
	if len(all) < 5 {
		t.Errorf("Expected at least 5 models in AllStats, got %d", len(all))
	}

	pro := all["gemini-3-pro-preview"]
	if pro.RPM != 1 {
		t.Errorf("Expected RPM 1 for pro, got %d", pro.RPM)
	}
	if pro.TPM != 1500 {
		t.Errorf("Expected TPM 1500 for pro, got %d", pro.TPM)
	}
}
