package forge

import (
	forgejo "codeberg.org/mvdkleijn/forgejo-sdk/forgejo/v2"

	"github.com/host-uk/core/pkg/log"
)

// ListOrgRepos returns all repositories for the given organisation.
func (c *Client) ListOrgRepos(org string) ([]*forgejo.Repository, error) {
	var all []*forgejo.Repository
	page := 1

	for {
		repos, resp, err := c.api.ListOrgRepos(org, forgejo.ListOrgReposOptions{
			ListOptions: forgejo.ListOptions{Page: page, PageSize: 50},
		})
		if err != nil {
			return nil, log.E("forge.ListOrgRepos", "failed to list org repos", err)
		}

		all = append(all, repos...)

		if resp == nil || page >= resp.LastPage {
			break
		}
		page++
	}

	return all, nil
}

// ListUserRepos returns all repositories for the authenticated user.
func (c *Client) ListUserRepos() ([]*forgejo.Repository, error) {
	var all []*forgejo.Repository
	page := 1

	for {
		repos, resp, err := c.api.ListMyRepos(forgejo.ListReposOptions{
			ListOptions: forgejo.ListOptions{Page: page, PageSize: 50},
		})
		if err != nil {
			return nil, log.E("forge.ListUserRepos", "failed to list user repos", err)
		}

		all = append(all, repos...)

		if resp == nil || page >= resp.LastPage {
			break
		}
		page++
	}

	return all, nil
}

// GetRepo returns a single repository by owner and name.
func (c *Client) GetRepo(owner, name string) (*forgejo.Repository, error) {
	repo, _, err := c.api.GetRepo(owner, name)
	if err != nil {
		return nil, log.E("forge.GetRepo", "failed to get repo", err)
	}

	return repo, nil
}

// CreateOrgRepo creates a new empty repository under an organisation.
func (c *Client) CreateOrgRepo(org string, opts forgejo.CreateRepoOption) (*forgejo.Repository, error) {
	repo, _, err := c.api.CreateOrgRepo(org, opts)
	if err != nil {
		return nil, log.E("forge.CreateOrgRepo", "failed to create org repo", err)
	}

	return repo, nil
}

// DeleteRepo deletes a repository from Forgejo.
func (c *Client) DeleteRepo(owner, name string) error {
	_, err := c.api.DeleteRepo(owner, name)
	if err != nil {
		return log.E("forge.DeleteRepo", "failed to delete repo", err)
	}

	return nil
}

// MigrateRepo migrates a repository from an external service using the Forgejo migration API.
// Unlike CreateMirror, this supports importing issues, labels, PRs, and more.
func (c *Client) MigrateRepo(opts forgejo.MigrateRepoOption) (*forgejo.Repository, error) {
	repo, _, err := c.api.MigrateRepo(opts)
	if err != nil {
		return nil, log.E("forge.MigrateRepo", "failed to migrate repo", err)
	}

	return repo, nil
}
