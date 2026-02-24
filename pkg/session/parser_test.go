package session

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"
	"time"
)

// ── truncate ───────────────────────────────────────────────────────

func TestTruncate_Good_Short(t *testing.T) {
	if got := truncate("hello", 10); got != "hello" {
		t.Fatalf("expected hello, got %s", got)
	}
}

func TestTruncate_Good_Exact(t *testing.T) {
	if got := truncate("12345", 5); got != "12345" {
		t.Fatalf("expected 12345, got %s", got)
	}
}

func TestTruncate_Good_Long(t *testing.T) {
	got := truncate("hello world", 5)
	if got != "hello..." {
		t.Fatalf("expected hello..., got %s", got)
	}
}

func TestTruncate_Good_Empty(t *testing.T) {
	if got := truncate("", 10); got != "" {
		t.Fatalf("expected empty, got %s", got)
	}
}

// ── shortID ────────────────────────────────────────────────────────

func TestShortID_Good_Long(t *testing.T) {
	got := shortID("f3fb074c-8c72-4da6-a15a-85bae652ccaa")
	if got != "f3fb074c" {
		t.Fatalf("expected f3fb074c, got %s", got)
	}
}

func TestShortID_Good_Short(t *testing.T) {
	if got := shortID("abc"); got != "abc" {
		t.Fatalf("expected abc, got %s", got)
	}
}

func TestShortID_Good_ExactEight(t *testing.T) {
	if got := shortID("12345678"); got != "12345678" {
		t.Fatalf("expected 12345678, got %s", got)
	}
}

// ── formatDuration ─────────────────────────────────────────────────

func TestFormatDuration_Good_Milliseconds(t *testing.T) {
	got := formatDuration(500 * time.Millisecond)
	if got != "500ms" {
		t.Fatalf("expected 500ms, got %s", got)
	}
}

func TestFormatDuration_Good_Seconds(t *testing.T) {
	got := formatDuration(3500 * time.Millisecond)
	if got != "3.5s" {
		t.Fatalf("expected 3.5s, got %s", got)
	}
}

func TestFormatDuration_Good_Minutes(t *testing.T) {
	got := formatDuration(2*time.Minute + 30*time.Second)
	if got != "2m30s" {
		t.Fatalf("expected 2m30s, got %s", got)
	}
}

func TestFormatDuration_Good_Hours(t *testing.T) {
	got := formatDuration(1*time.Hour + 15*time.Minute)
	if got != "1h15m" {
		t.Fatalf("expected 1h15m, got %s", got)
	}
}

// ── extractToolInput ───────────────────────────────────────────────

func TestExtractToolInput_Good_Bash(t *testing.T) {
	raw := json.RawMessage(`{"command":"go test ./...","description":"run tests"}`)
	got := extractToolInput("Bash", raw)
	if got != "go test ./... # run tests" {
		t.Fatalf("unexpected: %s", got)
	}
}

func TestExtractToolInput_Good_BashNoDesc(t *testing.T) {
	raw := json.RawMessage(`{"command":"ls"}`)
	got := extractToolInput("Bash", raw)
	if got != "ls" {
		t.Fatalf("expected ls, got %s", got)
	}
}

func TestExtractToolInput_Good_Read(t *testing.T) {
	raw := json.RawMessage(`{"file_path":"/tmp/test.go"}`)
	got := extractToolInput("Read", raw)
	if got != "/tmp/test.go" {
		t.Fatalf("expected /tmp/test.go, got %s", got)
	}
}

func TestExtractToolInput_Good_Edit(t *testing.T) {
	raw := json.RawMessage(`{"file_path":"/tmp/test.go","old_string":"foo","new_string":"bar"}`)
	got := extractToolInput("Edit", raw)
	if got != "/tmp/test.go (edit)" {
		t.Fatalf("expected /tmp/test.go (edit), got %s", got)
	}
}

