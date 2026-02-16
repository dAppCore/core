package handlers

import (
	"context"
	"fmt"
	"time"

	"forge.lthn.ai/core/go/pkg/forge"
	"forge.lthn.ai/core/go/pkg/jobrunner"
)

// PublishDraftHandler marks a draft PR as ready for review once its checks pass.
type PublishDraftHandler struct {
	forge *forge.Client
}

// NewPublishDraftHandler creates a handler that publishes draft PRs.
func NewPublishDraftHandler(f *forge.Client) *PublishDraftHandler {
	return &PublishDraftHandler{forge: f}
}

// Name returns the handler identifier.
func (h *PublishDraftHandler) Name() string {
	return "publish_draft"
}

// Match returns true when the PR is a draft, open, and all checks have passed.
func (h *PublishDraftHandler) Match(signal *jobrunner.PipelineSignal) bool {
	return signal.IsDraft &&
		signal.PRState == "OPEN" &&
		signal.CheckStatus == "SUCCESS"
}

// Execute marks the PR as no longer a draft.
func (h *PublishDraftHandler) Execute(ctx context.Context, signal *jobrunner.PipelineSignal) (*jobrunner.ActionResult, error) {
	start := time.Now()

	err := h.forge.SetPRDraft(signal.RepoOwner, signal.RepoName, int64(signal.PRNumber), false)

	result := &jobrunner.ActionResult{
		Action:    "publish_draft",
		RepoOwner: signal.RepoOwner,
		RepoName:  signal.RepoName,
		PRNumber:  signal.PRNumber,
		Success:   err == nil,
		Timestamp: time.Now(),
		Duration:  time.Since(start),
	}

	if err != nil {
		result.Error = fmt.Sprintf("publish draft failed: %v", err)
	}

	return result, nil
}
