package forge

import (
	"path"
	"strings"

	"forge.lthn.ai/core/cli/pkg/cli"
)

// splitOwnerRepo splits "owner/repo" into its parts.
func splitOwnerRepo(s string) (string, string, error) {
	parts := strings.SplitN(s, "/", 2)
	if len(parts) != 2 || parts[0] == "" || parts[1] == "" {
		return "", "", cli.Err("expected format: owner/repo (got %q)", s)
	}
	return parts[0], parts[1], nil
}

// strPtr returns a pointer to the given string.
func strPtr(s string) *string { return &s }

// extractRepoName extracts a repository name from a clone URL.
// e.g. "https://github.com/owner/repo.git" -> "repo"
func extractRepoName(cloneURL string) string {
	// Get the last path segment
	name := path.Base(cloneURL)
	// Strip .git suffix
	name = strings.TrimSuffix(name, ".git")
	if name == "" || name == "." || name == "/" {
		return ""
	}
	return name
}
