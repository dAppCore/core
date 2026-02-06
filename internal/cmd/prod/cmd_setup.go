package prod

import (
	"context"
	"fmt"
	"os"
	"time"

	"github.com/host-uk/core/pkg/cli"
	"github.com/host-uk/core/pkg/infra"
	"github.com/spf13/cobra"
)

var setupCmd = &cobra.Command{
	Use:   "setup",
	Short: "Phase 1: discover topology, create LB, configure DNS",
	Long: `Run the Phase 1 foundation setup:

  1. Discover Hetzner topology (Cloud + Robot servers)
  2. Create Hetzner managed load balancer
  3. Configure DNS records via CloudNS
  4. Verify connectivity to all hosts

Required environment variables:
  HCLOUD_TOKEN           Hetzner Cloud API token
  HETZNER_ROBOT_USER     Hetzner Robot username
  HETZNER_ROBOT_PASS     Hetzner Robot password
  CLOUDNS_AUTH_ID        CloudNS auth ID
  CLOUDNS_AUTH_PASSWORD  CloudNS auth password`,
	RunE: runSetup,
}

var (
	setupDryRun bool
	setupStep   string
)

func init() {
	setupCmd.Flags().BoolVar(&setupDryRun, "dry-run", false, "Show what would be done without making changes")
	setupCmd.Flags().StringVar(&setupStep, "step", "", "Run a specific step only (discover, lb, dns)")
}

func runSetup(cmd *cobra.Command, args []string) error {
	cfg, cfgPath, err := loadConfig()
	if err != nil {
		return err
	}

	cli.Print("%s Production setup from %s\n\n",
		cli.BoldStyle.Render("▶"),
		cli.DimStyle.Render(cfgPath))

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Minute)
	defer cancel()

	steps := []struct {
		name string
		fn   func(context.Context, *infra.Config) error
	}{
		{"discover", stepDiscover},
		{"lb", stepLoadBalancer},
		{"dns", stepDNS},
	}

	for _, step := range steps {
		if setupStep != "" && setupStep != step.name {
			continue
		}

		cli.Print("\n%s Step: %s\n", cli.BoldStyle.Render("━━"), cli.TitleStyle.Render(step.name))

		if err := step.fn(ctx, cfg); err != nil {
			cli.Print("  %s %s: %s\n", cli.ErrorStyle.Render("✗"), step.name, err)
			return fmt.Errorf("step %s failed: %w", step.name, err)
		}

		cli.Print("  %s %s complete\n", cli.SuccessStyle.Render("✓"), step.name)
	}

	cli.Print("\n%s Setup complete\n", cli.SuccessStyle.Render("✓"))
	return nil
}

func stepDiscover(ctx context.Context, cfg *infra.Config) error {
	// Discover HCloud servers
	hcloudToken := os.Getenv("HCLOUD_TOKEN")
	if hcloudToken != "" {
		cli.Print("  Discovering Hetzner Cloud servers...\n")

		hc := infra.NewHCloudClient(hcloudToken)
		servers, err := hc.ListServers(ctx)
		if err != nil {
			return fmt.Errorf("list HCloud servers: %w", err)
		}

		for _, s := range servers {
			cli.Print("    %s %s  %s  %s  %s\n",
				cli.SuccessStyle.Render("●"),
				cli.BoldStyle.Render(s.Name),
				s.PublicNet.IPv4.IP,
				s.ServerType.Name,
				cli.DimStyle.Render(s.Datacenter.Name))
		}
	} else {
		cli.Print("  %s HCLOUD_TOKEN not set — skipping Cloud discovery\n",
			cli.WarningStyle.Render("⚠"))
	}

	// Discover Robot servers
	robotUser := os.Getenv("HETZNER_ROBOT_USER")
	robotPass := os.Getenv("HETZNER_ROBOT_PASS")
	if robotUser != "" && robotPass != "" {
		cli.Print("  Discovering Hetzner Robot servers...\n")

		hr := infra.NewHRobotClient(robotUser, robotPass)
		servers, err := hr.ListServers(ctx)
		if err != nil {
			return fmt.Errorf("list Robot servers: %w", err)
		}

		for _, s := range servers {
			status := cli.SuccessStyle.Render("●")
			if s.Status != "ready" {
				status = cli.WarningStyle.Render("○")
			}
			cli.Print("    %s %s  %s  %s  %s\n",
				status,
				cli.BoldStyle.Render(s.ServerName),
				s.ServerIP,
				s.Product,
				cli.DimStyle.Render(s.Datacenter))
		}
	} else {
		cli.Print("  %s HETZNER_ROBOT_USER/PASS not set — skipping Robot discovery\n",
			cli.WarningStyle.Render("⚠"))
	}

	return nil
}

