package ml

import (
	"fmt"
	"os"

	"forge.lthn.ai/core/cli/pkg/cli"
	"forge.lthn.ai/core/cli/pkg/ml"
)

var statusCmd = &cli.Command{
	Use:   "status",
	Short: "Show training and generation progress",
	Long:  "Queries InfluxDB for training status, loss, and generation progress. Optionally shows DuckDB table counts.",
	RunE:  runStatus,
}

func runStatus(cmd *cli.Command, args []string) error {
	influx := ml.NewInfluxClient(influxURL, influxDB)

	if err := ml.PrintStatus(influx, os.Stdout); err != nil {
		return fmt.Errorf("status: %w", err)
	}

	path := dbPath
	if path == "" {
		path = os.Getenv("LEM_DB")
	}

	if path != "" {
		db, err := ml.OpenDB(path)
		if err != nil {
			return fmt.Errorf("open db: %w", err)
		}
		defer db.Close()

		counts, err := db.TableCounts()
		if err != nil {
			return fmt.Errorf("table counts: %w", err)
		}

		fmt.Println()
		fmt.Println("DuckDB:")
		order := []string{"golden_set", "expansion_prompts", "seeds", "training_examples",
			"prompts", "gemini_responses", "benchmark_questions", "benchmark_results", "validations"}
		for _, table := range order {
			if count, ok := counts[table]; ok {
				fmt.Fprintf(os.Stdout, "  %-22s %6d rows\n", table, count)
			}
		}
	}

	return nil
}
