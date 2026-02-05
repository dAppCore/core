# Core-IDE Job Runner Implementation Plan

> **For Claude:** REQUIRED SUB-SKILL: Use superpowers:executing-plans to implement this plan task-by-task.

**Goal:** Turn core-ide into an autonomous job runner that polls GitHub for pipeline work, executes it via typed handlers, and captures JSONL training data.

**Architecture:** Go workspace (`go.work`) linking root module + core-ide module. Pluggable `JobSource` interface with GitHub as first adapter. `JobHandler` interface for each pipeline action (publish draft, resolve threads, etc.). `Poller` orchestrates discovery and dispatch. `Journal` writes JSONL. Headless mode reuses existing `pkg/cli.Daemon` infrastructure. Handlers live in `pkg/jobrunner/` (root module), core-ide imports them via workspace.

**Tech Stack:** Go 1.25, GitHub REST API (via `oauth2`), `pkg/cli.Daemon` for headless, `testify/assert` + `httptest` for tests.

---

### Task 0: Set Up Go Workspace (`go.work`)

**Files:**
- Create: `go.work`

**Context:** The repo has two real modules — the root (`github.com/host-uk/core`) and core-ide (`github.com/host-uk/core/internal/core-ide`). Without a workspace, core-ide can't import `pkg/jobrunner` from the root module during local development without fragile `replace` directives. A `go.work` file makes cross-module imports resolve locally, keeps each module's `go.mod` clean, and lets CI build each variant independently.

**Step 1: Create the workspace file**

```bash
cd /Users/snider/Code/host-uk/core
go work init . ./internal/core-ide
```

This generates `go.work`:
```
go 1.25.5

use (
	.
	./internal/core-ide
)
```

**Step 2: Sync dependency versions across modules**

```bash
go work sync
```

This aligns shared dependency versions between the two modules.

**Step 3: Verify the workspace**

Run: `go build ./...`
Expected: Root module builds successfully.

Run: `cd internal/core-ide && go build .`
Expected: core-ide builds successfully.

Run: `go test ./pkg/... -count=1`
Expected: All existing tests pass (workspace doesn't change behaviour, just resolution).

**Step 4: Add go.work.sum to gitignore**

`go.work.sum` is generated and shouldn't be committed (it's machine-specific like `go.sum` but for the workspace). Check if `.gitignore` already excludes it:

```bash
grep -q 'go.work.sum' .gitignore || echo 'go.work.sum' >> .gitignore
```

**Note:** Whether to commit `go.work` itself is a choice. Committing it means all developers and CI share the same workspace layout. Since the module layout is fixed (root + core-ide), committing it is the right call — it documents the build variants explicitly.

**Step 5: Commit**

```bash
git add go.work .gitignore
git commit -m "build: add Go workspace for root + core-ide modules"
```

---

### Task 1: Core Types (`pkg/jobrunner/types.go`)

**Files:**
- Create: `pkg/jobrunner/types.go`
- Test: `pkg/jobrunner/types_test.go`

**Step 1: Write the test file**

```go
package jobrunner

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestPipelineSignal_RepoFullName_Good(t *testing.T) {
	s := &PipelineSignal{RepoOwner: "host-uk", RepoName: "core"}
	assert.Equal(t, "host-uk/core", s.RepoFullName())
}

func TestPipelineSignal_HasUnresolvedThreads_Good(t *testing.T) {
	s := &PipelineSignal{ThreadsTotal: 5, ThreadsResolved: 3}
	assert.True(t, s.HasUnresolvedThreads())
}

func TestPipelineSignal_HasUnresolvedThreads_Bad_AllResolved(t *testing.T) {
	s := &PipelineSignal{ThreadsTotal: 5, ThreadsResolved: 5}
	assert.False(t, s.HasUnresolvedThreads())
}

func TestActionResult_JSON_Good(t *testing.T) {
	r := &ActionResult{
		Action:    "publish_draft",
		RepoOwner: "host-uk",
		RepoName:  "core",
		PRNumber:  315,
		Success:   true,
		Timestamp: time.Date(2026, 2, 5, 12, 0, 0, 0, time.UTC),
	}
	assert.Equal(t, "publish_draft", r.Action)
	assert.True(t, r.Success)
}
```

**Step 2: Run test to verify it fails**

Run: `go test ./pkg/jobrunner/ -v -count=1`
Expected: FAIL — package does not exist yet.

**Step 3: Write the types**

```go
package jobrunner

import (
	"context"
	"time"
)

// PipelineSignal is the structural snapshot of a child issue/PR.
// Never contains comment bodies or free text — structural signals only.
type PipelineSignal struct {
	EpicNumber      int
	ChildNumber     int
	PRNumber        int
	RepoOwner       string
	RepoName        string
	PRState         string    // OPEN, MERGED, CLOSED
	IsDraft         bool
	Mergeable       string    // MERGEABLE, CONFLICTING, UNKNOWN
	CheckStatus     string    // SUCCESS, FAILURE, PENDING
	ThreadsTotal    int
	ThreadsResolved int
	LastCommitSHA   string
	LastCommitAt    time.Time
	LastReviewAt    time.Time
}

// RepoFullName returns "owner/repo".
func (s *PipelineSignal) RepoFullName() string {
	return s.RepoOwner + "/" + s.RepoName
}

// HasUnresolvedThreads returns true if there are unresolved review threads.
func (s *PipelineSignal) HasUnresolvedThreads() bool {
	return s.ThreadsTotal > s.ThreadsResolved
}

// ActionResult carries the outcome of a handler execution.
type ActionResult struct {
	Action     string        `json:"action"`
	RepoOwner  string        `json:"repo_owner"`
	RepoName   string        `json:"repo_name"`
	EpicNumber int           `json:"epic"`
	ChildNumber int          `json:"child"`
	PRNumber   int           `json:"pr"`
	Success    bool          `json:"success"`
	Error      string        `json:"error,omitempty"`
	Timestamp  time.Time     `json:"ts"`
	Duration   time.Duration `json:"duration_ms"`
	Cycle      int           `json:"cycle"`
}

// JobSource discovers actionable work from an external system.
type JobSource interface {
	Name() string
	Poll(ctx context.Context) ([]*PipelineSignal, error)
	Report(ctx context.Context, result *ActionResult) error
}

// JobHandler processes a single pipeline signal.
type JobHandler interface {
	Name() string
	Match(signal *PipelineSignal) bool
	Execute(ctx context.Context, signal *PipelineSignal) (*ActionResult, error)
}
```

**Step 4: Run tests**

Run: `go test ./pkg/jobrunner/ -v -count=1`
Expected: PASS (4 tests).

**Step 5: Commit**

```bash
git add pkg/jobrunner/types.go pkg/jobrunner/types_test.go
git commit -m "feat(jobrunner): add core types — PipelineSignal, ActionResult, JobSource, JobHandler"
```

---

### Task 2: Journal JSONL Writer (`pkg/jobrunner/journal.go`)

**Files:**
- Create: `pkg/jobrunner/journal.go`
- Test: `pkg/jobrunner/journal_test.go`

**Step 1: Write the test**