func TestExtractToolInput_Good_Write(t *testing.T) {
	raw := json.RawMessage(`{"file_path":"/tmp/out.go","content":"package main"}`)
	got := extractToolInput("Write", raw)
	if got != "/tmp/out.go (12 bytes)" {
		t.Fatalf("unexpected: %s", got)
	}
}

func TestExtractToolInput_Good_Grep(t *testing.T) {
	raw := json.RawMessage(`{"pattern":"TODO","path":"/src"}`)
	got := extractToolInput("Grep", raw)
	if got != "/TODO/ in /src" {
		t.Fatalf("unexpected: %s", got)
	}
}

func TestExtractToolInput_Good_GrepNoPath(t *testing.T) {
	raw := json.RawMessage(`{"pattern":"TODO"}`)
	got := extractToolInput("Grep", raw)
	if got != "/TODO/ in ." {
		t.Fatalf("unexpected: %s", got)
	}
}

func TestExtractToolInput_Good_Glob(t *testing.T) {
	raw := json.RawMessage(`{"pattern":"**/*.go"}`)
	got := extractToolInput("Glob", raw)
	if got != "**/*.go" {
		t.Fatalf("unexpected: %s", got)
	}
}

func TestExtractToolInput_Good_Task(t *testing.T) {
	raw := json.RawMessage(`{"prompt":"investigate the bug","description":"debug helper","subagent_type":"Explore"}`)
	got := extractToolInput("Task", raw)
	if got != "[Explore] debug helper" {
		t.Fatalf("unexpected: %s", got)
	}
}

func TestExtractToolInput_Good_TaskNoDesc(t *testing.T) {
	raw := json.RawMessage(`{"prompt":"investigate the bug","subagent_type":"Explore"}`)
	got := extractToolInput("Task", raw)
	if got != "[Explore] investigate the bug" {
		t.Fatalf("unexpected: %s", got)
	}
}

func TestExtractToolInput_Good_UnknownTool(t *testing.T) {
	raw := json.RawMessage(`{"alpha":"1","beta":"2"}`)
	got := extractToolInput("CustomTool", raw)
	if got != "alpha, beta" {
		t.Fatalf("unexpected: %s", got)
	}
}

func TestExtractToolInput_Good_NilInput(t *testing.T) {
	got := extractToolInput("Bash", nil)
	if got != "" {
		t.Fatalf("expected empty, got %s", got)
	}
}

func TestExtractToolInput_Bad_InvalidJSON(t *testing.T) {
	raw := json.RawMessage(`not json`)
	got := extractToolInput("Bash", raw)
	// Falls through to fallback, which also fails — returns empty.
	if got != "" {
		t.Fatalf("expected empty, got %s", got)
	}
}

// ── extractResultContent ───────────────────────────────────────────

func TestExtractResultContent_Good_String(t *testing.T) {
	got := extractResultContent("hello")
	if got != "hello" {
		t.Fatalf("expected hello, got %s", got)
	}
}

func TestExtractResultContent_Good_Slice(t *testing.T) {
	input := []any{
		map[string]any{"text": "line1"},
		map[string]any{"text": "line2"},
	}
	got := extractResultContent(input)
	if got != "line1\nline2" {
		t.Fatalf("unexpected: %s", got)
	}
}

func TestExtractResultContent_Good_Map(t *testing.T) {
	input := map[string]any{"text": "content"}
	got := extractResultContent(input)
	if got != "content" {
		t.Fatalf("expected content, got %s", got)
	}
}

func TestExtractResultContent_Good_MapNoText(t *testing.T) {
	input := map[string]any{"data": 42}
	got := extractResultContent(input)
	if got == "" {
		t.Fatal("expected non-empty fallback")
	}
}

func TestExtractResultContent_Good_Other(t *testing.T) {
	got := extractResultContent(42)
	if got != "42" {
		t.Fatalf("expected 42, got %s", got)
	}
}

