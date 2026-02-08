package forge

import (
	forgejo "codeberg.org/mvdkleijn/forgejo-sdk/forgejo/v2"

	"github.com/host-uk/core/pkg/log"
)

// CreateRepoWebhook creates a webhook on a repository.
func (c *Client) CreateRepoWebhook(owner, repo string, opts forgejo.CreateHookOption) (*forgejo.Hook, error) {
	hook, _, err := c.api.CreateRepoHook(owner, repo, opts)
	if err != nil {
		return nil, log.E("forge.CreateRepoWebhook", "failed to create repo webhook", err)
	}

	return hook, nil
}

// ListRepoWebhooks returns all webhooks for a repository.
func (c *Client) ListRepoWebhooks(owner, repo string) ([]*forgejo.Hook, error) {
	var all []*forgejo.Hook
	page := 1

	for {
		hooks, resp, err := c.api.ListRepoHooks(owner, repo, forgejo.ListHooksOptions{
			ListOptions: forgejo.ListOptions{Page: page, PageSize: 50},
		})
		if err != nil {
			return nil, log.E("forge.ListRepoWebhooks", "failed to list repo webhooks", err)
		}

		all = append(all, hooks...)

		if resp == nil || page >= resp.LastPage {
			break
		}
		page++
	}

	return all, nil
}