```go
package jobrunner

import (
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

	signal := &PipelineSignal{
		EpicNumber:  299,
		ChildNumber: 212,
		PRNumber:    316,
		RepoOwner:   "host-uk",
		RepoName:    "core",
		PRState:     "OPEN",
		IsDraft:     true,
		CheckStatus: "SUCCESS",
	}

	result := &ActionResult{
		Action:    "publish_draft",
		RepoOwner: "host-uk",
		RepoName:  "core",
		PRNumber:  316,
		Success:   true,
		Timestamp: time.Date(2026, 2, 5, 12, 0, 0, 0, time.UTC),
		Duration:  340 * time.Millisecond,
		Cycle:     1,
	}

	err = j.Append(signal, result)
	require.NoError(t, err)

	// Read the file back
	pattern := filepath.Join(dir, "host-uk", "core", "*.jsonl")
	files, _ := filepath.Glob(pattern)
	require.Len(t, files, 1)

	data, err := os.ReadFile(files[0])
	require.NoError(t, err)

	var entry JournalEntry
	err = json.Unmarshal([]byte(strings.TrimSpace(string(data))), &entry)
	require.NoError(t, err)

	assert.Equal(t, "publish_draft", entry.Action)
	assert.Equal(t, 316, entry.PR)
	assert.Equal(t, 299, entry.Epic)
	assert.True(t, entry.Result.Success)
}

func TestJournal_Append_Bad_NilSignal(t *testing.T) {
	dir := t.TempDir()
	j, err := NewJournal(dir)
	require.NoError(t, err)

	err = j.Append(nil, &ActionResult{})
	assert.Error(t, err)
}
```

**Step 2: Run test to verify it fails**

Run: `go test ./pkg/jobrunner/ -run TestJournal -v -count=1`
Expected: FAIL — `NewJournal` undefined.

**Step 3: Write the implementation**

```go
package jobrunner

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sync"
	"time"
)

// JournalEntry is a single JSONL record for training data.
type JournalEntry struct {
	Timestamp time.Time      `json:"ts"`
	Epic      int            `json:"epic"`
	Child     int            `json:"child"`
	PR        int            `json:"pr"`
	Repo      string         `json:"repo"`
	Action    string         `json:"action"`
	Signals   SignalSnapshot `json:"signals"`
	Result    ResultSnapshot `json:"result"`
	Cycle     int            `json:"cycle"`
}

// SignalSnapshot captures the structural state at action time.
type SignalSnapshot struct {
	PRState         string `json:"pr_state"`
	IsDraft         bool   `json:"is_draft"`
	CheckStatus     string `json:"check_status"`
	Mergeable       string `json:"mergeable"`
	ThreadsTotal    int    `json:"threads_total"`
	ThreadsResolved int    `json:"threads_resolved"`
}

// ResultSnapshot captures the action outcome.
type ResultSnapshot struct {
	Success    bool   `json:"success"`
	Error      string `json:"error,omitempty"`
	DurationMs int64  `json:"duration_ms"`
}

// Journal writes append-only JSONL files organised by repo and date.
type Journal struct {
	baseDir string
	mu      sync.Mutex
}

// NewJournal creates a journal writer rooted at baseDir.
// Files are written to baseDir/<owner>/<repo>/YYYY-MM-DD.jsonl.
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
		Timestamp: result.Timestamp,
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

	// Build path: baseDir/owner/repo/YYYY-MM-DD.jsonl
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
	defer f.Close()

	_, err = f.Write(data)
	return err
}
```

**Step 4: Run tests**

Run: `go test ./pkg/jobrunner/ -v -count=1`
Expected: PASS (all tests including Task 1).

**Step 5: Commit**

```bash
git add pkg/jobrunner/journal.go pkg/jobrunner/journal_test.go
git commit -m "feat(jobrunner): add JSONL journal writer for training data"
```

---

### Task 3: Poller and Dispatcher (`pkg/jobrunner/poller.go`)

**Files:**
- Create: `pkg/jobrunner/poller.go`
- Test: `pkg/jobrunner/poller_test.go`

**Step 1: Write the test**

```go
package jobrunner

import (
	"context"
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type mockSource struct {
	name    string
	signals []*PipelineSignal
	reports []*ActionResult
	mu      sync.Mutex
}

func (m *mockSource) Name() string { return m.name }
func (m *mockSource) Poll(_ context.Context) ([]*PipelineSignal, error) {
	return m.signals, nil
}
func (m *mockSource) Report(_ context.Context, r *ActionResult) error {
	m.mu.Lock()
	m.reports = append(m.reports, r)
	m.mu.Unlock()
	return nil
}

type mockHandler struct {
	name     string
	matchFn  func(*PipelineSignal) bool
	executed []*PipelineSignal
	mu       sync.Mutex
}

func (m *mockHandler) Name() string { return m.name }
func (m *mockHandler) Match(s *PipelineSignal) bool {
	if m.matchFn != nil {
		return m.matchFn(s)
	}
	return true
}
func (m *mockHandler) Execute(_ context.Context, s *PipelineSignal) (*ActionResult, error) {
	m.mu.Lock()
	m.executed = append(m.executed, s)
	m.mu.Unlock()
	return &ActionResult{
		Action:    m.name,
		Success:   true,
		Timestamp: time.Now().UTC(),
	}, nil
}

func TestPoller_RunOnce_Good(t *testing.T) {
	signal := &PipelineSignal{
		PRNumber:  315,
		RepoOwner: "host-uk",
		RepoName:  "core",
		IsDraft:   true,
		PRState:   "OPEN",
	}

	source := &mockSource{name: "test", signals: []*PipelineSignal{signal}}
	handler := &mockHandler{name: "publish_draft"}
	journal, err := NewJournal(t.TempDir())
	require.NoError(t, err)

	p := NewPoller(PollerConfig{
		Sources:      []JobSource{source},
		Handlers:     []JobHandler{handler},
		Journal:      journal,
		PollInterval: time.Second,
	})

	err = p.RunOnce(context.Background())
	require.NoError(t, err)

	handler.mu.Lock()
	assert.Len(t, handler.executed, 1)
	handler.mu.Unlock()
}

func TestPoller_RunOnce_Good_NoSignals(t *testing.T) {
	source := &mockSource{name: "test", signals: nil}
	handler := &mockHandler{name: "noop"}
	journal, err := NewJournal(t.TempDir())
	require.NoError(t, err)

	p := NewPoller(PollerConfig{
		Sources:  []JobSource{source},
		Handlers: []JobHandler{handler},
		Journal:  journal,
	})

	err = p.RunOnce(context.Background())
	require.NoError(t, err)

	handler.mu.Lock()
	assert.Len(t, handler.executed, 0)
	handler.mu.Unlock()
}

func TestPoller_RunOnce_Good_NoMatchingHandler(t *testing.T) {
	signal := &PipelineSignal{PRNumber: 1, RepoOwner: "a", RepoName: "b"}
	source := &mockSource{name: "test", signals: []*PipelineSignal{signal}}
	handler := &mockHandler{
		name:    "never_match",
		matchFn: func(*PipelineSignal) bool { return false },
	}
	journal, err := NewJournal(t.TempDir())
	require.NoError(t, err)

	p := NewPoller(PollerConfig{
		Sources:  []JobSource{source},
		Handlers: []JobHandler{handler},
		Journal:  journal,
	})

	err = p.RunOnce(context.Background())
	require.NoError(t, err)

	handler.mu.Lock()
	assert.Len(t, handler.executed, 0)
	handler.mu.Unlock()
}
```

**Step 2: Run test to verify it fails**

