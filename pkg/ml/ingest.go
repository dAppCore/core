package ml

import (
	"bufio"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"regexp"
	"strconv"
	"strings"
	"time"
)

// IngestConfig holds the configuration for a benchmark/training ingest run.
type IngestConfig struct {
	ContentFile    string
	CapabilityFile string
	TrainingLog    string
	Model          string
	RunID          string
	BatchSize      int
}

// contentScoreLine is the JSON structure for a content scores JSONL line.
type contentScoreLine struct {
	Label      string                                    `json:"label"`
	Aggregates map[string]interface{}                    `json:"aggregates"`
	Probes     map[string]contentScoreProbe              `json:"probes"`
}

// contentScoreProbe is the per-probe block within a content score line.
type contentScoreProbe struct {
	Scores map[string]interface{} `json:"scores"`
}

// capabilityScoreLine is the JSON structure for a capability scores JSONL line.
type capabilityScoreLine struct {
	Label      string                        `json:"label"`
	Accuracy   float64                       `json:"accuracy"`
	Correct    int                           `json:"correct"`
	Total      int                           `json:"total"`
	ByCategory map[string]capabilityCatBlock `json:"by_category"`
}

// capabilityCatBlock is the per-category block within a capability score line.
type capabilityCatBlock struct {
	Correct int `json:"correct"`
	Total   int `json:"total"`
}

// Training log regexes.
var (
	reValLoss   = regexp.MustCompile(`Iter (\d+): Val loss ([\d.]+)`)
	reTrainLoss = regexp.MustCompile(`Iter (\d+): Train loss ([\d.]+), Learning Rate ([\d.eE+-]+), It/sec ([\d.]+), Tokens/sec ([\d.]+)`)
)

// Ingest reads benchmark scores and training logs and writes them to InfluxDB.
// At least one of ContentFile, CapabilityFile, or TrainingLog must be set.
func Ingest(influx *InfluxClient, cfg IngestConfig, w io.Writer) error {
	if cfg.ContentFile == "" && cfg.CapabilityFile == "" && cfg.TrainingLog == "" {
		return fmt.Errorf("at least one of --content, --capability, or --training-log is required")
	}
	if cfg.Model == "" {
		return fmt.Errorf("--model is required")
	}
	if cfg.RunID == "" {
		cfg.RunID = cfg.Model
	}
	if cfg.BatchSize <= 0 {
		cfg.BatchSize = 100
	}

	var totalPoints int

	if cfg.ContentFile != "" {
		n, err := ingestContentScores(influx, cfg, w)
		if err != nil {
			return fmt.Errorf("ingest content scores: %w", err)
		}
		totalPoints += n
	}

	if cfg.CapabilityFile != "" {
		n, err := ingestCapabilityScores(influx, cfg, w)
		if err != nil {
			return fmt.Errorf("ingest capability scores: %w", err)
		}
		totalPoints += n
	}

	if cfg.TrainingLog != "" {
		n, err := ingestTrainingLog(influx, cfg, w)
		if err != nil {
			return fmt.Errorf("ingest training log: %w", err)
		}
		totalPoints += n
	}

	fmt.Fprintf(w, "Ingested %d total points into InfluxDB\n", totalPoints)
	return nil
}

