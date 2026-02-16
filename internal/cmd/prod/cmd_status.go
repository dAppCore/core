package prod

import (
	"context"
	"fmt"
	"os"
	"strings"
	"sync"
	"time"

	"forge.lthn.ai/core/cli/pkg/ansible"
	"forge.lthn.ai/core/cli/pkg/cli"
	"forge.lthn.ai/core/cli/pkg/infra"
	"github.com/spf13/cobra"
)

var statusCmd = &cobra.Command{
	Use:   "status",
	Short: "Show production infrastructure health",
	Long: `Check connectivity, services, and cluster health across all production hosts.

Tests:
  - SSH connectivity to all hosts
  - Docker daemon status
  - Coolify controller (noc)
  - Galera cluster state (de, de2)
  - Redis Sentinel status (de, de2)
  - Load balancer health (if HCLOUD_TOKEN set)`,
	RunE: runStatus,
}

type hostStatus struct {
	Name      string
	Host      *infra.Host
	Connected bool
	ConnTime  time.Duration
	OS        string
	Docker    string
	Services  map[string]string
	Error     error
}

func runStatus(cmd *cobra.Command, args []string) error {
	cfg, cfgPath, err := loadConfig()
	if err != nil {
		return err
	}

	cli.Print("%s Infrastructure status from %s\n\n",
		cli.BoldStyle.Render("▶"),
		cli.DimStyle.Render(cfgPath))

	ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
	defer cancel()

	// Check all hosts in parallel
	var (
		wg       sync.WaitGroup
		mu       sync.Mutex
		statuses []hostStatus
	)

	for name, host := range cfg.Hosts {
		wg.Add(1)
		go func(name string, host *infra.Host) {
			defer wg.Done()
			s := checkHost(ctx, name, host)
			mu.Lock()
			statuses = append(statuses, s)
			mu.Unlock()
		}(name, host)
	}

	wg.Wait()

	// Print results in consistent order
	order := []string{"noc", "de", "de2", "build"}
	for _, name := range order {
		for _, s := range statuses {
			if s.Name == name {
				printHostStatus(s)
				break
			}
		}
	}

	// Check LB if token available
	if token := os.Getenv("HCLOUD_TOKEN"); token != "" {
		fmt.Println()
		checkLoadBalancer(ctx, token)
	} else {
		fmt.Println()
		cli.Print("%s Load balancer: %s\n",
			cli.DimStyle.Render("  ○"),
			cli.DimStyle.Render("HCLOUD_TOKEN not set (skipped)"))
	}

	return nil
}

func checkHost(ctx context.Context, name string, host *infra.Host) hostStatus {
	s := hostStatus{
		Name:     name,
		Host:     host,
		Services: make(map[string]string),
	}

	sshCfg := ansible.SSHConfig{
		Host:    host.IP,
		Port:    host.SSH.Port,
		User:    host.SSH.User,
		KeyFile: host.SSH.Key,
		Timeout: 15 * time.Second,
	}

	client, err := ansible.NewSSHClient(sshCfg)
	if err != nil {
		s.Error = fmt.Errorf("create SSH client: %w", err)
		return s
	}
	defer func() { _ = client.Close() }()

	start := time.Now()
	if err := client.Connect(ctx); err != nil {
		s.Error = fmt.Errorf("SSH connect: %w", err)
		return s
	}
	s.Connected = true
	s.ConnTime = time.Since(start)

	// OS info
	stdout, _, _, _ := client.Run(ctx, "cat /etc/os-release 2>/dev/null | grep PRETTY_NAME | cut -d'\"' -f2")
	s.OS = strings.TrimSpace(stdout)

	// Docker
	stdout, _, _, err = client.Run(ctx, "docker --version 2>/dev/null | head -1")
	if err == nil && stdout != "" {
		s.Docker = strings.TrimSpace(stdout)
	}

	// Check each expected service
	for _, svc := range host.Services {
		status := checkService(ctx, client, svc)
		s.Services[svc] = status
	}

	return s
}

