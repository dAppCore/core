package forge

import (
	"fmt"

	"forge.lthn.ai/core/cli/pkg/cli"
	fg "forge.lthn.ai/core/cli/pkg/forge"
)

// addStatusCommand adds the 'status' subcommand for instance info.
func addStatusCommand(parent *cli.Command) {
	cmd := &cli.Command{
		Use:   "status",
		Short: "Show Forgejo instance status",
		Long:  "Display Forgejo instance version, authenticated user, and summary counts.",
		RunE: func(cmd *cli.Command, args []string) error {
			return runStatus()
		},
	}

	parent.AddCommand(cmd)
}

func runStatus() error {
	client, err := fg.NewFromConfig("", "")
	if err != nil {
		return err
	}

	// Get server version
	ver, _, err := client.API().ServerVersion()
	if err != nil {
		return cli.WrapVerb(err, "get", "server version")
	}

	// Get authenticated user
	user, _, err := client.API().GetMyUserInfo()
	if err != nil {
		return cli.WrapVerb(err, "get", "user info")
	}

	// Get org count
	orgs, err := client.ListMyOrgs()
	if err != nil {
		return cli.WrapVerb(err, "list", "organisations")
	}

	// Get repo count
	repos, err := client.ListUserRepos()
	if err != nil {
		return cli.WrapVerb(err, "list", "repositories")
	}

	cli.Blank()
	cli.Print("  %s %s\n", dimStyle.Render("Instance:"), valueStyle.Render(client.URL()))
	cli.Print("  %s %s\n", dimStyle.Render("Version:"), valueStyle.Render(ver))
	cli.Print("  %s %s\n", dimStyle.Render("User:"), valueStyle.Render(user.UserName))
	cli.Print("  %s %s\n", dimStyle.Render("Orgs:"), numberStyle.Render(fmt.Sprintf("%d", len(orgs))))
	cli.Print("  %s %s\n", dimStyle.Render("Repos:"), numberStyle.Render(fmt.Sprintf("%d", len(repos))))
	cli.Blank()

	return nil
}
