package lab

import (
	"os"
	"testing"
)

// ── LoadConfig defaults ────────────────────────────────────────────

func TestLoadConfig_Good_Defaults(t *testing.T) {
	cfg := LoadConfig()

	if cfg.Addr != ":8080" {
		t.Fatalf("expected :8080, got %s", cfg.Addr)
	}
	if cfg.PrometheusURL != "http://prometheus:9090" {
		t.Fatalf("unexpected PrometheusURL: %s", cfg.PrometheusURL)
	}
	if cfg.PrometheusInterval != 15 {
		t.Fatalf("expected 15, got %d", cfg.PrometheusInterval)
	}
	if cfg.ForgeURL != "https://forge.lthn.io" {
		t.Fatalf("unexpected ForgeURL: %s", cfg.ForgeURL)
	}
	if cfg.ForgeInterval != 60 {
		t.Fatalf("expected 60, got %d", cfg.ForgeInterval)
	}
	if cfg.HFAuthor != "lthn" {
		t.Fatalf("expected lthn, got %s", cfg.HFAuthor)
	}
	if cfg.HFInterval != 300 {
		t.Fatalf("expected 300, got %d", cfg.HFInterval)
	}
	if cfg.TrainingDataDir != "/data/training" {
		t.Fatalf("unexpected TrainingDataDir: %s", cfg.TrainingDataDir)
	}
	if cfg.InfluxDB != "training" {
		t.Fatalf("expected training, got %s", cfg.InfluxDB)
	}
}

// ── env override ───────────────────────────────────────────────────

func TestLoadConfig_Good_EnvOverride(t *testing.T) {
	os.Setenv("ADDR", ":9090")
	os.Setenv("FORGE_URL", "https://forge.lthn.ai")
	os.Setenv("HF_AUTHOR", "snider")
	defer func() {
		os.Unsetenv("ADDR")
		os.Unsetenv("FORGE_URL")
		os.Unsetenv("HF_AUTHOR")
	}()

	cfg := LoadConfig()
	if cfg.Addr != ":9090" {
		t.Fatalf("expected :9090, got %s", cfg.Addr)
	}
	if cfg.ForgeURL != "https://forge.lthn.ai" {
		t.Fatalf("expected forge.lthn.ai, got %s", cfg.ForgeURL)
	}
	if cfg.HFAuthor != "snider" {
		t.Fatalf("expected snider, got %s", cfg.HFAuthor)
	}
}

// ── envInt ─────────────────────────────────────────────────────────

func TestLoadConfig_Good_IntEnvOverride(t *testing.T) {
	os.Setenv("PROMETHEUS_INTERVAL", "30")
	defer os.Unsetenv("PROMETHEUS_INTERVAL")

	cfg := LoadConfig()
	if cfg.PrometheusInterval != 30 {
		t.Fatalf("expected 30, got %d", cfg.PrometheusInterval)
	}
}

func TestLoadConfig_Bad_InvalidIntFallsBack(t *testing.T) {
	os.Setenv("PROMETHEUS_INTERVAL", "not-a-number")
	defer os.Unsetenv("PROMETHEUS_INTERVAL")

	cfg := LoadConfig()
	if cfg.PrometheusInterval != 15 {
		t.Fatalf("expected fallback 15, got %d", cfg.PrometheusInterval)
	}
}

// ── env / envInt helpers directly ──────────────────────────────────

func TestEnv_Good(t *testing.T) {
	os.Setenv("TEST_LAB_KEY", "hello")
	defer os.Unsetenv("TEST_LAB_KEY")

	if got := env("TEST_LAB_KEY", "default"); got != "hello" {
		t.Fatalf("expected hello, got %s", got)
	}
}

func TestEnv_Good_Fallback(t *testing.T) {
	os.Unsetenv("TEST_LAB_MISSING")
	if got := env("TEST_LAB_MISSING", "fallback"); got != "fallback" {
		t.Fatalf("expected fallback, got %s", got)
	}
}

func TestEnvInt_Good(t *testing.T) {
	os.Setenv("TEST_LAB_INT", "42")
	defer os.Unsetenv("TEST_LAB_INT")

	if got := envInt("TEST_LAB_INT", 0); got != 42 {
		t.Fatalf("expected 42, got %d", got)
	}
}

func TestEnvInt_Bad_Fallback(t *testing.T) {
	os.Unsetenv("TEST_LAB_INT_MISSING")
	if got := envInt("TEST_LAB_INT_MISSING", 99); got != 99 {
		t.Fatalf("expected 99, got %d", got)
	}
}

func TestEnvInt_Bad_InvalidString(t *testing.T) {
	os.Setenv("TEST_LAB_INT_BAD", "xyz")
	defer os.Unsetenv("TEST_LAB_INT_BAD")

	if got := envInt("TEST_LAB_INT_BAD", 7); got != 7 {
		t.Fatalf("expected fallback 7, got %d", got)
	}
}
