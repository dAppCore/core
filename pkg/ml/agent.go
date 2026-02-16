package ml

import (
	"bufio"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"sort"
	"strings"
	"time"
)

// AgentConfig holds scoring agent configuration.
type AgentConfig struct {
	M3Host        string
	M3User        string
	M3SSHKey      string
	M3AdapterBase string
	InfluxURL     string
	InfluxDB      string
	DBPath        string
	APIURL        string
	JudgeURL      string
	JudgeModel    string
	Model         string
	BaseModel     string
	PollInterval  int
	WorkDir       string
	Filter        string
	Force         bool
	OneShot       bool
	DryRun        bool
}

// Checkpoint represents a discovered adapter checkpoint on M3.
type Checkpoint struct {
	RemoteDir string
	Filename  string
	Dirname   string
	Iteration int
	ModelTag  string
	Label     string
	RunID     string
}

// ProbeResult holds the result of running all probes against a checkpoint.
type ProbeResult struct {
	Accuracy   float64                      `json:"accuracy"`
	Correct    int                          `json:"correct"`
	Total      int                          `json:"total"`
	ByCategory map[string]CategoryResult    `json:"by_category"`
	Probes     map[string]SingleProbeResult `json:"probes"`
}

// CategoryResult holds pass/fail counts for a probe category.
type CategoryResult struct {
	Correct int `json:"correct"`
	Total   int `json:"total"`
}

// SingleProbeResult holds the result of a single probe.
type SingleProbeResult struct {
	Passed   bool   `json:"passed"`
	Response string `json:"response"`
}

// bufferEntry is a JSONL-buffered result for when InfluxDB is down.
type bufferEntry struct {
	Checkpoint Checkpoint  `json:"checkpoint"`
	Results    ProbeResult `json:"results"`
	Timestamp  string      `json:"timestamp"`
}

// BaseModelMap maps model tags to their HuggingFace/local model paths.
var BaseModelMap = map[string]string{
	"gemma-3-1b":  "mlx-community/gemma-3-1b-it-4bit",
	"gemma-3-4b":  "mlx-community/gemma-3-4b-it-4bit",
	"gemma-3-12b": "mlx-community/gemma-3-12b-it-4bit",
	"gemma-3-27b": "mlx-community/gemma-3-27b-it-qat-4bit",
	"gpt-oss-20b": "/Volumes/Data/lem/models/gpt-oss-20b-mlx",
}

// ModelFamilies identifies known model families from adapter directory names.
var ModelFamilies = []struct {
	DirPrefix string
	Tag       string
	Short     string
}{
	{"deepseek-r1-7b", "deepseek-r1-7b", "R1"},
	{"27b-", "gemma-3-27b", "G27"},
	{"27b", "gemma-3-27b", "G27"},
	{"15k/gemma-3-27b", "gemma-3-27b", "G27"},
	{"15k/gemma-3-12b", "gemma-3-12b", "G12"},
	{"15k/gemma-3-1b", "gemma-3-1b", "G1"},
	{"12b", "gemma-3-12b", "G12"},
	{"1b-", "gemma-3-1b", "G1"},
	{"1b", "gemma-3-1b", "G1"},
	{"4b", "gemma-3-4b", "G4"},
	{"vi-12b", "gemma-3-12b", "Vi12"},
	{"vi", "gemma-3-1b", "Vi1"},
	{"gpt-oss", "gpt-oss-20b", "GPT"},
	{"lem-gpt-oss", "gpt-oss-20b", "LGPT"},
	{"bench-1b", "gemma-3-1b", "B1"},
	{"book", "gemma-3-27b", "Book"},
	{"cross", "gemma-3-12b", "Cross"},
}

// AdapterMeta maps an adapter directory name to (model_tag, label_prefix, run_id_stem).
func AdapterMeta(dirname string) (string, string, string) {
	name := strings.TrimPrefix(dirname, "adapters-")

	for _, fam := range ModelFamilies {
		if strings.HasPrefix(name, fam.DirPrefix) {
			variant := strings.TrimPrefix(name, fam.DirPrefix)
			variant = strings.TrimLeft(variant, "-")
			if variant == "" {
				variant = "base"
			}
			short := fam.Short + "-" + variant
			if variant == "base" {
				short = fam.Short
			}
			stem := strings.ReplaceAll(name, "/", "-")
			return fam.Tag, short, stem
		}
	}

	short := name
	if len(short) > 10 {
		short = short[:10]
	}
	return name, short, name
}

