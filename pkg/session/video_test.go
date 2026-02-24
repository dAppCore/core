package session

import (
	"strings"
	"testing"
	"time"
)

// ── RenderMP4 ──────────────────────────────────────────────────────

func TestRenderMP4_Bad_VHSNotInstalled(t *testing.T) {
	// Save PATH, set to empty so vhs won't be found.
	origPath := t.TempDir() // use empty dir as PATH
	t.Setenv("PATH", origPath)

	sess := &Session{ID: "test", StartTime: time.Now(), Events: []Event{}}
	err := RenderMP4(sess, "/tmp/test.mp4")
	if err == nil {
		t.Fatal("expected error when vhs not installed")
	}
	if !strings.Contains(err.Error(), "vhs not installed") {
		t.Fatalf("unexpected error: %s", err)
	}
}

// ── extractCommand ─────────────────────────────────────────────────

func TestExtractCommand_Good_WithDesc(t *testing.T) {
	got := extractCommand("go test ./... # run tests")
	if got != "go test ./..." {
		t.Fatalf("expected 'go test ./...', got %s", got)
	}
}

func TestExtractCommand_Good_NoDesc(t *testing.T) {
	got := extractCommand("ls -la")
	if got != "ls -la" {
		t.Fatalf("expected ls -la, got %s", got)
	}
}

func TestExtractCommand_Good_Empty(t *testing.T) {
	got := extractCommand("")
	if got != "" {
		t.Fatalf("expected empty, got %s", got)
	}
}

func TestExtractCommand_Good_HashInMiddle(t *testing.T) {
	got := extractCommand(`echo "hello # world" # desc`)
	if got != `echo "hello` {
		// Note: the simple split finds first " # " occurrence.
		t.Fatalf("unexpected: %s", got)
	}
}

// ── generateTape ───────────────────────────────────────────────────

func TestGenerateTape_Good_BasicSession(t *testing.T) {
	sess := &Session{
		ID:        "f3fb074c-8c72-4da6-a15a-85bae652ccaa",
		StartTime: time.Date(2026, 2, 24, 10, 0, 0, 0, time.UTC),
		Events: []Event{
			{
				Type: "tool_use", Tool: "Bash",
				Input: "go test ./...", Output: "ok 1.2s",
				Success: true,
			},
			{
				Type: "tool_use", Tool: "Read",
				Input: "/tmp/test.go",
			},
			{
				Type: "tool_use", Tool: "Task",
				Input: "[Explore] find tests",
			},
		},
	}

	tape := generateTape(sess, "/tmp/out.mp4")
	if !strings.Contains(tape, "Output /tmp/out.mp4") {
		t.Fatal("missing output directive")
	}
	if !strings.Contains(tape, "f3fb074c") {
		t.Fatal("missing session ID")
	}
	if !strings.Contains(tape, "go test ./...") {
		t.Fatal("missing bash command")
	}
	if !strings.Contains(tape, "Read: /tmp/test.go") {
		t.Fatal("missing Read event")
	}
	if !strings.Contains(tape, "Agent:") {
		t.Fatal("missing Task/Agent event")
	}
	if !strings.Contains(tape, "OK") {
		t.Fatal("missing success indicator")
	}
}

func TestGenerateTape_Good_FailedCommand(t *testing.T) {
	sess := &Session{
		ID:        "fail-test",
		StartTime: time.Now(),
		Events: []Event{
			{
				Type: "tool_use", Tool: "Bash",
				Input: "false", Output: "exit 1",
				Success: false,
			},
		},
	}

	tape := generateTape(sess, "/tmp/fail.mp4")
	if !strings.Contains(tape, "FAILED") {
		t.Fatal("missing FAILED indicator")
	}
}

func TestGenerateTape_Good_EmptySession(t *testing.T) {
	sess := &Session{ID: "empty", StartTime: time.Now()}
	tape := generateTape(sess, "/tmp/empty.mp4")
	if !strings.Contains(tape, "Output /tmp/empty.mp4") {
		t.Fatal("missing output directive")
	}
	if !strings.Contains(tape, "Sleep 3s") {
		t.Fatal("missing final sleep")
	}
}

func TestGenerateTape_Good_LongOutput(t *testing.T) {
	longOutput := strings.Repeat("x", 300)
	sess := &Session{
		ID:        "long",
		StartTime: time.Now(),
		Events: []Event{
			{
				Type: "tool_use", Tool: "Bash",
				Input: "cat bigfile", Output: longOutput,
				Success: true,
			},
		},
	}

	tape := generateTape(sess, "/tmp/long.mp4")
	// Output should be truncated to 200 chars + "..."
	if strings.Contains(tape, strings.Repeat("x", 300)) {
		t.Fatal("output not truncated")
	}
}

func TestGenerateTape_Good_SkipsNonToolEvents(t *testing.T) {
	sess := &Session{
		ID:        "mixed",
		StartTime: time.Now(),
		Events: []Event{
			{Type: "user", Input: "hello"},
			{Type: "assistant", Input: "hi there"},
			{Type: "tool_use", Tool: "Bash", Input: "echo hi", Output: "hi", Success: true},
		},
	}

	tape := generateTape(sess, "/tmp/mixed.mp4")
	if strings.Contains(tape, "hello") {
		t.Fatal("user message should be skipped")
	}
	if strings.Contains(tape, "hi there") {
		t.Fatal("assistant message should be skipped")
	}
	if !strings.Contains(tape, "echo hi") {
		t.Fatal("bash command should be present")
	}
}

func TestGenerateTape_Good_EmptyBashCommand(t *testing.T) {
	sess := &Session{
		ID:        "empty-cmd",
		StartTime: time.Now(),
		Events: []Event{
			{Type: "tool_use", Tool: "Bash", Input: "", Success: true},
		},
	}

	tape := generateTape(sess, "/tmp/empty-cmd.mp4")
	// Empty command should be skipped — no "$ " line.
	if strings.Contains(tape, "$ ") {
		t.Fatal("empty command should be skipped")
	}
}
