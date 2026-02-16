package ml

import (
	"fmt"
	"io"
	"strings"
)

// SeedInfluxConfig holds options for the SeedInflux migration.
type SeedInfluxConfig struct {
	Force     bool
	BatchSize int
}

// SeedInflux migrates golden_set rows from DuckDB into InfluxDB as
// gold_gen measurement points. This is a one-time migration tool;
// it skips the write when InfluxDB already contains all records
// unless Force is set.
func SeedInflux(db *DB, influx *InfluxClient, cfg SeedInfluxConfig, w io.Writer) error {
	if cfg.BatchSize <= 0 {
		cfg.BatchSize = 500
	}

	// Count source rows in DuckDB.
	var total int
	if err := db.conn.QueryRow("SELECT count(*) FROM golden_set").Scan(&total); err != nil {
		return fmt.Errorf("no golden_set table: %w", err)
	}

	// Check how many distinct records InfluxDB already has.
	existing := 0
	rows, err := influx.QuerySQL("SELECT count(DISTINCT i) AS n FROM gold_gen")
	if err == nil && len(rows) > 0 {
		if n, ok := rows[0]["n"].(float64); ok {
			existing = int(n)
		}
	}

	fmt.Fprintf(w, "DuckDB has %d records, InfluxDB golden_gen has %d\n", total, existing)

	if existing >= total && !cfg.Force {
		fmt.Fprintln(w, "InfluxDB already has all records. Use --force to re-seed.")
		return nil
	}

	// Query all golden_set rows from DuckDB.
	dbRows, err := db.conn.Query(
		"SELECT idx, seed_id, domain, voice, gen_time, char_count FROM golden_set ORDER BY idx",
	)
	if err != nil {
		return fmt.Errorf("query golden_set: %w", err)
	}
	defer dbRows.Close()

	var batch []string
	written := 0

	for dbRows.Next() {
		var idx int
		var seedID, domain, voice string
		var genTime float64
		var charCount int

		if err := dbRows.Scan(&idx, &seedID, &domain, &voice, &genTime, &charCount); err != nil {
			return fmt.Errorf("scan row %d: %w", written, err)
		}

		// Build line protocol point.
		// Tags: i (idx), w (worker), d (domain), v (voice)
		// Fields: seed_id (string), gen_time (float), chars (integer)
		escapedSeedID := strings.ReplaceAll(seedID, `"`, `\"`)

		line := fmt.Sprintf(
			"gold_gen,i=%s,w=migration,d=%s,v=%s seed_id=\"%s\",gen_time=%v,chars=%di",
			EscapeLp(fmt.Sprintf("%d", idx)),
			EscapeLp(domain),
			EscapeLp(voice),
			escapedSeedID,
			genTime,
			charCount,
		)
		batch = append(batch, line)

		if len(batch) >= cfg.BatchSize {
			if err := influx.WriteLp(batch); err != nil {
				return fmt.Errorf("write batch at row %d: %w", written, err)
			}
			written += len(batch)
			batch = batch[:0]

			if written%2000 == 0 {
				fmt.Fprintf(w, "  wrote %d / %d\n", written, total)
			}
		}
	}

	if err := dbRows.Err(); err != nil {
		return fmt.Errorf("iterate golden_set rows: %w", err)
	}

	// Flush remaining batch.
	if len(batch) > 0 {
		if err := influx.WriteLp(batch); err != nil {
			return fmt.Errorf("write final batch: %w", err)
		}
		written += len(batch)
	}

	fmt.Fprintf(w, "Seeded %d records into InfluxDB golden_gen\n", written)
	return nil
}
