package collector

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"sort"
	"strings"
	"time"

	"forge.lthn.ai/core/cli/pkg/lab"
)

type InfluxDB struct {
	cfg   *lab.Config
	store *lab.Store
}

func NewInfluxDB(cfg *lab.Config, s *lab.Store) *InfluxDB {
	return &InfluxDB{cfg: cfg, store: s}
}

func (i *InfluxDB) Name() string { return "influxdb" }

func (i *InfluxDB) Collect(ctx context.Context) error {
	if i.cfg.InfluxURL == "" || i.cfg.InfluxToken == "" {
		return nil
	}

	data := lab.BenchmarkData{
		Loss:            make(map[string][]lab.LossPoint),
		Content:         make(map[string][]lab.ContentPoint),
		Capability:      make(map[string][]lab.CapabilityPoint),
		CapabilityJudge: make(map[string][]lab.CapabilityJudgePoint),
		UpdatedAt:       time.Now(),
	}

	// Collect all run identifiers from each measurement.
	runSet := map[string]lab.BenchmarkRun{}

	// Training loss data.
	if rows, err := i.query(ctx, "SELECT run_id, model, iteration, loss, loss_type, learning_rate, iterations_per_sec, tokens_per_sec FROM training_loss ORDER BY run_id, iteration"); err == nil {
		for _, row := range rows {
			rid := jsonStr(row["run_id"])
			mdl := jsonStr(row["model"])
			if rid == "" {
				continue
			}
			runSet[rid] = lab.BenchmarkRun{RunID: rid, Model: mdl, Type: "training"}
			data.Loss[rid] = append(data.Loss[rid], lab.LossPoint{
				Iteration:    jsonInt(row["iteration"]),
				Loss:         jsonFloat(row["loss"]),
				LossType:     jsonStr(row["loss_type"]),
				LearningRate: jsonFloat(row["learning_rate"]),
				TokensPerSec: jsonFloat(row["tokens_per_sec"]),
			})
		}
	}

	// Content scores.
	if rows, err := i.query(ctx, "SELECT run_id, model, label, dimension, score, iteration, has_kernel FROM content_score ORDER BY run_id, iteration, dimension"); err == nil {
		for _, row := range rows {
			rid := jsonStr(row["run_id"])
			mdl := jsonStr(row["model"])
			if rid == "" {
				continue
			}
			if _, ok := runSet[rid]; !ok {
				runSet[rid] = lab.BenchmarkRun{RunID: rid, Model: mdl, Type: "content"}
			}
			hk := jsonStr(row["has_kernel"])
			data.Content[rid] = append(data.Content[rid], lab.ContentPoint{
				Label:     jsonStr(row["label"]),
				Dimension: jsonStr(row["dimension"]),
				Score:     jsonFloat(row["score"]),
				Iteration: jsonInt(row["iteration"]),
				HasKernel: hk == "true" || hk == "True",
			})
		}
	}

	// Capability scores.
	if rows, err := i.query(ctx, "SELECT run_id, model, label, category, accuracy, correct, total, iteration FROM capability_score ORDER BY run_id, iteration, category"); err == nil {
		for _, row := range rows {
			rid := jsonStr(row["run_id"])
			mdl := jsonStr(row["model"])
			if rid == "" {
				continue
			}
			if _, ok := runSet[rid]; !ok {
				runSet[rid] = lab.BenchmarkRun{RunID: rid, Model: mdl, Type: "capability"}
			}
			data.Capability[rid] = append(data.Capability[rid], lab.CapabilityPoint{
				Label:     jsonStr(row["label"]),
				Category:  jsonStr(row["category"]),
				Accuracy:  jsonFloat(row["accuracy"]),
				Correct:   jsonInt(row["correct"]),
				Total:     jsonInt(row["total"]),
				Iteration: jsonInt(row["iteration"]),
			})
		}
	}

	// Capability judge scores (0-10 per probe).
	if rows, err := i.query(ctx, "SELECT run_id, model, label, probe_id, category, reasoning, correctness, clarity, avg, iteration FROM capability_judge ORDER BY run_id, iteration, probe_id"); err == nil {
		for _, row := range rows {
			rid := jsonStr(row["run_id"])
			if rid == "" {
				continue
			}
			data.CapabilityJudge[rid] = append(data.CapabilityJudge[rid], lab.CapabilityJudgePoint{
				Label:       jsonStr(row["label"]),
				ProbeID:     jsonStr(row["probe_id"]),
				Category:    jsonStr(row["category"]),
				Reasoning:   jsonFloat(row["reasoning"]),
				Correctness: jsonFloat(row["correctness"]),
				Clarity:     jsonFloat(row["clarity"]),
				Avg:         jsonFloat(row["avg"]),
				Iteration:   jsonInt(row["iteration"]),
			})
		}
	}

	// Build sorted runs list.
	for _, r := range runSet {
		data.Runs = append(data.Runs, r)
	}
	sort.Slice(data.Runs, func(i, j int) bool {
		return data.Runs[i].Model < data.Runs[j].Model || (data.Runs[i].Model == data.Runs[j].Model && data.Runs[i].RunID < data.Runs[j].RunID)
	})

	i.store.SetBenchmarks(data)

	// Live training run statuses.
	var runStatuses []lab.TrainingRunStatus
	if rows, err := i.query(ctx, "SELECT model, run_id, status, iteration, total_iters, pct FROM training_status ORDER BY time DESC LIMIT 50"); err == nil {
		// Deduplicate: keep only the latest status per run_id.
		seen := map[string]bool{}
		for _, row := range rows {
			rid := jsonStr(row["run_id"])
			if rid == "" || seen[rid] {
				continue
			}
			seen[rid] = true
			rs := lab.TrainingRunStatus{
				Model:      jsonStr(row["model"]),
				RunID:      rid,
				Status:     jsonStr(row["status"]),
				Iteration:  jsonInt(row["iteration"]),
				TotalIters: jsonInt(row["total_iters"]),
				Pct:        jsonFloat(row["pct"]),
			}
			// Find latest loss for this run from already-collected data.
			if lossPoints, ok := data.Loss[rid]; ok {
				for j := len(lossPoints) - 1; j >= 0; j-- {
					if lossPoints[j].LossType == "train" && rs.LastLoss == 0 {
						rs.LastLoss = lossPoints[j].Loss
						rs.TokensSec = lossPoints[j].TokensPerSec
					}
					if lossPoints[j].LossType == "val" && rs.ValLoss == 0 {
						rs.ValLoss = lossPoints[j].Loss
					}
					if rs.LastLoss > 0 && rs.ValLoss > 0 {
						break
					}
				}
			}
			runStatuses = append(runStatuses, rs)
		}
	}
	i.store.SetTrainingRuns(runStatuses)

	// Golden set data explorer — query gold_gen (real-time per-generation records).
	gs := lab.GoldenSetSummary{TargetTotal: 15000, UpdatedAt: time.Now()}

	// Try real-time gold_gen first (populated by lem_generate.py directly).
	if rows, err := i.query(ctx, "SELECT count(DISTINCT i) AS total, count(DISTINCT d) AS domains, count(DISTINCT v) AS voices, avg(gen_time) AS avg_t, avg(chars) AS avg_c FROM gold_gen"); err == nil && len(rows) > 0 {
		r := rows[0]
		total := jsonInt(r["total"])
		if total > 0 {
			gs.Available = true
			gs.TotalExamples = total
			gs.Domains = jsonInt(r["domains"])
			gs.Voices = jsonInt(r["voices"])
			gs.AvgGenTime = jsonFloat(r["avg_t"])
			gs.AvgResponseChars = jsonFloat(r["avg_c"])
			gs.CompletionPct = float64(total) / float64(gs.TargetTotal) * 100
		}
	}

	// Fallback to pipeline.py metrics if gold_gen isn't populated.
	if !gs.Available {
		if rows, err := i.query(ctx, "SELECT total_examples, domains, voices, avg_gen_time, avg_response_chars, completion_pct FROM golden_set_stats ORDER BY time DESC LIMIT 1"); err == nil && len(rows) > 0 {
			r := rows[0]
			gs.Available = true
			gs.TotalExamples = jsonInt(r["total_examples"])
			gs.Domains = jsonInt(r["domains"])
			gs.Voices = jsonInt(r["voices"])
			gs.AvgGenTime = jsonFloat(r["avg_gen_time"])
			gs.AvgResponseChars = jsonFloat(r["avg_response_chars"])
			gs.CompletionPct = jsonFloat(r["completion_pct"])
		}
	}

	if gs.Available {
		// Per-domain from gold_gen.
		if rows, err := i.query(ctx, "SELECT d, count(DISTINCT i) AS n, avg(gen_time) AS avg_t FROM gold_gen GROUP BY d ORDER BY n DESC"); err == nil && len(rows) > 0 {
			for _, r := range rows {
				gs.DomainStats = append(gs.DomainStats, lab.DomainStat{
					Domain:     jsonStr(r["d"]),
					Count:      jsonInt(r["n"]),
					AvgGenTime: jsonFloat(r["avg_t"]),
				})
			}
		}
		// Fallback to pipeline stats.
		if len(gs.DomainStats) == 0 {
			if rows, err := i.query(ctx, "SELECT DISTINCT domain, count, avg_gen_time FROM golden_set_domain ORDER BY count DESC"); err == nil {
				for _, r := range rows {
					gs.DomainStats = append(gs.DomainStats, lab.DomainStat{
						Domain:     jsonStr(r["domain"]),
						Count:      jsonInt(r["count"]),
						AvgGenTime: jsonFloat(r["avg_gen_time"]),
					})
				}
			}
		}

		// Per-voice from gold_gen.
		if rows, err := i.query(ctx, "SELECT v, count(DISTINCT i) AS n, avg(chars) AS avg_c, avg(gen_time) AS avg_t FROM gold_gen GROUP BY v ORDER BY n DESC"); err == nil && len(rows) > 0 {
			for _, r := range rows {
				gs.VoiceStats = append(gs.VoiceStats, lab.VoiceStat{
					Voice:      jsonStr(r["v"]),
					Count:      jsonInt(r["n"]),
					AvgChars:   jsonFloat(r["avg_c"]),
					AvgGenTime: jsonFloat(r["avg_t"]),
				})
			}
		}
		// Fallback.
		if len(gs.VoiceStats) == 0 {
			if rows, err := i.query(ctx, "SELECT DISTINCT voice, count, avg_chars, avg_gen_time FROM golden_set_voice ORDER BY count DESC"); err == nil {
				for _, r := range rows {
					gs.VoiceStats = append(gs.VoiceStats, lab.VoiceStat{
						Voice:      jsonStr(r["voice"]),
						Count:      jsonInt(r["count"]),
						AvgChars:   jsonFloat(r["avg_chars"]),
						AvgGenTime: jsonFloat(r["avg_gen_time"]),
					})
				}
			}
		}
	}
	// Worker activity.
	if rows, err := i.query(ctx, "SELECT w, count(DISTINCT i) AS n, max(time) AS last_seen FROM gold_gen GROUP BY w ORDER BY n DESC"); err == nil {
		for _, r := range rows {
			gs.Workers = append(gs.Workers, lab.WorkerStat{
				Worker: jsonStr(r["w"]),
				Count:  jsonInt(r["n"]),
			})
		}
	}

	i.store.SetGoldenSet(gs)

	// Dataset stats (from DuckDB, pushed as dataset_stats measurement).
	ds := lab.DatasetSummary{UpdatedAt: time.Now()}
	if rows, err := i.query(ctx, "SELECT table, rows FROM dataset_stats ORDER BY rows DESC"); err == nil && len(rows) > 0 {
		ds.Available = true
		for _, r := range rows {
			ds.Tables = append(ds.Tables, lab.DatasetTable{
				Name: jsonStr(r["table"]),
				Rows: jsonInt(r["rows"]),
			})
		}
	}
	i.store.SetDataset(ds)

	i.store.SetError("influxdb", nil)
	return nil
}

