package forge

import (
	forgejo "codeberg.org/mvdkleijn/forgejo-sdk/forgejo/v2"

	"github.com/host-uk/core/pkg/log"
)

// ListIssuesOpts configures issue listing.
type ListIssuesOpts struct {
	State string // "open", "closed", "all"
	Page  int
	Limit int
}

// ListIssues returns issues for the given repository.
func (c *Client) ListIssues(owner, repo string, opts ListIssuesOpts) ([]*forgejo.Issue, error) {
	state := forgejo.StateOpen
	switch opts.State {
	case "closed":
		state = forgejo.StateClosed
	case "all":
		state = forgejo.StateAll
	}

	limit := opts.Limit
	if limit == 0 {
		limit = 50
	}

	page := opts.Page
	if page == 0 {
		page = 1
	}

	issues, _, err := c.api.ListRepoIssues(owner, repo, forgejo.ListIssueOption{
		ListOptions: forgejo.ListOptions{Page: page, PageSize: limit},
		State:       state,
		Type:        forgejo.IssueTypeIssue,
	})
	if err != nil {
		return nil, log.E("forge.ListIssues", "failed to list issues", err)
	}

	return issues, nil
}

// GetIssue returns a single issue by number.
func (c *Client) GetIssue(owner, repo string, number int64) (*forgejo.Issue, error) {
	issue, _, err := c.api.GetIssue(owner, repo, number)
	if err != nil {
		return nil, log.E("forge.GetIssue", "failed to get issue", err)
	}

	return issue, nil
}

// CreateIssue creates a new issue in the given repository.
func (c *Client) CreateIssue(owner, repo string, opts forgejo.CreateIssueOption) (*forgejo.Issue, error) {
	issue, _, err := c.api.CreateIssue(owner, repo, opts)
	if err != nil {
		return nil, log.E("forge.CreateIssue", "failed to create issue", err)
	}

	return issue, nil
}

// EditIssue edits an existing issue.
func (c *Client) EditIssue(owner, repo string, number int64, opts forgejo.EditIssueOption) (*forgejo.Issue, error) {
	issue, _, err := c.api.EditIssue(owner, repo, number, opts)
	if err != nil {
		return nil, log.E("forge.EditIssue", "failed to edit issue", err)
	}

	return issue, nil
}

// ListPullRequests returns pull requests for the given repository.
func (c *Client) ListPullRequests(owner, repo string, state string) ([]*forgejo.PullRequest, error) {
	st := forgejo.StateOpen
	switch state {
	case "closed":
		st = forgejo.StateClosed
	case "all":
		st = forgejo.StateAll
	}

	var all []*forgejo.PullRequest
	page := 1

	for {
		prs, resp, err := c.api.ListRepoPullRequests(owner, repo, forgejo.ListPullRequestsOptions{
			ListOptions: forgejo.ListOptions{Page: page, PageSize: 50},
			State:       st,
		})
		if err != nil {
			return nil, log.E("forge.ListPullRequests", "failed to list pull requests", err)
		}

		all = append(all, prs...)

		if resp == nil || page >= resp.LastPage {
			break
		}
		page++
	}

	return all, nil
}

// GetPullRequest returns a single pull request by number.
func (c *Client) GetPullRequest(owner, repo string, number int64) (*forgejo.PullRequest, error) {
	pr, _, err := c.api.GetPullRequest(owner, repo, number)
	if err != nil {
		return nil, log.E("forge.GetPullRequest", "failed to get pull request", err)
	}

	return pr, nil
}