Run: `go test ./pkg/jobrunner/ -run TestPoller -v -count=1`
Expected: FAIL — `NewPoller` undefined.

**Step 3: Write the implementation**

```go
package jobrunner

import (
	"context"
	"fmt"
	"time"

	"github.com/host-uk/core/pkg/log"
)

// PollerConfig configures the job runner poller.
type PollerConfig struct {
	Sources      []JobSource
	Handlers     []JobHandler
	Journal      *Journal
	PollInterval time.Duration
	DryRun       bool
}

// Poller discovers and dispatches pipeline work.
type Poller struct {
	cfg   PollerConfig
	cycle int
}

// NewPoller creates a poller with the given configuration.
func NewPoller(cfg PollerConfig) *Poller {
	if cfg.PollInterval == 0 {
		cfg.PollInterval = 60 * time.Second
	}
	return &Poller{cfg: cfg}
}

// Run starts the polling loop. Blocks until context is cancelled.
func (p *Poller) Run(ctx context.Context) error {
	ticker := time.NewTicker(p.cfg.PollInterval)
	defer ticker.Stop()

	// Run once immediately
	if err := p.RunOnce(ctx); err != nil {
		log.Info("poller", "cycle_error", err)
	}

	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-ticker.C:
			if err := p.RunOnce(ctx); err != nil {
				log.Info("poller", "cycle_error", err)
			}
		}
	}
}

// RunOnce performs a single poll-dispatch cycle across all sources.
func (p *Poller) RunOnce(ctx context.Context) error {
	p.cycle++

	for _, source := range p.cfg.Sources {
		if err := ctx.Err(); err != nil {
			return err
		}

		signals, err := source.Poll(ctx)
		if err != nil {
			log.Info("poller", "source", source.Name(), "poll_error", err)
			continue
		}

		for _, signal := range signals {
			if err := ctx.Err(); err != nil {
				return err
			}
			p.dispatch(ctx, source, signal)
		}
	}

	return nil
}

// dispatch finds the first matching handler and executes it.
// One action per signal per cycle.
func (p *Poller) dispatch(ctx context.Context, source JobSource, signal *PipelineSignal) {
	for _, handler := range p.cfg.Handlers {
		if !handler.Match(signal) {
			continue
		}

		if p.cfg.DryRun {
			log.Info("poller",
				"dry_run", handler.Name(),
				"repo", signal.RepoFullName(),
				"pr", signal.PRNumber,
			)
			return
		}

		start := time.Now()
		result, err := handler.Execute(ctx, signal)
		if err != nil {
			log.Info("poller",
				"handler", handler.Name(),
				"error", err,
				"repo", signal.RepoFullName(),
				"pr", signal.PRNumber,
			)
			return
		}

		result.Cycle = p.cycle
		result.EpicNumber = signal.EpicNumber
		result.ChildNumber = signal.ChildNumber
		result.Duration = time.Since(start)

		// Write to journal
		if p.cfg.Journal != nil {
			if err := p.cfg.Journal.Append(signal, result); err != nil {
				log.Info("poller", "journal_error", err)
			}
		}

		// Report back to source
		if err := source.Report(ctx, result); err != nil {
			log.Info("poller", "report_error", err)
		}

		return // one action per signal per cycle
	}
}

// Cycle returns the current cycle count.
func (p *Poller) Cycle() int {
	return p.cycle
}

// DryRun returns whether the poller is in dry-run mode.
func (p *Poller) DryRun() bool {
	return p.cfg.DryRun
}

// SetDryRun enables or disables dry-run mode.
func (p *Poller) SetDryRun(v bool) {
	p.cfg.DryRun = v
}

// AddSource appends a job source to the poller.
func (p *Poller) AddSource(s JobSource) {
	p.cfg.Sources = append(p.cfg.Sources, s)
}

// AddHandler appends a job handler to the poller.
func (p *Poller) AddHandler(h JobHandler) {
	p.cfg.Handlers = append(p.cfg.Handlers, h)
}

_ = fmt.Sprintf // ensure fmt imported for future use
```

Wait — remove that last line. The `fmt` import is only needed if used. Let me correct: the implementation above doesn't use `fmt` directly, so remove it from imports. The `log` package import path is `github.com/host-uk/core/pkg/log`.

**Step 4: Run tests**

Run: `go test ./pkg/jobrunner/ -v -count=1`
Expected: PASS (all tests).

**Step 5: Commit**

```bash
git add pkg/jobrunner/poller.go pkg/jobrunner/poller_test.go
git commit -m "feat(jobrunner): add Poller with multi-source dispatch and journal integration"
```

---

### Task 4: GitHub Source — Signal Builder (`pkg/jobrunner/github/`)

**Files:**
- Create: `pkg/jobrunner/github/source.go`
- Create: `pkg/jobrunner/github/signals.go`
- Test: `pkg/jobrunner/github/source_test.go`

**Context:** This package lives in the root go.mod (`github.com/host-uk/core`), NOT in the core-ide module. It uses `oauth2` and the GitHub REST API (same pattern as `internal/cmd/updater/github.go`). Uses conditional requests (ETag/If-None-Match) to conserve rate limit.

**Step 1: Write the test**

```go
package github

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/host-uk/core/pkg/jobrunner"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestGitHubSource_Poll_Good(t *testing.T) {
	// Mock GitHub API: return one open PR that's a draft with passing checks
	mux := http.NewServeMux()

	// GET /repos/host-uk/core/issues?labels=epic&state=open
	mux.HandleFunc("/repos/host-uk/core/issues", func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Query().Get("labels") == "epic" {
			json.NewEncoder(w).Encode([]map[string]any{
				{
					"number": 299,
					"body":   "- [ ] #212\n- [x] #213",
					"state":  "open",
				},
			})
			return
		}
		json.NewEncoder(w).Encode([]map[string]any{})
	})

	// GET /repos/host-uk/core/pulls?state=open
	mux.HandleFunc("/repos/host-uk/core/pulls", func(w http.ResponseWriter, r *http.Request) {
		json.NewEncoder(w).Encode([]map[string]any{
			{
				"number":             316,
				"state":              "open",
				"draft":              true,
				"mergeable_state":    "clean",
				"body":               "Closes #212",
				"head":               map[string]any{"sha": "abc123"},
			},
		})
	})

	// GET /repos/host-uk/core/commits/abc123/check-suites
	mux.HandleFunc("/repos/host-uk/core/commits/", func(w http.ResponseWriter, r *http.Request) {
		json.NewEncoder(w).Encode(map[string]any{
			"check_suites": []map[string]any{
				{"conclusion": "success", "status": "completed"},
			},
		})
	})

	server := httptest.NewServer(mux)
	defer server.Close()

	src := NewGitHubSource(Config{
		Repos:  []string{"host-uk/core"},
		APIURL: server.URL,
	})

	signals, err := src.Poll(context.Background())
	require.NoError(t, err)
	require.NotEmpty(t, signals)

	assert.Equal(t, 316, signals[0].PRNumber)
	assert.True(t, signals[0].IsDraft)
	assert.Equal(t, "host-uk", signals[0].RepoOwner)
	assert.Equal(t, "core", signals[0].RepoName)
}

func TestGitHubSource_Name_Good(t *testing.T) {
	src := NewGitHubSource(Config{Repos: []string{"host-uk/core"}})
	assert.Equal(t, "github", src.Name())
}
```

**Step 2: Run test to verify it fails**

