package forge

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"

	forgejo "codeberg.org/mvdkleijn/forgejo-sdk/forgejo/v2"

	"forge.lthn.ai/core/cli/pkg/log"
)

// MergePullRequest merges a pull request with the given method ("squash", "rebase", "merge").
func (c *Client) MergePullRequest(owner, repo string, index int64, method string) error {
	style := forgejo.MergeStyleMerge
	switch method {
	case "squash":
		style = forgejo.MergeStyleSquash
	case "rebase":
		style = forgejo.MergeStyleRebase
	}

	merged, _, err := c.api.MergePullRequest(owner, repo, index, forgejo.MergePullRequestOption{
		Style:                  style,
		DeleteBranchAfterMerge: true,
	})
	if err != nil {
		return log.E("forge.MergePullRequest", "failed to merge pull request", err)
	}
	if !merged {
		return log.E("forge.MergePullRequest", fmt.Sprintf("merge returned false for %s/%s#%d", owner, repo, index), nil)
	}
	return nil
}

// SetPRDraft sets or clears the draft status on a pull request.
// The Forgejo SDK v2.2.0 doesn't expose the draft field on EditPullRequestOption,
// so we use a raw HTTP PATCH request.
func (c *Client) SetPRDraft(owner, repo string, index int64, draft bool) error {
	payload := map[string]bool{"draft": draft}
	body, err := json.Marshal(payload)
	if err != nil {
		return log.E("forge.SetPRDraft", "marshal payload", err)
	}

	url := fmt.Sprintf("%s/api/v1/repos/%s/%s/pulls/%d", c.url, owner, repo, index)
	req, err := http.NewRequest(http.MethodPatch, url, bytes.NewReader(body))
	if err != nil {
		return log.E("forge.SetPRDraft", "create request", err)
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "token "+c.token)

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return log.E("forge.SetPRDraft", "failed to update draft status", err)
	}
	defer func() { _ = resp.Body.Close() }()

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return log.E("forge.SetPRDraft", fmt.Sprintf("unexpected status %d", resp.StatusCode), nil)
	}
	return nil
}

// ListPRReviews returns all reviews for a pull request.
func (c *Client) ListPRReviews(owner, repo string, index int64) ([]*forgejo.PullReview, error) {
	var all []*forgejo.PullReview
	page := 1

	for {
		reviews, resp, err := c.api.ListPullReviews(owner, repo, index, forgejo.ListPullReviewsOptions{
			ListOptions: forgejo.ListOptions{Page: page, PageSize: 50},
		})
		if err != nil {
			return nil, log.E("forge.ListPRReviews", "failed to list reviews", err)
		}

		all = append(all, reviews...)

		if resp == nil || page >= resp.LastPage {
			break
		}
		page++
	}

	return all, nil
}

// GetCombinedStatus returns the combined commit status for a ref (SHA or branch).
func (c *Client) GetCombinedStatus(owner, repo string, ref string) (*forgejo.CombinedStatus, error) {
	status, _, err := c.api.GetCombinedStatus(owner, repo, ref)
	if err != nil {
		return nil, log.E("forge.GetCombinedStatus", "failed to get combined status", err)
	}
	return status, nil
}

// DismissReview dismisses a pull request review by ID.
func (c *Client) DismissReview(owner, repo string, index, reviewID int64, message string) error {
	_, err := c.api.DismissPullReview(owner, repo, index, reviewID, forgejo.DismissPullReviewOptions{
		Message: message,
	})
	if err != nil {
		return log.E("forge.DismissReview", "failed to dismiss review", err)
	}
	return nil
}
