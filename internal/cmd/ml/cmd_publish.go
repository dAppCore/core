package ml

import (
	"forge.lthn.ai/core/cli/pkg/cli"
	"forge.lthn.ai/core/cli/pkg/ml"
)

var (
	publishInputDir string
	publishRepo     string
	publishPublic   bool
	publishToken    string
	publishDryRun   bool
)

var publishCmd = &cli.Command{
	Use:   "publish",
	Short: "Upload Parquet dataset to HuggingFace Hub",
	Long:  "Uploads train/valid/test Parquet files and an optional dataset card to a HuggingFace dataset repository.",
	RunE:  runPublish,
}

func init() {
	publishCmd.Flags().StringVar(&publishInputDir, "input-dir", "", "Directory containing Parquet files (required)")
	publishCmd.Flags().StringVar(&publishRepo, "repo", "lthn/LEM-golden-set", "HuggingFace dataset repo ID")
	publishCmd.Flags().BoolVar(&publishPublic, "public", false, "Make dataset public")
	publishCmd.Flags().StringVar(&publishToken, "token", "", "HuggingFace API token (defaults to HF_TOKEN env)")
	publishCmd.Flags().BoolVar(&publishDryRun, "dry-run", false, "Show what would be uploaded without uploading")
	_ = publishCmd.MarkFlagRequired("input-dir")
}

func runPublish(cmd *cli.Command, args []string) error {
	return ml.Publish(ml.PublishConfig{
		InputDir: publishInputDir,
		Repo:     publishRepo,
		Public:   publishPublic,
		Token:    publishToken,
		DryRun:   publishDryRun,
	}, cmd.OutOrStdout())
}
