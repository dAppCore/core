package forge

import (
	"fmt"

	forgejo "codeberg.org/mvdkleijn/forgejo-sdk/forgejo/v2"

	"github.com/host-uk/core/pkg/cli"
	fg "github.com/host-uk/core/pkg/forge"
)

// Repos command flags.
var (
	reposOrg     string
	reposMirrors bool
)

// addReposCommand adds the 'repos' subcommand for listing repositories.
func addReposCommand(parent *cli.Command) {
	cmd := &cli.Command{
		Use:   "repos",
		Short: "List repositories",
		Long:  "List repositories from your Forgejo instance, optionally filtered by organisation or mirror status.",
		RunE: func(cmd *cli.Command, args []string) error {
			return runRepos()
		},
	}

	cmd.Flags().StringVar(&reposOrg, "org", "", "Filter by organisation")
	cmd.Flags().BoolVar(&reposMirrors, "mirrors", false, "Show only mirror repositories")

	parent.AddCommand(cmd)
}

func runRepos() error {
	client, err := fg.NewFromConfig("", "")
	if err != nil {
		return err
	}

	var repos []*forgejo.Repository
	if reposOrg != "" {
		repos, err = client.ListOrgRepos(reposOrg)
	} else {
		repos, err = client.ListUserRepos()
	}
	if err != nil {
		return err
	}

	// Filter mirrors if requested
	if reposMirrors {
		var filtered []*forgejo.Repository
		for _, r := range repos {
			if r.Mirror {
				filtered = append(filtered, r)
			}
		}
		repos = filtered
	}

	if len(repos) == 0 {
		cli.Text("No repositories found.")
		return nil
	}

	// Build table
	table := cli.NewTable("Name", "Type", "Visibility", "Stars")

	for _, r := range repos {
		repoType := "source"
		if r.Mirror {
			repoType = "mirror"
		}

		visibility := successStyle.Render("public")
		if r.Private {
			visibility = warningStyle.Render("private")
		}

		table.AddRow(
			repoStyle.Render(r.FullName),
			dimStyle.Render(repoType),
			visibility,
			fmt.Sprintf("%d", r.Stars),
		)
	}

	cli.Blank()
	cli.Print("  %s\n\n", fmt.Sprintf("%d repositories", len(repos)))
	table.Render()

	return nil
}
