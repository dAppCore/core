// Package ml provides ML inference, scoring, and training pipeline commands.
//
// Commands:
//   - core ml score: Score responses with heuristic and LLM judges
//   - core ml probe: Run capability and content probes against a model
//   - core ml export: Export golden set to training JSONL/Parquet
//   - core ml expand: Generate expansion responses
//   - core ml status: Show training and generation progress
//   - core ml gguf: Convert MLX LoRA adapter to GGUF format
//   - core ml convert: Convert MLX LoRA adapter to PEFT format
//   - core ml agent: Run the scoring agent daemon
//   - core ml worker: Run a distributed worker node
package ml

import (
	"forge.lthn.ai/core/cli/pkg/cli"
)

func init() {
	cli.RegisterCommands(AddMLCommands)
}

var mlCmd = &cli.Command{
	Use:   "ml",
	Short: "ML inference, scoring, and training pipeline",
	Long:  "Commands for ML model scoring, probe evaluation, data export, and format conversion.",
}

// AddMLCommands registers the 'ml' command and all subcommands.
func AddMLCommands(root *cli.Command) {
	initFlags()
	mlCmd.AddCommand(scoreCmd)
	mlCmd.AddCommand(probeCmd)
	mlCmd.AddCommand(exportCmd)
	mlCmd.AddCommand(expandCmd)
	mlCmd.AddCommand(statusCmd)
	mlCmd.AddCommand(ggufCmd)
	mlCmd.AddCommand(convertCmd)
	mlCmd.AddCommand(agentCmd)
	mlCmd.AddCommand(workerCmd)
	root.AddCommand(mlCmd)
}

// Shared persistent flags.
var (
	apiURL     string
	judgeURL   string
	judgeModel string
	influxURL  string
	influxDB   string
	dbPath     string
	modelName  string
)

func initFlags() {
	mlCmd.PersistentFlags().StringVar(&apiURL, "api-url", "http://10.69.69.108:8090", "OpenAI-compatible API URL")
	mlCmd.PersistentFlags().StringVar(&judgeURL, "judge-url", "http://10.69.69.108:11434", "Judge model API URL (Ollama)")
	mlCmd.PersistentFlags().StringVar(&judgeModel, "judge-model", "gemma3:27b", "Judge model name")
	mlCmd.PersistentFlags().StringVar(&influxURL, "influx", "", "InfluxDB URL (default http://10.69.69.165:8181)")
	mlCmd.PersistentFlags().StringVar(&influxDB, "influx-db", "", "InfluxDB database (default training)")
	mlCmd.PersistentFlags().StringVar(&dbPath, "db", "", "DuckDB database path (or set LEM_DB env)")
	mlCmd.PersistentFlags().StringVar(&modelName, "model", "", "Model name for API")
}