Run: `go test ./pkg/jobrunner/github/ -v -count=1`
Expected: FAIL — package does not exist.

**Step 3: Write `signals.go`** — PR/issue data structures and signal extraction

```go
package github

import (
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/host-uk/core/pkg/jobrunner"
)

// ghIssue is the minimal structure from GitHub Issues API.
type ghIssue struct {
	Number int    `json:"number"`
	Body   string `json:"body"`
	State  string `json:"state"`
}

// ghPR is the minimal structure from GitHub Pull Requests API.
type ghPR struct {
	Number        int       `json:"number"`
	State         string    `json:"state"`
	Draft         bool      `json:"draft"`
	MergeableState string  `json:"mergeable_state"`
	Body          string    `json:"body"`
	Head          ghRef     `json:"head"`
	UpdatedAt     time.Time `json:"updated_at"`
}

type ghRef struct {
	SHA string `json:"sha"`
}

// ghCheckSuites is the response from /commits/:sha/check-suites.
type ghCheckSuites struct {
	CheckSuites []ghCheckSuite `json:"check_suites"`
}

type ghCheckSuite struct {
	Conclusion string `json:"conclusion"`
	Status     string `json:"status"`
}

// ghReviewThread counts (from GraphQL or approximated from review comments).
type ghReviewCounts struct {
	Total    int
	Resolved int
}

// parseEpicChildren extracts unchecked child issue numbers from an epic body.
// Matches: - [ ] #123
var checklistRe = regexp.MustCompile(`- \[( |x)\] #(\d+)`)

func parseEpicChildren(body string) (unchecked []int, checked []int) {
	matches := checklistRe.FindAllStringSubmatch(body, -1)
	for _, m := range matches {
		num, _ := strconv.Atoi(m[2])
		if m[1] == "x" {
			checked = append(checked, num)
		} else {
			unchecked = append(unchecked, num)
		}
	}
	return
}

// findLinkedPR finds a PR that references an issue number in its body.
// Matches: Closes #123, Fixes #123, Resolves #123
func findLinkedPR(prs []ghPR, issueNumber int) *ghPR {
	pattern := strconv.Itoa(issueNumber)
	for i := range prs {
		if strings.Contains(prs[i].Body, "#"+pattern) {
			return &prs[i]
		}
	}
	return nil
}

// aggregateCheckStatus returns the overall check status from check suites.
func aggregateCheckStatus(suites []ghCheckSuite) string {
	if len(suites) == 0 {
		return "PENDING"
	}
	for _, s := range suites {
		if s.Status != "completed" {
			return "PENDING"
		}
		if s.Conclusion == "failure" || s.Conclusion == "timed_out" || s.Conclusion == "cancelled" {
			return "FAILURE"
		}
	}
	return "SUCCESS"
}

// mergeableToString normalises GitHub's mergeable_state to our enum.
func mergeableToString(state string) string {
	switch state {
	case "clean", "has_hooks", "unstable":
		return "MERGEABLE"
	case "dirty":
		return "CONFLICTING"
	default:
		return "UNKNOWN"
	}
}

// buildSignal creates a PipelineSignal from GitHub API data.
func buildSignal(owner, repo string, epic ghIssue, childNum int, pr ghPR, checks ghCheckSuites) *jobrunner.PipelineSignal {
	return &jobrunner.PipelineSignal{
		EpicNumber:    epic.Number,
		ChildNumber:   childNum,
		PRNumber:      pr.Number,
		RepoOwner:     owner,
		RepoName:      repo,
		PRState:       strings.ToUpper(pr.State),
		IsDraft:       pr.Draft,
		Mergeable:     mergeableToString(pr.MergeableState),
		CheckStatus:   aggregateCheckStatus(checks.CheckSuites),
		LastCommitSHA: pr.Head.SHA,
		LastCommitAt:  pr.UpdatedAt,
	}
}
```

**Step 4: Write `source.go`** — GitHubSource implementing JobSource

```go
package github

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"strings"

	"github.com/host-uk/core/pkg/jobrunner"
	"github.com/host-uk/core/pkg/log"
	"golang.org/x/oauth2"
)

// Config for the GitHub job source.
type Config struct {
	Repos  []string // "owner/repo" format
	APIURL string   // override for testing (default: https://api.github.com)
}

// GitHubSource polls GitHub for pipeline signals.
type GitHubSource struct {
	cfg    Config
	client *http.Client
	etags  map[string]string // URL -> ETag for conditional requests
}

// NewGitHubSource creates a GitHub job source.
func NewGitHubSource(cfg Config) *GitHubSource {
	if cfg.APIURL == "" {
		cfg.APIURL = "https://api.github.com"
	}

	var client *http.Client
	token := os.Getenv("GITHUB_TOKEN")
	if token != "" {
		ts := oauth2.StaticTokenSource(&oauth2.Token{AccessToken: token})
		client = oauth2.NewClient(context.Background(), ts)
	} else {
		client = http.DefaultClient
	}

	return &GitHubSource{
		cfg:    cfg,
		client: client,
		etags:  make(map[string]string),
	}
}

func (g *GitHubSource) Name() string { return "github" }

// Poll scans all configured repos for actionable pipeline signals.
func (g *GitHubSource) Poll(ctx context.Context) ([]*jobrunner.PipelineSignal, error) {
	var all []*jobrunner.PipelineSignal

	for _, repoSpec := range g.cfg.Repos {
		parts := strings.SplitN(repoSpec, "/", 2)
		if len(parts) != 2 {
			continue
		}
		owner, repo := parts[0], parts[1]

		signals, err := g.pollRepo(ctx, owner, repo)
		if err != nil {
			log.Info("github_source", "repo", repoSpec, "error", err)
			continue
		}
		all = append(all, signals...)
	}

	return all, nil
}

func (g *GitHubSource) pollRepo(ctx context.Context, owner, repo string) ([]*jobrunner.PipelineSignal, error) {
	// 1. Fetch epic issues
	epics, err := g.fetchEpics(ctx, owner, repo)
	if err != nil {
		return nil, err
	}

	// 2. Fetch open PRs
	prs, err := g.fetchPRs(ctx, owner, repo)
	if err != nil {
		return nil, err
	}

	var signals []*jobrunner.PipelineSignal

	for _, epic := range epics {
		unchecked, _ := parseEpicChildren(epic.Body)
		for _, childNum := range unchecked {
			pr := findLinkedPR(prs, childNum)
			if pr == nil {
				continue // no PR yet for this child
			}

			checks, err := g.fetchCheckSuites(ctx, owner, repo, pr.Head.SHA)
			if err != nil {
				log.Info("github_source", "pr", pr.Number, "check_error", err)
				checks = ghCheckSuites{}
			}

			signals = append(signals, buildSignal(owner, repo, epic, childNum, *pr, checks))
		}
	}

	return signals, nil
}

func (g *GitHubSource) fetchEpics(ctx context.Context, owner, repo string) ([]ghIssue, error) {
	url := fmt.Sprintf("%s/repos/%s/%s/issues?labels=epic&state=open&per_page=100", g.cfg.APIURL, owner, repo)
	var issues []ghIssue
	return issues, g.getJSON(ctx, url, &issues)
}

func (g *GitHubSource) fetchPRs(ctx context.Context, owner, repo string) ([]ghPR, error) {
	url := fmt.Sprintf("%s/repos/%s/%s/pulls?state=open&per_page=100", g.cfg.APIURL, owner, repo)
	var prs []ghPR
	return prs, g.getJSON(ctx, url, &prs)
}