// RunAgentLoop is the main scoring agent loop.
func RunAgentLoop(cfg *AgentConfig) {
	log.Println(strings.Repeat("=", 60))
	log.Println("ROCm Scoring Agent — Go Edition")
	log.Printf("M3: %s@%s", cfg.M3User, cfg.M3Host)
	log.Printf("Inference API: %s", cfg.APIURL)
	log.Printf("Judge API: %s (%s)", cfg.JudgeURL, cfg.JudgeModel)
	log.Printf("InfluxDB: %s/%s", cfg.InfluxURL, cfg.InfluxDB)
	if cfg.DBPath != "" {
		log.Printf("DuckDB: %s", cfg.DBPath)
	}
	log.Printf("Poll interval: %ds", cfg.PollInterval)
	log.Println(strings.Repeat("=", 60))

	influx := NewInfluxClient(cfg.InfluxURL, cfg.InfluxDB)
	os.MkdirAll(cfg.WorkDir, 0755)

	for {
		ReplayInfluxBuffer(cfg.WorkDir, influx)

		log.Println("Discovering checkpoints on M3...")
		checkpoints, err := DiscoverCheckpoints(cfg)
		if err != nil {
			log.Printf("Discovery failed: %v", err)
			if cfg.OneShot {
				return
			}
			time.Sleep(time.Duration(cfg.PollInterval) * time.Second)
			continue
		}
		log.Printf("Found %d total checkpoints", len(checkpoints))

		var unscored []Checkpoint
		if cfg.Force {
			unscored = checkpoints
			log.Printf("Force mode: scoring all %d checkpoints", len(unscored))
		} else {
			scored, err := GetScoredLabels(influx)
			if err != nil {
				log.Printf("InfluxDB query failed: %v", err)
			}
			log.Printf("Already scored: %d (run_id, label) pairs", len(scored))
			unscored = FindUnscored(checkpoints, scored)
			log.Printf("Unscored: %d checkpoints", len(unscored))
		}

		if len(unscored) == 0 {
			log.Printf("Nothing to score. Sleeping %ds...", cfg.PollInterval)
			if cfg.OneShot {
				return
			}
			time.Sleep(time.Duration(cfg.PollInterval) * time.Second)
			continue
		}

		targets := unscored
		if !cfg.Force {
			targets = unscored[:1]
		}

		for i, target := range targets {
			log.Printf("Grabbed: %s (%s) [%d/%d]", target.Label, target.Dirname, i+1, len(targets))

			if cfg.DryRun {
				log.Printf("[DRY RUN] Would process: %s/%s", target.Dirname, target.Filename)
				continue
			}

			if err := ProcessOne(cfg, influx, target); err != nil {
				log.Printf("Error processing %s: %v", target.Label, err)
			}
			time.Sleep(5 * time.Second)
		}

		if cfg.DryRun || cfg.OneShot {
			return
		}
	}
}

// DiscoverCheckpoints lists all adapter directories and checkpoint files on M3 via SSH.
func DiscoverCheckpoints(cfg *AgentConfig) ([]Checkpoint, error) {
	pattern := "adapters-*"
	if cfg.Filter != "" {
		pattern = "adapters-" + cfg.Filter + "*"
	}
	out, err := SSHCommand(cfg, fmt.Sprintf("ls -d %s/%s 2>/dev/null", cfg.M3AdapterBase, pattern))
	if err != nil {
		return nil, fmt.Errorf("list adapter dirs: %w", err)
	}

	var checkpoints []Checkpoint
	iterRe := regexp.MustCompile(`(\d+)`)

	var adapterDirs []string
	for _, dirpath := range strings.Split(strings.TrimSpace(out), "\n") {
		if dirpath == "" {
			continue
		}
		subOut, subErr := SSHCommand(cfg, fmt.Sprintf("ls -d %s/gemma-3-* 2>/dev/null", dirpath))
		if subErr == nil && strings.TrimSpace(subOut) != "" {
			for _, sub := range strings.Split(strings.TrimSpace(subOut), "\n") {
				if sub != "" {
					adapterDirs = append(adapterDirs, sub)
				}
			}
		} else {
			adapterDirs = append(adapterDirs, dirpath)
		}
	}

	for _, dirpath := range adapterDirs {
		dirname := strings.TrimPrefix(dirpath, cfg.M3AdapterBase+"/")

		filesOut, err := SSHCommand(cfg, fmt.Sprintf("ls %s/*_adapters.safetensors 2>/dev/null", dirpath))
		if err != nil {
			continue
		}

		for _, fp := range strings.Split(strings.TrimSpace(filesOut), "\n") {
			if fp == "" {
				continue
			}
			filename := fileBase(fp)

			match := iterRe.FindStringSubmatch(filename)
			if len(match) < 2 {
				continue
			}
			iteration := 0
			fmt.Sscanf(match[1], "%d", &iteration)

			modelTag, labelPrefix, stem := AdapterMeta(dirname)
			label := fmt.Sprintf("%s @%s", labelPrefix, match[1])
			runID := fmt.Sprintf("%s-capability-auto", stem)

			checkpoints = append(checkpoints, Checkpoint{
				RemoteDir: dirpath,
				Filename:  filename,
				Dirname:   dirname,
				Iteration: iteration,
				ModelTag:  modelTag,
				Label:     label,
				RunID:     runID,
			})
		}
	}

	return checkpoints, nil
}

