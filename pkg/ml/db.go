package ml

import (
	"database/sql"
	"fmt"

	_ "github.com/marcboeker/go-duckdb"
)

// DB wraps a DuckDB connection.
type DB struct {
	conn *sql.DB
	path string
}

// OpenDB opens a DuckDB database file in read-only mode to avoid locking
// issues with the Python pipeline.
func OpenDB(path string) (*DB, error) {
	conn, err := sql.Open("duckdb", path+"?access_mode=READ_ONLY")
	if err != nil {
		return nil, fmt.Errorf("open duckdb %s: %w", path, err)
	}
	if err := conn.Ping(); err != nil {
		conn.Close()
		return nil, fmt.Errorf("ping duckdb %s: %w", path, err)
	}
	return &DB{conn: conn, path: path}, nil
}

// OpenDBReadWrite opens a DuckDB database in read-write mode.
func OpenDBReadWrite(path string) (*DB, error) {
	conn, err := sql.Open("duckdb", path)
	if err != nil {
		return nil, fmt.Errorf("open duckdb %s: %w", path, err)
	}
	if err := conn.Ping(); err != nil {
		conn.Close()
		return nil, fmt.Errorf("ping duckdb %s: %w", path, err)
	}
	return &DB{conn: conn, path: path}, nil
}

// Close closes the database connection.
func (db *DB) Close() error {
	return db.conn.Close()
}

// Path returns the database file path.
func (db *DB) Path() string {
	return db.path
}

// Exec executes a query without returning rows.
func (db *DB) Exec(query string, args ...interface{}) error {
	_, err := db.conn.Exec(query, args...)
	return err
}

// QueryRowScan executes a query expected to return at most one row and scans
// the result into dest. It is a convenience wrapper around sql.DB.QueryRow.
func (db *DB) QueryRowScan(query string, dest interface{}, args ...interface{}) error {
	return db.conn.QueryRow(query, args...).Scan(dest)
}

// GoldenSetRow represents one row from the golden_set table.
type GoldenSetRow struct {
	Idx       int
	SeedID    string
	Domain    string
	Voice     string
	Prompt    string
	Response  string
	GenTime   float64
	CharCount int
}

// ExpansionPromptRow represents one row from the expansion_prompts table.
type ExpansionPromptRow struct {
	Idx      int64
	SeedID   string
	Region   string
	Domain   string
	Language string
	Prompt   string
	PromptEn string
	Priority int
	Status   string
}

// QueryGoldenSet returns all golden set rows with responses >= minChars.
func (db *DB) QueryGoldenSet(minChars int) ([]GoldenSetRow, error) {
	rows, err := db.conn.Query(
		"SELECT idx, seed_id, domain, voice, prompt, response, gen_time, char_count "+
			"FROM golden_set WHERE char_count >= ? ORDER BY idx",
		minChars,
	)
	if err != nil {
		return nil, fmt.Errorf("query golden_set: %w", err)
	}
	defer rows.Close()

	var result []GoldenSetRow
	for rows.Next() {
		var r GoldenSetRow
		if err := rows.Scan(&r.Idx, &r.SeedID, &r.Domain, &r.Voice,
			&r.Prompt, &r.Response, &r.GenTime, &r.CharCount); err != nil {
			return nil, fmt.Errorf("scan golden_set row: %w", err)
		}
		result = append(result, r)
	}
	return result, rows.Err()
}

// CountGoldenSet returns the total count of golden set rows.
func (db *DB) CountGoldenSet() (int, error) {
	var count int
	err := db.conn.QueryRow("SELECT COUNT(*) FROM golden_set").Scan(&count)
	if err != nil {
		return 0, fmt.Errorf("count golden_set: %w", err)
	}
	return count, nil
}

// QueryExpansionPrompts returns expansion prompts filtered by status.
func (db *DB) QueryExpansionPrompts(status string, limit int) ([]ExpansionPromptRow, error) {
	query := "SELECT idx, seed_id, region, domain, language, prompt, prompt_en, priority, status " +
		"FROM expansion_prompts"
	var args []interface{}

	if status != "" {
		query += " WHERE status = ?"
		args = append(args, status)
	}
	query += " ORDER BY priority, idx"

	if limit > 0 {
		query += fmt.Sprintf(" LIMIT %d", limit)
	}

	rows, err := db.conn.Query(query, args...)
	if err != nil {
		return nil, fmt.Errorf("query expansion_prompts: %w", err)
	}
	defer rows.Close()

	var result []ExpansionPromptRow
	for rows.Next() {
		var r ExpansionPromptRow
		if err := rows.Scan(&r.Idx, &r.SeedID, &r.Region, &r.Domain,
			&r.Language, &r.Prompt, &r.PromptEn, &r.Priority, &r.Status); err != nil {
			return nil, fmt.Errorf("scan expansion_prompt row: %w", err)
		}
		result = append(result, r)
	}
	return result, rows.Err()
}

