package ml

import (
	"context"
	"fmt"
	"os"

	"forge.lthn.ai/core/cli/pkg/cli"
	"forge.lthn.ai/core/cli/pkg/ml"
)

var (
	expandWorker  string
	expandOutput  string
	expandLimit   int
	expandDryRun  bool
)

var expandCmd = &cli.Command{
	Use:   "expand",
	Short: "Generate expansion responses from pending prompts",
	Long:  "Reads pending expansion prompts from DuckDB and generates responses via an OpenAI-compatible API.",
	RunE:  runExpand,
}

func init() {
	expandCmd.Flags().StringVar(&expandWorker, "worker", "", "Worker hostname (defaults to os.Hostname())")
	expandCmd.Flags().StringVar(&expandOutput, "output", ".", "Output directory for JSONL files")
	expandCmd.Flags().IntVar(&expandLimit, "limit", 0, "Max prompts to process (0 = all)")
	expandCmd.Flags().BoolVar(&expandDryRun, "dry-run", false, "Print plan and exit without generating")
}

func runExpand(cmd *cli.Command, args []string) error {
	if modelName == "" {
		return fmt.Errorf("--model is required")
	}

	path := dbPath
	if path == "" {
		path = os.Getenv("LEM_DB")
	}
	if path == "" {
		return fmt.Errorf("--db or LEM_DB env is required")
	}

	if expandWorker == "" {
		h, _ := os.Hostname()
		expandWorker = h
	}

	db, err := ml.OpenDBReadWrite(path)
	if err != nil {
		return fmt.Errorf("open db: %w", err)
	}
	defer db.Close()

	rows, err := db.QueryExpansionPrompts("pending", expandLimit)
	if err != nil {
		return fmt.Errorf("query expansion_prompts: %w", err)
	}
	fmt.Printf("Loaded %d pending prompts from %s\n", len(rows), path)

	var prompts []ml.Response
	for _, r := range rows {
		prompt := r.Prompt
		if prompt == "" && r.PromptEn != "" {
			prompt = r.PromptEn
		}
		prompts = append(prompts, ml.Response{
			ID:     r.SeedID,
			Domain: r.Domain,
			Prompt: prompt,
		})
	}

	ctx := context.Background()
	backend := ml.NewHTTPBackend(apiURL, modelName)
	influx := ml.NewInfluxClient(influxURL, influxDB)

	return ml.ExpandPrompts(ctx, backend, influx, prompts, modelName, expandWorker, expandOutput, expandDryRun, expandLimit)
}
