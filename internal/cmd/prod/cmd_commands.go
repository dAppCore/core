package prod

import (
	"forge.lthn.ai/core/cli/pkg/cli"
	"github.com/spf13/cobra"
)

func init() {
	cli.RegisterCommands(AddProdCommands)
}

// AddProdCommands registers the 'prod' command and all subcommands.
func AddProdCommands(root *cobra.Command) {
	root.AddCommand(Cmd)
}
