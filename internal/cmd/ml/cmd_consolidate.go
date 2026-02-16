package ml

import (
	"forge.lthn.ai/core/cli/pkg/cli"
	"forge.lthn.ai/core/cli/pkg/ml"
)

var (
	consolidateM3Host    string
	consolidateRemoteDir string
	consolidatePattern   string
	consolidateOutputDir string
	consolidateMergedOut string
)

var consolidateCmd = &cli.Command{
	Use:   "consolidate",
	Short: "Pull and merge response JSONL files from M3",
	Long:  "Pulls JSONL response files from M3 via SSH/SCP, merges them by idx, deduplicates, and writes a single merged JSONL output.",
	RunE:  runConsolidate,
}

func init() {
	consolidateCmd.Flags().StringVar(&consolidateM3Host, "m3-host", "m3", "M3 SSH host")
	consolidateCmd.Flags().StringVar(&consolidateRemoteDir, "remote", "/Volumes/Data/lem/responses", "Remote response directory")
	consolidateCmd.Flags().StringVar(&consolidatePattern, "pattern", "gold*.jsonl", "File glob pattern")
	consolidateCmd.Flags().StringVar(&consolidateOutputDir, "output", "", "Local output directory (default: responses)")
	consolidateCmd.Flags().StringVar(&consolidateMergedOut, "merged", "", "Merged output path (default: gold-merged.jsonl in parent of output dir)")
}

func runConsolidate(cmd *cli.Command, args []string) error {
	cfg := ml.ConsolidateConfig{
		M3Host:    consolidateM3Host,
		RemoteDir: consolidateRemoteDir,
		Pattern:   consolidatePattern,
		OutputDir: consolidateOutputDir,
		MergedOut: consolidateMergedOut,
	}

	return ml.Consolidate(cfg, cmd.OutOrStdout())
}
