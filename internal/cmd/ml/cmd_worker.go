package ml

import (
	"time"

	"forge.lthn.ai/core/cli/pkg/cli"
	"forge.lthn.ai/core/cli/pkg/ml"
)

var (
	workerAPIBase   string
	workerID        string
	workerName      string
	workerAPIKey    string
	workerGPU       string
	workerVRAM      int
	workerLangs     string
	workerModels    string
	workerInferURL  string
	workerTaskType  string
	workerBatchSize int
	workerPoll      time.Duration
	workerOneShot   bool
	workerDryRun    bool
)

var workerCmd = &cli.Command{
	Use:   "worker",
	Short: "Run a distributed worker node",
	Long:  "Polls the LEM API for tasks, runs local inference, and submits results.",
	RunE:  runWorker,
}

func init() {
	workerCmd.Flags().StringVar(&workerAPIBase, "api", ml.EnvOr("LEM_API", "https://infer.lthn.ai"), "LEM API base URL")
	workerCmd.Flags().StringVar(&workerID, "id", ml.EnvOr("LEM_WORKER_ID", ml.MachineID()), "Worker ID")
	workerCmd.Flags().StringVar(&workerName, "name", ml.EnvOr("LEM_WORKER_NAME", ml.Hostname()), "Worker display name")
	workerCmd.Flags().StringVar(&workerAPIKey, "key", ml.EnvOr("LEM_API_KEY", ""), "API key")
	workerCmd.Flags().StringVar(&workerGPU, "gpu", ml.EnvOr("LEM_GPU", ""), "GPU type")
	workerCmd.Flags().IntVar(&workerVRAM, "vram", ml.IntEnvOr("LEM_VRAM_GB", 0), "GPU VRAM in GB")
	workerCmd.Flags().StringVar(&workerLangs, "languages", ml.EnvOr("LEM_LANGUAGES", ""), "Comma-separated language codes")
	workerCmd.Flags().StringVar(&workerModels, "models", ml.EnvOr("LEM_MODELS", ""), "Comma-separated model names")
	workerCmd.Flags().StringVar(&workerInferURL, "infer", ml.EnvOr("LEM_INFER_URL", "http://localhost:8090"), "Local inference endpoint")
	workerCmd.Flags().StringVar(&workerTaskType, "type", "", "Filter by task type")
	workerCmd.Flags().IntVar(&workerBatchSize, "batch", 5, "Tasks per poll")
	workerCmd.Flags().DurationVar(&workerPoll, "poll", 30*time.Second, "Poll interval")
	workerCmd.Flags().BoolVar(&workerOneShot, "one-shot", false, "Process one batch and exit")
	workerCmd.Flags().BoolVar(&workerDryRun, "dry-run", false, "Fetch tasks but don't run inference")
}

func runWorker(cmd *cli.Command, args []string) error {
	if workerAPIKey == "" {
		workerAPIKey = ml.ReadKeyFile()
	}

	cfg := &ml.WorkerConfig{
		APIBase:      workerAPIBase,
		WorkerID:     workerID,
		Name:         workerName,
		APIKey:       workerAPIKey,
		GPUType:      workerGPU,
		VRAMGb:       workerVRAM,
		InferURL:     workerInferURL,
		TaskType:     workerTaskType,
		BatchSize:    workerBatchSize,
		PollInterval: workerPoll,
		OneShot:      workerOneShot,
		DryRun:       workerDryRun,
	}

	if workerLangs != "" {
		cfg.Languages = ml.SplitComma(workerLangs)
	}
	if workerModels != "" {
		cfg.Models = ml.SplitComma(workerModels)
	}

	ml.RunWorkerLoop(cfg)
	return nil
}
