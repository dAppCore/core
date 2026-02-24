package session

import (
	"path/filepath"
	"testing"
	"time"
)

func TestSearch_Good_MatchFound(t *testing.T) {
	dir := t.TempDir()

	ts1 := time.Date(2026, 2, 24, 10, 0, 0, 0, time.UTC)
	ts2 := time.Date(2026, 2, 24, 10, 0, 1, 0, time.UTC)

	writeJSONL(t, filepath.Join(dir, "search-test.jsonl"), []any{
		map[string]any{
			"type": "assistant", "timestamp": ts1.Format(time.RFC3339Nano),
			"message": map[string]any{
				"role": "assistant",
				"content": []any{
					map[string]any{
						"type": "tool_use", "id": "tu_1", "name": "Bash",
						"input": map[string]any{"command": "go test ./..."},
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
	})

	results, err := Search(dir, "go test")
	if err != nil {
		t.Fatal(err)
	}
	if len(results) != 1 {
		t.Fatalf("expected 1 result, got %d", len(results))
	}
	if results[0].Tool != "Bash" {
		t.Fatalf("expected Bash, got %s", results[0].Tool)
	}
}

func TestSearch_Good_CaseInsensitive(t *testing.T) {
	dir := t.TempDir()

	ts1 := time.Date(2026, 2, 24, 10, 0, 0, 0, time.UTC)
	ts2 := time.Date(2026, 2, 24, 10, 0, 1, 0, time.UTC)

	writeJSONL(t, filepath.Join(dir, "case.jsonl"), []any{
		map[string]any{
			"type": "assistant", "timestamp": ts1.Format(time.RFC3339Nano),
			"message": map[string]any{
				"role": "assistant",
				"content": []any{
					map[string]any{
						"type": "tool_use", "id": "tu_2", "name": "Bash",
						"input": map[string]any{"command": "GO TEST"},
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
						"type": "tool_result", "tool_use_id": "tu_2",
						"content": "ok",
					},
				},
			},
		},
	})

	results, err := Search(dir, "go test")
	if err != nil {
		t.Fatal(err)
	}
	if len(results) != 1 {
		t.Fatal("case insensitive search should match")
	}
}

func TestSearch_Good_NoMatch(t *testing.T) {
	dir := t.TempDir()

	ts1 := time.Date(2026, 2, 24, 10, 0, 0, 0, time.UTC)
	ts2 := time.Date(2026, 2, 24, 10, 0, 1, 0, time.UTC)

	writeJSONL(t, filepath.Join(dir, "nomatch.jsonl"), []any{
		map[string]any{
			"type": "assistant", "timestamp": ts1.Format(time.RFC3339Nano),
			"message": map[string]any{
				"role": "assistant",
				"content": []any{
					map[string]any{
						"type": "tool_use", "id": "tu_3", "name": "Bash",
						"input": map[string]any{"command": "ls"},
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
						"type": "tool_result", "tool_use_id": "tu_3",
						"content": "file.txt",
					},
				},
			},
		},
	})

	results, err := Search(dir, "nonexistent query")
	if err != nil {
		t.Fatal(err)
	}
	if len(results) != 0 {
		t.Fatalf("expected 0 results, got %d", len(results))
	}
}

func TestSearch_Good_EmptyDir(t *testing.T) {
	dir := t.TempDir()
	results, err := Search(dir, "anything")
	if err != nil {
		t.Fatal(err)
	}
	if len(results) != 0 {
		t.Fatalf("expected 0, got %d", len(results))
	}
}

func TestSearch_Good_SkipsNonToolEvents(t *testing.T) {
	dir := t.TempDir()

	ts := time.Date(2026, 2, 24, 10, 0, 0, 0, time.UTC)
	writeJSONL(t, filepath.Join(dir, "skip.jsonl"), []any{
		map[string]any{
			"type": "user", "timestamp": ts.Format(time.RFC3339Nano),
			"message": map[string]any{
				"role": "user",
				"content": []any{
					map[string]any{"type": "text", "text": "go test should find this"},
				},
			},
		},
	})

	results, err := Search(dir, "go test")
	if err != nil {
		t.Fatal(err)
	}
	if len(results) != 0 {
		t.Fatal("search should only match tool_use events")
	}
}