func (g *GitHubSource) fetchCheckSuites(ctx context.Context, owner, repo, sha string) (ghCheckSuites, error) {
	url := fmt.Sprintf("%s/repos/%s/%s/commits/%s/check-suites", g.cfg.APIURL, owner, repo, sha)
	var result ghCheckSuites
	return result, g.getJSON(ctx, url, &result)
}

// getJSON performs a GET with conditional request support.
func (g *GitHubSource) getJSON(ctx context.Context, url string, out any) error {
	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return err
	}
	req.Header.Set("Accept", "application/vnd.github+json")

	if etag, ok := g.etags[url]; ok {
		req.Header.Set("If-None-Match", etag)
	}

	resp, err := g.client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	// Store ETag for next request
	if etag := resp.Header.Get("ETag"); etag != "" {
		g.etags[url] = etag
	}

	if resp.StatusCode == http.StatusNotModified {
		return nil // no change since last poll
	}

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("HTTP %d for %s", resp.StatusCode, url)
	}

	return json.NewDecoder(resp.Body).Decode(out)
}

// Report is a no-op for GitHub (actions are performed directly via API).
func (g *GitHubSource) Report(_ context.Context, _ *jobrunner.ActionResult) error {
	return nil
}
```

**Step 5: Run tests**

Run: `go test ./pkg/jobrunner/github/ -v -count=1`
Expected: PASS.

**Step 6: Commit**

```bash
git add pkg/jobrunner/github/
git commit -m "feat(jobrunner): add GitHub source adapter with ETag conditional requests"
```

---

### Task 5: Publish Draft Handler (`pkg/jobrunner/handlers/`)

**Files:**
- Create: `pkg/jobrunner/handlers/publish_draft.go`
- Test: `pkg/jobrunner/handlers/publish_draft_test.go`

**Context:** Handlers live in `pkg/jobrunner/handlers/` (root module). They use `net/http` to call GitHub REST API directly. Each handler implements `jobrunner.JobHandler`.

**Step 1: Write the test**

```go
package handlers

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/host-uk/core/pkg/jobrunner"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestPublishDraft_Match_Good(t *testing.T) {
	h := NewPublishDraft(nil)
	signal := &jobrunner.PipelineSignal{
		IsDraft:     true,
		PRState:     "OPEN",
		CheckStatus: "SUCCESS",
	}
	assert.True(t, h.Match(signal))
}

func TestPublishDraft_Match_Bad_NotDraft(t *testing.T) {
	h := NewPublishDraft(nil)
	signal := &jobrunner.PipelineSignal{
		IsDraft:     false,
		PRState:     "OPEN",
		CheckStatus: "SUCCESS",
	}
	assert.False(t, h.Match(signal))
}

func TestPublishDraft_Match_Bad_ChecksFailing(t *testing.T) {
	h := NewPublishDraft(nil)
	signal := &jobrunner.PipelineSignal{
		IsDraft:     true,
		PRState:     "OPEN",
		CheckStatus: "FAILURE",
	}
	assert.False(t, h.Match(signal))
}

func TestPublishDraft_Execute_Good(t *testing.T) {
	var calledURL string
	var calledMethod string
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		calledURL = r.URL.Path
		calledMethod = r.Method
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"number":316}`))
	}))
	defer server.Close()

	h := NewPublishDraft(&http.Client{})
	h.apiURL = server.URL

	signal := &jobrunner.PipelineSignal{
		PRNumber:  316,
		RepoOwner: "host-uk",
		RepoName:  "core",
		IsDraft:   true,
		PRState:   "OPEN",
	}

	result, err := h.Execute(context.Background(), signal)
	require.NoError(t, err)
	assert.True(t, result.Success)
	assert.Equal(t, "publish_draft", result.Action)
	assert.Equal(t, "/repos/host-uk/core/pulls/316", calledURL)
	assert.Equal(t, "PATCH", calledMethod)
}
```

**Step 2: Run test to verify it fails**

Run: `go test ./pkg/jobrunner/handlers/ -run TestPublishDraft -v -count=1`
Expected: FAIL — package does not exist.

**Step 3: Write the implementation**

```go
package handlers

import (
	"bytes"
	"context"
	"fmt"
	"net/http"
	"time"

	"github.com/host-uk/core/pkg/jobrunner"
)

// PublishDraft marks a draft PR as ready for review.
type PublishDraft struct {
	client *http.Client
	apiURL string
}

// NewPublishDraft creates a publish_draft handler.
// Pass nil client to use http.DefaultClient.
func NewPublishDraft(client *http.Client) *PublishDraft {
	if client == nil {
		client = http.DefaultClient
	}
	return &PublishDraft{
		client: client,
		apiURL: "https://api.github.com",
	}
}

func (h *PublishDraft) Name() string { return "publish_draft" }

// Match returns true for open draft PRs with passing checks.
func (h *PublishDraft) Match(s *jobrunner.PipelineSignal) bool {
	return s.IsDraft && s.PRState == "OPEN" && s.CheckStatus == "SUCCESS"
}

// Execute calls PATCH /repos/:owner/:repo/pulls/:number with draft=false.
func (h *PublishDraft) Execute(ctx context.Context, s *jobrunner.PipelineSignal) (*jobrunner.ActionResult, error) {
	url := fmt.Sprintf("%s/repos/%s/%s/pulls/%d", h.apiURL, s.RepoOwner, s.RepoName, s.PRNumber)
	body := bytes.NewBufferString(`{"draft":false}`)

	req, err := http.NewRequestWithContext(ctx, "PATCH", url, body)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Accept", "application/vnd.github+json")
	req.Header.Set("Content-Type", "application/json")

	resp, err := h.client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	result := &jobrunner.ActionResult{
		Action:    "publish_draft",
		RepoOwner: s.RepoOwner,
		RepoName:  s.RepoName,
		PRNumber:  s.PRNumber,
		Timestamp: time.Now().UTC(),
	}

	if resp.StatusCode >= 200 && resp.StatusCode < 300 {
		result.Success = true
	} else {
		result.Error = fmt.Sprintf("HTTP %d", resp.StatusCode)
	}

	return result, nil
}
```

**Step 4: Run tests**

Run: `go test ./pkg/jobrunner/handlers/ -v -count=1`
Expected: PASS.

**Step 5: Commit**

```bash
git add pkg/jobrunner/handlers/
git commit -m "feat(jobrunner): add publish_draft handler"
```

---

### Task 6: Send Fix Command Handler

**Files:**
- Create: `pkg/jobrunner/handlers/send_fix_command.go`
- Test: `pkg/jobrunner/handlers/send_fix_command_test.go`

**Step 1: Write the test**