// GetScoredLabels returns all (run_id, label) pairs already scored in InfluxDB.
func GetScoredLabels(influx *InfluxClient) (map[[2]string]bool, error) {
	rows, err := influx.QuerySQL("SELECT DISTINCT run_id, label FROM capability_score")
	if err != nil {
		return nil, err
	}

	scored := make(map[[2]string]bool)
	for _, row := range rows {
		runID, _ := row["run_id"].(string)
		label, _ := row["label"].(string)
		if runID != "" && label != "" {
			scored[[2]string{runID, label}] = true
		}
	}
	return scored, nil
}

// FindUnscored filters checkpoints to only unscored ones, sorted by (dirname, iteration).
func FindUnscored(checkpoints []Checkpoint, scored map[[2]string]bool) []Checkpoint {
	var unscored []Checkpoint
	for _, c := range checkpoints {
		if !scored[[2]string{c.RunID, c.Label}] {
			unscored = append(unscored, c)
		}
	}
	sort.Slice(unscored, func(i, j int) bool {
		if unscored[i].Dirname != unscored[j].Dirname {
			return unscored[i].Dirname < unscored[j].Dirname
		}
		return unscored[i].Iteration < unscored[j].Iteration
	})
	return unscored
}

// isMLXNative returns true if this model can be served directly on M3 via
// mlx_lm.server with --adapter, avoiding the MLX→PEFT conversion step.
func isMLXNative(modelTag string) bool {
	return strings.HasPrefix(modelTag, "gemma-3-") || strings.HasPrefix(modelTag, "gpt-oss")
}

// ProcessOne fetches, converts, scores, and pushes one checkpoint.
func ProcessOne(cfg *AgentConfig, influx *InfluxClient, cp Checkpoint) error {
	log.Println(strings.Repeat("=", 60))
	log.Printf("Processing: %s / %s [%s]", cp.Dirname, cp.Filename, cp.ModelTag)
	log.Println(strings.Repeat("=", 60))

	if isMLXNative(cp.ModelTag) {
		return processMLXNative(cfg, influx, cp)
	}
	return processWithConversion(cfg, influx, cp)
}

// processMLXNative scores a checkpoint using Ollama on M3.
func processMLXNative(cfg *AgentConfig, influx *InfluxClient, cp Checkpoint) error {
	ollamaBase, ok := OllamaBaseModelMap[cp.ModelTag]
	if !ok {
		return fmt.Errorf("unknown Ollama model for tag %s", cp.ModelTag)
	}
	hfBase := HFBaseModelMap[cp.ModelTag]
	if hfBase == "" {
		hfBase = ollamaBase
	}

	tempModel := fmt.Sprintf("lem-%s-%d", cp.ModelTag, cp.Iteration)
	localAdapterDir := filepath.Join(cfg.WorkDir, "adapter-"+cp.Dirname)
	peftDir := filepath.Join(cfg.WorkDir, "peft-"+cp.Dirname)

	os.MkdirAll(localAdapterDir, 0755)

	defer func() {
		os.RemoveAll(localAdapterDir)
		os.RemoveAll(peftDir)
		OllamaDeleteModel(cfg.JudgeURL, tempModel)
	}()

	log.Printf("Fetching adapter from M3 (%s)...", cp.Filename)
	remoteSF := fmt.Sprintf("%s/%s", cp.RemoteDir, cp.Filename)
	remoteCfg := fmt.Sprintf("%s/adapter_config.json", cp.RemoteDir)
	localSF := filepath.Join(localAdapterDir, cp.Filename)
	localCfg := filepath.Join(localAdapterDir, "adapter_config.json")

	if err := SCPFrom(cfg, remoteSF, localSF); err != nil {
		return fmt.Errorf("scp safetensors: %w", err)
	}
	if err := SCPFrom(cfg, remoteCfg, localCfg); err != nil {
		return fmt.Errorf("scp config: %w", err)
	}

	log.Println("Converting MLX → PEFT format...")
	if err := ConvertMLXtoPEFT(localSF, localCfg, peftDir, hfBase); err != nil {
		return fmt.Errorf("convert adapter: %w", err)
	}

	log.Printf("Creating Ollama model %s (base: %s)...", tempModel, ollamaBase)
	if err := OllamaCreateModel(cfg.JudgeURL, tempModel, ollamaBase, peftDir); err != nil {
		return fmt.Errorf("ollama create: %w", err)
	}
	log.Printf("Ollama model %s ready", tempModel)

	ctx := context.Background()
	probeBackend := NewHTTPBackend(cfg.JudgeURL, tempModel)

	const baseTS int64 = 1739577600
	results, fullResponses := RunCapabilityProbesFull(ctx, probeBackend, func(probeID, category string, passed bool, response string, correct, total int) {
		passedInt := 0
		if passed {
			passedInt = 1
		}
		ts := (baseTS + int64(cp.Iteration)*1000 + int64(total+100)) * 1_000_000_000
		line := fmt.Sprintf(
			"probe_score,model=%s,run_id=%s,label=%s,probe_id=%s passed=%di,iteration=%di %d",
			EscapeLp(cp.ModelTag), EscapeLp(cp.RunID), EscapeLp(cp.Label), EscapeLp(probeID),
			passedInt, cp.Iteration, ts,
		)
		if err := influx.WriteLp([]string{line}); err != nil {
			log.Printf("  [%s] InfluxDB stream failed: %v", probeID, err)
		}
	})

	log.Printf("Capability: %s -- %.1f%% (%d/%d)",
		cp.Label, results.Accuracy, results.Correct, results.Total)

	if err := PushCapabilitySummary(influx, cp, results); err != nil {
		log.Printf("InfluxDB summary push failed, buffering: %v", err)
		BufferInfluxResult(cfg.WorkDir, cp, results)
	}
	PushCapabilityResultsDB(cfg.DBPath, cp, results)

	judgeBackend := NewHTTPBackend(cfg.JudgeURL, cfg.JudgeModel)
	judge := NewJudge(judgeBackend)

	log.Println("Judging 23 capability responses (0-10 quality scoring)...")
	ScoreCapabilityAndPush(ctx, judge, influx, cp, fullResponses)

	log.Println("Running 6 content probes (0-10 judge scoring)...")
	contentResponses := RunContentProbesViaAPI(ctx, probeBackend)
	if len(contentResponses) > 0 {
		contentRunID := strings.Replace(cp.RunID, "-capability-", "-content-", 1)
		ScoreContentAndPush(ctx, judge, influx, cp, contentRunID, contentResponses)
	}

	return nil
}

