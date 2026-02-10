package handlers

import (
	"context"
	"fmt"
	"time"

	"github.com/host-uk/core/pkg/forge"
	"github.com/host-uk/core/pkg/jobrunner"
)

const (
	ColorAgentComplete = "#0e8a16" // Green
)

// CompletionHandler manages issue state when an agent finishes work.
type CompletionHandler struct {
	forge *forge.Client
}

// NewCompletionHandler creates a handler for agent completion events.
func NewCompletionHandler(client *forge.Client) *CompletionHandler {
	return &CompletionHandler{
		forge: client,
	}
}

// Name returns the handler identifier.
func (h *CompletionHandler) Name() string {
	return "completion"
}

// Match returns true if the signal indicates an agent has finished a task.
func (h *CompletionHandler) Match(signal *jobrunner.PipelineSignal) bool {
	return signal.Type == "agent_completion"
}

// Execute updates the issue labels based on the completion status.
func (h *CompletionHandler) Execute(ctx context.Context, signal *jobrunner.PipelineSignal) (*jobrunner.ActionResult, error) {
	start := time.Now()

	// Remove in-progress label.
	if inProgressLabel, err := h.forge.GetLabelByName(signal.RepoOwner, signal.RepoName, LabelInProgress); err == nil {
		_ = h.forge.RemoveIssueLabel(signal.RepoOwner, signal.RepoName, int64(signal.ChildNumber), inProgressLabel.ID)
	}

	if signal.Success {
		completeLabel, err := h.forge.EnsureLabel(signal.RepoOwner, signal.RepoName, LabelAgentComplete, ColorAgentComplete)
		if err != nil {
			return nil, fmt.Errorf("ensure label %s: %w", LabelAgentComplete, err)
		}

		if err := h.forge.AddIssueLabels(signal.RepoOwner, signal.RepoName, int64(signal.ChildNumber), []int64{completeLabel.ID}); err != nil {
			return nil, fmt.Errorf("add completed label: %w", err)
		}

		if signal.Message != "" {
			_ = h.forge.CreateIssueComment(signal.RepoOwner, signal.RepoName, int64(signal.ChildNumber), signal.Message)
		}
	} else {
		failedLabel, err := h.forge.EnsureLabel(signal.RepoOwner, signal.RepoName, LabelAgentFailed, ColorAgentFailed)
		if err != nil {
			return nil, fmt.Errorf("ensure label %s: %w", LabelAgentFailed, err)
		}

		if err := h.forge.AddIssueLabels(signal.RepoOwner, signal.RepoName, int64(signal.ChildNumber), []int64{failedLabel.ID}); err != nil {
			return nil, fmt.Errorf("add failed label: %w", err)
		}

		msg := "Agent reported failure."
		if signal.Error != "" {
			msg += fmt.Sprintf("\n\nError: %s", signal.Error)
		}
		_ = h.forge.CreateIssueComment(signal.RepoOwner, signal.RepoName, int64(signal.ChildNumber), msg)
	}

	return &jobrunner.ActionResult{
		Action:      "completion",
		RepoOwner:   signal.RepoOwner,
		RepoName:    signal.RepoName,
		EpicNumber:  signal.EpicNumber,
		ChildNumber: signal.ChildNumber,
		Success:     true,
		Timestamp:   time.Now(),
		Duration:    time.Since(start),
	}, nil
}