```go
package handlers

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/host-uk/core/pkg/jobrunner"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestSendFixCommand_Match_Good_Conflicting(t *testing.T) {
	h := NewSendFixCommand(nil)
	signal := &jobrunner.PipelineSignal{
		PRState:   "OPEN",
		Mergeable: "CONFLICTING",
	}
	assert.True(t, h.Match(signal))
}

func TestSendFixCommand_Match_Good_UnresolvedThreads(t *testing.T) {
	h := NewSendFixCommand(nil)
	signal := &jobrunner.PipelineSignal{
		PRState:         "OPEN",
		Mergeable:       "MERGEABLE",
		ThreadsTotal:    3,
		ThreadsResolved: 1,
		CheckStatus:     "FAILURE",
	}
	assert.True(t, h.Match(signal))
}

func TestSendFixCommand_Match_Bad_Clean(t *testing.T) {
	h := NewSendFixCommand(nil)
	signal := &jobrunner.PipelineSignal{
		PRState:         "OPEN",
		Mergeable:       "MERGEABLE",
		CheckStatus:     "SUCCESS",
		ThreadsTotal:    0,
		ThreadsResolved: 0,
	}
	assert.False(t, h.Match(signal))
}

func TestSendFixCommand_Execute_Good_Conflict(t *testing.T) {
	var postedBody map[string]string
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		json.NewDecoder(r.Body).Decode(&postedBody)
		w.WriteHeader(http.StatusCreated)
		w.Write([]byte(`{"id":1}`))
	}))
	defer server.Close()

	h := NewSendFixCommand(&http.Client{})
	h.apiURL = server.URL

	signal := &jobrunner.PipelineSignal{
		PRNumber:  296,
		RepoOwner: "host-uk",
		RepoName:  "core",
		PRState:   "OPEN",
		Mergeable: "CONFLICTING",
	}

	result, err := h.Execute(context.Background(), signal)
	require.NoError(t, err)
	assert.True(t, result.Success)
	assert.Contains(t, postedBody["body"], "fix the merge conflict")
}
```

**Step 2: Run test to verify it fails**

Run: `go test ./pkg/jobrunner/handlers/ -run TestSendFixCommand -v -count=1`
Expected: FAIL — `NewSendFixCommand` undefined.

**Step 3: Write the implementation**

```go
package handlers

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/host-uk/core/pkg/jobrunner"
)

// SendFixCommand comments on a PR to request a fix.
type SendFixCommand struct {
	client *http.Client
	apiURL string
}

func NewSendFixCommand(client *http.Client) *SendFixCommand {
	if client == nil {
		client = http.DefaultClient
	}
	return &SendFixCommand{client: client, apiURL: "https://api.github.com"}
}

func (h *SendFixCommand) Name() string { return "send_fix_command" }

// Match returns true for open PRs that are conflicting OR have unresolved
// review threads with failing checks (indicating reviews need fixing).
func (h *SendFixCommand) Match(s *jobrunner.PipelineSignal) bool {
	if s.PRState != "OPEN" {
		return false
	}
	if s.Mergeable == "CONFLICTING" {
		return true
	}
	if s.HasUnresolvedThreads() && s.CheckStatus == "FAILURE" {
		return true
	}
	return false
}

// Execute posts a comment with the appropriate fix command.
func (h *SendFixCommand) Execute(ctx context.Context, s *jobrunner.PipelineSignal) (*jobrunner.ActionResult, error) {
	msg := "Can you fix the code reviews?"
	if s.Mergeable == "CONFLICTING" {
		msg = "Can you fix the merge conflict?"
	}

	url := fmt.Sprintf("%s/repos/%s/%s/issues/%d/comments", h.apiURL, s.RepoOwner, s.RepoName, s.PRNumber)
	payload, _ := json.Marshal(map[string]string{"body": msg})

	req, err := http.NewRequestWithContext(ctx, "POST", url, bytes.NewReader(payload))
	if err != nil {
		return nil, err
	}
	req.Header.Set("Accept", "application/vnd.github+json")
	req.Header.Set("Content-Type", "application/json")

	resp, err := h.client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	result := &jobrunner.ActionResult{
		Action:    "send_fix_command",
		RepoOwner: s.RepoOwner,
		RepoName:  s.RepoName,
		PRNumber:  s.PRNumber,
		Timestamp: time.Now().UTC(),
	}

	if resp.StatusCode == http.StatusCreated {
		result.Success = true
	} else {
		result.Error = fmt.Sprintf("HTTP %d", resp.StatusCode)
	}

	return result, nil
}
```

**Step 4: Run tests**

Run: `go test ./pkg/jobrunner/handlers/ -v -count=1`
Expected: PASS.

**Step 5: Commit**

```bash
git add pkg/jobrunner/handlers/send_fix_command.go pkg/jobrunner/handlers/send_fix_command_test.go
git commit -m "feat(jobrunner): add send_fix_command handler"
```

---

### Task 7: Remaining Handlers (enable_auto_merge, tick_parent, close_child)

**Files:**
- Create: `pkg/jobrunner/handlers/enable_auto_merge.go` + test
- Create: `pkg/jobrunner/handlers/tick_parent.go` + test
- Create: `pkg/jobrunner/handlers/close_child.go` + test

**Context:** Same pattern as Tasks 5-6. Each handler: Match checks signal conditions, Execute calls GitHub REST API. Tests use httptest.

**Step 1: Write tests for all three** (one test file per handler, same pattern as above)

**enable_auto_merge:**
- Match: `PRState=OPEN && Mergeable=MERGEABLE && CheckStatus=SUCCESS && !IsDraft && ThreadsTotal==ThreadsResolved`
- Execute: `PUT /repos/:owner/:repo/pulls/:number/merge` with `merge_method=squash` — actually, auto-merge uses `gh api` to enable. For REST: `POST /repos/:owner/:repo/pulls/:number/merge` — No. Auto-merge is enabled via GraphQL `enablePullRequestAutoMerge`. For REST fallback, use: `PATCH /repos/:owner/:repo/pulls/:number` — No, that's not right either.

Actually, auto-merge via REST requires: `PUT /repos/:owner/:repo/pulls/:number/auto_merge`. This is not a standard GitHub REST endpoint. Auto-merge is enabled via the GraphQL API:

```graphql
mutation { enablePullRequestAutoMerge(input: {pullRequestId: "..."}) { ... } }
```

**Simpler approach:** Shell out to `gh pr merge --auto <number> -R owner/repo`. This is what the pipeline flow does today. Let's use `os/exec` with the `gh` CLI.

```go
// enable_auto_merge.go
package handlers

import (
	"context"
	"fmt"
	"os/exec"
	"time"

	"github.com/host-uk/core/pkg/jobrunner"
)

type EnableAutoMerge struct{}

func NewEnableAutoMerge() *EnableAutoMerge { return &EnableAutoMerge{} }

func (h *EnableAutoMerge) Name() string { return "enable_auto_merge" }

func (h *EnableAutoMerge) Match(s *jobrunner.PipelineSignal) bool {
	return s.PRState == "OPEN" &&
		!s.IsDraft &&
		s.Mergeable == "MERGEABLE" &&
		s.CheckStatus == "SUCCESS" &&
		!s.HasUnresolvedThreads()
}

func (h *EnableAutoMerge) Execute(ctx context.Context, s *jobrunner.PipelineSignal) (*jobrunner.ActionResult, error) {
	cmd := exec.CommandContext(ctx, "gh", "pr", "merge", "--auto",
		fmt.Sprintf("%d", s.PRNumber),
		"-R", s.RepoFullName(),
	)
	output, err := cmd.CombinedOutput()

	result := &jobrunner.ActionResult{
		Action:    "enable_auto_merge",
		RepoOwner: s.RepoOwner,
		RepoName:  s.RepoName,
		PRNumber:  s.PRNumber,
		Timestamp: time.Now().UTC(),
	}

	if err != nil {
		result.Error = fmt.Sprintf("%v: %s", err, string(output))
	} else {
		result.Success = true
	}

	return result, nil
}
```