// ── ParseTranscript ────────────────────────────────────────────────

func writeJSONL(t *testing.T, path string, entries []any) {
	t.Helper()
	f, err := os.Create(path)
	if err != nil {
		t.Fatal(err)
	}
	defer f.Close()
	enc := json.NewEncoder(f)
	for _, e := range entries {
		if err := enc.Encode(e); err != nil {
			t.Fatal(err)
		}
	}
}

func TestParseTranscript_Good_BasicFlow(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "test-session.jsonl")

	ts1 := time.Date(2026, 2, 24, 10, 0, 0, 0, time.UTC)
	ts2 := time.Date(2026, 2, 24, 10, 0, 1, 0, time.UTC)
	ts3 := time.Date(2026, 2, 24, 10, 0, 2, 0, time.UTC)

	entries := []any{
		map[string]any{
			"type": "assistant", "timestamp": ts1.Format(time.RFC3339Nano),
			"message": map[string]any{
				"role": "assistant",
				"content": []any{
					map[string]any{
						"type": "tool_use", "id": "tu_1", "name": "Bash",
						"input": map[string]any{"command": "go test ./...", "description": "run tests"},
					},
				},
			},
		},
		map[string]any{
			"type": "user", "timestamp": ts2.Format(time.RFC3339Nano),
			"message": map[string]any{
				"role": "user",
				"content": []any{
					map[string]any{
						"type": "tool_result", "tool_use_id": "tu_1",
						"content": "ok  forge.lthn.ai/core/go  1.2s",
					},
				},
			},
		},
		map[string]any{
			"type": "user", "timestamp": ts3.Format(time.RFC3339Nano),
			"message": map[string]any{
				"role": "user",
				"content": []any{
					map[string]any{
						"type": "text", "text": "nice work",
					},
				},
			},
		},
	}

	writeJSONL(t, path, entries)

	sess, err := ParseTranscript(path)
	if err != nil {
		t.Fatal(err)
	}

	if sess.ID != "test-session" {
		t.Fatalf("expected test-session, got %s", sess.ID)
	}
	if len(sess.Events) != 2 {
		t.Fatalf("expected 2 events, got %d", len(sess.Events))
	}

	// Tool use event.
	tool := sess.Events[0]
	if tool.Type != "tool_use" {
		t.Fatalf("expected tool_use, got %s", tool.Type)
	}
	if tool.Tool != "Bash" {
		t.Fatalf("expected Bash, got %s", tool.Tool)
	}
	if !tool.Success {
		t.Fatal("expected success")
	}
	if tool.Duration != time.Second {
		t.Fatalf("expected 1s duration, got %s", tool.Duration)
	}

	// User message.
	user := sess.Events[1]
	if user.Type != "user" {
		t.Fatalf("expected user, got %s", user.Type)
	}
	if user.Input != "nice work" {
		t.Fatalf("unexpected input: %s", user.Input)
	}
}

func TestParseTranscript_Good_ToolError(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "err-session.jsonl")

	ts1 := time.Date(2026, 2, 24, 10, 0, 0, 0, time.UTC)
	ts2 := time.Date(2026, 2, 24, 10, 0, 1, 0, time.UTC)
	isError := true

	entries := []any{
		map[string]any{
			"type": "assistant", "timestamp": ts1.Format(time.RFC3339Nano),
			"message": map[string]any{
				"role": "assistant",
				"content": []any{
					map[string]any{
						"type": "tool_use", "id": "tu_err", "name": "Bash",
						"input": map[string]any{"command": "rm -rf /"},
					},
				},
			},
		},
		map[string]any{
			"type": "user", "timestamp": ts2.Format(time.RFC3339Nano),
			"message": map[string]any{
				"role": "user",
				"content": []any{
					map[string]any{
						"type": "tool_result", "tool_use_id": "tu_err",
						"content": "permission denied", "is_error": &isError,
					},
				},
			},
		},
	}

	writeJSONL(t, path, entries)

	sess, err := ParseTranscript(path)
	if err != nil {
		t.Fatal(err)
	}

	if len(sess.Events) != 1 {
		t.Fatalf("expected 1 event, got %d", len(sess.Events))
	}
	if sess.Events[0].Success {
		t.Fatal("expected failure")
	}
	if sess.Events[0].ErrorMsg != "permission denied" {
		t.Fatalf("unexpected error: %s", sess.Events[0].ErrorMsg)
	}
}

