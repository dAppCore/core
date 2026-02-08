package forge

import (
	"fmt"

	forgejo "codeberg.org/mvdkleijn/forgejo-sdk/forgejo/v2"

	"github.com/host-uk/core/pkg/cli"
	fg "github.com/host-uk/core/pkg/forge"
)

// Migrate command flags.
var (
	migrateOrg     string
	migrateService string
	migrateToken   string
	migrateMirror  bool
)

// addMigrateCommand adds the 'migrate' subcommand for importing repos from external services.
func addMigrateCommand(parent *cli.Command) {
	cmd := &cli.Command{
		Use:   "migrate <clone-url>",
		Short: "Migrate a repo from an external service",
		Long: `Migrate a repository from GitHub, GitLab, Gitea, or other services into Forgejo.

Unlike a simple mirror, migration imports issues, labels, pull requests, releases, and more.

Examples:
  core forge migrate https://github.com/owner/repo --org MyOrg --service github
  core forge migrate https://gitea.example.com/owner/repo --service gitea --token TOKEN`,
		Args: cli.ExactArgs(1),
		RunE: func(cmd *cli.Command, args []string) error {
			return runMigrate(args[0])
		},
	}

	cmd.Flags().StringVar(&migrateOrg, "org", "", "Forgejo organisation to migrate into (default: your user account)")
	cmd.Flags().StringVar(&migrateService, "service", "github", "Source service type (github, gitlab, gitea, forgejo, gogs, git)")
	cmd.Flags().StringVar(&migrateToken, "token", "", "Auth token for the source service")
	cmd.Flags().BoolVar(&migrateMirror, "mirror", false, "Set up as a mirror (periodic sync)")

	parent.AddCommand(cmd)
}

func runMigrate(cloneURL string) error {
	client, err := fg.NewFromConfig("", "")
	if err != nil {
		return err
	}

	// Determine target owner on Forgejo
	targetOwner := migrateOrg
	if targetOwner == "" {
		user, _, err := client.API().GetMyUserInfo()
		if err != nil {
			return cli.WrapVerb(err, "get", "current user")
		}
		targetOwner = user.UserName
	}

	// Extract repo name from clone URL
	repoName := extractRepoName(cloneURL)
	if repoName == "" {
		return cli.Err("could not extract repo name from URL: %s", cloneURL)
	}

	// Map service flag to SDK type
	service := mapServiceType(migrateService)

	cli.Print("  Migrating %s -> %s/%s on Forgejo...\n", cloneURL, targetOwner, repoName)

	opts := forgejo.MigrateRepoOption{
		RepoName:     repoName,
		RepoOwner:    targetOwner,
		CloneAddr:    cloneURL,
		Service:      service,
		Mirror:       migrateMirror,
		AuthToken:    migrateToken,
		Issues:       true,
		Labels:       true,
		PullRequests: true,
		Releases:     true,
		Milestones:   true,
		Wiki:         true,
		Description:  "Migrated from " + cloneURL,
	}

	repo, err := client.MigrateRepo(opts)
	if err != nil {
		return err
	}

	cli.Blank()
	cli.Success(fmt.Sprintf("Migration complete: %s", repo.FullName))
	cli.Print("  %s %s\n", dimStyle.Render("URL:"), valueStyle.Render(repo.HTMLURL))
	cli.Print("  %s %s\n", dimStyle.Render("Clone:"), valueStyle.Render(repo.CloneURL))
	if migrateMirror {
		cli.Print("  %s %s\n", dimStyle.Render("Type:"), dimStyle.Render("mirror (periodic sync)"))
	}
	cli.Blank()

	return nil
}

func mapServiceType(s string) forgejo.GitServiceType {
	switch s {
	case "github":
		return forgejo.GitServiceGithub
	case "gitlab":
		return forgejo.GitServiceGitlab
	case "gitea":
		return forgejo.GitServiceGitea
	case "forgejo":
		return forgejo.GitServiceForgejo
	case "gogs":
		return forgejo.GitServiceGogs
	default:
		return forgejo.GitServicePlain
	}
}