**tick_parent and close_child** follow the same `gh` CLI pattern:
- `tick_parent`: Reads epic issue body, checks the child's checkbox, updates via `gh issue edit`
- `close_child`: `gh issue close <number> -R owner/repo`

**Step 2-5:** Same TDD cycle as Tasks 5-6. Write test, verify fail, implement, verify pass, commit.

For brevity, the exact test code follows the same pattern. Key test assertions:
- `tick_parent`: Verify `gh issue edit` is called with updated body
- `close_child`: Verify `gh issue close` is called
- `enable_auto_merge`: Verify `gh pr merge --auto` is called

**Testability:** Use a command factory variable for mocking `exec.Command`:

```go
// In each handler file:
var execCommand = exec.CommandContext

// In tests:
originalExecCommand := execCommand
defer func() { execCommand = originalExecCommand }()
execCommand = func(ctx context.Context, name string, args ...string) *exec.Cmd {
    // return a mock command
}
```

**Step 6: Commit**

```bash
git add pkg/jobrunner/handlers/enable_auto_merge.go pkg/jobrunner/handlers/enable_auto_merge_test.go
git add pkg/jobrunner/handlers/tick_parent.go pkg/jobrunner/handlers/tick_parent_test.go
git add pkg/jobrunner/handlers/close_child.go pkg/jobrunner/handlers/close_child_test.go
git commit -m "feat(jobrunner): add enable_auto_merge, tick_parent, close_child handlers"
```

---

### Task 8: Resolve Threads Handler

**Files:**
- Create: `pkg/jobrunner/handlers/resolve_threads.go`
- Test: `pkg/jobrunner/handlers/resolve_threads_test.go`

**Context:** This handler is special — it needs GraphQL to resolve review threads (no REST endpoint exists). Use a minimal GraphQL client (raw `net/http` POST to `https://api.github.com/graphql`).

**Step 1: Write the test**

```go
package handlers

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/host-uk/core/pkg/jobrunner"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestResolveThreads_Match_Good(t *testing.T) {
	h := NewResolveThreads(nil)
	signal := &jobrunner.PipelineSignal{
		PRState:         "OPEN",
		ThreadsTotal:    3,
		ThreadsResolved: 1,
	}
	assert.True(t, h.Match(signal))
}

func TestResolveThreads_Match_Bad_AllResolved(t *testing.T) {
	h := NewResolveThreads(nil)
	signal := &jobrunner.PipelineSignal{
		PRState:         "OPEN",
		ThreadsTotal:    3,
		ThreadsResolved: 3,
	}
	assert.False(t, h.Match(signal))
}

func TestResolveThreads_Execute_Good(t *testing.T) {
	callCount := 0
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		callCount++
		var req map[string]any
		json.NewDecoder(r.Body).Decode(&req)

		query := req["query"].(string)

		// First call: fetch threads
		if callCount == 1 {
			json.NewEncoder(w).Encode(map[string]any{
				"data": map[string]any{
					"repository": map[string]any{
						"pullRequest": map[string]any{
							"reviewThreads": map[string]any{
								"nodes": []map[string]any{
									{"id": "PRRT_1", "isResolved": false},
									{"id": "PRRT_2", "isResolved": true},
								},
							},
						},
					},
				},
			})
			return
		}

		// Subsequent calls: resolve thread
		json.NewEncoder(w).Encode(map[string]any{
			"data": map[string]any{
				"resolveReviewThread": map[string]any{
					"thread": map[string]any{"isResolved": true},
				},
			},
		})
	}))
	defer server.Close()

	h := NewResolveThreads(&http.Client{})
	h.graphqlURL = server.URL

	signal := &jobrunner.PipelineSignal{
		PRNumber:        315,
		RepoOwner:       "host-uk",
		RepoName:        "core",
		PRState:         "OPEN",
		ThreadsTotal:    2,
		ThreadsResolved: 1,
	}

	result, err := h.Execute(context.Background(), signal)
	require.NoError(t, err)
	assert.True(t, result.Success)
	assert.Equal(t, 2, callCount) // 1 fetch + 1 resolve (only PRRT_1 unresolved)
}
```

**Step 2: Run test to verify it fails**

Run: `go test ./pkg/jobrunner/handlers/ -run TestResolveThreads -v -count=1`
Expected: FAIL — `NewResolveThreads` undefined.

**Step 3: Write the implementation**

```go
package handlers

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/host-uk/core/pkg/jobrunner"
)

// ResolveThreads resolves all unresolved review threads on a PR.
type ResolveThreads struct {
	client     *http.Client
	graphqlURL string
}

func NewResolveThreads(client *http.Client) *ResolveThreads {
	if client == nil {
		client = http.DefaultClient
	}
	return &ResolveThreads{
		client:     client,
		graphqlURL: "https://api.github.com/graphql",
	}
}

func (h *ResolveThreads) Name() string { return "resolve_threads" }

func (h *ResolveThreads) Match(s *jobrunner.PipelineSignal) bool {
	return s.PRState == "OPEN" && s.HasUnresolvedThreads()
}

func (h *ResolveThreads) Execute(ctx context.Context, s *jobrunner.PipelineSignal) (*jobrunner.ActionResult, error) {
	// 1. Fetch unresolved thread IDs
	threadIDs, err := h.fetchUnresolvedThreads(ctx, s.RepoOwner, s.RepoName, s.PRNumber)
	if err != nil {
		return nil, fmt.Errorf("fetch threads: %w", err)
	}

	// 2. Resolve each thread
	resolved := 0
	for _, id := range threadIDs {
		if err := h.resolveThread(ctx, id); err != nil {
			// Log but continue — some threads may not be resolvable
			continue
		}
		resolved++
	}

	result := &jobrunner.ActionResult{
		Action:    "resolve_threads",
		RepoOwner: s.RepoOwner,
		RepoName:  s.RepoName,
		PRNumber:  s.PRNumber,
		Success:   resolved > 0,
		Timestamp: time.Now().UTC(),
	}

	if resolved == 0 && len(threadIDs) > 0 {
		result.Error = fmt.Sprintf("0/%d threads resolved", len(threadIDs))
	}

	return result, nil
}

func (h *ResolveThreads) fetchUnresolvedThreads(ctx context.Context, owner, repo string, pr int) ([]string, error) {
	query := fmt.Sprintf(`{
		repository(owner: %q, name: %q) {
			pullRequest(number: %d) {
				reviewThreads(first: 100) {
					nodes { id isResolved }
				}
			}
		}
	}`, owner, repo, pr)

	resp, err := h.graphql(ctx, query)
	if err != nil {
		return nil, err
	}

	type thread struct {
		ID         string `json:"id"`
		IsResolved bool   `json:"isResolved"`
	}
	var result struct {
		Data struct {
			Repository struct {
				PullRequest struct {
					ReviewThreads struct {
						Nodes []thread `json:"nodes"`
					} `json:"reviewThreads"`
				} `json:"pullRequest"`
			} `json:"repository"`
		} `json:"data"`
	}

	if err := json.Unmarshal(resp, &result); err != nil {
		return nil, err
	}

	var ids []string
	for _, t := range result.Data.Repository.PullRequest.ReviewThreads.Nodes {
		if !t.IsResolved {
			ids = append(ids, t.ID)
		}
	}
	return ids, nil
}

func (h *ResolveThreads) resolveThread(ctx context.Context, threadID string) error {
	mutation := fmt.Sprintf(`mutation {
		resolveReviewThread(input: {threadId: %q}) {
			thread { isResolved }
		}
	}`, threadID)

	_, err := h.graphql(ctx, mutation)
	return err
}

func (h *ResolveThreads) graphql(ctx context.Context, query string) (json.RawMessage, error) {
	payload, _ := json.Marshal(map[string]string{"query": query})

	req, err := http.NewRequestWithContext(ctx, "POST", h.graphqlURL, bytes.NewReader(payload))
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := h.client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("GraphQL HTTP %d", resp.StatusCode)
	}

	var raw json.RawMessage
	err = json.NewDecoder(resp.Body).Decode(&raw)
	return raw, err
}
```

