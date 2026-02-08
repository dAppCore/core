package forge

import (
	"fmt"

	"github.com/host-uk/core/pkg/cli"
	fg "github.com/host-uk/core/pkg/forge"
)

// Auth command flags.
var (
	authURL   string
	authToken string
)

// addAuthCommand adds the 'auth' subcommand for authentication status and login.
func addAuthCommand(parent *cli.Command) {
	cmd := &cli.Command{
		Use:   "auth",
		Short: "Show authentication status",
		Long:  "Show the current Forgejo authentication status, or log in with a new token.",
		RunE: func(cmd *cli.Command, args []string) error {
			return runAuth()
		},
	}

	cmd.Flags().StringVar(&authURL, "url", "", "Forgejo instance URL")
	cmd.Flags().StringVar(&authToken, "token", "", "API token (create at <url>/user/settings/applications)")

	parent.AddCommand(cmd)
}

func runAuth() error {
	// If credentials provided, save them first
	if authURL != "" || authToken != "" {
		if err := fg.SaveConfig(authURL, authToken); err != nil {
			return err
		}
		if authURL != "" {
			cli.Success(fmt.Sprintf("URL set to %s", authURL))
		}
		if authToken != "" {
			cli.Success("Token saved")
		}
	}

	// Always show current auth status
	url, token, err := fg.ResolveConfig(authURL, authToken)
	if err != nil {
		return err
	}

	if token == "" {
		cli.Blank()
		cli.Print("  %s %s\n", dimStyle.Render("URL:"), valueStyle.Render(url))
		cli.Print("  %s %s\n", dimStyle.Render("Auth:"), warningStyle.Render("not authenticated"))
		cli.Print("  %s %s\n", dimStyle.Render("Hint:"), dimStyle.Render(fmt.Sprintf("core forge auth --token TOKEN (create at %s/user/settings/applications)", url)))
		cli.Blank()
		return nil
	}

	client, err := fg.NewFromConfig(authURL, authToken)
	if err != nil {
		return err
	}

	user, _, err := client.API().GetMyUserInfo()
	if err != nil {
		cli.Blank()
		cli.Print("  %s %s\n", dimStyle.Render("URL:"), valueStyle.Render(url))
		cli.Print("  %s %s\n", dimStyle.Render("Auth:"), errorStyle.Render("token invalid or expired"))
		cli.Blank()
		return nil
	}

	cli.Blank()
	cli.Success(fmt.Sprintf("Authenticated to %s", client.URL()))
	cli.Print("  %s %s\n", dimStyle.Render("User:"), valueStyle.Render(user.UserName))
	cli.Print("  %s %s\n", dimStyle.Render("Email:"), valueStyle.Render(user.Email))
	if user.IsAdmin {
		cli.Print("  %s %s\n", dimStyle.Render("Role:"), infoStyle.Render("admin"))
	}
	cli.Blank()

	return nil
}
