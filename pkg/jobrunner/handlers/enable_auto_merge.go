package handlers

import (
	"context"
	"fmt"
	"time"

	"github.com/host-uk/core/pkg/jobrunner"
)

// EnableAutoMergeHandler enables squash auto-merge on a PR that is ready.
type EnableAutoMergeHandler struct{}

// NewEnableAutoMergeHandler creates a handler that enables auto-merge.
func NewEnableAutoMergeHandler() *EnableAutoMergeHandler {
	return &EnableAutoMergeHandler{}
}

// Name returns the handler identifier.
func (h *EnableAutoMergeHandler) Name() string {
	return "enable_auto_merge"
}

// Match returns true when the PR is open, not a draft, mergeable, checks
// are passing, and there are no unresolved review threads.
func (h *EnableAutoMergeHandler) Match(signal *jobrunner.PipelineSignal) bool {
	return signal.PRState == "OPEN" &&
		!signal.IsDraft &&
		signal.Mergeable == "MERGEABLE" &&
		signal.CheckStatus == "SUCCESS" &&
		!signal.HasUnresolvedThreads()
}

// Execute shells out to gh to enable auto-merge with squash strategy.
func (h *EnableAutoMergeHandler) Execute(ctx context.Context, signal *jobrunner.PipelineSignal) (*jobrunner.ActionResult, error) {
	start := time.Now()

	repoFlag := fmt.Sprintf("%s/%s", signal.RepoOwner, signal.RepoName)
	prNumber := fmt.Sprintf("%d", signal.PRNumber)

	cmd := execCommand(ctx, "gh", "pr", "merge", "--auto", "--squash", prNumber, "-R", repoFlag)
	output, err := cmd.CombinedOutput()

	result := &jobrunner.ActionResult{
		Action:    "enable_auto_merge",
		RepoOwner: signal.RepoOwner,
		RepoName:  signal.RepoName,
		PRNumber:  signal.PRNumber,
		Success:   err == nil,
		Timestamp: time.Now(),
		Duration:  time.Since(start),
	}

	if err != nil {
		result.Error = fmt.Sprintf("gh pr merge failed: %v: %s", err, string(output))
	}

	return result, nil
}
