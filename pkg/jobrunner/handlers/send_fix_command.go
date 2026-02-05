package handlers

import (
	"bytes"
	"context"
	"fmt"
	"net/http"
	"time"

	"github.com/host-uk/core/pkg/jobrunner"
)

// SendFixCommandHandler posts a comment on a PR asking for conflict or
// review fixes.
type SendFixCommandHandler struct {
	client *http.Client
	apiURL string
}

// NewSendFixCommandHandler creates a handler that posts fix commands.
// If client is nil, http.DefaultClient is used.
// If apiURL is empty, the default GitHub API URL is used.
func NewSendFixCommandHandler(client *http.Client, apiURL string) *SendFixCommandHandler {
	if client == nil {
		client = http.DefaultClient
	}
	if apiURL == "" {
		apiURL = defaultAPIURL
	}
	return &SendFixCommandHandler{client: client, apiURL: apiURL}
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

// Execute posts a comment on the PR issue asking for a fix.
func (h *SendFixCommandHandler) Execute(ctx context.Context, signal *jobrunner.PipelineSignal) (*jobrunner.ActionResult, error) {
	start := time.Now()

	var message string
	if signal.Mergeable == "CONFLICTING" {
		message = "Can you fix the merge conflict?"
	} else {
		message = "Can you fix the code reviews?"
	}

	url := fmt.Sprintf("%s/repos/%s/%s/issues/%d/comments", h.apiURL, signal.RepoOwner, signal.RepoName, signal.PRNumber)
	bodyStr := fmt.Sprintf(`{"body":%q}`, message)
	body := bytes.NewBufferString(bodyStr)

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, url, body)
	if err != nil {
		return nil, fmt.Errorf("send_fix_command: create request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Accept", "application/vnd.github+json")

	resp, err := h.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("send_fix_command: execute request: %w", err)
	}
	defer func() { _ = resp.Body.Close() }()

	success := resp.StatusCode >= 200 && resp.StatusCode < 300
	result := &jobrunner.ActionResult{
		Action:    "send_fix_command",
		RepoOwner: signal.RepoOwner,
		RepoName:  signal.RepoName,
		PRNumber:  signal.PRNumber,
		Success:   success,
		Timestamp: time.Now(),
		Duration:  time.Since(start),
	}

	if !success {
		result.Error = fmt.Sprintf("unexpected status %d", resp.StatusCode)
	}

	return result, nil
}