// ingestContentScores reads a content scores JSONL file and writes content_score
// and probe_score measurements to InfluxDB.
func ingestContentScores(influx *InfluxClient, cfg IngestConfig, w io.Writer) (int, error) {
	f, err := os.Open(cfg.ContentFile)
	if err != nil {
		return 0, fmt.Errorf("open %s: %w", cfg.ContentFile, err)
	}
	defer f.Close()

	scanner := bufio.NewScanner(f)
	scanner.Buffer(make([]byte, 1024*1024), 1024*1024)

	var lines []string
	var totalPoints int
	lineNum := 0

	for scanner.Scan() {
		lineNum++
		raw := strings.TrimSpace(scanner.Text())
		if raw == "" {
			continue
		}

		var entry contentScoreLine
		if err := json.Unmarshal([]byte(raw), &entry); err != nil {
			return totalPoints, fmt.Errorf("line %d: parse json: %w", lineNum, err)
		}

		label := entry.Label
		iteration := extractIteration(label)
		hasKernel := "false"
		if strings.Contains(strings.ToLower(label), "kernel") || strings.Contains(label, "LEK") {
			hasKernel = "true"
		}
		ts := time.Now().UnixNano()

		// Write aggregate content_score — one point per dimension.
		for dim, val := range entry.Aggregates {
			score, ok := toFloat64(val)
			if !ok {
				continue
			}
			line := fmt.Sprintf(
				"content_score,model=%s,run_id=%s,label=%s,dimension=%s,has_kernel=%s score=%.6f,iteration=%di %d",
				EscapeLp(cfg.Model), EscapeLp(cfg.RunID), EscapeLp(label),
				EscapeLp(dim), hasKernel, score, iteration, ts,
			)
			lines = append(lines, line)
			totalPoints++
		}

		// Write per-probe probe_score — one point per probe per dimension.
		for probeID, probe := range entry.Probes {
			for dim, val := range probe.Scores {
				score, ok := toFloat64(val)
				if !ok {
					continue
				}
				line := fmt.Sprintf(
					"probe_score,model=%s,run_id=%s,label=%s,probe_id=%s,dimension=%s,has_kernel=%s score=%.6f,iteration=%di %d",
					EscapeLp(cfg.Model), EscapeLp(cfg.RunID), EscapeLp(label),
					EscapeLp(probeID), EscapeLp(dim), hasKernel, score, iteration, ts,
				)
				lines = append(lines, line)
				totalPoints++
			}
		}

		// Flush batch if needed.
		if len(lines) >= cfg.BatchSize {
			if err := influx.WriteLp(lines); err != nil {
				return totalPoints, fmt.Errorf("write batch: %w", err)
			}
			lines = lines[:0]
		}
	}

	if err := scanner.Err(); err != nil {
		return totalPoints, fmt.Errorf("scan %s: %w", cfg.ContentFile, err)
	}

	// Flush remaining lines.
	if len(lines) > 0 {
		if err := influx.WriteLp(lines); err != nil {
			return totalPoints, fmt.Errorf("write final batch: %w", err)
		}
	}

	fmt.Fprintf(w, "  content scores: %d points from %d lines\n", totalPoints, lineNum)
	return totalPoints, nil
}

// ingestCapabilityScores reads a capability scores JSONL file and writes
// capability_score measurements to InfluxDB.
func ingestCapabilityScores(influx *InfluxClient, cfg IngestConfig, w io.Writer) (int, error) {
	f, err := os.Open(cfg.CapabilityFile)
	if err != nil {
		return 0, fmt.Errorf("open %s: %w", cfg.CapabilityFile, err)
	}
	defer f.Close()

	scanner := bufio.NewScanner(f)
	scanner.Buffer(make([]byte, 1024*1024), 1024*1024)

	var lines []string
	var totalPoints int
	lineNum := 0

	for scanner.Scan() {
		lineNum++
		raw := strings.TrimSpace(scanner.Text())
		if raw == "" {
			continue
		}

		var entry capabilityScoreLine
		if err := json.Unmarshal([]byte(raw), &entry); err != nil {
			return totalPoints, fmt.Errorf("line %d: parse json: %w", lineNum, err)
		}

		label := entry.Label
		iteration := extractIteration(label)
		ts := time.Now().UnixNano()

		// Overall capability score.
		line := fmt.Sprintf(
			"capability_score,model=%s,run_id=%s,label=%s,category=overall accuracy=%.6f,correct=%di,total=%di,iteration=%di %d",
			EscapeLp(cfg.Model), EscapeLp(cfg.RunID), EscapeLp(label),
			entry.Accuracy, entry.Correct, entry.Total, iteration, ts,
		)
		lines = append(lines, line)
		totalPoints++

		// Per-category breakdown.
		for cat, block := range entry.ByCategory {
			var catAccuracy float64
			if block.Total > 0 {
				catAccuracy = float64(block.Correct) / float64(block.Total)
			}
			line := fmt.Sprintf(
				"capability_score,model=%s,run_id=%s,label=%s,category=%s accuracy=%.6f,correct=%di,total=%di,iteration=%di %d",
				EscapeLp(cfg.Model), EscapeLp(cfg.RunID), EscapeLp(label),
				EscapeLp(cat), catAccuracy, block.Correct, block.Total, iteration, ts,
			)
			lines = append(lines, line)
			totalPoints++
		}

		// Flush batch if needed.
		if len(lines) >= cfg.BatchSize {
			if err := influx.WriteLp(lines); err != nil {
				return totalPoints, fmt.Errorf("write batch: %w", err)
			}
			lines = lines[:0]
		}
	}

	if err := scanner.Err(); err != nil {
		return totalPoints, fmt.Errorf("scan %s: %w", cfg.CapabilityFile, err)
	}

	// Flush remaining lines.
	if len(lines) > 0 {
		if err := influx.WriteLp(lines); err != nil {
			return totalPoints, fmt.Errorf("write final batch: %w", err)
		}
	}

	fmt.Fprintf(w, "  capability scores: %d points from %d lines\n", totalPoints, lineNum)
	return totalPoints, nil
}

