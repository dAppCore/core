package forge

import (
	"os"

	"github.com/host-uk/core/pkg/config"
	"github.com/host-uk/core/pkg/log"
)

const (
	// ConfigKeyURL is the config key for the Forgejo instance URL.
	ConfigKeyURL = "forge.url"
	// ConfigKeyToken is the config key for the Forgejo API token.
	ConfigKeyToken = "forge.token"

	// DefaultURL is the default Forgejo instance URL.
	DefaultURL = "http://localhost:4000"
)

// NewFromConfig creates a Forgejo client using the standard config resolution:
//
//  1. ~/.core/config.yaml keys: forge.token, forge.url
//  2. FORGE_TOKEN + FORGE_URL environment variables (override config file)
//  3. Provided flag overrides (highest priority; pass empty to skip)
func NewFromConfig(flagURL, flagToken string) (*Client, error) {
	url, token, err := ResolveConfig(flagURL, flagToken)
	if err != nil {
		return nil, err
	}

	if token == "" {
		return nil, log.E("forge.NewFromConfig", "no API token configured (set FORGE_TOKEN or run: core forge config --token TOKEN)", nil)
	}

	return New(url, token)
}

// ResolveConfig resolves the Forgejo URL and token from all config sources.
// Flag values take highest priority, then env vars, then config file.
func ResolveConfig(flagURL, flagToken string) (url, token string, err error) {
	// Start with config file values
	cfg, cfgErr := config.New()
	if cfgErr == nil {
		_ = cfg.Get(ConfigKeyURL, &url)
		_ = cfg.Get(ConfigKeyToken, &token)
	}

	// Overlay environment variables
	if envURL := os.Getenv("FORGE_URL"); envURL != "" {
		url = envURL
	}
	if envToken := os.Getenv("FORGE_TOKEN"); envToken != "" {
		token = envToken
	}

	// Overlay flag values (highest priority)
	if flagURL != "" {
		url = flagURL
	}
	if flagToken != "" {
		token = flagToken
	}

	// Default URL if nothing configured
	if url == "" {
		url = DefaultURL
	}

	return url, token, nil
}

// SaveConfig persists the Forgejo URL and/or token to the config file.
func SaveConfig(url, token string) error {
	cfg, err := config.New()
	if err != nil {
		return log.E("forge.SaveConfig", "failed to load config", err)
	}

	if url != "" {
		if err := cfg.Set(ConfigKeyURL, url); err != nil {
			return log.E("forge.SaveConfig", "failed to save URL", err)
		}
	}

	if token != "" {
		if err := cfg.Set(ConfigKeyToken, token); err != nil {
			return log.E("forge.SaveConfig", "failed to save token", err)
		}
	}

	return nil
}
