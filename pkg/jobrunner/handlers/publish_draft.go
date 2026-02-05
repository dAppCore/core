package handlers

import (
	"bytes"
	"context"
	"fmt"
	"net/http"
	"time"

	"github.com/host-uk/core/pkg/jobrunner"
)

const defaultAPIURL = "https://api.github.com"

// PublishDraftHandler marks a draft PR as ready for review once its checks pass.
type PublishDraftHandler struct {
	client *http.Client
	apiURL string
}

// NewPublishDraftHandler creates a handler that publishes draft PRs.
// If client is nil, http.DefaultClient is used.
// If apiURL is empty, the default GitHub API URL is used.
func NewPublishDraftHandler(client *http.Client, apiURL string) *PublishDraftHandler {
	if client == nil {
		client = http.DefaultClient
	}
	if apiURL == "" {
		apiURL = defaultAPIURL
	}
	return &PublishDraftHandler{client: client, apiURL: apiURL}
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

// Execute patches the PR to mark it as no longer a draft.
func (h *PublishDraftHandler) Execute(ctx context.Context, signal *jobrunner.PipelineSignal) (*jobrunner.ActionResult, error) {
	start := time.Now()
	url := fmt.Sprintf("%s/repos/%s/%s/pulls/%d", h.apiURL, signal.RepoOwner, signal.RepoName, signal.PRNumber)

	body := bytes.NewBufferString(`{"draft":false}`)
	req, err := http.NewRequestWithContext(ctx, http.MethodPatch, url, body)
	if err != nil {
		return nil, fmt.Errorf("publish_draft: create request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Accept", "application/vnd.github+json")

	resp, err := h.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("publish_draft: execute request: %w", err)
	}
	defer func() { _ = resp.Body.Close() }()

	success := resp.StatusCode >= 200 && resp.StatusCode < 300
	result := &jobrunner.ActionResult{
		Action:    "publish_draft",
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
