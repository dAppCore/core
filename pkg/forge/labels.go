package forge

import (
	forgejo "codeberg.org/mvdkleijn/forgejo-sdk/forgejo/v2"

	"github.com/host-uk/core/pkg/log"
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
