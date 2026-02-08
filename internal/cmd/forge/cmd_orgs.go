package forge

import (
	"fmt"

	"github.com/host-uk/core/pkg/cli"
	fg "github.com/host-uk/core/pkg/forge"
)

// addOrgsCommand adds the 'orgs' subcommand for listing organisations.
func addOrgsCommand(parent *cli.Command) {
	cmd := &cli.Command{
		Use:   "orgs",
		Short: "List organisations",
		Long:  "List all organisations the authenticated user belongs to.",
		RunE: func(cmd *cli.Command, args []string) error {
			return runOrgs()
		},
	}

	parent.AddCommand(cmd)
}

func runOrgs() error {
	client, err := fg.NewFromConfig("", "")
	if err != nil {
		return err
	}

	orgs, err := client.ListMyOrgs()
	if err != nil {
		return err
	}

	if len(orgs) == 0 {
		cli.Text("No organisations found.")
		return nil
	}

	cli.Blank()
	cli.Print("  %s\n\n", fmt.Sprintf("%d organisations", len(orgs)))

	table := cli.NewTable("Name", "Visibility", "Description")

	for _, org := range orgs {
		visibility := successStyle.Render(org.Visibility)
		if org.Visibility == "private" {
			visibility = warningStyle.Render(org.Visibility)
		}

		desc := cli.Truncate(org.Description, 50)
		if desc == "" {
			desc = dimStyle.Render("-")
		}

		table.AddRow(
			repoStyle.Render(org.UserName),
			visibility,
			desc,
		)
	}

	table.Render()

	return nil
}
