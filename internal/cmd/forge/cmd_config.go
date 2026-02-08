package forge

import (
	"fmt"

	"github.com/host-uk/core/pkg/cli"
	fg "github.com/host-uk/core/pkg/forge"
)

// Config command flags.
var (
	configURL   string
	configToken string
	configTest  bool
)

// addConfigCommand adds the 'config' subcommand for Forgejo connection setup.
func addConfigCommand(parent *cli.Command) {
	cmd := &cli.Command{
		Use:   "config",
		Short: "Configure Forgejo connection",
		Long:  "Set the Forgejo instance URL and API token, or test the current connection.",
		RunE: func(cmd *cli.Command, args []string) error {
			return runConfig()
		},
	}

	cmd.Flags().StringVar(&configURL, "url", "", "Forgejo instance URL")
	cmd.Flags().StringVar(&configToken, "token", "", "Forgejo API token")
	cmd.Flags().BoolVar(&configTest, "test", false, "Test the current connection")

	parent.AddCommand(cmd)
}

func runConfig() error {
	// If setting values, save them first
	if configURL != "" || configToken != "" {
		if err := fg.SaveConfig(configURL, configToken); err != nil {
			return err
		}

		if configURL != "" {
			cli.Success(fmt.Sprintf("Forgejo URL set to %s", configURL))
		}
		if configToken != "" {
			cli.Success("Forgejo token saved")
		}
	}

	// If testing, verify the connection
	if configTest {
		return runConfigTest()
	}

	// If no flags, show current config
	if configURL == "" && configToken == "" && !configTest {
		return showConfig()
	}

	return nil
}

func showConfig() error {
	url, token, err := fg.ResolveConfig("", "")
	if err != nil {
		return err
	}

	cli.Blank()
	cli.Print("  %s %s\n", dimStyle.Render("URL:"), valueStyle.Render(url))

	if token != "" {
		masked := token
		if len(token) >= 8 {
			masked = token[:4] + "..." + token[len(token)-4:]
		}
		cli.Print("  %s %s\n", dimStyle.Render("Token:"), valueStyle.Render(masked))
	} else {
		cli.Print("  %s %s\n", dimStyle.Render("Token:"), warningStyle.Render("not set"))
	}

	cli.Blank()

	return nil
}

func runConfigTest() error {
	client, err := fg.NewFromConfig(configURL, configToken)
	if err != nil {
		return err
	}

	user, _, err := client.API().GetMyUserInfo()
	if err != nil {
		cli.Error("Connection failed")
		return cli.WrapVerb(err, "connect to", "Forgejo")
	}

	cli.Blank()
	cli.Success(fmt.Sprintf("Connected to %s", client.URL()))
	cli.Print("  %s %s\n", dimStyle.Render("User:"), valueStyle.Render(user.UserName))
	cli.Print("  %s %s\n", dimStyle.Render("Email:"), valueStyle.Render(user.Email))
	cli.Blank()

	return nil
}