// processWithConversion fetches adapter locally, converts MLX→PEFT, and scores.
func processWithConversion(cfg *AgentConfig, influx *InfluxClient, cp Checkpoint) error {
	localAdapterDir := filepath.Join(cfg.WorkDir, cp.Dirname)
	os.MkdirAll(localAdapterDir, 0755)

	localSF := filepath.Join(localAdapterDir, cp.Filename)
	localCfg := filepath.Join(localAdapterDir, "adapter_config.json")

	defer func() {
		os.Remove(localSF)
		os.Remove(localCfg)
		peftDir := filepath.Join(cfg.WorkDir, fmt.Sprintf("peft_%07d", cp.Iteration))
		os.RemoveAll(peftDir)
	}()

	log.Println("Fetching adapter from M3...")
	remoteSF := fmt.Sprintf("%s/%s", cp.RemoteDir, cp.Filename)
	remoteCfg := fmt.Sprintf("%s/adapter_config.json", cp.RemoteDir)

	if err := SCPFrom(cfg, remoteSF, localSF); err != nil {
		return fmt.Errorf("scp safetensors: %w", err)
	}
	if err := SCPFrom(cfg, remoteCfg, localCfg); err != nil {
		return fmt.Errorf("scp config: %w", err)
	}

	log.Println("Converting MLX to PEFT format...")
	peftDir := filepath.Join(cfg.WorkDir, fmt.Sprintf("peft_%07d", cp.Iteration))
	if err := ConvertMLXtoPEFT(localSF, localCfg, peftDir, cfg.BaseModel); err != nil {
		return fmt.Errorf("convert adapter: %w", err)
	}

	log.Println("Running 23 capability probes...")
	ctx := context.Background()
	modelName := cfg.Model
	if modelName == "" {
		modelName = cp.ModelTag
	}
	backend := NewHTTPBackend(cfg.APIURL, modelName)

	results := RunCapabilityProbes(ctx, backend)

	log.Printf("Result: %s -- %.1f%% (%d/%d)",
		cp.Label, results.Accuracy, results.Correct, results.Total)

	if err := PushCapabilityResults(influx, cp, results); err != nil {
		log.Printf("InfluxDB push failed, buffering: %v", err)
		BufferInfluxResult(cfg.WorkDir, cp, results)
	}
	PushCapabilityResultsDB(cfg.DBPath, cp, results)

	return nil
}

// ProbeCallback is called after each probe completes for real-time streaming.
type ProbeCallback func(probeID, category string, passed bool, response string, correct, total int)

// RunCapabilityProbes runs all 23 probes against a backend.
func RunCapabilityProbes(ctx context.Context, backend Backend) ProbeResult {
	results := ProbeResult{
		ByCategory: make(map[string]CategoryResult),
		Probes:     make(map[string]SingleProbeResult),
	}

	correct := 0
	total := 0

	for _, probe := range CapabilityProbes {
		response, err := backend.Generate(ctx, probe.Prompt, GenOpts{Temperature: 0.1, MaxTokens: 500})
		if err != nil {
			log.Printf("  [%s] ERROR: %v", probe.ID, err)
			results.Probes[probe.ID] = SingleProbeResult{Passed: false, Response: err.Error()}
			total++
			cat := results.ByCategory[probe.Category]
			cat.Total++
			results.ByCategory[probe.Category] = cat
			continue
		}

		clean := StripThinkBlocks(response)
		passed := probe.Check(clean)
		total++
		if passed {
			correct++
		}

		cat := results.ByCategory[probe.Category]
		cat.Total++
		if passed {
			cat.Correct++
		}
		results.ByCategory[probe.Category] = cat

		stored := clean
		if len(stored) > 300 {
			stored = stored[:300]
		}
		results.Probes[probe.ID] = SingleProbeResult{Passed: passed, Response: stored}

		status := "FAIL"
		if passed {
			status = "PASS"
		}
		log.Printf("  [%s] %s (expected: %s)", probe.ID, status, probe.Answer)
	}

	if total > 0 {
		results.Accuracy = float64(correct) / float64(total) * 100
	}
	results.Correct = correct
	results.Total = total

	return results
}

