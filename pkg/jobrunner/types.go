package jobrunner

import (
	"context"
	"time"
)

// PipelineSignal is the structural snapshot of a child issue/PR.
// Carries structural state plus issue title/body for dispatch prompts.
type PipelineSignal struct {
	EpicNumber      int
	ChildNumber     int
	PRNumber        int
	RepoOwner       string
	RepoName        string
	PRState         string // OPEN, MERGED, CLOSED
	IsDraft         bool
	Mergeable       string // MERGEABLE, CONFLICTING, UNKNOWN
	CheckStatus     string // SUCCESS, FAILURE, PENDING
	ThreadsTotal    int
	ThreadsResolved int
	LastCommitSHA   string
	LastCommitAt    time.Time
	LastReviewAt    time.Time
	NeedsCoding     bool   // true if child has no PR (work not started)
	Assignee        string // issue assignee username (for dispatch)
	IssueTitle      string // child issue title (for dispatch prompt)
	IssueBody       string // child issue body (for dispatch prompt)
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
	Action      string        `json:"action"`
	RepoOwner   string        `json:"repo_owner"`
	RepoName    string        `json:"repo_name"`
	EpicNumber  int           `json:"epic"`
	ChildNumber int           `json:"child"`
	PRNumber    int           `json:"pr"`
	Success     bool          `json:"success"`
	Error       string        `json:"error,omitempty"`
	Timestamp   time.Time     `json:"ts"`
	Duration    time.Duration `json:"duration_ms"`
	Cycle       int           `json:"cycle"`
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