**Step 4: Run tests**

Run: `go test ./pkg/jobrunner/handlers/ -v -count=1`
Expected: PASS.

**Step 5: Commit**

```bash
git add pkg/jobrunner/handlers/resolve_threads.go pkg/jobrunner/handlers/resolve_threads_test.go
git commit -m "feat(jobrunner): add resolve_threads handler with GraphQL"
```

---

### Task 9: Headless Mode in core-ide

**Files:**
- Modify: `internal/core-ide/main.go`

**Context:** core-ide currently always creates a Wails app. We need to branch: headless starts the poller + MCP bridge directly; desktop mode keeps the existing Wails app with poller as an optional service.

Note: core-ide has its own `go.mod` (`github.com/host-uk/core/internal/core-ide`). The jobrunner package lives in the root module. We need to add the root module as a dependency of core-ide, OR move the handler wiring into the root module. **Simplest approach:** core-ide imports `github.com/host-uk/core/pkg/jobrunner` — this requires adding the root module as a dependency in core-ide's go.mod.

**Step 1: Update core-ide go.mod**

Run: `cd /Users/snider/Code/host-uk/core/internal/core-ide && go get github.com/host-uk/core/pkg/jobrunner`

If this fails because the package isn't published yet, use a `replace` directive temporarily:

```
replace github.com/host-uk/core => ../..
```

Then `go mod tidy`.

**Step 2: Modify main.go**

Add `--headless` flag parsing, `hasDisplay()` detection, and the headless startup path.

The headless path:
1. Create `cli.Daemon` with PID file + health server
2. Create `Journal` at `~/.core/journal/`
3. Create `GitHubSource` with repos from config/env
4. Create all handlers
5. Create `Poller` with sources + handlers + journal
6. Start daemon, run poller in goroutine, block on `daemon.Run(ctx)`

The desktop path:
- Existing Wails app code, unchanged for now
- Poller can be added as a Wails service later

```go
// At top of main():
headless := false
for _, arg := range os.Args[1:] {
    if arg == "--headless" {
        headless = true
    }
}

if headless || !hasDisplay() {
    startHeadless()
    return
}
// ... existing Wails app code ...
```

**Step 3: Run core-ide with --headless --dry-run to verify**

Run: `cd /Users/snider/Code/host-uk/core/internal/core-ide && go run . --headless --dry-run`
Expected: Starts, logs poll cycle, exits cleanly on Ctrl+C.

**Step 4: Commit**

```bash
git add internal/core-ide/main.go internal/core-ide/go.mod internal/core-ide/go.sum
git commit -m "feat(core-ide): add headless mode with job runner poller"
```

---

### Task 10: Register Handlers as MCP Tools

**Files:**
- Modify: `internal/core-ide/mcp_bridge.go`

**Context:** Register each JobHandler as an MCP tool so they're callable via the HTTP API (POST /mcp/call). This lets external tools invoke handlers manually.

**Step 1: Add handler registration to MCPBridge**

Add a `handlers` field and register them in `ServiceStartup`. Add a `job_*` prefix to distinguish from webview tools.

```go
// In handleMCPTools — append job handler tools to the tool list
// In handleMCPCall — add a job_* dispatch path
```

**Step 2: Test via curl**

Run: `curl -X POST http://localhost:9877/mcp/call -d '{"tool":"job_publish_draft","params":{"pr":316,"owner":"host-uk","repo":"core"}}'`
Expected: Returns handler result JSON.

**Step 3: Commit**

```bash
git add internal/core-ide/mcp_bridge.go
git commit -m "feat(core-ide): register job handlers as MCP tools"
```

---

### Task 11: Updater Integration in core-ide

**Files:**
- Modify: `internal/core-ide/main.go` (headless startup path)

**Context:** Wire the existing `internal/cmd/updater` package into core-ide's headless startup. Check for updates on startup, auto-apply in headless mode.

**Step 1: Add updater to headless startup**

```go
// In startHeadless(), before starting poller:
updaterSvc, err := updater.NewUpdateService(updater.UpdateServiceConfig{
    RepoURL:        "https://github.com/host-uk/core",
    Channel:        "alpha",
    CheckOnStartup: updater.CheckAndUpdateOnStartup,
})
if err == nil {
    _ = updaterSvc.Start() // will auto-update and restart if newer version exists
}
```

**Step 2: Test by running headless**

Run: `core-ide --headless` — should check for updates on startup, then start polling.

**Step 3: Commit**

```bash
git add internal/core-ide/main.go
git commit -m "feat(core-ide): integrate updater for headless auto-update"
```

---

### Task 12: Systemd Service File

**Files:**
- Create: `internal/core-ide/build/linux/core-ide.service`

**Step 1: Write the systemd unit**

```ini
[Unit]
Description=Core IDE Job Runner
After=network-online.target
Wants=network-online.target

[Service]
Type=simple
ExecStart=/usr/local/bin/core-ide --headless
Restart=always
RestartSec=10
Environment=CORE_DAEMON=1
Environment=GITHUB_TOKEN=

[Install]
WantedBy=multi-user.target
```

**Step 2: Add to nfpm.yaml** so it's included in the Linux package:

In `internal/core-ide/build/linux/nfpm/nfpm.yaml`, add to `contents`:
```yaml
- src: ../core-ide.service
  dst: /etc/systemd/system/core-ide.service
  type: config
```

**Step 3: Commit**

```bash
git add internal/core-ide/build/linux/core-ide.service internal/core-ide/build/linux/nfpm/nfpm.yaml
git commit -m "feat(core-ide): add systemd service for headless mode"
```

---

### Task 13: Run Full Test Suite

**Step 1: Run all jobrunner tests**

Run: `go test ./pkg/jobrunner/... -v -count=1`
Expected: All tests pass.

**Step 2: Run core-ide build**

Run: `cd /Users/snider/Code/host-uk/core/internal/core-ide && go build -o /dev/null .`
Expected: Builds without errors.

**Step 3: Run dry-run integration test**

Run: `cd /Users/snider/Code/host-uk/core/internal/core-ide && go run . --headless --dry-run`
Expected: Polls GitHub, logs signals, takes no actions, exits on Ctrl+C.

---

## Batch Execution Plan

| Batch | Tasks | Description |
|-------|-------|-------------|
| 0 | 0 | Go workspace setup |
| 1 | 1-2 | Core types + Journal |
| 2 | 3-4 | Poller + GitHub Source |
| 3 | 5-8 | All handlers |
| 4 | 9-11 | core-ide integration (headless, MCP, updater) |
| 5 | 12-13 | Systemd + verification |
