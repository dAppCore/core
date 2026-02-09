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

var lbCmd = &cobra.Command{
	Use:   "lb",
	Short: "Manage Hetzner load balancer",
	Long: `View and manage the Hetzner Cloud managed load balancer.

Requires: HCLOUD_TOKEN`,
}

var lbStatusCmd = &cobra.Command{
	Use:   "status",
	Short: "Show load balancer status and target health",
	RunE:  runLBStatus,
}

var lbCreateCmd = &cobra.Command{
	Use:   "create",
	Short: "Create load balancer from infra.yaml",
	RunE:  runLBCreate,
}

func init() {
	lbCmd.AddCommand(lbStatusCmd)
	lbCmd.AddCommand(lbCreateCmd)
}

func getHCloudClient() (*infra.HCloudClient, error) {
	token := os.Getenv("HCLOUD_TOKEN")
	if token == "" {
		return nil, fmt.Errorf("HCLOUD_TOKEN environment variable required")
	}
	return infra.NewHCloudClient(token), nil
}

func runLBStatus(cmd *cobra.Command, args []string) error {
	hc, err := getHCloudClient()
	if err != nil {
		return err
	}

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	lbs, err := hc.ListLoadBalancers(ctx)
	if err != nil {
		return fmt.Errorf("list load balancers: %w", err)
	}

	if len(lbs) == 0 {
		cli.Print("No load balancers found\n")
		return nil
	}

	for _, lb := range lbs {
		cli.Print("%s %s\n", cli.BoldStyle.Render("▶"), cli.TitleStyle.Render(lb.Name))
		cli.Print("  ID:        %d\n", lb.ID)
		cli.Print("  IP:        %s\n", lb.PublicNet.IPv4.IP)
		cli.Print("  Algorithm: %s\n", lb.Algorithm.Type)
		cli.Print("  Location:  %s\n", lb.Location.Name)

		if len(lb.Services) > 0 {
			cli.Print("\n  Services:\n")
			for _, s := range lb.Services {
				cli.Print("    %s :%d -> :%d  proxy_protocol=%v\n",
					s.Protocol, s.ListenPort, s.DestinationPort, s.Proxyprotocol)
			}
		}

		if len(lb.Targets) > 0 {
			cli.Print("\n  Targets:\n")
			for _, t := range lb.Targets {
				ip := ""
				if t.IP != nil {
					ip = t.IP.IP
				}
				for _, hs := range t.HealthStatus {
					icon := cli.SuccessStyle.Render("●")
					if hs.Status != "healthy" {
						icon = cli.ErrorStyle.Render("○")
					}
					cli.Print("    %s %s :%d %s\n", icon, ip, hs.ListenPort, hs.Status)
				}
			}
		}
		fmt.Println()
	}

	return nil
}

func runLBCreate(cmd *cobra.Command, args []string) error {
	cfg, _, err := loadConfig()
	if err != nil {
		return err
	}

	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Minute)
	defer cancel()

	return stepLoadBalancer(ctx, cfg)
}
