package bugseti

import (
	"os/exec"
	"strings"
	"testing"
)

func TestCheckGHCLI_Good(t *testing.T) {
	// Only run if gh is actually available (CI-friendly skip)
	if _, err := exec.LookPath("gh"); err != nil {
		t.Skip("gh CLI not installed, skipping")
	}

	err := CheckGHCLI()
	// We can't guarantee auth status in all environments,
	// but if gh is present the function should at least not panic.
	if err != nil {
		t.Logf("CheckGHCLI returned error (may be expected if not authenticated): %v", err)
	}
}

func TestCheckGHCLI_Bad_MissingBinary(t *testing.T) {
	// Save and clear PATH to simulate missing gh
	t.Setenv("PATH", t.TempDir())

	err := CheckGHCLI()
	if err == nil {
		t.Fatal("expected error when gh is not in PATH")
	}
	if !strings.Contains(err.Error(), "gh CLI not found") {
		t.Errorf("error should mention 'gh CLI not found', got: %v", err)
	}
	if !strings.Contains(err.Error(), "https://cli.github.com") {
		t.Errorf("error should include install URL, got: %v", err)
	}
}
