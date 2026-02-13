package bugseti

import (
	"github.com/host-uk/core/pkg/forge"
)

// CheckForge verifies that the Forgejo API is configured and reachable.
// Returns nil if a token is configured and the API responds, or an error
// with actionable instructions for the user.
func CheckForge() (*forge.Client, error) {
	client, err := forge.NewFromConfig("", "")
	if err != nil {
		return nil, err
	}

	// Verify the token works by fetching the current user
	if _, err := client.GetCurrentUser(); err != nil {
		return nil, err
	}

	return client, nil
}
