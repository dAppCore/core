package prod

import (
	"github.com/spf13/cobra"
)

var (
	infraFile string
)

// Cmd is the root prod command.
var Cmd = &cobra.Command{
	Use:   "prod",
	Short: "Production infrastructure management",
	Long: `Manage the Host UK production infrastructure.

Commands:
  status    Show infrastructure health and connectivity
  setup     Phase 1: discover topology, create LB, configure DNS
  dns       Manage DNS records via CloudNS
  lb        Manage Hetzner load balancer
  ssh       SSH into a production host

Configuration is read from infra.yaml in the project root.`,
}

func init() {
	Cmd.PersistentFlags().StringVar(&infraFile, "config", "", "Path to infra.yaml (auto-discovered if not set)")

	Cmd.AddCommand(statusCmd)
	Cmd.AddCommand(setupCmd)
	Cmd.AddCommand(dnsCmd)
	Cmd.AddCommand(lbCmd)
	Cmd.AddCommand(sshCmd)
}
