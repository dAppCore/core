package handlers

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/host-uk/core/pkg/jobrunner"
)

// TickParentHandler ticks a child checkbox in the parent epic issue body
// after the child's PR has been merged.
type TickParentHandler struct{}

// NewTickParentHandler creates a handler that ticks parent epic checkboxes.
func NewTickParentHandler() *TickParentHandler {
	return &TickParentHandler{}
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
// child issue with a checked one, and updates the epic.
func (h *TickParentHandler) Execute(ctx context.Context, signal *jobrunner.PipelineSignal) (*jobrunner.ActionResult, error) {
	start := time.Now()
	repoFlag := signal.RepoFullName()

	// Fetch the epic issue body.
	viewCmd := execCommand(ctx, "gh", "issue", "view",
		fmt.Sprintf("%d", signal.EpicNumber),
		"-R", repoFlag,
		"--json", "body",
		"-q", ".body",
	)
	bodyBytes, err := viewCmd.Output()
	if err != nil {
		return nil, fmt.Errorf("tick_parent: fetch epic body: %w", err)
	}

	oldBody := string(bodyBytes)
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

	editCmd := execCommand(ctx, "gh", "issue", "edit",
		fmt.Sprintf("%d", signal.EpicNumber),
		"-R", repoFlag,
		"--body", newBody,
	)
	editOutput, err := editCmd.CombinedOutput()
	if err != nil {
		return &jobrunner.ActionResult{
			Action:    "tick_parent",
			RepoOwner: signal.RepoOwner,
			RepoName:  signal.RepoName,
			PRNumber:  signal.PRNumber,
			Error:     fmt.Sprintf("gh issue edit failed: %v: %s", err, string(editOutput)),
			Timestamp: time.Now(),
			Duration:  time.Since(start),
		}, nil
	}

	// Also close the child issue (design steps 8+9 combined).
	closeCmd := execCommand(ctx, "gh", "issue", "close",
		fmt.Sprintf("%d", signal.ChildNumber),
		"-R", repoFlag,
	)
	closeOutput, err := closeCmd.CombinedOutput()

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
		result.Error = fmt.Sprintf("gh issue close failed: %v: %s", err, string(closeOutput))
	}

	return result, nil
}
