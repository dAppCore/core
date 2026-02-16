package ml

import (
	"forge.lthn.ai/core/cli/pkg/cli"
	"forge.lthn.ai/core/cli/pkg/ml"
)

var (
	agentM3Host        string
	agentM3User        string
	agentM3SSHKey      string
	agentM3AdapterBase string
	agentBaseModel     string
	agentPollInterval  int
	agentWorkDir       string
	agentFilter        string
	agentForce         bool
	agentOneShot       bool
	agentDryRun        bool
)

var agentCmd = &cli.Command{
	Use:   "agent",
	Short: "Run the scoring agent daemon",
	Long:  "Polls M3 for unscored LoRA checkpoints, converts, probes, and pushes results to InfluxDB.",
	RunE:  runAgent,
}

func init() {
	agentCmd.Flags().StringVar(&agentM3Host, "m3-host", ml.EnvOr("M3_HOST", "10.69.69.108"), "M3 host address")
	agentCmd.Flags().StringVar(&agentM3User, "m3-user", ml.EnvOr("M3_USER", "claude"), "M3 SSH user")
	agentCmd.Flags().StringVar(&agentM3SSHKey, "m3-ssh-key", ml.EnvOr("M3_SSH_KEY", ml.ExpandHome("~/.ssh/id_ed25519")), "SSH key for M3")
	agentCmd.Flags().StringVar(&agentM3AdapterBase, "m3-adapter-base", ml.EnvOr("M3_ADAPTER_BASE", "/Volumes/Data/lem"), "Adapter base dir on M3")
	agentCmd.Flags().StringVar(&agentBaseModel, "base-model", ml.EnvOr("BASE_MODEL", "deepseek-ai/DeepSeek-R1-Distill-Qwen-7B"), "HuggingFace base model ID")
	agentCmd.Flags().IntVar(&agentPollInterval, "poll", ml.IntEnvOr("POLL_INTERVAL", 300), "Poll interval in seconds")
	agentCmd.Flags().StringVar(&agentWorkDir, "work-dir", ml.EnvOr("WORK_DIR", "/tmp/scoring-agent"), "Working directory for adapters")
	agentCmd.Flags().StringVar(&agentFilter, "filter", "", "Filter adapter dirs by prefix")
	agentCmd.Flags().BoolVar(&agentForce, "force", false, "Re-score already-scored checkpoints")
	agentCmd.Flags().BoolVar(&agentOneShot, "one-shot", false, "Process one checkpoint and exit")
	agentCmd.Flags().BoolVar(&agentDryRun, "dry-run", false, "Discover and plan but don't execute")
}

func runAgent(cmd *cli.Command, args []string) error {
	cfg := &ml.AgentConfig{
		M3Host:        agentM3Host,
		M3User:        agentM3User,
		M3SSHKey:      agentM3SSHKey,
		M3AdapterBase: agentM3AdapterBase,
		InfluxURL:     influxURL,
		InfluxDB:      influxDB,
		DBPath:        dbPath,
		APIURL:        apiURL,
		JudgeURL:      judgeURL,
		JudgeModel:    judgeModel,
		Model:         modelName,
		BaseModel:     agentBaseModel,
		PollInterval:  agentPollInterval,
		WorkDir:       agentWorkDir,
		Filter:        agentFilter,
		Force:         agentForce,
		OneShot:       agentOneShot,
		DryRun:        agentDryRun,
	}

	ml.RunAgentLoop(cfg)
	return nil
}
