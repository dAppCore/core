package ml

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"runtime"
	"time"
)

// WorkerConfig holds the worker's runtime configuration.
type WorkerConfig struct {
	APIBase      string
	WorkerID     string
	Name         string
	APIKey       string
	GPUType      string
	VRAMGb       int
	Languages    []string
	Models       []string
	InferURL     string
	TaskType     string
	BatchSize    int
	PollInterval time.Duration
	OneShot      bool
	DryRun       bool
}

// APITask represents a task from the LEM API.
type APITask struct {
	ID         int    `json:"id"`
	TaskType   string `json:"task_type"`
	Status     string `json:"status"`
	Language   string `json:"language"`
	Domain     string `json:"domain"`
	ModelName  string `json:"model_name"`
	PromptID   string `json:"prompt_id"`
	PromptText string `json:"prompt_text"`
	Config     *struct {
		Temperature float64 `json:"temperature,omitempty"`
		MaxTokens   int     `json:"max_tokens,omitempty"`
	} `json:"config"`
	Priority int `json:"priority"`
}

// RunWorkerLoop is the main worker loop that polls for tasks and processes them.
func RunWorkerLoop(cfg *WorkerConfig) {
	log.Printf("LEM Worker starting")
	log.Printf("  ID:       %s", cfg.WorkerID)
	log.Printf("  Name:     %s", cfg.Name)
	log.Printf("  API:      %s", cfg.APIBase)
	log.Printf("  Infer:    %s", cfg.InferURL)
	log.Printf("  GPU:      %s (%d GB)", cfg.GPUType, cfg.VRAMGb)
	log.Printf("  Langs:    %v", cfg.Languages)
	log.Printf("  Models:   %v", cfg.Models)
	log.Printf("  Batch:    %d", cfg.BatchSize)
	log.Printf("  Dry-run:  %v", cfg.DryRun)

	if err := workerRegister(cfg); err != nil {
		log.Fatalf("Registration failed: %v", err)
	}
	log.Println("Registered with LEM API")

	for {
		processed := workerPoll(cfg)

		if cfg.OneShot {
			log.Printf("One-shot mode: processed %d tasks, exiting", processed)
			return
		}

		if processed == 0 {
			log.Printf("No tasks available, sleeping %v", cfg.PollInterval)
			time.Sleep(cfg.PollInterval)
		}

		workerHeartbeat(cfg)
	}
}

func workerRegister(cfg *WorkerConfig) error {
	body := map[string]interface{}{
		"worker_id": cfg.WorkerID,
		"name":      cfg.Name,
		"version":   "0.1.0",
		"os":        runtime.GOOS,
		"arch":      runtime.GOARCH,
	}
	if cfg.GPUType != "" {
		body["gpu_type"] = cfg.GPUType
	}
	if cfg.VRAMGb > 0 {
		body["vram_gb"] = cfg.VRAMGb
	}
	if len(cfg.Languages) > 0 {
		body["languages"] = cfg.Languages
	}
	if len(cfg.Models) > 0 {
		body["supported_models"] = cfg.Models
	}

	_, err := apiPost(cfg, "/api/lem/workers/register", body)
	return err
}

func workerHeartbeat(cfg *WorkerConfig) {
	body := map[string]interface{}{
		"worker_id": cfg.WorkerID,
	}
	apiPost(cfg, "/api/lem/workers/heartbeat", body)
}

func workerPoll(cfg *WorkerConfig) int {
	url := fmt.Sprintf("/api/lem/tasks/next?worker_id=%s&limit=%d", cfg.WorkerID, cfg.BatchSize)
	if cfg.TaskType != "" {
		url += "&type=" + cfg.TaskType
	}

	resp, err := apiGet(cfg, url)
	if err != nil {
		log.Printf("Error fetching tasks: %v", err)
		return 0
	}

	var result struct {
		Tasks []APITask `json:"tasks"`
		Count int       `json:"count"`
	}
	if err := json.Unmarshal(resp, &result); err != nil {
		log.Printf("Error parsing tasks: %v", err)
		return 0
	}

	if result.Count == 0 {
		return 0
	}

	log.Printf("Got %d tasks", result.Count)
	processed := 0

	for _, task := range result.Tasks {
		if err := workerProcessTask(cfg, task); err != nil {
			log.Printf("Task %d failed: %v", task.ID, err)
			apiDelete(cfg, fmt.Sprintf("/api/lem/tasks/%d/claim", task.ID), map[string]interface{}{
				"worker_id": cfg.WorkerID,
			})
			continue
		}
		processed++
	}

	return processed
}

func workerProcessTask(cfg *WorkerConfig, task APITask) error {
	log.Printf("Processing task %d: %s [%s/%s] %d chars prompt",
		task.ID, task.TaskType, task.Language, task.Domain, len(task.PromptText))

	_, err := apiPost(cfg, fmt.Sprintf("/api/lem/tasks/%d/claim", task.ID), map[string]interface{}{
		"worker_id": cfg.WorkerID,
	})
	if err != nil {
		return fmt.Errorf("claim: %w", err)
	}

	apiPatch(cfg, fmt.Sprintf("/api/lem/tasks/%d/status", task.ID), map[string]interface{}{
		"worker_id": cfg.WorkerID,
		"status":    "in_progress",
	})

	if cfg.DryRun {
		log.Printf("  [DRY-RUN] Would generate response for: %.80s...", task.PromptText)
		return nil
	}

	start := time.Now()
	response, err := workerInfer(cfg, task)
	genTime := time.Since(start)

	if err != nil {
		apiPatch(cfg, fmt.Sprintf("/api/lem/tasks/%d/status", task.ID), map[string]interface{}{
			"worker_id": cfg.WorkerID,
			"status":    "abandoned",
		})
		return fmt.Errorf("inference: %w", err)
	}

	modelUsed := task.ModelName
	if modelUsed == "" {
		modelUsed = "default"
	}

	_, err = apiPost(cfg, fmt.Sprintf("/api/lem/tasks/%d/result", task.ID), map[string]interface{}{
		"worker_id":     cfg.WorkerID,
		"response_text": response,
		"model_used":    modelUsed,
		"gen_time_ms":   int(genTime.Milliseconds()),
	})
	if err != nil {
		return fmt.Errorf("submit result: %w", err)
	}

	log.Printf("  Completed: %d chars in %v", len(response), genTime.Round(time.Millisecond))
	return nil
}

