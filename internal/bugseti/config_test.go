package bugseti

import (
	"os"
	"testing"
)

func TestConfigPermissions(t *testing.T) {
	// Get a temporary file path
	f, err := os.CreateTemp("", "bugseti-config-*.json")
	if err != nil {
		t.Fatal(err)
	}
	name := f.Name()
	f.Close()
	os.Remove(name) // Ensure it doesn't exist
	defer os.Remove(name)

	c := &ConfigService{
		path: name,
		config: &Config{},
	}

	if err := c.Save(); err != nil {
		t.Fatalf("Save failed: %v", err)
	}

	info, err := os.Stat(name)
	if err != nil {
		t.Fatal(err)
	}

	mode := info.Mode().Perm()
	if mode != 0600 {
		t.Errorf("expected file permissions 0600, got %04o", mode)
	}
}
