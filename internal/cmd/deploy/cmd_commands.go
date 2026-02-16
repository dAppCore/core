package deploy

import (
	"forge.lthn.ai/core/cli/pkg/cli"
	"github.com/spf13/cobra"
)

func init() {
	cli.RegisterCommands(AddDeployCommands)
}

// AddDeployCommands registers the 'deploy' command and all subcommands.
func AddDeployCommands(root *cobra.Command) {
	root.AddCommand(Cmd)
}
