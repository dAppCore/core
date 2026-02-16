package ml

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"time"
)

// ExpandOutput is the JSONL output structure for expansion generation.
type ExpandOutput struct {
	ID             string  `json:"id"`
	Domain         string  `json:"domain,omitempty"`
	Prompt         string  `json:"prompt"`
	Response       string  `json:"response"`
	Model          string  `json:"model"`
	ElapsedSeconds float64 `json:"elapsed_seconds"`
	Chars          int     `json:"chars"`
}

// GetCompletedIDs queries InfluxDB for prompt IDs that have already been
// processed in the expansion_gen measurement.
func GetCompletedIDs(influx *InfluxClient) (map[string]bool, error) {
	rows, err := influx.QuerySQL("SELECT DISTINCT seed_id FROM expansion_gen")
	if err != nil {
		return nil, fmt.Errorf("query expansion_gen: %w", err)
	}

	ids := make(map[string]bool, len(rows))
	for _, row := range rows {
		id := strVal(row, "seed_id")
		if id != "" {
			ids[id] = true
		}
	}

	return ids, nil
}

// ExpandPrompts generates responses for expansion prompts using the given
// backend and reports progress to InfluxDB. Already-completed prompts (per
// InfluxDB) are skipped. API errors for individual prompts are logged and
// skipped. InfluxDB reporting is best-effort.
func ExpandPrompts(ctx context.Context, backend Backend, influx *InfluxClient, prompts []Response,
	modelName, worker, outputDir string, dryRun bool, limit int) error {

	remaining := prompts

	// Check InfluxDB for already-completed IDs.
	completed, err := GetCompletedIDs(influx)
	if err != nil {
		log.Printf("warning: could not check completed IDs: %v", err)
	} else {
		remaining = nil
		for _, p := range prompts {
			if !completed[p.ID] {
				remaining = append(remaining, p)
			}
		}

		skipped := len(prompts) - len(remaining)
		if skipped > 0 {
			log.Printf("skipping %d already-completed prompts, %d remaining", skipped, len(remaining))
		}
	}

	if limit > 0 && limit < len(remaining) {
		remaining = remaining[:limit]
	}

	if len(remaining) == 0 {
		log.Println("all prompts already completed, nothing to do")
		return nil
	}

	if dryRun {
		log.Printf("dry-run: would process %d prompts with model %s (worker: %s)", len(remaining), modelName, worker)
		for i, p := range remaining {
			if i >= 10 {
				log.Printf("  ... and %d more", len(remaining)-10)
				break
			}
			log.Printf("  %s (domain: %s)", p.ID, p.Domain)
		}
		return nil
	}

	outputPath := filepath.Join(outputDir, fmt.Sprintf("expand-%s.jsonl", worker))
	f, err := os.OpenFile(outputPath, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		return fmt.Errorf("open output file: %w", err)
	}
	defer f.Close()

	total := len(remaining)
	completedCount := 0

	for idx, p := range remaining {
		start := time.Now()
		response, err := backend.Generate(ctx, p.Prompt, GenOpts{Temperature: 0.7, MaxTokens: 2048})
		elapsed := time.Since(start).Seconds()

		if err != nil {
			log.Printf("[%d/%d] id=%s ERROR: %v", idx+1, total, p.ID, err)
			continue
		}

		chars := len(response)
		completedCount++

		out := ExpandOutput{
			ID:             p.ID,
			Domain:         p.Domain,
			Prompt:         p.Prompt,
			Response:       response,
			Model:          modelName,
			ElapsedSeconds: elapsed,
			Chars:          chars,
		}

		line, err := json.Marshal(out)
		if err != nil {
			log.Printf("[%d/%d] id=%s marshal error: %v", idx+1, total, p.ID, err)
			continue
		}

		if _, err := f.Write(append(line, '\n')); err != nil {
			log.Printf("[%d/%d] id=%s write error: %v", idx+1, total, p.ID, err)
			continue
		}

		genLine := fmt.Sprintf("expansion_gen,i=%d,w=%s,d=%s seed_id=\"%s\",gen_time=%f,chars=%di,model=\"%s\"",
			idx, EscapeLp(worker), EscapeLp(p.Domain),
			p.ID, elapsed, chars, modelName)

		pct := float64(completedCount) / float64(total) * 100.0
		progressLine := fmt.Sprintf("expansion_progress,worker=%s completed=%di,target=%di,pct=%f",
			EscapeLp(worker), completedCount, total, pct)

		if writeErr := influx.WriteLp([]string{genLine, progressLine}); writeErr != nil {
			log.Printf("[%d/%d] id=%s influx write error: %v", idx+1, total, p.ID, writeErr)
		}

		log.Printf("[%d/%d] id=%s chars=%d time=%.1fs", idx+1, total, p.ID, chars, elapsed)
	}

	log.Printf("expand complete: %d/%d prompts generated, output: %s", completedCount, total, outputPath)

	return nil
}
