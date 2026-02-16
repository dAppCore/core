package forge

import (
	"fmt"
	"strings"

	forgejo "codeberg.org/mvdkleijn/forgejo-sdk/forgejo/v2"

	"forge.lthn.ai/core/cli/pkg/log"
)

// ListOrgLabels returns all labels for repos in the given organisation.
// Note: The Forgejo SDK does not have a dedicated org-level labels endpoint.
// This lists labels from the first repo found, which works when orgs use shared label sets.
// For org-wide label management, use ListRepoLabels with a specific repo.
func (c *Client) ListOrgLabels(org string) ([]*forgejo.Label, error) {
	// Forgejo doesn't expose org-level labels via SDK — list repos and aggregate unique labels.
	repos, err := c.ListOrgRepos(org)
	if err != nil {
		return nil, err
	}

	if len(repos) == 0 {
		return nil, nil
	}

	// Use the first repo's labels as representative of the org's label set.
	return c.ListRepoLabels(repos[0].Owner.UserName, repos[0].Name)
}

// ListRepoLabels returns all labels for a repository.
func (c *Client) ListRepoLabels(owner, repo string) ([]*forgejo.Label, error) {
	var all []*forgejo.Label
	page := 1

	for {
		labels, resp, err := c.api.ListRepoLabels(owner, repo, forgejo.ListLabelsOptions{
			ListOptions: forgejo.ListOptions{Page: page, PageSize: 50},
		})
		if err != nil {
			return nil, log.E("forge.ListRepoLabels", "failed to list repo labels", err)
		}

		all = append(all, labels...)

		if resp == nil || page >= resp.LastPage {
			break
		}
		page++
	}

	return all, nil
}

// CreateRepoLabel creates a label on a repository.
func (c *Client) CreateRepoLabel(owner, repo string, opts forgejo.CreateLabelOption) (*forgejo.Label, error) {
	label, _, err := c.api.CreateLabel(owner, repo, opts)
	if err != nil {
		return nil, log.E("forge.CreateRepoLabel", "failed to create repo label", err)
	}

	return label, nil
}

// GetLabelByName retrieves a specific label by name from a repository.
func (c *Client) GetLabelByName(owner, repo, name string) (*forgejo.Label, error) {
	labels, err := c.ListRepoLabels(owner, repo)
	if err != nil {
		return nil, err
	}

	for _, l := range labels {
		if strings.EqualFold(l.Name, name) {
			return l, nil
		}
	}

	return nil, fmt.Errorf("label %s not found in %s/%s", name, owner, repo)
}

// EnsureLabel checks if a label exists, and creates it if it doesn't.
func (c *Client) EnsureLabel(owner, repo, name, color string) (*forgejo.Label, error) {
	label, err := c.GetLabelByName(owner, repo, name)
	if err == nil {
		return label, nil
	}

	return c.CreateRepoLabel(owner, repo, forgejo.CreateLabelOption{
		Name:  name,
		Color: color,
	})
}

// AddIssueLabels adds labels to an issue.
func (c *Client) AddIssueLabels(owner, repo string, number int64, labelIDs []int64) error {
	_, _, err := c.api.AddIssueLabels(owner, repo, number, forgejo.IssueLabelsOption{
		Labels: labelIDs,
	})
	if err != nil {
		return log.E("forge.AddIssueLabels", "failed to add labels to issue", err)
	}
	return nil
}

// RemoveIssueLabel removes a label from an issue.
func (c *Client) RemoveIssueLabel(owner, repo string, number int64, labelID int64) error {
	_, err := c.api.DeleteIssueLabel(owner, repo, number, labelID)
	if err != nil {
		return log.E("forge.RemoveIssueLabel", "failed to remove label from issue", err)
	}
	return nil
}