// CapResponseEntry holds a capability probe response with its metadata for judge scoring.
type CapResponseEntry struct {
	ProbeID  string
	Category string
	Prompt   string
	Answer   string
	Response string
	Passed   bool
}

// RunCapabilityProbesFull runs all probes via a backend and returns both
// aggregate results and full responses for judge scoring.
func RunCapabilityProbesFull(ctx context.Context, backend Backend, onProbe ProbeCallback) (ProbeResult, []CapResponseEntry) {
	results := ProbeResult{
		ByCategory: make(map[string]CategoryResult),
		Probes:     make(map[string]SingleProbeResult),
	}
	var fullResponses []CapResponseEntry

	correct := 0
	total := 0

	for _, probe := range CapabilityProbes {
		response, err := backend.Generate(ctx, probe.Prompt, GenOpts{Temperature: 0.1, MaxTokens: 500})
		if err != nil {
			log.Printf("  [%s] ERROR: %v", probe.ID, err)
			response = fmt.Sprintf("ERROR: %v", err)
		}

		clean := StripThinkBlocks(response)
		passed := probe.Check(clean)
		total++
		if passed {
			correct++
		}

		cat := results.ByCategory[probe.Category]
		cat.Total++
		if passed {
			cat.Correct++
		}
		results.ByCategory[probe.Category] = cat

		stored := clean
		if len(stored) > 300 {
			stored = stored[:300]
		}
		results.Probes[probe.ID] = SingleProbeResult{Passed: passed, Response: stored}

		fullResponses = append(fullResponses, CapResponseEntry{
			ProbeID:  probe.ID,
			Category: probe.Category,
			Prompt:   probe.Prompt,
			Answer:   probe.Answer,
			Response: clean,
			Passed:   passed,
		})

		status := "FAIL"
		if passed {
			status = "PASS"
		}
		log.Printf("  [%s] %s (expected: %s)", probe.ID, status, probe.Answer)

		if onProbe != nil {
			onProbe(probe.ID, probe.Category, passed, stored, correct, total)
		}
	}

	if total > 0 {
		results.Accuracy = float64(correct) / float64(total) * 100
	}
	results.Correct = correct
	results.Total = total

	return results, fullResponses
}

// ContentResponse holds a content probe response for later judging.
type ContentResponse struct {
	Probe    ContentProbe
	Response string
}

// RunContentProbesViaAPI runs content probes via a backend.
func RunContentProbesViaAPI(ctx context.Context, backend Backend) []ContentResponse {
	var responses []ContentResponse

	for _, probe := range ContentProbes {
		reply, err := backend.Generate(ctx, probe.Prompt, GenOpts{Temperature: 0.7, MaxTokens: 1000})
		if err != nil {
			log.Printf("  [content:%s] ERROR: %v", probe.ID, err)
			continue
		}

		reply = StripThinkBlocks(reply)
		log.Printf("  [content:%s] got %d chars", probe.ID, len(reply))

		responses = append(responses, ContentResponse{
			Probe:    probe,
			Response: reply,
		})
	}

	return responses
}

// RunContentProbesViaRunner sends content probes through an SSH probe runner.
func RunContentProbesViaRunner(stdin io.WriteCloser, scanner *bufio.Scanner) []ContentResponse {
	var responses []ContentResponse

	for _, probe := range ContentProbes {
		req := map[string]interface{}{
			"prompt":     probe.Prompt,
			"max_tokens": 1000,
			"temp":       0.7,
		}
		reqJSON, _ := json.Marshal(req)
		fmt.Fprintf(stdin, "%s\n", reqJSON)

		var response string
		if scanner.Scan() {
			var resp probeRunnerResponse
			if err := json.Unmarshal(scanner.Bytes(), &resp); err != nil {
				log.Printf("  [content:%s] parse error: %v", probe.ID, err)
				continue
			} else if resp.Error != "" {
				log.Printf("  [content:%s] ERROR: %s", probe.ID, resp.Error)
				continue
			} else {
				response = resp.Response
			}
		} else {
			log.Printf("  [content:%s] no response from runner", probe.ID)
			continue
		}

		response = StripThinkBlocks(response)
		log.Printf("  [content:%s] got %d chars", probe.ID, len(response))

		responses = append(responses, ContentResponse{
			Probe:    probe,
			Response: response,
		})
	}

	return responses
}

