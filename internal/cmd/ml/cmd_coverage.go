package ml

import (
	"fmt"
	"os"

	"forge.lthn.ai/core/cli/pkg/cli"
	"forge.lthn.ai/core/cli/pkg/ml"
)

var coverageCmd = &cli.Command{
	Use:   "coverage",
	Short: "Analyze seed coverage by region and domain",
	Long:  "Queries seeds by region and domain, renders ASCII bar charts, and highlights underrepresented areas.",
	RunE:  runCoverage,
}

func runCoverage(cmd *cli.Command, args []string) error {
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

	return ml.PrintCoverage(db, cmd.OutOrStdout())
}
