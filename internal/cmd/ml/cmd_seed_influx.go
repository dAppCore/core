package ml

import (
	"fmt"
	"os"

	"forge.lthn.ai/core/cli/pkg/cli"
	"forge.lthn.ai/core/cli/pkg/ml"
)

var seedInfluxCmd = &cli.Command{
	Use:   "seed-influx",
	Short: "Seed InfluxDB golden_gen from DuckDB golden_set",
	Long:  "One-time migration: batch-loads DuckDB golden_set records into InfluxDB golden_gen measurement.",
	RunE:  runSeedInflux,
}

var (
	seedInfluxForce     bool
	seedInfluxBatchSize int
)

func init() {
	seedInfluxCmd.Flags().BoolVar(&seedInfluxForce, "force", false, "Re-seed even if InfluxDB already has data")
	seedInfluxCmd.Flags().IntVar(&seedInfluxBatchSize, "batch-size", 500, "Lines per InfluxDB write batch")
}

func runSeedInflux(cmd *cli.Command, args []string) error {
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

	return ml.SeedInflux(db, influx, ml.SeedInfluxConfig{
		Force:     seedInfluxForce,
		BatchSize: seedInfluxBatchSize,
	}, os.Stdout)
}
