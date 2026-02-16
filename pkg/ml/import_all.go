package ml

import (
	"bufio"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

// ImportConfig holds options for the import-all operation.
type ImportConfig struct {
	SkipM3  bool
	DataDir string
	M3Host  string
}

// ImportAll imports all LEM data into DuckDB from M3 and local files.
func ImportAll(db *DB, cfg ImportConfig, w io.Writer) error {
	m3Host := cfg.M3Host
	if m3Host == "" {
		m3Host = "m3"
	}

	totals := make(map[string]int)

	// ── 1. Golden set ──
	goldenPath := filepath.Join(cfg.DataDir, "gold-15k.jsonl")
	if !cfg.SkipM3 {
		fmt.Fprintln(w, "  Pulling golden set from M3...")
		scpCmd := exec.Command("scp", fmt.Sprintf("%s:/Volumes/Data/lem/responses/gold-15k.jsonl", m3Host), goldenPath)
		if err := scpCmd.Run(); err != nil {
			fmt.Fprintf(w, "  WARNING: could not pull golden set from M3: %v\n", err)
		}
	}
	if _, err := os.Stat(goldenPath); err == nil {
		db.Exec("DROP TABLE IF EXISTS golden_set")
		err := db.Exec(fmt.Sprintf(`
			CREATE TABLE golden_set AS
			SELECT
				idx::INT AS idx,
				seed_id::VARCHAR AS seed_id,
				domain::VARCHAR AS domain,
				voice::VARCHAR AS voice,
				prompt::VARCHAR AS prompt,
				response::VARCHAR AS response,
				gen_time::DOUBLE AS gen_time,
				length(response)::INT AS char_count,
				length(response) - length(replace(response, ' ', '')) + 1 AS word_count
			FROM read_json_auto('%s', maximum_object_size=1048576)
		`, escapeSQLPath(goldenPath)))
		if err != nil {
			fmt.Fprintf(w, "  WARNING: golden set import failed: %v\n", err)
		} else {
			var n int
			db.QueryRowScan("SELECT count(*) FROM golden_set", &n)
			totals["golden_set"] = n
			fmt.Fprintf(w, "  golden_set: %d rows\n", n)
		}
	}

	// ── 2. Training examples ──
	trainingDirs := []struct {
		name  string
		files []string
	}{
		{"training", []string{"training/train.jsonl", "training/valid.jsonl", "training/test.jsonl"}},
		{"training-2k", []string{"training-2k/train.jsonl", "training-2k/valid.jsonl", "training-2k/test.jsonl"}},
		{"training-expanded", []string{"training-expanded/train.jsonl", "training-expanded/valid.jsonl"}},
		{"training-book", []string{"training-book/train.jsonl", "training-book/valid.jsonl", "training-book/test.jsonl"}},
		{"training-conv", []string{"training-conv/train.jsonl", "training-conv/valid.jsonl", "training-conv/test.jsonl"}},
		{"gold-full", []string{"gold-full/train.jsonl", "gold-full/valid.jsonl"}},
		{"sovereignty-gold", []string{"sovereignty-gold/train.jsonl", "sovereignty-gold/valid.jsonl"}},
		{"composure-lessons", []string{"composure-lessons/train.jsonl", "composure-lessons/valid.jsonl"}},
		{"watts-full", []string{"watts-full/train.jsonl", "watts-full/valid.jsonl"}},
		{"watts-expanded", []string{"watts-expanded/train.jsonl", "watts-expanded/valid.jsonl"}},
		{"watts-composure", []string{"watts-composure-merged/train.jsonl", "watts-composure-merged/valid.jsonl"}},
		{"western-fresh", []string{"western-fresh/train.jsonl", "western-fresh/valid.jsonl"}},
		{"deepseek-soak", []string{"deepseek-western-soak/train.jsonl", "deepseek-western-soak/valid.jsonl"}},
		{"russian-bridge", []string{"russian-bridge/train.jsonl", "russian-bridge/valid.jsonl"}},
	}

	trainingLocal := filepath.Join(cfg.DataDir, "training")
	os.MkdirAll(trainingLocal, 0755)

	if !cfg.SkipM3 {
		fmt.Fprintln(w, "  Pulling training sets from M3...")
		for _, td := range trainingDirs {
			for _, rel := range td.files {
				local := filepath.Join(trainingLocal, rel)
				os.MkdirAll(filepath.Dir(local), 0755)
				scpCmd := exec.Command("scp", fmt.Sprintf("%s:/Volumes/Data/lem/%s", m3Host, rel), local)
				scpCmd.Run() // ignore errors, file might not exist
			}
		}
	}

	db.Exec("DROP TABLE IF EXISTS training_examples")
	db.Exec(`
		CREATE TABLE training_examples (
			source VARCHAR,
			split VARCHAR,
			prompt TEXT,
			response TEXT,
			num_turns INT,
			full_messages TEXT,
			char_count INT
		)
	`)

	trainingTotal := 0
	for _, td := range trainingDirs {
		for _, rel := range td.files {
			local := filepath.Join(trainingLocal, rel)
			if _, err := os.Stat(local); os.IsNotExist(err) {
				continue
			}

			split := "train"
			if strings.Contains(rel, "valid") {
				split = "valid"
			} else if strings.Contains(rel, "test") {
				split = "test"
			}

			n := importTrainingFile(db, local, td.name, split)
			trainingTotal += n
		}
	}
	totals["training_examples"] = trainingTotal
	fmt.Fprintf(w, "  training_examples: %d rows\n", trainingTotal)

	// ── 3. Benchmark results ──
	benchLocal := filepath.Join(cfg.DataDir, "benchmarks")
	os.MkdirAll(benchLocal, 0755)

	if !cfg.SkipM3 {
		fmt.Fprintln(w, "  Pulling benchmarks from M3...")
		for _, bname := range []string{"truthfulqa", "gsm8k", "do_not_answer", "toxigen"} {
			scpCmd := exec.Command("scp",
				fmt.Sprintf("%s:/Volumes/Data/lem/benchmarks/%s.jsonl", m3Host, bname),
				filepath.Join(benchLocal, bname+".jsonl"))
			scpCmd.Run()
		}
		for _, subdir := range []string{"results", "scale_results", "cross_arch_results", "deepseek-r1-7b"} {
			localSub := filepath.Join(benchLocal, subdir)
			os.MkdirAll(localSub, 0755)
			scpCmd := exec.Command("scp", "-r",
				fmt.Sprintf("%s:/Volumes/Data/lem/benchmarks/%s/", m3Host, subdir),
				filepath.Join(benchLocal)+"/")
			scpCmd.Run()
		}
	}

	db.Exec("DROP TABLE IF EXISTS benchmark_results")
	db.Exec(`
		CREATE TABLE benchmark_results (
			source VARCHAR, id VARCHAR, benchmark VARCHAR, model VARCHAR,
			prompt TEXT, response TEXT, elapsed_seconds DOUBLE, domain VARCHAR
		)
	`)

	benchTotal := 0
	for _, subdir := range []string{"results", "scale_results", "cross_arch_results", "deepseek-r1-7b"} {
		resultDir := filepath.Join(benchLocal, subdir)
		matches, _ := filepath.Glob(filepath.Join(resultDir, "*.jsonl"))
		for _, jf := range matches {
			n := importBenchmarkFile(db, jf, subdir)
			benchTotal += n
		}
	}

	// Also import standalone benchmark files.
	for _, bfile := range []string{"lem_bench", "lem_ethics", "lem_ethics_allen", "instruction_tuned", "abliterated", "base_pt"} {
		local := filepath.Join(benchLocal, bfile+".jsonl")
		if _, err := os.Stat(local); os.IsNotExist(err) {
			if !cfg.SkipM3 {
				scpCmd := exec.Command("scp",
					fmt.Sprintf("%s:/Volumes/Data/lem/benchmark/%s.jsonl", m3Host, bfile), local)
				scpCmd.Run()
			}
		}
		if _, err := os.Stat(local); err == nil {
			n := importBenchmarkFile(db, local, "benchmark")
			benchTotal += n
		}
	}
	totals["benchmark_results"] = benchTotal
	fmt.Fprintf(w, "  benchmark_results: %d rows\n", benchTotal)

	// ── 4. Benchmark questions ──
	db.Exec("DROP TABLE IF EXISTS benchmark_questions")
	db.Exec(`
		CREATE TABLE benchmark_questions (
			benchmark VARCHAR, id VARCHAR, question TEXT,
			best_answer TEXT, correct_answers TEXT, incorrect_answers TEXT, category VARCHAR
		)
	`)

	benchQTotal := 0
	for _, bname := range []string{"truthfulqa", "gsm8k", "do_not_answer", "toxigen"} {
		local := filepath.Join(benchLocal, bname+".jsonl")
		if _, err := os.Stat(local); err == nil {
			n := importBenchmarkQuestions(db, local, bname)
			benchQTotal += n
		}
	}
	totals["benchmark_questions"] = benchQTotal
	fmt.Fprintf(w, "  benchmark_questions: %d rows\n", benchQTotal)

	// ── 5. Seeds ──
	db.Exec("DROP TABLE IF EXISTS seeds")
	db.Exec(`
		CREATE TABLE seeds (
			source_file VARCHAR, region VARCHAR, seed_id VARCHAR, domain VARCHAR, prompt TEXT
		)
	`)

	seedTotal := 0
	seedDirs := []string{filepath.Join(cfg.DataDir, "seeds"), "/tmp/lem-data/seeds", "/tmp/lem-repo/seeds"}
	for _, seedDir := range seedDirs {
		if _, err := os.Stat(seedDir); os.IsNotExist(err) {
			continue
		}
		n := importSeeds(db, seedDir)
		seedTotal += n
	}
	totals["seeds"] = seedTotal
	fmt.Fprintf(w, "  seeds: %d rows\n", seedTotal)

	// ── Summary ──
	grandTotal := 0
	fmt.Fprintf(w, "\n%s\n", strings.Repeat("=", 50))
	fmt.Fprintln(w, "LEM Database Import Complete")
	fmt.Fprintln(w, strings.Repeat("=", 50))
	for table, count := range totals {
		fmt.Fprintf(w, "  %-25s %8d\n", table, count)
		grandTotal += count
	}
	fmt.Fprintf(w, "  %s\n", strings.Repeat("-", 35))
	fmt.Fprintf(w, "  %-25s %8d\n", "TOTAL", grandTotal)
	fmt.Fprintf(w, "\nDatabase: %s\n", db.Path())

	return nil
}

func importTrainingFile(db *DB, path, source, split string) int {
	f, err := os.Open(path)
	if err != nil {
		return 0
	}
	defer f.Close()

	count := 0
	scanner := bufio.NewScanner(f)
	scanner.Buffer(make([]byte, 1024*1024), 1024*1024)

	for scanner.Scan() {
		var rec struct {
			Messages []ChatMessage `json:"messages"`
		}
		if err := json.Unmarshal(scanner.Bytes(), &rec); err != nil {
			continue
		}

		prompt := ""
		response := ""
		assistantCount := 0
		for _, m := range rec.Messages {
			if m.Role == "user" && prompt == "" {
				prompt = m.Content
			}
			if m.Role == "assistant" {
				if response == "" {
					response = m.Content
				}
				assistantCount++
			}
		}

		msgsJSON, _ := json.Marshal(rec.Messages)
		db.Exec(`INSERT INTO training_examples VALUES (?, ?, ?, ?, ?, ?, ?)`,
			source, split, prompt, response, assistantCount, string(msgsJSON), len(response))
		count++
	}
	return count
}

func importBenchmarkFile(db *DB, path, source string) int {
	f, err := os.Open(path)
	if err != nil {
		return 0
	}
	defer f.Close()

	count := 0
	scanner := bufio.NewScanner(f)
	scanner.Buffer(make([]byte, 1024*1024), 1024*1024)

	for scanner.Scan() {
		var rec map[string]interface{}
		if err := json.Unmarshal(scanner.Bytes(), &rec); err != nil {
			continue
		}

		db.Exec(`INSERT INTO benchmark_results VALUES (?, ?, ?, ?, ?, ?, ?, ?)`,
			source,
			fmt.Sprintf("%v", rec["id"]),
			strOrEmpty(rec, "benchmark"),
			strOrEmpty(rec, "model"),
			strOrEmpty(rec, "prompt"),
			strOrEmpty(rec, "response"),
			floatOrZero(rec, "elapsed_seconds"),
			strOrEmpty(rec, "domain"),
		)
		count++
	}
	return count
}

func importBenchmarkQuestions(db *DB, path, benchmark string) int {
	f, err := os.Open(path)
	if err != nil {
		return 0
	}
	defer f.Close()

	count := 0
	scanner := bufio.NewScanner(f)
	scanner.Buffer(make([]byte, 1024*1024), 1024*1024)

	for scanner.Scan() {
		var rec map[string]interface{}
		if err := json.Unmarshal(scanner.Bytes(), &rec); err != nil {
			continue
		}

		correctJSON, _ := json.Marshal(rec["correct_answers"])
		incorrectJSON, _ := json.Marshal(rec["incorrect_answers"])

		db.Exec(`INSERT INTO benchmark_questions VALUES (?, ?, ?, ?, ?, ?, ?)`,
			benchmark,
			fmt.Sprintf("%v", rec["id"]),
			strOrEmpty(rec, "question"),
			strOrEmpty(rec, "best_answer"),
			string(correctJSON),
			string(incorrectJSON),
			strOrEmpty(rec, "category"),
		)
		count++
	}
	return count
}

func importSeeds(db *DB, seedDir string) int {
	count := 0
	filepath.Walk(seedDir, func(path string, info os.FileInfo, err error) error {
		if err != nil || info.IsDir() || !strings.HasSuffix(path, ".json") {
			return nil
		}

		data, err := os.ReadFile(path)
		if err != nil {
			return nil
		}

		rel, _ := filepath.Rel(seedDir, path)
		region := strings.TrimSuffix(filepath.Base(path), ".json")

		// Try parsing as array or object with prompts/seeds field.
		var seedsList []interface{}
		var raw interface{}
		if err := json.Unmarshal(data, &raw); err != nil {
			return nil
		}

		switch v := raw.(type) {
		case []interface{}:
			seedsList = v
		case map[string]interface{}:
			if prompts, ok := v["prompts"].([]interface{}); ok {
				seedsList = prompts
			} else if seeds, ok := v["seeds"].([]interface{}); ok {
				seedsList = seeds
			}
		}

		for _, s := range seedsList {
			switch seed := s.(type) {
			case map[string]interface{}:
				prompt := strOrEmpty(seed, "prompt")
				if prompt == "" {
					prompt = strOrEmpty(seed, "text")
				}
				if prompt == "" {
					prompt = strOrEmpty(seed, "question")
				}
				db.Exec(`INSERT INTO seeds VALUES (?, ?, ?, ?, ?)`,
					rel, region,
					strOrEmpty(seed, "seed_id"),
					strOrEmpty(seed, "domain"),
					prompt,
				)
				count++
			case string:
				db.Exec(`INSERT INTO seeds VALUES (?, ?, ?, ?, ?)`,
					rel, region, "", "", seed)
				count++
			}
		}
		return nil
	})
	return count
}

func strOrEmpty(m map[string]interface{}, key string) string {
	if v, ok := m[key]; ok {
		return fmt.Sprintf("%v", v)
	}
	return ""
}

func floatOrZero(m map[string]interface{}, key string) float64 {
	if v, ok := m[key]; ok {
		if f, ok := v.(float64); ok {
			return f
		}
	}
	return 0
}

func escapeSQLPath(p string) string {
	return strings.ReplaceAll(p, "'", "''")
}
