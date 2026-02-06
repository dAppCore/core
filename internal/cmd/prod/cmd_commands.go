package prod

import (
	"github.com/host-uk/core/pkg/cli"
	"github.com/spf13/cobra"
)

func init() {
	cli.RegisterCommands(AddProdCommands)
}

// AddProdCommands registers the 'prod' command and all subcommands.
func AddProdCommands(root *cobra.Command) {
	root.AddCommand(Cmd)
}
