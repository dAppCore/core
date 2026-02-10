package bugseti

import (
	"fmt"
	"os/exec"
)

// CheckGHCLI verifies that the gh CLI is installed and authenticated.
// Returns nil if gh is available and logged in, or an error with
// actionable instructions for the user.
func CheckGHCLI() error {
	// Check if gh is in PATH
	if _, err := exec.LookPath("gh"); err != nil {
		return fmt.Errorf("gh CLI not found in PATH: %w\n\n"+
			"BugSETI requires the GitHub CLI (gh) to fetch issues and submit PRs.\n"+
			"Install it from: https://cli.github.com\n\n"+
			"  macOS:   brew install gh\n"+
			"  Linux:   https://github.com/cli/cli/blob/trunk/docs/install_linux.md\n"+
			"  Windows: winget install --id GitHub.cli", err)
	}

	// Check if gh is authenticated
	cmd := exec.Command("gh", "auth", "status")
	if out, err := cmd.CombinedOutput(); err != nil {
		return fmt.Errorf("gh CLI is not authenticated: %w\n%s\n\n"+
			"Run 'gh auth login' to authenticate with GitHub.", err, out)
	}

	return nil
}