// probeRunnerResponse is the JSON response from the Python probe runner.
type probeRunnerResponse struct {
	Response string  `json:"response"`
	Error    string  `json:"error"`
	Elapsed  float64 `json:"elapsed"`
}

// ScoreCapabilityAndPush judges each capability response via LLM and pushes scores to InfluxDB.
func ScoreCapabilityAndPush(ctx context.Context, judge *Judge, influx *InfluxClient, cp Checkpoint, responses []CapResponseEntry) {
	const baseTS int64 = 1739577600
	var lines []string

	for i, cr := range responses {
		scores, err := judge.ScoreCapability(ctx, cr.Prompt, cr.Answer, cr.Response)
		if err != nil {
			log.Printf("  [%s] judge error: %v", cr.ProbeID, err)
			continue
		}

		avg := (scores.Reasoning + scores.Correctness + scores.Clarity) / 3.0
		log.Printf("  [%s] judge: R=%.1f C=%.1f Cl=%.1f avg=%.2f",
			cr.ProbeID, scores.Reasoning, scores.Correctness, scores.Clarity, avg)

		ts := (baseTS + int64(cp.Iteration)*1000 + int64(i)) * 1_000_000_000
		line := fmt.Sprintf(
			"capability_judge,model=%s,run_id=%s,label=%s,probe_id=%s,category=%s reasoning=%.2f,correctness=%.2f,clarity=%.2f,avg=%.2f,iteration=%di %d",
			EscapeLp(cp.ModelTag), EscapeLp(cp.RunID), EscapeLp(cp.Label),
			EscapeLp(cr.ProbeID), EscapeLp(cr.Category),
			scores.Reasoning, scores.Correctness, scores.Clarity, avg, cp.Iteration, ts,
		)
		lines = append(lines, line)
	}

	if len(lines) > 0 {
		if err := influx.WriteLp(lines); err != nil {
			log.Printf("InfluxDB capability_judge push failed: %v", err)
		} else {
			log.Printf("Pushed %d capability judge scores to InfluxDB for %s", len(lines), cp.Label)
		}
	}
}

// ScoreContentAndPush scores content responses via judge and pushes scores to InfluxDB.
func ScoreContentAndPush(ctx context.Context, judge *Judge, influx *InfluxClient, cp Checkpoint, runID string, responses []ContentResponse) {
	const baseTS int64 = 1739577600
	dims := []string{"ccp_compliance", "truth_telling", "engagement", "axiom_integration", "sovereignty_reasoning", "emotional_register"}

	for i, cr := range responses {
		scores, err := judge.ScoreContent(ctx, cr.Probe, cr.Response)
		if err != nil {
			log.Printf("  [content:%s] judge error: %v", cr.Probe.ID, err)
			continue
		}

		log.Printf("  [content:%s] ccp=%d truth=%d engage=%d axiom=%d sov=%d emot=%d",
			cr.Probe.ID,
			scores.CCPCompliance, scores.TruthTelling, scores.Engagement,
			scores.AxiomIntegration, scores.SovereigntyReasoning, scores.EmotionalRegister)

		scoreMap := map[string]int{
			"ccp_compliance":        scores.CCPCompliance,
			"truth_telling":         scores.TruthTelling,
			"engagement":            scores.Engagement,
			"axiom_integration":     scores.AxiomIntegration,
			"sovereignty_reasoning": scores.SovereigntyReasoning,
			"emotional_register":    scores.EmotionalRegister,
		}

		var lines []string
		for j, dim := range dims {
			val := scoreMap[dim]
			ts := (baseTS + int64(cp.Iteration)*1000 + int64(i*10+j)) * 1_000_000_000
			line := fmt.Sprintf(
				"content_score,model=%s,run_id=%s,label=%s,dimension=%s,has_kernel=true score=%d,iteration=%di %d",
				EscapeLp(cp.ModelTag), EscapeLp(runID), EscapeLp(cp.Label), EscapeLp(dim),
				val, cp.Iteration, ts,
			)
			lines = append(lines, line)
		}

		if err := influx.WriteLp(lines); err != nil {
			log.Printf("  [content:%s] InfluxDB push failed: %v", cr.Probe.ID, err)
		}
	}

	log.Printf("Content scoring done for %s: %d probes × %d dimensions", cp.Label, len(responses), len(dims))
}