func workerInfer(cfg *WorkerConfig, task APITask) (string, error) {
	messages := []map[string]string{
		{"role": "user", "content": task.PromptText},
	}

	temp := 0.7
	maxTokens := 2048
	if task.Config != nil {
		if task.Config.Temperature > 0 {
			temp = task.Config.Temperature
		}
		if task.Config.MaxTokens > 0 {
			maxTokens = task.Config.MaxTokens
		}
	}

	reqBody := map[string]interface{}{
		"model":       task.ModelName,
		"messages":    messages,
		"temperature": temp,
		"max_tokens":  maxTokens,
	}

	data, err := json.Marshal(reqBody)
	if err != nil {
		return "", err
	}

	req, err := http.NewRequest("POST", cfg.InferURL+"/v1/chat/completions", bytes.NewReader(data))
	if err != nil {
		return "", err
	}
	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{Timeout: 5 * time.Minute}
	resp, err := client.Do(req)
	if err != nil {
		return "", fmt.Errorf("inference request: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("read response: %w", err)
	}

	if resp.StatusCode != 200 {
		return "", fmt.Errorf("inference HTTP %d: %s", resp.StatusCode, truncStr(string(body), 200))
	}

	var chatResp struct {
		Choices []struct {
			Message struct {
				Content string `json:"content"`
			} `json:"message"`
		} `json:"choices"`
	}
	if err := json.Unmarshal(body, &chatResp); err != nil {
		return "", fmt.Errorf("parse response: %w", err)
	}

	if len(chatResp.Choices) == 0 {
		return "", fmt.Errorf("no choices in response")
	}

	content := chatResp.Choices[0].Message.Content
	if len(content) < 10 {
		return "", fmt.Errorf("response too short: %d chars", len(content))
	}

	return content, nil
}

// HTTP helpers for the LEM API.

func apiGet(cfg *WorkerConfig, path string) ([]byte, error) {
	req, err := http.NewRequest("GET", cfg.APIBase+path, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Authorization", "Bearer "+cfg.APIKey)

	client := &http.Client{Timeout: 15 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	if resp.StatusCode >= 400 {
		return nil, fmt.Errorf("HTTP %d: %s", resp.StatusCode, truncStr(string(body), 200))
	}

	return body, nil
}

func apiPost(cfg *WorkerConfig, path string, data map[string]interface{}) ([]byte, error) {
	return apiRequest(cfg, "POST", path, data)
}

func apiPatch(cfg *WorkerConfig, path string, data map[string]interface{}) ([]byte, error) {
	return apiRequest(cfg, "PATCH", path, data)
}

func apiDelete(cfg *WorkerConfig, path string, data map[string]interface{}) ([]byte, error) {
	return apiRequest(cfg, "DELETE", path, data)
}

func apiRequest(cfg *WorkerConfig, method, path string, data map[string]interface{}) ([]byte, error) {
	jsonData, err := json.Marshal(data)
	if err != nil {
		return nil, err
	}

	req, err := http.NewRequest(method, cfg.APIBase+path, bytes.NewReader(jsonData))
	if err != nil {
		return nil, err
	}
	req.Header.Set("Authorization", "Bearer "+cfg.APIKey)
	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{Timeout: 15 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	if resp.StatusCode >= 400 {
		return nil, fmt.Errorf("HTTP %d: %s", resp.StatusCode, truncStr(string(body), 200))
	}

	return body, nil
}

// MachineID returns the machine ID from /etc/machine-id or hostname fallback.
func MachineID() string {
	if data, err := os.ReadFile("/etc/machine-id"); err == nil {
		id := string(bytes.TrimSpace(data))
		if len(id) > 0 {
			return id
		}
	}
	h, _ := os.Hostname()
	return h
}

// Hostname returns the system hostname.
func Hostname() string {
	h, _ := os.Hostname()
	return h
}

// ReadKeyFile reads the LEM API key from ~/.config/lem/api_key.
func ReadKeyFile() string {
	home, _ := os.UserHomeDir()
	path := filepath.Join(home, ".config", "lem", "api_key")
	data, err := os.ReadFile(path)
	if err != nil {
		return ""
	}
	return string(bytes.TrimSpace(data))
}

// SplitComma splits a comma-separated string into trimmed parts.
func SplitComma(s string) []string {
	var result []string
	for _, part := range bytes.Split([]byte(s), []byte(",")) {
		trimmed := bytes.TrimSpace(part)
		if len(trimmed) > 0 {
			result = append(result, string(trimmed))
		}
	}
	return result
}

func truncStr(s string, n int) string {
	if len(s) <= n {
		return s
	}
	return s[:n] + "..."
}
