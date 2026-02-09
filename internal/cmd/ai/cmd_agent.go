package ai

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/host-uk/core/pkg/agentci"
	"github.com/host-uk/core/pkg/cli"
	"github.com/host-uk/core/pkg/config"
)

// AddAgentCommands registers the 'agent' subcommand group under 'ai'.
func AddAgentCommands(parent *cli.Command) {
	agentCmd := &cli.Command{
		Use:   "agent",
		Short: "Manage AgentCI dispatch targets",
	}

	agentCmd.AddCommand(agentAddCmd())
	agentCmd.AddCommand(agentListCmd())
	agentCmd.AddCommand(agentStatusCmd())
	agentCmd.AddCommand(agentLogsCmd())
	agentCmd.AddCommand(agentSetupCmd())
	agentCmd.AddCommand(agentRemoveCmd())

	parent.AddCommand(agentCmd)
}

func loadConfig() (*config.Config, error) {
	return config.New()
}

func agentAddCmd() *cli.Command {
	cmd := &cli.Command{
		Use:   "add <name> <user@host>",
		Short: "Add an agent to the config",
		Args:  cli.ExactArgs(2),
		RunE: func(cmd *cli.Command, args []string) error {
			name := args[0]
			host := args[1]

			forgejoUser, _ := cmd.Flags().GetString("forgejo-user")
			if forgejoUser == "" {
				forgejoUser = name
			}
			queueDir, _ := cmd.Flags().GetString("queue-dir")
			if queueDir == "" {
				queueDir = "/home/claude/ai-work/queue"
			}

			// Test SSH connectivity.
			fmt.Printf("Testing SSH to %s... ", host)
			out, err := exec.Command("ssh",
				"-o", "StrictHostKeyChecking=accept-new",
				"-o", "ConnectTimeout=10",
				host, "echo ok").CombinedOutput()
			if err != nil {
				fmt.Println(errorStyle.Render("FAILED"))
				return fmt.Errorf("SSH failed: %s", strings.TrimSpace(string(out)))
			}
			fmt.Println(successStyle.Render("OK"))

			cfg, err := loadConfig()
			if err != nil {
				return err
			}

			ac := agentci.AgentConfig{
				Host:        host,
				QueueDir:    queueDir,
				ForgejoUser: forgejoUser,
				Active:      true,
			}
			if err := agentci.SaveAgent(cfg, name, ac); err != nil {
				return err
			}

			fmt.Printf("Agent %s added (%s)\n", successStyle.Render(name), host)
			return nil
		},
	}
	cmd.Flags().String("forgejo-user", "", "Forgejo username (defaults to agent name)")
	cmd.Flags().String("queue-dir", "", "Remote queue directory (default: /home/claude/ai-work/queue)")
	return cmd
}

func agentListCmd() *cli.Command {
	return &cli.Command{
		Use:   "list",
		Short: "List configured agents",
		RunE: func(cmd *cli.Command, args []string) error {
			cfg, err := loadConfig()
			if err != nil {
				return err
			}

			agents, err := agentci.ListAgents(cfg)
			if err != nil {
				return err
			}

			if len(agents) == 0 {
				fmt.Println(dimStyle.Render("No agents configured. Use 'core ai agent add' to add one."))
				return nil
			}

			table := cli.NewTable("NAME", "HOST", "FORGEJO USER", "ACTIVE", "QUEUE")
			for name, ac := range agents {
				active := dimStyle.Render("no")
				if ac.Active {
					active = successStyle.Render("yes")
				}

				// Quick SSH check for queue depth.
				queue := dimStyle.Render("-")
				out, err := exec.Command("ssh",
					"-o", "StrictHostKeyChecking=accept-new",
					"-o", "ConnectTimeout=5",
					ac.Host,
					fmt.Sprintf("ls %s/ticket-*.json 2>/dev/null | wc -l", ac.QueueDir),
				).Output()
				if err == nil {
					n := strings.TrimSpace(string(out))
					if n != "0" {
						queue = n
					} else {
						queue = "0"
					}
				}

				table.AddRow(name, ac.Host, ac.ForgejoUser, active, queue)
			}
			table.Render()
			return nil
		},
	}
}

