// Package forge provides CLI commands for managing a Forgejo instance.
//
// Commands:
//   - config: Configure Forgejo connection (URL, token)
//   - status: Show instance status and version
//   - repos: List repositories
//   - issues: List and create issues
//   - prs: List pull requests
//   - migrate: Migrate repos from external services
//   - sync: Sync GitHub repos to Forgejo upstream branches
//   - orgs: List organisations
//   - labels: List and create labels
package forge

import (
	"github.com/host-uk/core/pkg/cli"
)

func init() {
	cli.RegisterCommands(AddForgeCommands)
}

// Style aliases from shared package.
var (
	successStyle = cli.SuccessStyle
	errorStyle   = cli.ErrorStyle
	warningStyle = cli.WarningStyle
	dimStyle     = cli.DimStyle
	valueStyle   = cli.ValueStyle
	repoStyle    = cli.RepoStyle
	numberStyle  = cli.NumberStyle
	infoStyle    = cli.InfoStyle
)

// AddForgeCommands registers the 'forge' command and all subcommands.
func AddForgeCommands(root *cli.Command) {
	forgeCmd := &cli.Command{
		Use:   "forge",
		Short: "Forgejo instance management",
		Long:  "Manage repositories, issues, pull requests, and organisations on your Forgejo instance.",
	}
	root.AddCommand(forgeCmd)

	addConfigCommand(forgeCmd)
	addStatusCommand(forgeCmd)
	addReposCommand(forgeCmd)
	addIssuesCommand(forgeCmd)
	addPRsCommand(forgeCmd)
	addMigrateCommand(forgeCmd)
	addSyncCommand(forgeCmd)
	addOrgsCommand(forgeCmd)
	addLabelsCommand(forgeCmd)
}
