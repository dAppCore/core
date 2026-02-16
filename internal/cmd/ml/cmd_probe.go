package ml

import (
	"context"
	"encoding/json"
	"fmt"
	"os"

	"forge.lthn.ai/core/cli/pkg/cli"
	"forge.lthn.ai/core/cli/pkg/ml"
)

var (
	probeOutput string
)

var probeCmd = &cli.Command{
	Use:   "probe",
	Short: "Run capability and content probes against a model",
	Long:  "Runs 23 capability probes and 6 content probes against an OpenAI-compatible API.",
	RunE:  runProbe,
}

func init() {
	probeCmd.Flags().StringVar(&probeOutput, "output", "", "Output JSON file for probe results")
}

func runProbe(cmd *cli.Command, args []string) error {
	if apiURL == "" {
		return fmt.Errorf("--api-url is required")
	}

	model := modelName
	if model == "" {
		model = "default"
	}

	ctx := context.Background()
	backend := ml.NewHTTPBackend(apiURL, model)

	fmt.Printf("Running %d capability probes against %s...\n", len(ml.CapabilityProbes), apiURL)
	results := ml.RunCapabilityProbes(ctx, backend)

	fmt.Printf("\nResults: %.1f%% (%d/%d)\n", results.Accuracy, results.Correct, results.Total)

	for cat, data := range results.ByCategory {
		catAcc := 0.0
		if data.Total > 0 {
			catAcc = float64(data.Correct) / float64(data.Total) * 100
		}
		fmt.Printf("  %-20s %d/%d (%.0f%%)\n", cat, data.Correct, data.Total, catAcc)
	}

	if probeOutput != "" {
		data, err := json.MarshalIndent(results, "", "  ")
		if err != nil {
			return fmt.Errorf("marshal results: %w", err)
		}
		if err := os.WriteFile(probeOutput, data, 0644); err != nil {
			return fmt.Errorf("write output: %w", err)
		}
		fmt.Printf("\nResults written to %s\n", probeOutput)
	}

	return nil
}
