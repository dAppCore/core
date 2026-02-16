package ml

import (
	"fmt"
	"os"

	"forge.lthn.ai/core/cli/pkg/cli"
	"forge.lthn.ai/core/cli/pkg/ml"
)

var inventoryCmd = &cli.Command{
	Use:   "inventory",
	Short: "Show DuckDB table inventory with stats",
	Long:  "Queries all DuckDB tables and prints row counts with per-table detail breakdowns.",
	RunE:  runInventory,
}

func runInventory(cmd *cli.Command, args []string) error {
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

	return ml.PrintInventory(db, os.Stdout)
}
