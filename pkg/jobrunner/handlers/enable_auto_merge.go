package handlers

import (
	"context"
	"fmt"
	"time"

	"forge.lthn.ai/core/go/pkg/forge"
	"forge.lthn.ai/core/go/pkg/jobrunner"
)

// EnableAutoMergeHandler merges a PR that is ready using squash strategy.
type EnableAutoMergeHandler struct {
	forge *forge.Client
}

// NewEnableAutoMergeHandler creates a handler that merges ready PRs.
func NewEnableAutoMergeHandler(f *forge.Client) *EnableAutoMergeHandler {
	return &EnableAutoMergeHandler{forge: f}
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

// Execute merges the pull request with squash strategy.
func (h *EnableAutoMergeHandler) Execute(ctx context.Context, signal *jobrunner.PipelineSignal) (*jobrunner.ActionResult, error) {
	start := time.Now()

	err := h.forge.MergePullRequest(signal.RepoOwner, signal.RepoName, int64(signal.PRNumber), "squash")

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
		result.Error = fmt.Sprintf("merge failed: %v", err)
	}

	return result, nil
}
