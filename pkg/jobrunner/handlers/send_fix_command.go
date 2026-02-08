package handlers

import (
	"context"
	"fmt"
	"time"

	"github.com/host-uk/core/pkg/forge"
	"github.com/host-uk/core/pkg/jobrunner"
)

// SendFixCommandHandler posts a comment on a PR asking for conflict or
// review fixes.
type SendFixCommandHandler struct {
	forge *forge.Client
}

// NewSendFixCommandHandler creates a handler that posts fix commands.
func NewSendFixCommandHandler(f *forge.Client) *SendFixCommandHandler {
	return &SendFixCommandHandler{forge: f}
}

// Name returns the handler identifier.
func (h *SendFixCommandHandler) Name() string {
	return "send_fix_command"
}

// Match returns true when the PR is open and either has merge conflicts or
// has unresolved threads with failing checks.
func (h *SendFixCommandHandler) Match(signal *jobrunner.PipelineSignal) bool {
	if signal.PRState != "OPEN" {
		return false
	}
	if signal.Mergeable == "CONFLICTING" {
		return true
	}
	if signal.HasUnresolvedThreads() && signal.CheckStatus == "FAILURE" {
		return true
	}
	return false
}

// Execute posts a comment on the PR asking for a fix.
func (h *SendFixCommandHandler) Execute(ctx context.Context, signal *jobrunner.PipelineSignal) (*jobrunner.ActionResult, error) {
	start := time.Now()

	var message string
	if signal.Mergeable == "CONFLICTING" {
		message = "Can you fix the merge conflict?"
	} else {
		message = "Can you fix the code reviews?"
	}

	err := h.forge.CreateIssueComment(
		signal.RepoOwner, signal.RepoName,
		int64(signal.PRNumber), message,
	)

	result := &jobrunner.ActionResult{
		Action:    "send_fix_command",
		RepoOwner: signal.RepoOwner,
		RepoName:  signal.RepoName,
		PRNumber:  signal.PRNumber,
		Success:   err == nil,
		Timestamp: time.Now(),
		Duration:  time.Since(start),
	}

	if err != nil {
		result.Error = fmt.Sprintf("post comment failed: %v", err)
	}

	return result, nil
}
