package session

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

func TestRenderHTML_Good_BasicSession(t *testing.T) {
	dir := t.TempDir()
	out := filepath.Join(dir, "session.html")

	sess := &Session{
		ID:        "f3fb074c-8c72-4da6-a15a-85bae652ccaa",
		StartTime: time.Date(2026, 2, 24, 10, 0, 0, 0, time.UTC),
		EndTime:   time.Date(2026, 2, 24, 10, 5, 0, 0, time.UTC),
		Events: []Event{
			{
				Timestamp: time.Date(2026, 2, 24, 10, 0, 5, 0, time.UTC),
				Type:      "tool_use",
				Tool:      "Bash",
				Input:     "go test ./...",
				Output:    "ok  forge.lthn.ai/core/go  1.2s",
				Duration:  time.Second,
				Success:   true,
			},
			{
				Timestamp: time.Date(2026, 2, 24, 10, 1, 0, 0, time.UTC),
				Type:      "tool_use",
				Tool:      "Read",
				Input:     "/tmp/test.go",
				Output:    "package main",
				Duration:  200 * time.Millisecond,
				Success:   true,
			},
			{
				Timestamp: time.Date(2026, 2, 24, 10, 2, 0, 0, time.UTC),
				Type:      "user",
				Input:     "looks good",
			},
		},
	}

	if err := RenderHTML(sess, out); err != nil {
		t.Fatal(err)
	}

	data, err := os.ReadFile(out)
	if err != nil {
		t.Fatal(err)
	}

	html := string(data)
	if !strings.Contains(html, "f3fb074c") {
		t.Fatal("missing session ID")
	}
	if !strings.Contains(html, "go test ./...") {
		t.Fatal("missing bash command")
	}
	if !strings.Contains(html, "2 tool calls") {
		t.Fatal("missing tool count")
	}
	if !strings.Contains(html, "filterEvents") {
		t.Fatal("missing JS filter function")
	}
}

func TestRenderHTML_Good_WithErrors(t *testing.T) {
	dir := t.TempDir()
	out := filepath.Join(dir, "errors.html")

	sess := &Session{
		ID:        "err-session",
		StartTime: time.Date(2026, 2, 24, 10, 0, 0, 0, time.UTC),
		EndTime:   time.Date(2026, 2, 24, 10, 1, 0, 0, time.UTC),
		Events: []Event{
			{
				Type: "tool_use", Tool: "Bash",
				Timestamp: time.Date(2026, 2, 24, 10, 0, 0, 0, time.UTC),
				Input: "bad command", Output: "error", Success: false,
			},
		},
	}

	if err := RenderHTML(sess, out); err != nil {
		t.Fatal(err)
	}

	data, _ := os.ReadFile(out)
	html := string(data)
	if !strings.Contains(html, "1 errors") {
		t.Fatal("missing error count")
	}
	if !strings.Contains(html, `class="event error"`) {
		t.Fatal("missing error class")
	}
	if !strings.Contains(html, "&#10007;") {
		t.Fatal("missing failure icon")
	}
}

func TestRenderHTML_Good_AssistantEvent(t *testing.T) {
	dir := t.TempDir()
	out := filepath.Join(dir, "asst.html")

	sess := &Session{
		ID:        "asst-test",
		StartTime: time.Date(2026, 2, 24, 10, 0, 0, 0, time.UTC),
		EndTime:   time.Date(2026, 2, 24, 10, 0, 5, 0, time.UTC),
		Events: []Event{
			{
				Type:      "assistant",
				Timestamp: time.Date(2026, 2, 24, 10, 0, 0, 0, time.UTC),
				Input:     "Let me check that.",
			},
		},
	}

	if err := RenderHTML(sess, out); err != nil {
		t.Fatal(err)
	}

	data, _ := os.ReadFile(out)
	if !strings.Contains(string(data), "Claude") {
		t.Fatal("missing Claude label for assistant")
	}
}

func TestRenderHTML_Good_EmptySession(t *testing.T) {
	dir := t.TempDir()
	out := filepath.Join(dir, "empty.html")

	sess := &Session{
		ID:        "empty",
		StartTime: time.Now(),
		EndTime:   time.Now(),
	}

	if err := RenderHTML(sess, out); err != nil {
		t.Fatal(err)
	}

	info, err := os.Stat(out)
	if err != nil {
		t.Fatal(err)
	}
	if info.Size() == 0 {
		t.Fatal("HTML file is empty")
	}
}

func TestRenderHTML_Bad_InvalidPath(t *testing.T) {
	sess := &Session{ID: "test", StartTime: time.Now(), EndTime: time.Now()}
	err := RenderHTML(sess, "/nonexistent/dir/out.html")
	if err == nil {
		t.Fatal("expected error for invalid path")
	}
}

func TestRenderHTML_Good_XSSEscaping(t *testing.T) {
	dir := t.TempDir()
	out := filepath.Join(dir, "xss.html")

	sess := &Session{
		ID:        "xss-test",
		StartTime: time.Now(),
		EndTime:   time.Now(),
		Events: []Event{
			{
				Type:      "tool_use",
				Tool:      "Bash",
				Timestamp: time.Now(),
				Input:     `echo "<script>alert('xss')</script>"`,
				Output:    `<img onerror=alert(1)>`,
				Success:   true,
			},
		},
	}

	if err := RenderHTML(sess, out); err != nil {
		t.Fatal(err)
	}

	data, _ := os.ReadFile(out)
	html := string(data)
	if strings.Contains(html, "<script>alert") {
		t.Fatal("XSS: unescaped script tag in HTML output")
	}
	if strings.Contains(html, "<img onerror") {
		t.Fatal("XSS: unescaped img tag in HTML output")
	}
}
