package ml

import (
	"fmt"
	"os"

	"forge.lthn.ai/core/cli/pkg/cli"
	"forge.lthn.ai/core/cli/pkg/ml"
)

var metricsCmd = &cli.Command{
	Use:   "metrics",
	Short: "Push golden set stats to InfluxDB",
	Long:  "Queries golden_set stats from DuckDB and pushes summary, per-domain, and per-voice metrics to InfluxDB.",
	RunE:  runMetrics,
}

func runMetrics(cmd *cli.Command, args []string) error {
	path := dbPath
	if path == "" {
		path = os.Getenv("LEM_DB")
	}
	if path == "" {
		return fmt.Errorf("--db or LEM_DB required")
	}

	db, err := ml.OpenDB(path)
	if err != nil {
		return fmt.Errorf("open db: %w", err)
	}
	defer db.Close()

	influx := ml.NewInfluxClient(influxURL, influxDB)

	return ml.PushMetrics(db, influx, os.Stdout)
}
