// Package forge provides a thin wrapper around the Forgejo Go SDK
// for managing repositories, issues, and pull requests on a Forgejo instance.
//
// Authentication is resolved from config file, environment variables, or flag overrides:
//
//  1. ~/.core/config.yaml keys: forge.token, forge.url
//  2. FORGE_TOKEN + FORGE_URL environment variables (override config file)
//  3. Flag overrides via core forge config --url/--token (highest priority)
package forge

import (
	forgejo "codeberg.org/mvdkleijn/forgejo-sdk/forgejo/v2"

	"github.com/host-uk/core/pkg/log"
)

// Client wraps the Forgejo SDK client with config-based auth.
type Client struct {
	api *forgejo.Client
	url string
}

// New creates a new Forgejo API client for the given URL and token.
func New(url, token string) (*Client, error) {
	api, err := forgejo.NewClient(url, forgejo.SetToken(token))
	if err != nil {
		return nil, log.E("forge.New", "failed to create client", err)
	}

	return &Client{api: api, url: url}, nil
}

// API exposes the underlying SDK client for direct access.
func (c *Client) API() *forgejo.Client { return c.api }

// URL returns the Forgejo instance URL.
func (c *Client) URL() string { return c.url }
