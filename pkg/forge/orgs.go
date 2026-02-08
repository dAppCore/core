package forge

import (
	forgejo "codeberg.org/mvdkleijn/forgejo-sdk/forgejo/v2"

	"github.com/host-uk/core/pkg/log"
)

// ListMyOrgs returns all organisations for the authenticated user.
func (c *Client) ListMyOrgs() ([]*forgejo.Organization, error) {
	var all []*forgejo.Organization
	page := 1

	for {
		orgs, resp, err := c.api.ListMyOrgs(forgejo.ListOrgsOptions{
			ListOptions: forgejo.ListOptions{Page: page, PageSize: 50},
		})
		if err != nil {
			return nil, log.E("forge.ListMyOrgs", "failed to list orgs", err)
		}

		all = append(all, orgs...)

		if resp == nil || page >= resp.LastPage {
			break
		}
		page++
	}

	return all, nil
}

// GetOrg returns a single organisation by name.
func (c *Client) GetOrg(name string) (*forgejo.Organization, error) {
	org, _, err := c.api.GetOrg(name)
	if err != nil {
		return nil, log.E("forge.GetOrg", "failed to get org", err)
	}

	return org, nil
}

// CreateOrg creates a new organisation.
func (c *Client) CreateOrg(opts forgejo.CreateOrgOption) (*forgejo.Organization, error) {
	org, _, err := c.api.CreateOrg(opts)
	if err != nil {
		return nil, log.E("forge.CreateOrg", "failed to create org", err)
	}

	return org, nil
}
