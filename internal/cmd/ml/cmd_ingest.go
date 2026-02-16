package ml

import (
	"fmt"
	"os"

	"forge.lthn.ai/core/cli/pkg/cli"
	"forge.lthn.ai/core/cli/pkg/ml"
)

var ingestCmd = &cli.Command{
	Use:   "ingest",
	Short: "Ingest benchmark scores and training logs into InfluxDB",
	Long:  "Reads content score, capability score, and training log files and writes measurements to InfluxDB for the lab dashboard.",
	RunE:  runIngest,
}

var (
	ingestContent    string
	ingestCapability string
	ingestTraining   string
	ingestRunID      string
	ingestBatchSize  int
)

func init() {
	ingestCmd.Flags().StringVar(&ingestContent, "content", "", "Content scores JSONL file")
	ingestCmd.Flags().StringVar(&ingestCapability, "capability", "", "Capability scores JSONL file")
	ingestCmd.Flags().StringVar(&ingestTraining, "training-log", "", "MLX LoRA training log file")
	ingestCmd.Flags().StringVar(&ingestRunID, "run-id", "", "Run ID tag (defaults to model name)")
	ingestCmd.Flags().IntVar(&ingestBatchSize, "batch-size", 100, "Lines per InfluxDB write batch")
}

func runIngest(cmd *cli.Command, args []string) error {
	if modelName == "" {
		return fmt.Errorf("--model is required")
	}
	if ingestContent == "" && ingestCapability == "" && ingestTraining == "" {
		return fmt.Errorf("at least one of --content, --capability, or --training-log is required")
	}

	influx := ml.NewInfluxClient(influxURL, influxDB)

	cfg := ml.IngestConfig{
		ContentFile:    ingestContent,
		CapabilityFile: ingestCapability,
		TrainingLog:    ingestTraining,
		Model:          modelName,
		RunID:          ingestRunID,
		BatchSize:      ingestBatchSize,
	}

	return ml.Ingest(influx, cfg, os.Stdout)
}
