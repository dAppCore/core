package infra

import (
	"os"
	"path/filepath"
	"testing"
)

func TestLoad_Good(t *testing.T) {
	// Find infra.yaml relative to test
	// Walk up from test dir to find it
	dir, err := os.Getwd()
	if err != nil {
		t.Fatal(err)
	}

	cfg, path, err := Discover(dir)
	if err != nil {
		t.Skipf("infra.yaml not found from %s: %v", dir, err)
	}

	t.Logf("Loaded %s", path)

	if len(cfg.Hosts) == 0 {
		t.Error("expected at least one host")
	}

	// Check required hosts exist
	for _, name := range []string{"noc", "de", "de2", "build"} {
		if _, ok := cfg.Hosts[name]; !ok {
			t.Errorf("expected host %q in config", name)
		}
	}

	// Check de host details
	de := cfg.Hosts["de"]
	if de.IP != "116.202.82.115" {
		t.Errorf("de IP = %q, want 116.202.82.115", de.IP)
	}
	if de.Role != "app" {
		t.Errorf("de role = %q, want app", de.Role)
	}

	// Check LB config
	if cfg.LoadBalancer.Name != "hermes" {
		t.Errorf("LB name = %q, want hermes", cfg.LoadBalancer.Name)
	}
	if cfg.LoadBalancer.Type != "lb11" {
		t.Errorf("LB type = %q, want lb11", cfg.LoadBalancer.Type)
	}
	if len(cfg.LoadBalancer.Backends) != 2 {
		t.Errorf("LB backends = %d, want 2", len(cfg.LoadBalancer.Backends))
	}

	// Check app servers helper
	apps := cfg.AppServers()
	if len(apps) != 2 {
		t.Errorf("AppServers() = %d, want 2", len(apps))
	}
}

func TestLoad_Bad(t *testing.T) {
	_, err := Load("/nonexistent/infra.yaml")
	if err == nil {
		t.Error("expected error for nonexistent file")
	}
}

func TestLoad_Ugly(t *testing.T) {
	// Invalid YAML
	tmp := filepath.Join(t.TempDir(), "infra.yaml")
	if err := os.WriteFile(tmp, []byte("{{invalid yaml"), 0644); err != nil {
		t.Fatal(err)
	}

	_, err := Load(tmp)
	if err == nil {
		t.Error("expected error for invalid YAML")
	}
}

func TestExpandPath(t *testing.T) {
	home, _ := os.UserHomeDir()

	tests := []struct {
		input string
		want  string
	}{
		{"~/.ssh/id_rsa", filepath.Join(home, ".ssh/id_rsa")},
		{"/absolute/path", "/absolute/path"},
		{"relative/path", "relative/path"},
	}

	for _, tt := range tests {
		got := expandPath(tt.input)
		if got != tt.want {
			t.Errorf("expandPath(%q) = %q, want %q", tt.input, got, tt.want)
		}
	}
}
