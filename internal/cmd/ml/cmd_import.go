package ml

import (
	"fmt"
	"os"
	"path/filepath"

	"forge.lthn.ai/core/cli/pkg/cli"
	"forge.lthn.ai/core/cli/pkg/ml"
)

var importCmd = &cli.Command{
	Use:   "import-all",
	Short: "Import all LEM data into DuckDB",
	Long:  "Imports golden set, training examples, benchmark results, benchmark questions, and seeds into DuckDB from M3 and local files.",
	RunE:  runImportAll,
}

var (
	importSkipM3 bool
	importDataDir string
	importM3Host  string
)

func init() {
	importCmd.Flags().BoolVar(&importSkipM3, "skip-m3", false, "Skip pulling data from M3")
	importCmd.Flags().StringVar(&importDataDir, "data-dir", "", "Local data directory (defaults to db directory)")
	importCmd.Flags().StringVar(&importM3Host, "m3-host", "m3", "M3 SSH host alias")
}

func runImportAll(cmd *cli.Command, args []string) error {
	path := dbPath
	if path == "" {
		path = os.Getenv("LEM_DB")
	}
	if path == "" {
		return fmt.Errorf("--db or LEM_DB required")
	}

	dataDir := importDataDir
	if dataDir == "" {
		dataDir = filepath.Dir(path)
	}

	db, err := ml.OpenDBReadWrite(path)
	if err != nil {
		return fmt.Errorf("open db: %w", err)
	}
	defer db.Close()

	cfg := ml.ImportConfig{
		SkipM3:  importSkipM3,
		DataDir: dataDir,
		M3Host:  importM3Host,
	}

	return ml.ImportAll(db, cfg, cmd.OutOrStdout())
}