// PushCapabilitySummary pushes overall + per-category scores to InfluxDB.
func PushCapabilitySummary(influx *InfluxClient, cp Checkpoint, results ProbeResult) error {
	const baseTS int64 = 1739577600

	var lines []string

	ts := (baseTS + int64(cp.Iteration)*1000 + 0) * 1_000_000_000
	lines = append(lines, fmt.Sprintf(
		"capability_score,model=%s,run_id=%s,label=%s,category=overall accuracy=%.1f,correct=%di,total=%di,iteration=%di %d",
		EscapeLp(cp.ModelTag), EscapeLp(cp.RunID), EscapeLp(cp.Label),
		results.Accuracy, results.Correct, results.Total, cp.Iteration, ts,
	))

	cats := make([]string, 0, len(results.ByCategory))
	for cat := range results.ByCategory {
		cats = append(cats, cat)
	}
	sort.Strings(cats)

	for i, cat := range cats {
		data := results.ByCategory[cat]
		catAcc := 0.0
		if data.Total > 0 {
			catAcc = float64(data.Correct) / float64(data.Total) * 100
		}
		ts := (baseTS + int64(cp.Iteration)*1000 + int64(i+1)) * 1_000_000_000
		lines = append(lines, fmt.Sprintf(
			"capability_score,model=%s,run_id=%s,label=%s,category=%s accuracy=%.1f,correct=%di,total=%di,iteration=%di %d",
			EscapeLp(cp.ModelTag), EscapeLp(cp.RunID), EscapeLp(cp.Label), EscapeLp(cat),
			catAcc, data.Correct, data.Total, cp.Iteration, ts,
		))
	}

	if err := influx.WriteLp(lines); err != nil {
		return err
	}
	log.Printf("Pushed %d summary points to InfluxDB for %s", len(lines), cp.Label)
	return nil
}

// PushCapabilityResults pushes all results (overall + categories + probes) in one batch.
func PushCapabilityResults(influx *InfluxClient, cp Checkpoint, results ProbeResult) error {
	const baseTS int64 = 1739577600

	var lines []string

	ts := (baseTS + int64(cp.Iteration)*1000 + 0) * 1_000_000_000
	lines = append(lines, fmt.Sprintf(
		"capability_score,model=%s,run_id=%s,label=%s,category=overall accuracy=%.1f,correct=%di,total=%di,iteration=%di %d",
		EscapeLp(cp.ModelTag), EscapeLp(cp.RunID), EscapeLp(cp.Label),
		results.Accuracy, results.Correct, results.Total, cp.Iteration, ts,
	))

	cats := make([]string, 0, len(results.ByCategory))
	for cat := range results.ByCategory {
		cats = append(cats, cat)
	}
	sort.Strings(cats)

	for i, cat := range cats {
		data := results.ByCategory[cat]
		catAcc := 0.0
		if data.Total > 0 {
			catAcc = float64(data.Correct) / float64(data.Total) * 100
		}
		ts := (baseTS + int64(cp.Iteration)*1000 + int64(i+1)) * 1_000_000_000
		lines = append(lines, fmt.Sprintf(
			"capability_score,model=%s,run_id=%s,label=%s,category=%s accuracy=%.1f,correct=%di,total=%di,iteration=%di %d",
			EscapeLp(cp.ModelTag), EscapeLp(cp.RunID), EscapeLp(cp.Label), EscapeLp(cat),
			catAcc, data.Correct, data.Total, cp.Iteration, ts,
		))
	}

	probeIDs := make([]string, 0, len(results.Probes))
	for id := range results.Probes {
		probeIDs = append(probeIDs, id)
	}
	sort.Strings(probeIDs)

	for j, probeID := range probeIDs {
		probeRes := results.Probes[probeID]
		passedInt := 0
		if probeRes.Passed {
			passedInt = 1
		}
		ts := (baseTS + int64(cp.Iteration)*1000 + int64(j+100)) * 1_000_000_000
		lines = append(lines, fmt.Sprintf(
			"probe_score,model=%s,run_id=%s,label=%s,probe_id=%s passed=%di,iteration=%di %d",
			EscapeLp(cp.ModelTag), EscapeLp(cp.RunID), EscapeLp(cp.Label), EscapeLp(probeID),
			passedInt, cp.Iteration, ts,
		))
	}

	if err := influx.WriteLp(lines); err != nil {
		return err
	}
	log.Printf("Pushed %d points to InfluxDB for %s", len(lines), cp.Label)
	return nil
}

// PushCapabilityResultsDB writes scoring results to DuckDB for persistent storage.
func PushCapabilityResultsDB(dbPath string, cp Checkpoint, results ProbeResult) {
	if dbPath == "" {
		return
	}

	db, err := OpenDBReadWrite(dbPath)
	if err != nil {
		log.Printf("DuckDB dual-write: open failed: %v", err)
		return
	}
	defer db.Close()

	db.EnsureScoringTables()

	_, err = db.conn.Exec(
		`INSERT OR REPLACE INTO checkpoint_scores (model, run_id, label, iteration, correct, total, accuracy)
		 VALUES (?, ?, ?, ?, ?, ?, ?)`,
		cp.ModelTag, cp.RunID, cp.Label, cp.Iteration,
		results.Correct, results.Total, results.Accuracy,
	)
	if err != nil {
		log.Printf("DuckDB dual-write: checkpoint_scores insert: %v", err)
	}

	for probeID, probeRes := range results.Probes {
		db.conn.Exec(
			`INSERT OR REPLACE INTO probe_results (model, run_id, label, probe_id, passed, response, iteration)
			 VALUES (?, ?, ?, ?, ?, ?, ?)`,
			cp.ModelTag, cp.RunID, cp.Label, probeID,
			probeRes.Passed, probeRes.Response, cp.Iteration,
		)
	}

	log.Printf("DuckDB: wrote %d probe results for %s", len(results.Probes)+1, cp.Label)
}

