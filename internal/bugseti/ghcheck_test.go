package bugseti

import (
	"os"
	"testing"
)

func TestCheckForge_Bad_MissingConfig(t *testing.T) {
	// Clear any env-based forge config to ensure CheckForge fails
	t.Setenv("FORGE_TOKEN", "")
	t.Setenv("FORGE_URL", "")

	// Point HOME to a temp dir so no config file is found
	t.Setenv("HOME", t.TempDir())
	if xdg := os.Getenv("XDG_CONFIG_HOME"); xdg != "" {
		t.Setenv("XDG_CONFIG_HOME", t.TempDir())
	}

	_, err := CheckForge()
	if err == nil {
		t.Fatal("expected error when forge is not configured")
	}
}
