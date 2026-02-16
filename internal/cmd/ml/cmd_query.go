package ml

import (
	"encoding/json"
	"fmt"
	"os"
	"strings"

	"forge.lthn.ai/core/cli/pkg/cli"
	"forge.lthn.ai/core/cli/pkg/ml"
)

var queryCmd = &cli.Command{
	Use:   "query [sql]",
	Short: "Run ad-hoc SQL against DuckDB",
	Long:  "Executes arbitrary SQL against the DuckDB database. Non-SELECT queries are auto-wrapped as golden_set WHERE clauses.",
	Example: `  core ml query "SELECT COUNT(*) FROM golden_set"
  core ml query "domain = 'ethics'"
  core ml query --json "SHOW TABLES"`,
	Args: cli.MinimumNArgs(1),
	RunE: runQuery,
}

var queryJSON bool

func init() {
	queryCmd.Flags().BoolVar(&queryJSON, "json", false, "Output as JSON")
}

func runQuery(cmd *cli.Command, args []string) error {
	path := dbPath
	if path == "" {
		path = os.Getenv("LEM_DB")
	}
	if path == "" {
		return fmt.Errorf("--db or LEM_DB env is required")
	}

	db, err := ml.OpenDB(path)
	if err != nil {
		return fmt.Errorf("open db: %w", err)
	}
	defer db.Close()

	sql := strings.Join(args, " ")

	// Auto-wrap non-SELECT queries as golden_set WHERE clauses.
	trimmed := strings.TrimSpace(strings.ToUpper(sql))
	if !strings.HasPrefix(trimmed, "SELECT") && !strings.HasPrefix(trimmed, "SHOW") &&
		!strings.HasPrefix(trimmed, "DESCRIBE") && !strings.HasPrefix(trimmed, "EXPLAIN") {
		sql = "SELECT * FROM golden_set WHERE " + sql + " LIMIT 20"
	}

	rows, err := db.QueryRows(sql)
	if err != nil {
		return fmt.Errorf("query: %w", err)
	}

	if queryJSON {
		enc := json.NewEncoder(os.Stdout)
		enc.SetIndent("", "  ")
		if err := enc.Encode(rows); err != nil {
			return fmt.Errorf("encode json: %w", err)
		}
		fmt.Fprintf(os.Stderr, "\n(%d rows)\n", len(rows))
		return nil
	}

	if len(rows) == 0 {
		fmt.Println("(0 rows)")
		return nil
	}

	// Collect column names in stable order from first row.
	var cols []string
	for col := range rows[0] {
		cols = append(cols, col)
	}

	// Calculate column widths (capped at 60).
	const maxWidth = 60
	widths := make([]int, len(cols))
	for i, col := range cols {
		widths[i] = len(col)
	}
	for _, row := range rows {
		for i, col := range cols {
			val := formatValue(row[col])
			if l := len(val); l > widths[i] {
				widths[i] = l
			}
		}
	}
	for i := range widths {
		if widths[i] > maxWidth {
			widths[i] = maxWidth
		}
	}

	// Print header.
	for i, col := range cols {
		if i > 0 {
			fmt.Print(" | ")
		}
		fmt.Printf("%-*s", widths[i], truncate(col, widths[i]))
	}
	fmt.Println()

	// Print separator.
	for i := range cols {
		if i > 0 {
			fmt.Print("-+-")
		}
		fmt.Print(strings.Repeat("-", widths[i]))
	}
	fmt.Println()

	// Print rows.
	for _, row := range rows {
		for i, col := range cols {
			if i > 0 {
				fmt.Print(" | ")
			}
			fmt.Printf("%-*s", widths[i], truncate(formatValue(row[col]), widths[i]))
		}
		fmt.Println()
	}

	fmt.Printf("\n(%d rows)\n", len(rows))
	return nil
}

func formatValue(v interface{}) string {
	if v == nil {
		return "NULL"
	}
	return fmt.Sprintf("%v", v)
}

func truncate(s string, max int) string {
	if len(s) <= max {
		return s
	}
	if max <= 3 {
		return s[:max]
	}
	return s[:max-3] + "..."
}
