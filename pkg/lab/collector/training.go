package collector

import (
	"bufio"
	"context"
	"encoding/json"
	"net/http"
	"os"
	"path/filepath"
	"time"

	"forge.lthn.ai/core/go/pkg/lab"
)

type Training struct {
	cfg   *lab.Config
	store *lab.Store
}

func NewTraining(cfg *lab.Config, s *lab.Store) *Training {
	return &Training{cfg: cfg, store: s}
}

func (t *Training) Name() string { return "training" }

func (t *Training) Collect(ctx context.Context) error {
	summary := lab.TrainingSummary{
		GoldTarget: 15000,
	}

	// Fetch from M3 lab-helper API
	if t.cfg.M3APIURL != "" {
		t.fetchM3API(ctx, &summary)
	}

	// Parse local intercept JSONL files
	interceptDir := t.cfg.TrainingDataDir
	if interceptDir != "" {
		count, lastTime := countJSONLFiles(filepath.Join(interceptDir, "command-intercepts"))
		summary.InterceptCount = count
		summary.LastIntercept = lastTime
	}

	// Count QA sessions
	sessDir := filepath.Join(t.cfg.TrainingDataDir, "qa-epic-verification", "sessions")
	if entries, err := os.ReadDir(sessDir); err == nil {
		summary.SessionCount = len(entries)
	}

	t.store.SetTraining(summary)
	t.store.SetError("training", nil)
	return nil
}

type m3TrainingResponse struct {
	GoldGenerated int      `json:"gold_generated"`
	GoldTarget    int      `json:"gold_target"`
	GoldPercent   float64  `json:"gold_percent"`
	SeedsComplete int      `json:"seeds_complete"`
	GGUFCount     int      `json:"gguf_count"`
	GGUFFiles     []string `json:"gguf_files"`
	AdapterCount  int      `json:"adapter_count"`
}

func (t *Training) fetchM3API(ctx context.Context, summary *lab.TrainingSummary) {
	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	req, err := http.NewRequestWithContext(ctx, "GET", t.cfg.M3APIURL+"/api/training", nil)
	if err != nil {
		return
	}

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		t.store.SetError("m3-api", err)
		return
	}
	defer resp.Body.Close()

	var data m3TrainingResponse
	if err := json.NewDecoder(resp.Body).Decode(&data); err != nil {
		return
	}

	summary.GoldGenerated = data.GoldGenerated
	summary.GoldAvailable = true
	summary.GoldPercent = data.GoldPercent
	summary.GGUFCount = data.GGUFCount
	summary.GGUFFiles = data.GGUFFiles
	summary.AdapterCount = data.AdapterCount
	t.store.SetError("m3-api", nil)
}

func countJSONLFiles(dir string) (int, time.Time) {
	var total int
	var lastTime time.Time

	files, err := filepath.Glob(filepath.Join(dir, "*.jsonl"))
	if err != nil {
		return 0, lastTime
	}

	for _, f := range files {
		file, err := os.Open(f)
		if err != nil {
			continue
		}
		scanner := bufio.NewScanner(file)
		for scanner.Scan() {
			total++
			var ev struct {
				Timestamp time.Time `json:"timestamp"`
			}
			if json.Unmarshal(scanner.Bytes(), &ev) == nil && ev.Timestamp.After(lastTime) {
				lastTime = ev.Timestamp
			}
		}
		file.Close()
	}

	return total, lastTime
}
