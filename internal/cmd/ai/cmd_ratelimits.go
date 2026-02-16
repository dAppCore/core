package ai

import (
	"fmt"
	"os"
	"strconv"
	"text/tabwriter"
	"time"

	"forge.lthn.ai/core/cli/pkg/cli"
	"forge.lthn.ai/core/cli/pkg/config"
	"forge.lthn.ai/core/cli/pkg/ratelimit"
)

// AddRateLimitCommands registers the 'ratelimits' subcommand group under 'ai'.
func AddRateLimitCommands(parent *cli.Command) {
	rlCmd := &cli.Command{
		Use:   "ratelimits",
		Short: "Manage Gemini API rate limits",
	}

	rlCmd.AddCommand(rlShowCmd())
	rlCmd.AddCommand(rlResetCmd())
	rlCmd.AddCommand(rlCountCmd())
	rlCmd.AddCommand(rlConfigCmd())
	rlCmd.AddCommand(rlCheckCmd())

	parent.AddCommand(rlCmd)
}

func rlShowCmd() *cli.Command {
	return &cli.Command{
		Use:   "show",
		Short: "Show current rate limit usage",
		RunE: func(cmd *cli.Command, args []string) error {
			rl, err := ratelimit.New()
			if err != nil {
				return err
			}
			if err := rl.Load(); err != nil {
				return err
			}

			stats := rl.AllStats()

			w := tabwriter.NewWriter(os.Stdout, 0, 0, 3, ' ', 0)
			fmt.Fprintln(w, "MODEL\tRPM\tTPM\tRPD\tSTATUS")

			for model, s := range stats {
				rpmStr := fmt.Sprintf("%d/%s", s.RPM, formatLimit(s.MaxRPM))
				tpmStr := fmt.Sprintf("%d/%s", s.TPM, formatLimit(s.MaxTPM))
				rpdStr := fmt.Sprintf("%d/%s", s.RPD, formatLimit(s.MaxRPD))

				status := "OK"
				if (s.MaxRPM > 0 && s.RPM >= s.MaxRPM) ||
					(s.MaxTPM > 0 && s.TPM >= s.MaxTPM) ||
					(s.MaxRPD > 0 && s.RPD >= s.MaxRPD) {
					status = "LIMITED"
				}

				fmt.Fprintf(w, "%s\t%s\t%s\t%s\t%s\n", model, rpmStr, tpmStr, rpdStr, status)
			}
			w.Flush()
			return nil
		},
	}
}

func rlResetCmd() *cli.Command {
	return &cli.Command{
		Use:   "reset [model]",
		Short: "Reset usage counters for a model (or all)",
		RunE: func(cmd *cli.Command, args []string) error {
			rl, err := ratelimit.New()
			if err != nil {
				return err
			}
			if err := rl.Load(); err != nil {
				return err
			}

			model := ""
			if len(args) > 0 {
				model = args[0]
			}

			rl.Reset(model)
			if err := rl.Persist(); err != nil {
				return err
			}

			if model == "" {
				fmt.Println("Reset stats for all models.")
			} else {
				fmt.Printf("Reset stats for model %q.\n", model)
			}
			return nil
		},
	}
}

func rlCountCmd() *cli.Command {
	return &cli.Command{
		Use:   "count <model> <text>",
		Short: "Count tokens for text using Gemini API",
		Args:  cli.ExactArgs(2),
		RunE: func(cmd *cli.Command, args []string) error {
			model := args[0]
			text := args[1]

			cfg, err := config.New()
			if err != nil {
				return err
			}

			var apiKey string
			if err := cfg.Get("agentci.gemini_api_key", &apiKey); err != nil || apiKey == "" {
				apiKey = os.Getenv("GEMINI_API_KEY")
			}
			if apiKey == "" {
				return fmt.Errorf("GEMINI_API_KEY not found in config or env")
			}

			count, err := ratelimit.CountTokens(apiKey, model, text)
			if err != nil {
				return err
			}

			fmt.Printf("Model: %s\nTokens: %d\n", model, count)
			return nil
		},
	}
}

func rlConfigCmd() *cli.Command {
	return &cli.Command{
		Use:   "config",
		Short: "Show configured quotas",
		RunE: func(cmd *cli.Command, args []string) error {
			rl, err := ratelimit.New()
			if err != nil {
				return err
			}

			w := tabwriter.NewWriter(os.Stdout, 0, 0, 3, ' ', 0)
			fmt.Fprintln(w, "MODEL\tMAX RPM\tMAX TPM\tMAX RPD")

			for model, q := range rl.Quotas {
				fmt.Fprintf(w, "%s\t%s\t%s\t%s\n",
					model,
					formatLimit(q.MaxRPM),
					formatLimit(q.MaxTPM),
					formatLimit(q.MaxRPD))
			}
			w.Flush()
			return nil
		},
	}
}

func rlCheckCmd() *cli.Command {
	return &cli.Command{
		Use:   "check <model> <estimated-tokens>",
		Short: "Check rate limit capacity for a model",
		Args:  cli.ExactArgs(2),
		RunE: func(cmd *cli.Command, args []string) error {
			model := args[0]
			tokens, err := strconv.Atoi(args[1])
			if err != nil {
				return fmt.Errorf("invalid token count: %w", err)
			}

			rl, err := ratelimit.New()
			if err != nil {
				return err
			}
			if err := rl.Load(); err != nil {
				fmt.Printf("Warning: could not load existing state: %v\n", err)
			}

			stats := rl.Stats(model)
			canSend := rl.CanSend(model, tokens)

			status := "RATE LIMITED"
			if canSend {
				status = "OK"
			}

			fmt.Printf("Model:        %s\n", model)
			fmt.Printf("Request Cost: %d tokens\n", tokens)
			fmt.Printf("Status:       %s\n", status)
			fmt.Printf("\nCurrent Usage (1m window):\n")
			fmt.Printf("  RPM: %d / %s\n", stats.RPM, formatLimit(stats.MaxRPM))
			fmt.Printf("  TPM: %d / %s\n", stats.TPM, formatLimit(stats.MaxTPM))
			fmt.Printf("  RPD: %d / %s (reset: %s)\n", stats.RPD, formatLimit(stats.MaxRPD), stats.DayStart.Format(time.RFC3339))

			return nil
		},
	}
}

func formatLimit(limit int) string {
	if limit == 0 {
		return "∞"
	}
	if limit >= 1000000 {
		return fmt.Sprintf("%dM", limit/1000000)
	}
	if limit >= 1000 {
		return fmt.Sprintf("%dK", limit/1000)
	}
	return fmt.Sprintf("%d", limit)
}
