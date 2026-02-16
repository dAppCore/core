package forge

import (
	"fmt"

	forgejo "codeberg.org/mvdkleijn/forgejo-sdk/forgejo/v2"

	"forge.lthn.ai/core/cli/pkg/cli"
	fg "forge.lthn.ai/core/cli/pkg/forge"
)

// Labels command flags.
var (
	labelsCreate string
	labelsColor  string
	labelsRepo   string
)

// addLabelsCommand adds the 'labels' subcommand for listing and creating labels.
func addLabelsCommand(parent *cli.Command) {
	cmd := &cli.Command{
		Use:   "labels <org>",
		Short: "List and manage labels",
		Long: `List labels from an organisation's repos, or create a new label.

Labels are listed from the first repo in the organisation. Use --repo to target a specific repo.

Examples:
  core forge labels Private-Host-UK
  core forge labels Private-Host-UK --create "feature" --color "00aabb"
  core forge labels Private-Host-UK --repo Enchantrix`,
		Args: cli.ExactArgs(1),
		RunE: func(cmd *cli.Command, args []string) error {
			if labelsCreate != "" {
				return runCreateLabel(args[0])
			}
			return runListLabels(args[0])
		},
	}

	cmd.Flags().StringVar(&labelsCreate, "create", "", "Create a label with this name")
	cmd.Flags().StringVar(&labelsColor, "color", "0075ca", "Label colour (hex, e.g. 00aabb)")
	cmd.Flags().StringVar(&labelsRepo, "repo", "", "Target a specific repo (default: first org repo)")

	parent.AddCommand(cmd)
}

func runListLabels(org string) error {
	client, err := fg.NewFromConfig("", "")
	if err != nil {
		return err
	}

	var labels []*forgejo.Label
	if labelsRepo != "" {
		labels, err = client.ListRepoLabels(org, labelsRepo)
	} else {
		labels, err = client.ListOrgLabels(org)
	}
	if err != nil {
		return err
	}

	if len(labels) == 0 {
		cli.Text("No labels found.")
		return nil
	}

	cli.Blank()
	cli.Print("  %s\n\n", fmt.Sprintf("%d labels", len(labels)))

	table := cli.NewTable("Name", "Color", "Description")

	for _, l := range labels {
		table.AddRow(
			warningStyle.Render(l.Name),
			dimStyle.Render("#"+l.Color),
			cli.Truncate(l.Description, 50),
		)
	}

	table.Render()

	return nil
}

func runCreateLabel(org string) error {
	client, err := fg.NewFromConfig("", "")
	if err != nil {
		return err
	}

	// Determine target repo
	repo := labelsRepo
	if repo == "" {
		repos, err := client.ListOrgRepos(org)
		if err != nil {
			return err
		}
		if len(repos) == 0 {
			return cli.Err("no repos in org %s to create label on", org)
		}
		repo = repos[0].Name
		org = repos[0].Owner.UserName
	}

	label, err := client.CreateRepoLabel(org, repo, forgejo.CreateLabelOption{
		Name:  labelsCreate,
		Color: "#" + labelsColor,
	})
	if err != nil {
		return err
	}

	cli.Blank()
	cli.Success(fmt.Sprintf("Created label %q on %s/%s", label.Name, org, repo))
	cli.Blank()

	return nil
}