func stepLoadBalancer(ctx context.Context, cfg *infra.Config) error {
	hcloudToken := os.Getenv("HCLOUD_TOKEN")
	if hcloudToken == "" {
		return fmt.Errorf("HCLOUD_TOKEN required for load balancer management")
	}

	hc := infra.NewHCloudClient(hcloudToken)

	// Check if LB already exists
	lbs, err := hc.ListLoadBalancers(ctx)
	if err != nil {
		return fmt.Errorf("list load balancers: %w", err)
	}

	for _, lb := range lbs {
		if lb.Name == cfg.LoadBalancer.Name {
			cli.Print("  Load balancer '%s' already exists (ID: %d, IP: %s)\n",
				lb.Name, lb.ID, lb.PublicNet.IPv4.IP)
			return nil
		}
	}

	if setupDryRun {
		cli.Print("  [dry-run] Would create load balancer '%s' (%s) in %s\n",
			cfg.LoadBalancer.Name, cfg.LoadBalancer.Type, cfg.LoadBalancer.Location)
		for _, b := range cfg.LoadBalancer.Backends {
			if host, ok := cfg.Hosts[b.Host]; ok {
				cli.Print("  [dry-run] Backend: %s (%s:%d)\n", b.Host, host.IP, b.Port)
			}
		}
		return nil
	}

	// Build targets from config
	targets := make([]infra.HCloudLBCreateTarget, 0, len(cfg.LoadBalancer.Backends))
	for _, b := range cfg.LoadBalancer.Backends {
		host, ok := cfg.Hosts[b.Host]
		if !ok {
			return fmt.Errorf("backend host '%s' not found in config", b.Host)
		}
		targets = append(targets, infra.HCloudLBCreateTarget{
			Type: "ip",
			IP:   &infra.HCloudLBTargetIP{IP: host.IP},
		})
	}

	// Build services
	services := make([]infra.HCloudLBService, 0, len(cfg.LoadBalancer.Listeners))
	for _, l := range cfg.LoadBalancer.Listeners {
		svc := infra.HCloudLBService{
			Protocol:        l.Protocol,
			ListenPort:      l.Frontend,
			DestinationPort: l.Backend,
			Proxyprotocol:   l.ProxyProtocol,
			HealthCheck: &infra.HCloudLBHealthCheck{
				Protocol: cfg.LoadBalancer.Health.Protocol,
				Port:     l.Backend,
				Interval: cfg.LoadBalancer.Health.Interval,
				Timeout:  10,
				Retries:  3,
				HTTP: &infra.HCloudLBHCHTTP{
					Path:       cfg.LoadBalancer.Health.Path,
					StatusCode: "2??",
				},
			},
		}
		services = append(services, svc)
	}

	req := infra.HCloudLBCreateRequest{
		Name:             cfg.LoadBalancer.Name,
		LoadBalancerType: cfg.LoadBalancer.Type,
		Location:         cfg.LoadBalancer.Location,
		Algorithm:        infra.HCloudLBAlgorithm{Type: cfg.LoadBalancer.Algorithm},
		Services:         services,
		Targets:          targets,
		Labels: map[string]string{
			"project": "host-uk",
			"managed": "core-cli",
		},
	}

	cli.Print("  Creating load balancer '%s'...\n", cfg.LoadBalancer.Name)

	lb, err := hc.CreateLoadBalancer(ctx, req)
	if err != nil {
		return fmt.Errorf("create load balancer: %w", err)
	}

	cli.Print("  Created: %s (ID: %d, IP: %s)\n",
		cli.BoldStyle.Render(lb.Name), lb.ID, lb.PublicNet.IPv4.IP)

	return nil
}

func stepDNS(ctx context.Context, cfg *infra.Config) error {
	authID := os.Getenv("CLOUDNS_AUTH_ID")
	authPass := os.Getenv("CLOUDNS_AUTH_PASSWORD")
	if authID == "" || authPass == "" {
		return fmt.Errorf("CLOUDNS_AUTH_ID and CLOUDNS_AUTH_PASSWORD required")
	}

	dns := infra.NewCloudNSClient(authID, authPass)

	for zoneName, zone := range cfg.DNS.Zones {
		cli.Print("  Zone: %s\n", cli.BoldStyle.Render(zoneName))

		for _, rec := range zone.Records {
			value := rec.Value
			// Skip templated values (need LB IP first)
			if value == "{{.lb_ip}}" {
				cli.Print("    %s %s %s %s — %s\n",
					cli.WarningStyle.Render("⚠"),
					rec.Name, rec.Type, value,
					cli.DimStyle.Render("needs LB IP (run setup --step=lb first)"))
				continue
			}

			if setupDryRun {
				cli.Print("    [dry-run] %s %s -> %s (TTL: %d)\n",
					rec.Type, rec.Name, value, rec.TTL)
				continue
			}

			changed, err := dns.EnsureRecord(ctx, zoneName, rec.Name, rec.Type, value, rec.TTL)
			if err != nil {
				cli.Print("    %s %s %s: %s\n", cli.ErrorStyle.Render("✗"), rec.Type, rec.Name, err)
				continue
			}

			if changed {
				cli.Print("    %s %s %s -> %s\n",
					cli.SuccessStyle.Render("✓"),
					rec.Type, rec.Name, value)
			} else {
				cli.Print("    %s %s %s (no change)\n",
					cli.DimStyle.Render("·"),
					rec.Type, rec.Name)
			}
		}
	}

	return nil
}
