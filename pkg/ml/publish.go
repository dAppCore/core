package ml

import (
	"bytes"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"
)

// PublishConfig holds options for the publish operation.
type PublishConfig struct {
	InputDir string
	Repo     string
	Public   bool
	Token    string
	DryRun   bool
}

// uploadEntry pairs a local file path with its remote destination.
type uploadEntry struct {
	local  string
	remote string
}

// Publish uploads Parquet files to HuggingFace Hub.
//
// It looks for train.parquet, valid.parquet, and test.parquet in InputDir,
// plus an optional dataset_card.md in the parent directory (uploaded as README.md).
// The token is resolved from PublishConfig.Token, the HF_TOKEN environment variable,
// or ~/.huggingface/token, in that order.
func Publish(cfg PublishConfig, w io.Writer) error {
	if cfg.InputDir == "" {
		return fmt.Errorf("input directory is required")
	}

	token := resolveHFToken(cfg.Token)
	if token == "" && !cfg.DryRun {
		return fmt.Errorf("HuggingFace token required (--token, HF_TOKEN env, or ~/.huggingface/token)")
	}

	files, err := collectUploadFiles(cfg.InputDir)
	if err != nil {
		return err
	}
	if len(files) == 0 {
		return fmt.Errorf("no Parquet files found in %s", cfg.InputDir)
	}

	if cfg.DryRun {
		fmt.Fprintf(w, "Dry run: would publish to %s\n", cfg.Repo)
		if cfg.Public {
			fmt.Fprintln(w, "  Visibility: public")
		} else {
			fmt.Fprintln(w, "  Visibility: private")
		}
		for _, f := range files {
			info, err := os.Stat(f.local)
			if err != nil {
				return fmt.Errorf("stat %s: %w", f.local, err)
			}
			sizeMB := float64(info.Size()) / 1024 / 1024
			fmt.Fprintf(w, "  %s -> %s (%.1f MB)\n", filepath.Base(f.local), f.remote, sizeMB)
		}
		return nil
	}

	fmt.Fprintf(w, "Publishing to https://huggingface.co/datasets/%s\n", cfg.Repo)

	for _, f := range files {
		if err := uploadFileToHF(token, cfg.Repo, f.local, f.remote); err != nil {
			return fmt.Errorf("upload %s: %w", filepath.Base(f.local), err)
		}
		fmt.Fprintf(w, "  Uploaded %s -> %s\n", filepath.Base(f.local), f.remote)
	}

	fmt.Fprintf(w, "\nPublished to https://huggingface.co/datasets/%s\n", cfg.Repo)
	return nil
}

// resolveHFToken returns a HuggingFace API token from the given value,
// HF_TOKEN env var, or ~/.huggingface/token file.
func resolveHFToken(explicit string) string {
	if explicit != "" {
		return explicit
	}
	if env := os.Getenv("HF_TOKEN"); env != "" {
		return env
	}
	home, err := os.UserHomeDir()
	if err != nil {
		return ""
	}
	data, err := os.ReadFile(filepath.Join(home, ".huggingface", "token"))
	if err != nil {
		return ""
	}
	return strings.TrimSpace(string(data))
}

// collectUploadFiles finds Parquet split files and an optional dataset card.
func collectUploadFiles(inputDir string) ([]uploadEntry, error) {
	splits := []string{"train", "valid", "test"}
	var files []uploadEntry

	for _, split := range splits {
		path := filepath.Join(inputDir, split+".parquet")
		if _, err := os.Stat(path); os.IsNotExist(err) {
			continue
		} else if err != nil {
			return nil, fmt.Errorf("stat %s: %w", path, err)
		}
		files = append(files, uploadEntry{path, fmt.Sprintf("data/%s.parquet", split)})
	}

	// Check for dataset card in parent directory.
	cardPath := filepath.Join(inputDir, "..", "dataset_card.md")
	if _, err := os.Stat(cardPath); err == nil {
		files = append(files, uploadEntry{cardPath, "README.md"})
	}

	return files, nil
}

// uploadFileToHF uploads a single file to a HuggingFace dataset repo via the Hub API.
func uploadFileToHF(token, repoID, localPath, remotePath string) error {
	data, err := os.ReadFile(localPath)
	if err != nil {
		return fmt.Errorf("read %s: %w", localPath, err)
	}

	url := fmt.Sprintf("https://huggingface.co/api/datasets/%s/upload/main/%s", repoID, remotePath)

	req, err := http.NewRequest(http.MethodPut, url, bytes.NewReader(data))
	if err != nil {
		return fmt.Errorf("create request: %w", err)
	}
	req.Header.Set("Authorization", "Bearer "+token)
	req.Header.Set("Content-Type", "application/octet-stream")

	client := &http.Client{Timeout: 120 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return fmt.Errorf("upload request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 300 {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("upload failed: HTTP %d: %s", resp.StatusCode, string(body))
	}

	return nil
}