func TestParseTranscript_Good_AssistantText(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "asst.jsonl")

	ts := time.Date(2026, 2, 24, 10, 0, 0, 0, time.UTC)
	entries := []any{
		map[string]any{
			"type": "assistant", "timestamp": ts.Format(time.RFC3339Nano),
			"message": map[string]any{
				"role": "assistant",
				"content": []any{
					map[string]any{"type": "text", "text": "Let me check that."},
				},
			},
		},
	}

	writeJSONL(t, path, entries)

	sess, err := ParseTranscript(path)
	if err != nil {
		t.Fatal(err)
	}
	if len(sess.Events) != 1 {
		t.Fatalf("expected 1 event, got %d", len(sess.Events))
	}
	if sess.Events[0].Type != "assistant" {
		t.Fatalf("expected assistant, got %s", sess.Events[0].Type)
	}
}

func TestParseTranscript_Bad_MissingFile(t *testing.T) {
	_, err := ParseTranscript("/nonexistent/path.jsonl")
	if err == nil {
		t.Fatal("expected error for missing file")
	}
}

func TestParseTranscript_Good_EmptyFile(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "empty.jsonl")
	os.WriteFile(path, []byte{}, 0644)

	sess, err := ParseTranscript(path)
	if err != nil {
		t.Fatal(err)
	}
	if len(sess.Events) != 0 {
		t.Fatalf("expected 0 events, got %d", len(sess.Events))
	}
}

func TestParseTranscript_Good_MalformedLines(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "bad.jsonl")
	os.WriteFile(path, []byte("not json\n{also bad\n"), 0644)

	sess, err := ParseTranscript(path)
	if err != nil {
		t.Fatal(err)
	}
	if len(sess.Events) != 0 {
		t.Fatalf("expected 0 events from bad lines, got %d", len(sess.Events))
	}
}

// ── ListSessions ───────────────────────────────────────────────────

func TestListSessions_Good(t *testing.T) {
	dir := t.TempDir()

	ts1 := time.Date(2026, 2, 24, 10, 0, 0, 0, time.UTC)
	ts2 := time.Date(2026, 2, 24, 11, 0, 0, 0, time.UTC)

	writeJSONL(t, filepath.Join(dir, "sess-a.jsonl"), []any{
		map[string]any{"type": "assistant", "timestamp": ts1.Format(time.RFC3339Nano),
			"message": map[string]any{"role": "assistant", "content": []any{}}},
	})
	writeJSONL(t, filepath.Join(dir, "sess-b.jsonl"), []any{
		map[string]any{"type": "assistant", "timestamp": ts2.Format(time.RFC3339Nano),
			"message": map[string]any{"role": "assistant", "content": []any{}}},
	})

	sessions, err := ListSessions(dir)
	if err != nil {
		t.Fatal(err)
	}
	if len(sessions) != 2 {
		t.Fatalf("expected 2 sessions, got %d", len(sessions))
	}
	// Sorted newest first.
	if sessions[0].ID != "sess-b" {
		t.Fatalf("expected sess-b first, got %s", sessions[0].ID)
	}
}

func TestListSessions_Good_EmptyDir(t *testing.T) {
	dir := t.TempDir()
	sessions, err := ListSessions(dir)
	if err != nil {
		t.Fatal(err)
	}
	if len(sessions) != 0 {
		t.Fatalf("expected 0, got %d", len(sessions))
	}
}