func (i *InfluxDB) query(ctx context.Context, sql string) ([]map[string]any, error) {
	ctx, cancel := context.WithTimeout(ctx, 10*time.Second)
	defer cancel()

	body := fmt.Sprintf(`{"db":%q,"q":%q}`, i.cfg.InfluxDB, sql)
	req, err := http.NewRequestWithContext(ctx, "POST", i.cfg.InfluxURL+"/api/v3/query_sql", strings.NewReader(body))
	if err != nil {
		return nil, err
	}
	req.Header.Set("Authorization", "Bearer "+i.cfg.InfluxToken)
	req.Header.Set("Content-Type", "application/json")

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		i.store.SetError("influxdb", err)
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		err := fmt.Errorf("influxdb query returned %d", resp.StatusCode)
		i.store.SetError("influxdb", err)
		return nil, err
	}

	var rows []map[string]any
	if err := json.NewDecoder(resp.Body).Decode(&rows); err != nil {
		return nil, err
	}
	return rows, nil
}

// JSON value helpers — InfluxDB 3 returns typed JSON values.

func jsonStr(v any) string {
	if v == nil {
		return ""
	}
	if s, ok := v.(string); ok {
		return s
	}
	return fmt.Sprintf("%v", v)
}

func jsonFloat(v any) float64 {
	if v == nil {
		return 0
	}
	switch n := v.(type) {
	case float64:
		return n
	case json.Number:
		f, _ := n.Float64()
		return f
	}
	return 0
}

func jsonInt(v any) int {
	if v == nil {
		return 0
	}
	switch n := v.(type) {
	case float64:
		return int(n)
	case json.Number:
		i, _ := n.Int64()
		return int(i)
	}
	return 0
}
