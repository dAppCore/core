package ml

import (
	"fmt"
	"io"
	"time"
)

// PushMetrics queries golden_set stats from DuckDB and writes them to InfluxDB
// as golden_set_stats, golden_set_domain, and golden_set_voice measurements.
func PushMetrics(db *DB, influx *InfluxClient, w io.Writer) error {
	// Overall stats.
	var total, domains, voices int
	var avgGenTime, avgChars float64
	err := db.conn.QueryRow(
		"SELECT count(*), count(DISTINCT domain), count(DISTINCT voice), " +
			"coalesce(avg(gen_time), 0), coalesce(avg(char_count), 0) FROM golden_set",
	).Scan(&total, &domains, &voices, &avgGenTime, &avgChars)
	if err != nil {
		return fmt.Errorf("query golden_set stats: %w", err)
	}

	if total == 0 {
		fmt.Fprintln(w, "golden_set is empty, nothing to push")
		return nil
	}

	completionPct := float64(total) / float64(TargetTotal) * 100.0
	ts := time.Now().UnixNano()

	var lines []string

	// Overall stats point.
	lines = append(lines, fmt.Sprintf(
		"golden_set_stats total_examples=%di,domains=%di,voices=%di,avg_gen_time=%.2f,avg_response_chars=%.0f,completion_pct=%.1f %d",
		total, domains, voices, avgGenTime, avgChars, completionPct, ts,
	))

	// Per-domain breakdown.
	domainRows, err := db.conn.Query(
		"SELECT domain, count(*) AS cnt, coalesce(avg(gen_time), 0) AS avg_gt FROM golden_set GROUP BY domain ORDER BY domain",
	)
	if err != nil {
		return fmt.Errorf("query golden_set domains: %w", err)
	}
	defer domainRows.Close()

	for domainRows.Next() {
		var domain string
		var count int
		var avgGT float64
		if err := domainRows.Scan(&domain, &count, &avgGT); err != nil {
			return fmt.Errorf("scan domain row: %w", err)
		}
		lines = append(lines, fmt.Sprintf(
			"golden_set_domain,domain=%s count=%di,avg_gen_time=%.2f %d",
			EscapeLp(domain), count, avgGT, ts,
		))
	}
	if err := domainRows.Err(); err != nil {
		return fmt.Errorf("iterate domain rows: %w", err)
	}

	// Per-voice breakdown.
	voiceRows, err := db.conn.Query(
		"SELECT voice, count(*) AS cnt, coalesce(avg(char_count), 0) AS avg_cc, coalesce(avg(gen_time), 0) AS avg_gt FROM golden_set GROUP BY voice ORDER BY voice",
	)
	if err != nil {
		return fmt.Errorf("query golden_set voices: %w", err)
	}
	defer voiceRows.Close()

	for voiceRows.Next() {
		var voice string
		var count int
		var avgCC, avgGT float64
		if err := voiceRows.Scan(&voice, &count, &avgCC, &avgGT); err != nil {
			return fmt.Errorf("scan voice row: %w", err)
		}
		lines = append(lines, fmt.Sprintf(
			"golden_set_voice,voice=%s count=%di,avg_chars=%.0f,avg_gen_time=%.2f %d",
			EscapeLp(voice), count, avgCC, avgGT, ts,
		))
	}
	if err := voiceRows.Err(); err != nil {
		return fmt.Errorf("iterate voice rows: %w", err)
	}

	// Write all points to InfluxDB.
	if err := influx.WriteLp(lines); err != nil {
		return fmt.Errorf("write metrics to influxdb: %w", err)
	}

	fmt.Fprintf(w, "Pushed %d points to InfluxDB\n", len(lines))
	fmt.Fprintf(w, "  total=%d  domains=%d  voices=%d  completion=%.1f%%\n",
		total, domains, voices, completionPct)
	fmt.Fprintf(w, "  avg_gen_time=%.2fs  avg_chars=%.0f\n", avgGenTime, avgChars)

	return nil
}
