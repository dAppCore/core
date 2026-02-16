package handlers

import (
	"context"
	"fmt"
	"strings"
	"time"

	forgejosdk "codeberg.org/mvdkleijn/forgejo-sdk/forgejo/v2"

	"forge.lthn.ai/core/go/pkg/forge"
	"forge.lthn.ai/core/go/pkg/jobrunner"
)

// TickParentHandler ticks a child checkbox in the parent epic issue body
// after the child's PR has been merged.
type TickParentHandler struct {
	forge *forge.Client
}

// NewTickParentHandler creates a handler that ticks parent epic checkboxes.
func NewTickParentHandler(f *forge.Client) *TickParentHandler {
	return &TickParentHandler{forge: f}
}

// Name returns the handler identifier.
func (h *TickParentHandler) Name() string {
	return "tick_parent"
}

// Match returns true when the child PR has been merged.
func (h *TickParentHandler) Match(signal *jobrunner.PipelineSignal) bool {
	return signal.PRState == "MERGED"
}

// Execute fetches the epic body, replaces the unchecked checkbox for the
// child issue with a checked one, updates the epic, and closes the child issue.
func (h *TickParentHandler) Execute(ctx context.Context, signal *jobrunner.PipelineSignal) (*jobrunner.ActionResult, error) {
	start := time.Now()

	// Fetch the epic issue body.
	epic, err := h.forge.GetIssue(signal.RepoOwner, signal.RepoName, int64(signal.EpicNumber))
	if err != nil {
		return nil, fmt.Errorf("tick_parent: fetch epic: %w", err)
	}

	oldBody := epic.Body
	unchecked := fmt.Sprintf("- [ ] #%d", signal.ChildNumber)
	checked := fmt.Sprintf("- [x] #%d", signal.ChildNumber)

	if !strings.Contains(oldBody, unchecked) {
		// Already ticked or not found -- nothing to do.
		return &jobrunner.ActionResult{
			Action:    "tick_parent",
			RepoOwner: signal.RepoOwner,
			RepoName:  signal.RepoName,
			PRNumber:  signal.PRNumber,
			Success:   true,
			Timestamp: time.Now(),
			Duration:  time.Since(start),
		}, nil
	}

	newBody := strings.Replace(oldBody, unchecked, checked, 1)

	// Update the epic body.
	_, err = h.forge.EditIssue(signal.RepoOwner, signal.RepoName, int64(signal.EpicNumber), forgejosdk.EditIssueOption{
		Body: &newBody,
	})
	if err != nil {
		return &jobrunner.ActionResult{
			Action:    "tick_parent",
			RepoOwner: signal.RepoOwner,
			RepoName:  signal.RepoName,
			PRNumber:  signal.PRNumber,
			Error:     fmt.Sprintf("edit epic failed: %v", err),
			Timestamp: time.Now(),
			Duration:  time.Since(start),
		}, nil
	}

	// Close the child issue.
	err = h.forge.CloseIssue(signal.RepoOwner, signal.RepoName, int64(signal.ChildNumber))

	result := &jobrunner.ActionResult{
		Action:    "tick_parent",
		RepoOwner: signal.RepoOwner,
		RepoName:  signal.RepoName,
		PRNumber:  signal.PRNumber,
		Success:   err == nil,
		Timestamp: time.Now(),
		Duration:  time.Since(start),
	}

	if err != nil {
		result.Error = fmt.Sprintf("close child issue failed: %v", err)
	}

	return result, nil
}
