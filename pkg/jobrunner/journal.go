package jobrunner

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sync"
)

// JournalEntry is a single line in the JSONL audit log.
type JournalEntry struct {
	Timestamp string         `json:"ts"`
	Epic      int            `json:"epic"`
	Child     int            `json:"child"`
	PR        int            `json:"pr"`
	Repo      string         `json:"repo"`
	Action    string         `json:"action"`
	Signals   SignalSnapshot `json:"signals"`
	Result    ResultSnapshot `json:"result"`
	Cycle     int            `json:"cycle"`
}

// SignalSnapshot captures the structural state of a PR at the time of action.
type SignalSnapshot struct {
	PRState         string `json:"pr_state"`
	IsDraft         bool   `json:"is_draft"`
	CheckStatus     string `json:"check_status"`
	Mergeable       string `json:"mergeable"`
	ThreadsTotal    int    `json:"threads_total"`
	ThreadsResolved int    `json:"threads_resolved"`
}

// ResultSnapshot captures the outcome of an action.
type ResultSnapshot struct {
	Success    bool   `json:"success"`
	Error      string `json:"error,omitempty"`
	DurationMs int64  `json:"duration_ms"`
}

// Journal writes ActionResult entries to date-partitioned JSONL files.
type Journal struct {
	baseDir string
	mu      sync.Mutex
}

// NewJournal creates a new Journal rooted at baseDir.
func NewJournal(baseDir string) (*Journal, error) {
	if baseDir == "" {
		return nil, fmt.Errorf("journal base directory is required")
	}
	return &Journal{baseDir: baseDir}, nil
}

// Append writes a journal entry for the given signal and result.
func (j *Journal) Append(signal *PipelineSignal, result *ActionResult) error {
	if signal == nil {
		return fmt.Errorf("signal is required")
	}
	if result == nil {
		return fmt.Errorf("result is required")
	}

	entry := JournalEntry{
		Timestamp: result.Timestamp.UTC().Format("2006-01-02T15:04:05Z"),
		Epic:      signal.EpicNumber,
		Child:     signal.ChildNumber,
		PR:        signal.PRNumber,
		Repo:      signal.RepoFullName(),
		Action:    result.Action,
		Signals: SignalSnapshot{
			PRState:         signal.PRState,
			IsDraft:         signal.IsDraft,
			CheckStatus:     signal.CheckStatus,
			Mergeable:       signal.Mergeable,
			ThreadsTotal:    signal.ThreadsTotal,
			ThreadsResolved: signal.ThreadsResolved,
		},
		Result: ResultSnapshot{
			Success:    result.Success,
			Error:      result.Error,
			DurationMs: result.Duration.Milliseconds(),
		},
		Cycle: result.Cycle,
	}

	data, err := json.Marshal(entry)
	if err != nil {
		return fmt.Errorf("marshal journal entry: %w", err)
	}
	data = append(data, '\n')

	date := result.Timestamp.UTC().Format("2006-01-02")
	dir := filepath.Join(j.baseDir, signal.RepoOwner, signal.RepoName)

	j.mu.Lock()
	defer j.mu.Unlock()

	if err := os.MkdirAll(dir, 0o755); err != nil {
		return fmt.Errorf("create journal directory: %w", err)
	}

	path := filepath.Join(dir, date+".jsonl")
	f, err := os.OpenFile(path, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0o644)
	if err != nil {
		return fmt.Errorf("open journal file: %w", err)
	}
	defer func() { _ = f.Close() }()

	_, err = f.Write(data)
	return err
}
