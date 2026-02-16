package handlers

import (
	"context"
	"fmt"
	"time"

	forgejosdk "codeberg.org/mvdkleijn/forgejo-sdk/forgejo/v2"

	"forge.lthn.ai/core/go/pkg/forge"
	"forge.lthn.ai/core/go/pkg/jobrunner"
)

// DismissReviewsHandler dismisses stale "request changes" reviews on a PR.
// This replaces the GitHub-only ResolveThreadsHandler because Forgejo does
// not have a thread resolution API.
type DismissReviewsHandler struct {
	forge *forge.Client
}

// NewDismissReviewsHandler creates a handler that dismisses stale reviews.
func NewDismissReviewsHandler(f *forge.Client) *DismissReviewsHandler {
	return &DismissReviewsHandler{forge: f}
}

// Name returns the handler identifier.
func (h *DismissReviewsHandler) Name() string {
	return "dismiss_reviews"
}

// Match returns true when the PR is open and has unresolved review threads.
func (h *DismissReviewsHandler) Match(signal *jobrunner.PipelineSignal) bool {
	return signal.PRState == "OPEN" && signal.HasUnresolvedThreads()
}

// Execute dismisses stale "request changes" reviews on the PR.
func (h *DismissReviewsHandler) Execute(ctx context.Context, signal *jobrunner.PipelineSignal) (*jobrunner.ActionResult, error) {
	start := time.Now()

	reviews, err := h.forge.ListPRReviews(signal.RepoOwner, signal.RepoName, int64(signal.PRNumber))
	if err != nil {
		return nil, fmt.Errorf("dismiss_reviews: list reviews: %w", err)
	}

	var dismissErrors []string
	dismissed := 0
	for _, review := range reviews {
		if review.State != forgejosdk.ReviewStateRequestChanges || review.Dismissed || !review.Stale {
			continue
		}

		if err := h.forge.DismissReview(
			signal.RepoOwner, signal.RepoName,
			int64(signal.PRNumber), review.ID,
			"Automatically dismissed: review is stale after new commits",
		); err != nil {
			dismissErrors = append(dismissErrors, err.Error())
		} else {
			dismissed++
		}
	}

	result := &jobrunner.ActionResult{
		Action:    "dismiss_reviews",
		RepoOwner: signal.RepoOwner,
		RepoName:  signal.RepoName,
		PRNumber:  signal.PRNumber,
		Success:   len(dismissErrors) == 0,
		Timestamp: time.Now(),
		Duration:  time.Since(start),
	}

	if len(dismissErrors) > 0 {
		result.Error = fmt.Sprintf("failed to dismiss %d review(s): %s",
			len(dismissErrors), dismissErrors[0])
	}

	return result, nil
}