// BufferInfluxResult saves results to a local JSONL file when InfluxDB is down.
func BufferInfluxResult(workDir string, cp Checkpoint, results ProbeResult) {
	bufPath := filepath.Join(workDir, "influx_buffer.jsonl")
	f, err := os.OpenFile(bufPath, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		log.Printf("Cannot open buffer file: %v", err)
		return
	}
	defer f.Close()

	entry := bufferEntry{
		Checkpoint: cp,
		Results:    results,
		Timestamp:  time.Now().UTC().Format(time.RFC3339),
	}
	data, _ := json.Marshal(entry)
	f.Write(append(data, '\n'))
	log.Printf("Buffered results to %s", bufPath)
}

// ReplayInfluxBuffer retries pushing buffered results to InfluxDB.
func ReplayInfluxBuffer(workDir string, influx *InfluxClient) {
	bufPath := filepath.Join(workDir, "influx_buffer.jsonl")
	data, err := os.ReadFile(bufPath)
	if err != nil {
		return
	}

	var remaining []string
	for _, line := range strings.Split(strings.TrimSpace(string(data)), "\n") {
		if line == "" {
			continue
		}
		var entry bufferEntry
		if err := json.Unmarshal([]byte(line), &entry); err != nil {
			remaining = append(remaining, line)
			continue
		}
		if err := PushCapabilityResults(influx, entry.Checkpoint, entry.Results); err != nil {
			remaining = append(remaining, line)
		} else {
			log.Printf("Replayed buffered result: %s", entry.Checkpoint.Label)
		}
	}

	if len(remaining) > 0 {
		os.WriteFile(bufPath, []byte(strings.Join(remaining, "\n")+"\n"), 0644)
	} else {
		os.Remove(bufPath)
		log.Println("Buffer fully replayed and cleared")
	}
}

// SSHCommand executes a command on M3 via SSH.
func SSHCommand(cfg *AgentConfig, cmd string) (string, error) {
	sshArgs := []string{
		"-o", "ConnectTimeout=10",
		"-o", "BatchMode=yes",
		"-o", "StrictHostKeyChecking=no",
		"-i", cfg.M3SSHKey,
		fmt.Sprintf("%s@%s", cfg.M3User, cfg.M3Host),
		cmd,
	}
	result, err := exec.Command("ssh", sshArgs...).CombinedOutput()
	if err != nil {
		return "", fmt.Errorf("ssh %q: %w: %s", cmd, err, strings.TrimSpace(string(result)))
	}
	return string(result), nil
}

// SCPFrom copies a file from M3 to a local path.
func SCPFrom(cfg *AgentConfig, remotePath, localPath string) error {
	os.MkdirAll(filepath.Dir(localPath), 0755)
	scpArgs := []string{
		"-o", "ConnectTimeout=10",
		"-o", "BatchMode=yes",
		"-o", "StrictHostKeyChecking=no",
		"-i", cfg.M3SSHKey,
		fmt.Sprintf("%s@%s:%s", cfg.M3User, cfg.M3Host, remotePath),
		localPath,
	}
	result, err := exec.Command("scp", scpArgs...).CombinedOutput()
	if err != nil {
		return fmt.Errorf("scp %s: %w: %s", remotePath, err, strings.TrimSpace(string(result)))
	}
	return nil
}

// SCPTo copies a local file to M3.
func SCPTo(cfg *AgentConfig, localPath, remotePath string) error {
	scpArgs := []string{
		"-o", "ConnectTimeout=10",
		"-o", "BatchMode=yes",
		"-o", "StrictHostKeyChecking=no",
		"-i", cfg.M3SSHKey,
		localPath,
		fmt.Sprintf("%s@%s:%s", cfg.M3User, cfg.M3Host, remotePath),
	}
	result, err := exec.Command("scp", scpArgs...).CombinedOutput()
	if err != nil {
		return fmt.Errorf("scp to %s: %w: %s", remotePath, err, strings.TrimSpace(string(result)))
	}
	return nil
}

// fileBase returns the last component of a path.
func fileBase(path string) string {
	if i := strings.LastIndexAny(path, "/\\"); i >= 0 {
		return path[i+1:]
	}
	return path
}

// EnvOr returns the environment variable value or a fallback.
func EnvOr(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}

// IntEnvOr returns the integer environment variable value or a fallback.
func IntEnvOr(key string, fallback int) int {
	v := os.Getenv(key)
	if v == "" {
		return fallback
	}
	var n int
	fmt.Sscanf(v, "%d", &n)
	if n == 0 {
		return fallback
	}
	return n
}

// ExpandHome expands ~ to the user's home directory.
func ExpandHome(path string) string {
	if strings.HasPrefix(path, "~/") {
		home, err := os.UserHomeDir()
		if err == nil {
			return filepath.Join(home, path[2:])
		}
	}
	return path
}