func agentStatusCmd() *cli.Command {
	return &cli.Command{
		Use:   "status <name>",
		Short: "Check agent status via SSH",
		Args:  cli.ExactArgs(1),
		RunE: func(cmd *cli.Command, args []string) error {
			name := args[0]
			cfg, err := loadConfig()
			if err != nil {
				return err
			}

			agents, err := agentci.ListAgents(cfg)
			if err != nil {
				return err
			}
			ac, ok := agents[name]
			if !ok {
				return fmt.Errorf("agent %q not found", name)
			}

			script := `
				echo "=== Queue ==="
				ls ~/ai-work/queue/ticket-*.json 2>/dev/null | wc -l
				echo "=== Active ==="
				ls ~/ai-work/active/ticket-*.json 2>/dev/null || echo "none"
				echo "=== Done ==="
				ls ~/ai-work/done/ticket-*.json 2>/dev/null | wc -l
				echo "=== Lock ==="
				if [ -f ~/ai-work/.runner.lock ]; then
					PID=$(cat ~/ai-work/.runner.lock)
					if kill -0 "$PID" 2>/dev/null; then
						echo "RUNNING (PID $PID)"
					else
						echo "STALE (PID $PID)"
					fi
				else
					echo "IDLE"
				fi
			`

			sshCmd := exec.Command("ssh",
				"-o", "StrictHostKeyChecking=accept-new",
				"-o", "ConnectTimeout=10",
				ac.Host, script)
			sshCmd.Stdout = os.Stdout
			sshCmd.Stderr = os.Stderr
			return sshCmd.Run()
		},
	}
}

func agentLogsCmd() *cli.Command {
	cmd := &cli.Command{
		Use:   "logs <name>",
		Short: "Stream agent runner logs",
		Args:  cli.ExactArgs(1),
		RunE: func(cmd *cli.Command, args []string) error {
			name := args[0]
			follow, _ := cmd.Flags().GetBool("follow")
			lines, _ := cmd.Flags().GetInt("lines")

			cfg, err := loadConfig()
			if err != nil {
				return err
			}

			agents, err := agentci.ListAgents(cfg)
			if err != nil {
				return err
			}
			ac, ok := agents[name]
			if !ok {
				return fmt.Errorf("agent %q not found", name)
			}

			tailArgs := []string{
				"-o", "StrictHostKeyChecking=accept-new",
				"-o", "ConnectTimeout=10",
				ac.Host,
			}
			if follow {
				tailArgs = append(tailArgs, fmt.Sprintf("tail -f -n %d ~/ai-work/logs/runner.log", lines))
			} else {
				tailArgs = append(tailArgs, fmt.Sprintf("tail -n %d ~/ai-work/logs/runner.log", lines))
			}

			sshCmd := exec.Command("ssh", tailArgs...)
			sshCmd.Stdout = os.Stdout
			sshCmd.Stderr = os.Stderr
			sshCmd.Stdin = os.Stdin
			return sshCmd.Run()
		},
	}
	cmd.Flags().BoolP("follow", "f", false, "Follow log output")
	cmd.Flags().IntP("lines", "n", 50, "Number of lines to show")
	return cmd
}

func agentSetupCmd() *cli.Command {
	return &cli.Command{
		Use:   "setup <name>",
		Short: "Bootstrap agent machine (create dirs, copy runner, install cron)",
		Args:  cli.ExactArgs(1),
		RunE: func(cmd *cli.Command, args []string) error {
			name := args[0]
			cfg, err := loadConfig()
			if err != nil {
				return err
			}

			agents, err := agentci.ListAgents(cfg)
			if err != nil {
				return err
			}
			ac, ok := agents[name]
			if !ok {
				return fmt.Errorf("agent %q not found — use 'core ai agent add' first", name)
			}

			// Find the setup script relative to the binary or in known locations.
			scriptPath := findSetupScript()
			if scriptPath == "" {
				return fmt.Errorf("agent-setup.sh not found — expected in scripts/ directory")
			}

			fmt.Printf("Setting up %s on %s...\n", name, ac.Host)
			setupCmd := exec.Command("bash", scriptPath, ac.Host)
			setupCmd.Stdout = os.Stdout
			setupCmd.Stderr = os.Stderr
			if err := setupCmd.Run(); err != nil {
				return fmt.Errorf("setup failed: %w", err)
			}

			fmt.Println(successStyle.Render("Setup complete!"))
			return nil
		},
	}
}

func agentRemoveCmd() *cli.Command {
	return &cli.Command{
		Use:   "remove <name>",
		Short: "Remove an agent from config",
		Args:  cli.ExactArgs(1),
		RunE: func(cmd *cli.Command, args []string) error {
			name := args[0]
			cfg, err := loadConfig()
			if err != nil {
				return err
			}

			if err := agentci.RemoveAgent(cfg, name); err != nil {
				return err
			}

			fmt.Printf("Agent %s removed.\n", name)
			return nil
		},
	}
}

// findSetupScript looks for agent-setup.sh in common locations.
func findSetupScript() string {
	// Relative to executable.
	exe, _ := os.Executable()
	if exe != "" {
		dir := filepath.Dir(exe)
		candidates := []string{
			filepath.Join(dir, "scripts", "agent-setup.sh"),
			filepath.Join(dir, "..", "scripts", "agent-setup.sh"),
		}
		for _, c := range candidates {
			if _, err := os.Stat(c); err == nil {
				return c
			}
		}
	}

	// Working directory.
	cwd, _ := os.Getwd()
	if cwd != "" {
		p := filepath.Join(cwd, "scripts", "agent-setup.sh")
		if _, err := os.Stat(p); err == nil {
			return p
		}
	}

	return ""
}

