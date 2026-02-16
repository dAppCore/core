package ml

import (
	"bufio"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"sort"
	"strings"
)

// ConsolidateConfig holds options for the consolidate operation.
type ConsolidateConfig struct {
	M3Host    string
	RemoteDir string
	Pattern   string
	OutputDir string
	MergedOut string
}

// Consolidate pulls JSONL response files from M3 via SSH, merges them by idx,
// deduplicates, and writes a single merged JSONL output.
func Consolidate(cfg ConsolidateConfig, w io.Writer) error {
	if cfg.OutputDir == "" {
		cfg.OutputDir = "responses"
	}
	if err := os.MkdirAll(cfg.OutputDir, 0755); err != nil {
		return fmt.Errorf("create output dir: %w", err)
	}

	// List remote files via SSH.
	fmt.Fprintln(w, "Pulling responses from remote...")
	listCmd := exec.Command("ssh", cfg.M3Host, fmt.Sprintf("ls %s/%s", cfg.RemoteDir, cfg.Pattern))
	listOutput, err := listCmd.Output()
	if err != nil {
		return fmt.Errorf("list remote files: %w", err)
	}

	remoteFiles := strings.Split(strings.TrimSpace(string(listOutput)), "\n")
	var validFiles []string
	for _, f := range remoteFiles {
		f = strings.TrimSpace(f)
		if f != "" {
			validFiles = append(validFiles, f)
		}
	}
	fmt.Fprintf(w, "  Found %d JSONL files on %s\n", len(validFiles), cfg.M3Host)

	// Pull each file via SCP.
	for _, rf := range validFiles {
		local := filepath.Join(cfg.OutputDir, filepath.Base(rf))
		scpCmd := exec.Command("scp", fmt.Sprintf("%s:%s", cfg.M3Host, rf), local)
		if err := scpCmd.Run(); err != nil {
			fmt.Fprintf(w, "  warning: failed to pull %s: %v\n", rf, err)
			continue
		}

		lines, err := countLines(local)
		if err == nil {
			fmt.Fprintf(w, "  %s: %d records\n", filepath.Base(rf), lines)
		}
	}

	// Merge and deduplicate on idx (first occurrence wins).
	seen := make(map[int]json.RawMessage)
	skipped := 0

	matches, _ := filepath.Glob(filepath.Join(cfg.OutputDir, cfg.Pattern))
	sort.Strings(matches)

	for _, local := range matches {
		f, err := os.Open(local)
		if err != nil {
			continue
		}
		scanner := bufio.NewScanner(f)
		scanner.Buffer(make([]byte, 1024*1024), 1024*1024)
		for scanner.Scan() {
			line := scanner.Text()
			var rec struct {
				Idx *int `json:"idx"`
			}
			if err := json.Unmarshal([]byte(line), &rec); err != nil {
				skipped++
				continue
			}
			if rec.Idx == nil {
				skipped++
				continue
			}
			if _, exists := seen[*rec.Idx]; !exists {
				seen[*rec.Idx] = json.RawMessage(line)
			}
		}
		f.Close()
	}

	if skipped > 0 {
		fmt.Fprintf(w, "  Skipped %d records without idx\n", skipped)
	}

	// Sort by idx and write merged file.
	mergedPath := cfg.MergedOut
	if mergedPath == "" {
		mergedPath = filepath.Join(cfg.OutputDir, "..", "gold-merged.jsonl")
	}

	idxs := make([]int, 0, len(seen))
	for idx := range seen {
		idxs = append(idxs, idx)
	}
	sort.Ints(idxs)

	out, err := os.Create(mergedPath)
	if err != nil {
		return fmt.Errorf("create merged file: %w", err)
	}
	defer out.Close()

	bw := bufio.NewWriter(out)
	for _, idx := range idxs {
		bw.Write(seen[idx])
		bw.WriteString("\n")
	}
	if err := bw.Flush(); err != nil {
		return fmt.Errorf("flush merged file: %w", err)
	}

	fmt.Fprintf(w, "\nMerged: %d unique examples -> %s\n", len(seen), mergedPath)
	return nil
}

// countLines returns the number of lines in a file.
func countLines(path string) (int, error) {
	f, err := os.Open(path)
	if err != nil {
		return 0, err
	}
	defer f.Close()

	count := 0
	scanner := bufio.NewScanner(f)
	for scanner.Scan() {
		count++
	}
	return count, scanner.Err()
}