// CountExpansionPrompts returns counts by status.
func (db *DB) CountExpansionPrompts() (total int, pending int, err error) {
	err = db.conn.QueryRow("SELECT COUNT(*) FROM expansion_prompts").Scan(&total)
	if err != nil {
		return 0, 0, fmt.Errorf("count expansion_prompts: %w", err)
	}
	err = db.conn.QueryRow("SELECT COUNT(*) FROM expansion_prompts WHERE status = 'pending'").Scan(&pending)
	if err != nil {
		return total, 0, fmt.Errorf("count pending expansion_prompts: %w", err)
	}
	return total, pending, nil
}

// UpdateExpansionStatus updates the status of an expansion prompt by idx.
func (db *DB) UpdateExpansionStatus(idx int64, status string) error {
	_, err := db.conn.Exec("UPDATE expansion_prompts SET status = ? WHERE idx = ?", status, idx)
	if err != nil {
		return fmt.Errorf("update expansion_prompt %d: %w", idx, err)
	}
	return nil
}

// QueryRows executes an arbitrary SQL query and returns results as maps.
func (db *DB) QueryRows(query string, args ...interface{}) ([]map[string]interface{}, error) {
	rows, err := db.conn.Query(query, args...)
	if err != nil {
		return nil, fmt.Errorf("query: %w", err)
	}
	defer rows.Close()

	cols, err := rows.Columns()
	if err != nil {
		return nil, fmt.Errorf("columns: %w", err)
	}

	var result []map[string]interface{}
	for rows.Next() {
		values := make([]interface{}, len(cols))
		ptrs := make([]interface{}, len(cols))
		for i := range values {
			ptrs[i] = &values[i]
		}
		if err := rows.Scan(ptrs...); err != nil {
			return nil, fmt.Errorf("scan: %w", err)
		}
		row := make(map[string]interface{}, len(cols))
		for i, col := range cols {
			row[col] = values[i]
		}
		result = append(result, row)
	}
	return result, rows.Err()
}

// EnsureScoringTables creates the scoring tables if they don't exist.
func (db *DB) EnsureScoringTables() {
	db.conn.Exec(`CREATE TABLE IF NOT EXISTS checkpoint_scores (
		model TEXT, run_id TEXT, label TEXT, iteration INTEGER,
		correct INTEGER, total INTEGER, accuracy DOUBLE,
		scored_at TIMESTAMP DEFAULT current_timestamp,
		PRIMARY KEY (run_id, label)
	)`)
	db.conn.Exec(`CREATE TABLE IF NOT EXISTS probe_results (
		model TEXT, run_id TEXT, label TEXT, probe_id TEXT,
		passed BOOLEAN, response TEXT, iteration INTEGER,
		scored_at TIMESTAMP DEFAULT current_timestamp,
		PRIMARY KEY (run_id, label, probe_id)
	)`)
	db.conn.Exec(`CREATE TABLE IF NOT EXISTS scoring_results (
		model TEXT, prompt_id TEXT, suite TEXT,
		dimension TEXT, score DOUBLE,
		scored_at TIMESTAMP DEFAULT current_timestamp
	)`)
}

// WriteScoringResult writes a single scoring dimension result to DuckDB.
func (db *DB) WriteScoringResult(model, promptID, suite, dimension string, score float64) error {
	_, err := db.conn.Exec(
		`INSERT INTO scoring_results (model, prompt_id, suite, dimension, score) VALUES (?, ?, ?, ?, ?)`,
		model, promptID, suite, dimension, score,
	)
	return err
}

// TableCounts returns row counts for all known tables.
func (db *DB) TableCounts() (map[string]int, error) {
	tables := []string{"golden_set", "expansion_prompts", "seeds", "prompts",
		"training_examples", "gemini_responses", "benchmark_questions", "benchmark_results", "validations",
		"checkpoint_scores", "probe_results", "scoring_results"}

	counts := make(map[string]int)
	for _, t := range tables {
		var count int
		err := db.conn.QueryRow(fmt.Sprintf("SELECT COUNT(*) FROM %s", t)).Scan(&count)
		if err != nil {
			continue
		}
		counts[t] = count
	}
	return counts, nil
}