func checkService(ctx context.Context, client *ansible.SSHClient, service string) string {
	switch service {
	case "coolify":
		stdout, _, _, _ := client.Run(ctx, "docker ps --format '{{.Names}}' 2>/dev/null | grep -c coolify")
		if strings.TrimSpace(stdout) != "0" && strings.TrimSpace(stdout) != "" {
			return "running"
		}
		return "not running"

	case "traefik":
		stdout, _, _, _ := client.Run(ctx, "docker ps --format '{{.Names}}' 2>/dev/null | grep -c traefik")
		if strings.TrimSpace(stdout) != "0" && strings.TrimSpace(stdout) != "" {
			return "running"
		}
		return "not running"

	case "galera":
		// Check Galera cluster state
		stdout, _, _, _ := client.Run(ctx,
			"docker exec $(docker ps -q --filter name=mariadb 2>/dev/null || echo none) "+
				"mariadb -u root -e \"SHOW STATUS LIKE 'wsrep_cluster_size'\" --skip-column-names 2>/dev/null | awk '{print $2}'")
		size := strings.TrimSpace(stdout)
		if size != "" && size != "0" {
			return fmt.Sprintf("cluster_size=%s", size)
		}
		// Try non-Docker
		stdout, _, _, _ = client.Run(ctx,
			"mariadb -u root -e \"SHOW STATUS LIKE 'wsrep_cluster_size'\" --skip-column-names 2>/dev/null | awk '{print $2}'")
		size = strings.TrimSpace(stdout)
		if size != "" && size != "0" {
			return fmt.Sprintf("cluster_size=%s", size)
		}
		return "not running"

	case "redis":
		stdout, _, _, _ := client.Run(ctx,
			"docker exec $(docker ps -q --filter name=redis 2>/dev/null || echo none) "+
				"redis-cli ping 2>/dev/null")
		if strings.TrimSpace(stdout) == "PONG" {
			return "running"
		}
		stdout, _, _, _ = client.Run(ctx, "redis-cli ping 2>/dev/null")
		if strings.TrimSpace(stdout) == "PONG" {
			return "running"
		}
		return "not running"

	case "forgejo-runner":
		stdout, _, _, _ := client.Run(ctx, "systemctl is-active forgejo-runner 2>/dev/null || docker ps --format '{{.Names}}' 2>/dev/null | grep -c runner")
		val := strings.TrimSpace(stdout)
		if val == "active" || (val != "0" && val != "") {
			return "running"
		}
		return "not running"

	default:
		// Generic docker container check
		stdout, _, _, _ := client.Run(ctx,
			fmt.Sprintf("docker ps --format '{{.Names}}' 2>/dev/null | grep -c %s", service))
		if strings.TrimSpace(stdout) != "0" && strings.TrimSpace(stdout) != "" {
			return "running"
		}
		return "not running"
	}
}

func printHostStatus(s hostStatus) {
	// Host header
	roleStyle := cli.DimStyle
	switch s.Host.Role {
	case "app":
		roleStyle = cli.SuccessStyle
	case "bastion":
		roleStyle = cli.WarningStyle
	case "builder":
		roleStyle = cli.InfoStyle
	}

	cli.Print("  %s %s  %s  %s\n",
		cli.BoldStyle.Render(s.Name),
		cli.DimStyle.Render(s.Host.IP),
		roleStyle.Render(s.Host.Role),
		cli.DimStyle.Render(s.Host.FQDN))

	if s.Error != nil {
		cli.Print("    %s %s\n", cli.ErrorStyle.Render("✗"), s.Error)
		return
	}

	if !s.Connected {
		cli.Print("    %s SSH unreachable\n", cli.ErrorStyle.Render("✗"))
		return
	}

	// Connection info
	cli.Print("    %s SSH %s",
		cli.SuccessStyle.Render("✓"),
		cli.DimStyle.Render(s.ConnTime.Round(time.Millisecond).String()))
	if s.OS != "" {
		cli.Print("  %s", cli.DimStyle.Render(s.OS))
	}
	fmt.Println()

	if s.Docker != "" {
		cli.Print("    %s %s\n", cli.SuccessStyle.Render("✓"), cli.DimStyle.Render(s.Docker))
	}

	// Services
	for _, svc := range s.Host.Services {
		status, ok := s.Services[svc]
		if !ok {
			continue
		}

		icon := cli.SuccessStyle.Render("●")
		style := cli.SuccessStyle
		if status == "not running" {
			icon = cli.ErrorStyle.Render("○")
			style = cli.ErrorStyle
		}

		cli.Print("    %s %s %s\n", icon, svc, style.Render(status))
	}

	fmt.Println()
}

func checkLoadBalancer(ctx context.Context, token string) {
	hc := infra.NewHCloudClient(token)
	lbs, err := hc.ListLoadBalancers(ctx)
	if err != nil {
		cli.Print("  %s Load balancer: %s\n", cli.ErrorStyle.Render("✗"), err)
		return
	}

	if len(lbs) == 0 {
		cli.Print("  %s No load balancers found\n", cli.DimStyle.Render("○"))
		return
	}

	for _, lb := range lbs {
		cli.Print("  %s LB: %s  IP: %s  Targets: %d\n",
			cli.SuccessStyle.Render("●"),
			cli.BoldStyle.Render(lb.Name),
			lb.PublicNet.IPv4.IP,
			len(lb.Targets))

		for _, t := range lb.Targets {
			for _, hs := range t.HealthStatus {
				icon := cli.SuccessStyle.Render("●")
				if hs.Status != "healthy" {
					icon = cli.ErrorStyle.Render("○")
				}
				ip := ""
				if t.IP != nil {
					ip = t.IP.IP
				}
				cli.Print("    %s :%d %s %s\n", icon, hs.ListenPort, hs.Status, cli.DimStyle.Render(ip))
			}
		}
	}
}

func loadConfig() (*infra.Config, string, error) {
	if infraFile != "" {
		cfg, err := infra.Load(infraFile)
		return cfg, infraFile, err
	}

	cwd, err := os.Getwd()
	if err != nil {
		return nil, "", err
	}

	return infra.Discover(cwd)
}
