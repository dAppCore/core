package ml

import (
	"fmt"
	"io"
	"sort"
)

// trainingRow holds deduplicated training status + loss for a single model.
type trainingRow struct {
	model      string
	status     string
	iteration  int
	totalIters int
	pct        float64
	loss       float64
	hasLoss    bool
}

// genRow holds deduplicated generation progress for a single worker.
type genRow struct {
	worker    string
	completed int
	target    int
	pct       float64
}

// PrintStatus queries InfluxDB for training and generation progress and writes
// a formatted summary to w.
func PrintStatus(influx *InfluxClient, w io.Writer) error {
	statusRows, err := influx.QuerySQL(
		"SELECT model, run_id, status, iteration, total_iters, pct FROM training_status ORDER BY time DESC LIMIT 10",
	)
	if err != nil {
		statusRows = nil
	}

	lossRows, err := influx.QuerySQL(
		"SELECT model, loss_type, loss, iteration, tokens_per_sec FROM training_loss WHERE loss_type = 'train' ORDER BY time DESC LIMIT 10",
	)
	if err != nil {
		lossRows = nil
	}

	goldenRows, err := influx.QuerySQL(
		"SELECT worker, completed, target, pct FROM golden_gen_progress ORDER BY time DESC LIMIT 5",
	)
	if err != nil {
		goldenRows = nil
	}

	expansionRows, err := influx.QuerySQL(
		"SELECT worker, completed, target, pct FROM expansion_progress ORDER BY time DESC LIMIT 5",
	)
	if err != nil {
		expansionRows = nil
	}

	training := dedupeTraining(statusRows, lossRows)
	golden := dedupeGeneration(goldenRows)
	expansion := dedupeGeneration(expansionRows)

	fmt.Fprintln(w, "Training:")
	if len(training) == 0 {
		fmt.Fprintln(w, "  (no data)")
	} else {
		for _, tr := range training {
			progress := fmt.Sprintf("%d/%d", tr.iteration, tr.totalIters)
			pct := fmt.Sprintf("%.1f%%", tr.pct)
			if tr.hasLoss {
				fmt.Fprintf(w, "  %-13s %-9s %9s %7s  loss=%.3f\n",
					tr.model, tr.status, progress, pct, tr.loss)
			} else {
				fmt.Fprintf(w, "  %-13s %-9s %9s %7s\n",
					tr.model, tr.status, progress, pct)
			}
		}
	}

	fmt.Fprintln(w)
	fmt.Fprintln(w, "Generation:")

	hasGenData := false

	if len(golden) > 0 {
		hasGenData = true
		for _, g := range golden {
			progress := fmt.Sprintf("%d/%d", g.completed, g.target)
			pct := fmt.Sprintf("%.1f%%", g.pct)
			fmt.Fprintf(w, "  %-13s %11s %7s  (%s)\n", "golden", progress, pct, g.worker)
		}
	}

	if len(expansion) > 0 {
		hasGenData = true
		for _, g := range expansion {
			progress := fmt.Sprintf("%d/%d", g.completed, g.target)
			pct := fmt.Sprintf("%.1f%%", g.pct)
			fmt.Fprintf(w, "  %-13s %11s %7s  (%s)\n", "expansion", progress, pct, g.worker)
		}
	}

	if !hasGenData {
		fmt.Fprintln(w, "  (no data)")
	}

	return nil
}

// dedupeTraining merges training status and loss rows, keeping only the first
// (latest) row per model.
func dedupeTraining(statusRows, lossRows []map[string]interface{}) []trainingRow {
	lossMap := make(map[string]float64)
	lossSeenMap := make(map[string]bool)
	for _, row := range lossRows {
		model := strVal(row, "model")
		if model == "" || lossSeenMap[model] {
			continue
		}
		lossSeenMap[model] = true
		lossMap[model] = floatVal(row, "loss")
	}

	seen := make(map[string]bool)
	var rows []trainingRow
	for _, row := range statusRows {
		model := strVal(row, "model")
		if model == "" || seen[model] {
			continue
		}
		seen[model] = true

		tr := trainingRow{
			model:      model,
			status:     strVal(row, "status"),
			iteration:  intVal(row, "iteration"),
			totalIters: intVal(row, "total_iters"),
			pct:        floatVal(row, "pct"),
		}

		if loss, ok := lossMap[model]; ok {
			tr.loss = loss
			tr.hasLoss = true
		}

		rows = append(rows, tr)
	}

	sort.Slice(rows, func(i, j int) bool {
		return rows[i].model < rows[j].model
	})

	return rows
}

// dedupeGeneration deduplicates generation progress rows by worker.
func dedupeGeneration(rows []map[string]interface{}) []genRow {
	seen := make(map[string]bool)
	var result []genRow
	for _, row := range rows {
		worker := strVal(row, "worker")
		if worker == "" || seen[worker] {
			continue
		}
		seen[worker] = true

		result = append(result, genRow{
			worker:    worker,
			completed: intVal(row, "completed"),
			target:    intVal(row, "target"),
			pct:       floatVal(row, "pct"),
		})
	}

	sort.Slice(result, func(i, j int) bool {
		return result[i].worker < result[j].worker
	})

	return result
}

// strVal extracts a string value from a row map.
func strVal(row map[string]interface{}, key string) string {
	v, ok := row[key]
	if !ok {
		return ""
	}
	s, ok := v.(string)
	if !ok {
		return ""
	}
	return s
}

// floatVal extracts a float64 value from a row map.
func floatVal(row map[string]interface{}, key string) float64 {
	v, ok := row[key]
	if !ok {
		return 0
	}
	f, ok := v.(float64)
	if !ok {
		return 0
	}
	return f
}

// intVal extracts an integer value from a row map. InfluxDB JSON returns all
// numbers as float64, so this truncates to int.
func intVal(row map[string]interface{}, key string) int {
	return int(floatVal(row, key))
}