// ingestTrainingLog reads an MLX LoRA training log and writes training_loss
// measurements to InfluxDB for both training and validation loss entries.
func ingestTrainingLog(influx *InfluxClient, cfg IngestConfig, w io.Writer) (int, error) {
	f, err := os.Open(cfg.TrainingLog)
	if err != nil {
		return 0, fmt.Errorf("open %s: %w", cfg.TrainingLog, err)
	}
	defer f.Close()

	scanner := bufio.NewScanner(f)
	scanner.Buffer(make([]byte, 1024*1024), 1024*1024)

	var lines []string
	var totalPoints int
	lineNum := 0

	for scanner.Scan() {
		lineNum++
		text := scanner.Text()

		// Try validation loss first (shorter regex, less common).
		if m := reValLoss.FindStringSubmatch(text); m != nil {
			iter, _ := strconv.Atoi(m[1])
			loss, _ := strconv.ParseFloat(m[2], 64)
			ts := time.Now().UnixNano()

			line := fmt.Sprintf(
				"training_loss,model=%s,run_id=%s,loss_type=val loss=%.6f,iteration=%di %d",
				EscapeLp(cfg.Model), EscapeLp(cfg.RunID), loss, iter, ts,
			)
			lines = append(lines, line)
			totalPoints++
		}

		// Try training loss.
		if m := reTrainLoss.FindStringSubmatch(text); m != nil {
			iter, _ := strconv.Atoi(m[1])
			loss, _ := strconv.ParseFloat(m[2], 64)
			lr, _ := strconv.ParseFloat(m[3], 64)
			itPerSec, _ := strconv.ParseFloat(m[4], 64)
			tokPerSec, _ := strconv.ParseFloat(m[5], 64)
			ts := time.Now().UnixNano()

			line := fmt.Sprintf(
				"training_loss,model=%s,run_id=%s,loss_type=train loss=%.6f,iteration=%di,learning_rate=%.10f,it_per_sec=%.4f,tokens_per_sec=%.2f %d",
				EscapeLp(cfg.Model), EscapeLp(cfg.RunID), loss, iter, lr, itPerSec, tokPerSec, ts,
			)
			lines = append(lines, line)
			totalPoints++
		}

		// Flush batch if needed.
		if len(lines) >= cfg.BatchSize {
			if err := influx.WriteLp(lines); err != nil {
				return totalPoints, fmt.Errorf("write batch: %w", err)
			}
			lines = lines[:0]
		}
	}

	if err := scanner.Err(); err != nil {
		return totalPoints, fmt.Errorf("scan %s: %w", cfg.TrainingLog, err)
	}

	// Flush remaining lines.
	if len(lines) > 0 {
		if err := influx.WriteLp(lines); err != nil {
			return totalPoints, fmt.Errorf("write final batch: %w", err)
		}
	}

	fmt.Fprintf(w, "  training log: %d points from %d lines\n", totalPoints, lineNum)
	return totalPoints, nil
}

// extractIteration extracts an iteration number from a label like "model@200".
// Returns 0 if no iteration is found.
func extractIteration(label string) int {
	idx := strings.LastIndex(label, "@")
	if idx < 0 || idx+1 >= len(label) {
		return 0
	}
	n, err := strconv.Atoi(label[idx+1:])
	if err != nil {
		return 0
	}
	return n
}

// toFloat64 converts a JSON-decoded interface{} value to float64.
// Handles float64 (standard json.Unmarshal), json.Number, and string values.
func toFloat64(v interface{}) (float64, bool) {
	switch val := v.(type) {
	case float64:
		return val, true
	case int:
		return float64(val), true
	case int64:
		return float64(val), true
	case json.Number:
		f, err := val.Float64()
		return f, err == nil
	case string:
		f, err := strconv.ParseFloat(val, 64)
		return f, err == nil
	default:
		return 0, false
	}
}
