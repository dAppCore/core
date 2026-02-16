package prod

import (
	"fmt"
	"os"
	"os/exec"
	"syscall"

	"forge.lthn.ai/core/cli/pkg/cli"
	"github.com/spf13/cobra"
)

var sshCmd = &cobra.Command{
	Use:   "ssh <host>",
	Short: "SSH into a production host",
	Long: `Open an SSH session to a production host defined in infra.yaml.

Examples:
  core prod ssh noc
  core prod ssh de
  core prod ssh de2
  core prod ssh build`,
	Args: cobra.ExactArgs(1),
	RunE: runSSH,
}

func runSSH(cmd *cobra.Command, args []string) error {
	cfg, _, err := loadConfig()
	if err != nil {
		return err
	}

	name := args[0]
	host, ok := cfg.Hosts[name]
	if !ok {
		// List available hosts
		cli.Print("Unknown host '%s'. Available:\n", name)
		for n, h := range cfg.Hosts {
			cli.Print("  %s  %s  (%s)\n", cli.BoldStyle.Render(n), h.IP, h.Role)
		}
		return fmt.Errorf("host '%s' not found in infra.yaml", name)
	}

	sshArgs := []string{
		"ssh",
		"-i", host.SSH.Key,
		"-p", fmt.Sprintf("%d", host.SSH.Port),
		"-o", "StrictHostKeyChecking=accept-new",
		fmt.Sprintf("%s@%s", host.SSH.User, host.IP),
	}

	cli.Print("%s %s@%s (%s)\n",
		cli.BoldStyle.Render("▶"),
		host.SSH.User, host.FQDN,
		cli.DimStyle.Render(host.IP))

	sshPath, err := exec.LookPath("ssh")
	if err != nil {
		return fmt.Errorf("ssh not found: %w", err)
	}

	// Replace current process with SSH
	return syscall.Exec(sshPath, sshArgs, os.Environ())
}
