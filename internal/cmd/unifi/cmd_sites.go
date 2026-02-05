package unifi

import (
	"fmt"

	"github.com/host-uk/core/pkg/cli"
	uf "github.com/host-uk/core/pkg/unifi"
)

// addSitesCommand adds the 'sites' subcommand for listing UniFi sites.
func addSitesCommand(parent *cli.Command) {
	cmd := &cli.Command{
		Use:   "sites",
		Short: "List controller sites",
		Long:  "List all sites configured on the UniFi controller.",
		RunE: func(cmd *cli.Command, args []string) error {
			return runSites()
		},
	}

	parent.AddCommand(cmd)
}

func runSites() error {
	client, err := uf.NewFromConfig("", "", "", "")
	if err != nil {
		return err
	}

	sites, err := client.GetSites()
	if err != nil {
		return err
	}

	if len(sites) == 0 {
		cli.Text("No sites found.")
		return nil
	}

	table := cli.NewTable("Name", "Description")

	for _, s := range sites {
		table.AddRow(
			valueStyle.Render(s.Name),
			dimStyle.Render(s.Desc),
		)
	}

	cli.Blank()
	cli.Print("  %s\n\n", fmt.Sprintf("%d sites", len(sites)))
	table.Render()

	return nil
}
