package jobrunner

import (
	"bufio"
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestJournal_Append_Good(t *testing.T) {
	dir := t.TempDir()

	j, err := NewJournal(dir)
	require.NoError(t, err)

	ts := time.Date(2026, 2, 5, 14, 30, 0, 0, time.UTC)

	signal := &PipelineSignal{
		EpicNumber:      10,
		ChildNumber:     3,
		PRNumber:        55,
		RepoOwner:       "host-uk",
		RepoName:        "core-tenant",
		PRState:         "OPEN",
		IsDraft:         false,
		Mergeable:       "MERGEABLE",
		CheckStatus:     "SUCCESS",
		ThreadsTotal:    2,
		ThreadsResolved: 1,
		LastCommitSHA:   "abc123",
		LastCommitAt:    ts,
		LastReviewAt:    ts,
	}

	result := &ActionResult{
		Action:      "merge",
		RepoOwner:   "host-uk",
		RepoName:    "core-tenant",
		EpicNumber:  10,
		ChildNumber: 3,
		PRNumber:    55,
		Success:     true,
		Timestamp:   ts,
		Duration:    1200 * time.Millisecond,
		Cycle:       1,
	}

	err = j.Append(signal, result)
	require.NoError(t, err)

	// Read the file back.
	expectedPath := filepath.Join(dir, "host-uk", "core-tenant", "2026-02-05.jsonl")
	f, err := os.Open(expectedPath)
	require.NoError(t, err)
	defer func() { _ = f.Close() }()

	scanner := bufio.NewScanner(f)
	require.True(t, scanner.Scan(), "expected at least one line in JSONL file")

	var entry JournalEntry
	err = json.Unmarshal(scanner.Bytes(), &entry)
	require.NoError(t, err)

	assert.Equal(t, "2026-02-05T14:30:00Z", entry.Timestamp)
	assert.Equal(t, 10, entry.Epic)
	assert.Equal(t, 3, entry.Child)
	assert.Equal(t, 55, entry.PR)
	assert.Equal(t, "host-uk/core-tenant", entry.Repo)
	assert.Equal(t, "merge", entry.Action)
	assert.Equal(t, 1, entry.Cycle)

	// Verify signal snapshot.
	assert.Equal(t, "OPEN", entry.Signals.PRState)
	assert.Equal(t, false, entry.Signals.IsDraft)
	assert.Equal(t, "SUCCESS", entry.Signals.CheckStatus)
	assert.Equal(t, "MERGEABLE", entry.Signals.Mergeable)
	assert.Equal(t, 2, entry.Signals.ThreadsTotal)
	assert.Equal(t, 1, entry.Signals.ThreadsResolved)

	// Verify result snapshot.
	assert.Equal(t, true, entry.Result.Success)
	assert.Equal(t, "", entry.Result.Error)
	assert.Equal(t, int64(1200), entry.Result.DurationMs)

	// Append a second entry and verify two lines exist.
	result2 := &ActionResult{
		Action:    "comment",
		RepoOwner: "host-uk",
		RepoName:  "core-tenant",
		Success:   false,
		Error:     "rate limited",
		Timestamp: ts,
		Duration:  50 * time.Millisecond,
		Cycle:     2,
	}
	err = j.Append(signal, result2)
	require.NoError(t, err)

	data, err := os.ReadFile(expectedPath)
	require.NoError(t, err)

	lines := 0
	sc := bufio.NewScanner(strings.NewReader(string(data)))
	for sc.Scan() {
		lines++
	}
	assert.Equal(t, 2, lines, "expected two JSONL lines after two appends")
}

func TestJournal_Append_Bad_NilSignal(t *testing.T) {
	dir := t.TempDir()

	j, err := NewJournal(dir)
	require.NoError(t, err)

	result := &ActionResult{
		Action:    "merge",
		Timestamp: time.Now(),
	}

	err = j.Append(nil, result)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "signal is required")
}

func TestJournal_Append_Bad_NilResult(t *testing.T) {
	dir := t.TempDir()

	j, err := NewJournal(dir)
	require.NoError(t, err)

	signal := &PipelineSignal{
		RepoOwner: "host-uk",
		RepoName:  "core-php",
	}

	err = j.Append(signal, nil)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "result is required")
}
