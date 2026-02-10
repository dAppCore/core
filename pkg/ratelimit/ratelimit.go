package ratelimit

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"sync"
	"time"

	"gopkg.in/yaml.v3"
)

// ModelQuota defines the rate limits for a specific model.
type ModelQuota struct {
	MaxRPM int `yaml:"max_rpm"` // Requests per minute
	MaxTPM int `yaml:"max_tpm"` // Tokens per minute
	MaxRPD int `yaml:"max_rpd"` // Requests per day (0 = unlimited)
}

// TokenEntry records a token usage event.
type TokenEntry struct {
	Time  time.Time `yaml:"time"`
	Count int       `yaml:"count"`
}

// UsageStats tracks usage history for a model.
type UsageStats struct {
	Requests []time.Time  `yaml:"requests"` // Sliding window (1m)
	Tokens   []TokenEntry `yaml:"tokens"`   // Sliding window (1m)
	DayStart time.Time    `yaml:"day_start"`
	DayCount int          `yaml:"day_count"`
}

// RateLimiter manages rate limits across multiple models.
type RateLimiter struct {
	mu       sync.RWMutex
	Quotas   map[string]ModelQuota  `yaml:"quotas"`
	State    map[string]*UsageStats `yaml:"state"`
	filePath string
}

// New creates a new RateLimiter with default quotas.
func New() (*RateLimiter, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return nil, err
	}

	rl := &RateLimiter{
		Quotas:   make(map[string]ModelQuota),
		State:    make(map[string]*UsageStats),
		filePath: filepath.Join(home, ".core", "ratelimits.yaml"),
	}

	// Default quotas based on Tier 1 observations (Feb 2026)
	rl.Quotas["gemini-3-pro-preview"] = ModelQuota{MaxRPM: 150, MaxTPM: 1000000, MaxRPD: 1000}
	rl.Quotas["gemini-3-flash-preview"] = ModelQuota{MaxRPM: 150, MaxTPM: 1000000, MaxRPD: 1000}
	rl.Quotas["gemini-2.5-pro"] = ModelQuota{MaxRPM: 150, MaxTPM: 1000000, MaxRPD: 1000}
	rl.Quotas["gemini-2.0-flash"] = ModelQuota{MaxRPM: 150, MaxTPM: 1000000, MaxRPD: 0} // Unlimited RPD
	rl.Quotas["gemini-2.0-flash-lite"] = ModelQuota{MaxRPM: 0, MaxTPM: 0, MaxRPD: 0}   // Unlimited

	return rl, nil
}

// Load reads the state from disk.
func (rl *RateLimiter) Load() error {
	rl.mu.Lock()
	defer rl.mu.Unlock()

	data, err := os.ReadFile(rl.filePath)
	if os.IsNotExist(err) {
		return nil
	}
	if err != nil {
		return err
	}

	return yaml.Unmarshal(data, rl)
}

// Persist writes the state to disk.
func (rl *RateLimiter) Persist() error {
	rl.mu.RLock()
	defer rl.mu.RUnlock()

	data, err := yaml.Marshal(rl)
	if err != nil {
		return err
	}

	dir := filepath.Dir(rl.filePath)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return err
	}

	return os.WriteFile(rl.filePath, data, 0644)
}

// prune removes entries older than the sliding window (1 minute).
// Caller must hold lock.
func (rl *RateLimiter) prune(model string) {
	stats, ok := rl.State[model]
	if !ok {
		return
	}

	now := time.Now()
	window := now.Add(-1 * time.Minute)

	// Prune requests
	validReqs := 0
	for _, t := range stats.Requests {
		if t.After(window) {
			stats.Requests[validReqs] = t
			validReqs++
		}
	}
	stats.Requests = stats.Requests[:validReqs]

	// Prune tokens
	validTokens := 0
	for _, t := range stats.Tokens {
		if t.Time.After(window) {
			stats.Tokens[validTokens] = t
			validTokens++
		}
	}
	stats.Tokens = stats.Tokens[:validTokens]

	// Reset daily counter if day has passed
	if now.Sub(stats.DayStart) >= 24*time.Hour {
		stats.DayStart = now
		stats.DayCount = 0
	}
}

// CanSend checks if a request can be sent without violating limits.
func (rl *RateLimiter) CanSend(model string, estimatedTokens int) bool {
	rl.mu.Lock()
	defer rl.mu.Unlock()

	quota, ok := rl.Quotas[model]
	if !ok {
		return true // Unknown models are allowed
	}

	// Unlimited check
	if quota.MaxRPM == 0 && quota.MaxTPM == 0 && quota.MaxRPD == 0 {
		return true
	}

	// Ensure state exists
	if _, ok := rl.State[model]; !ok {
		rl.State[model] = &UsageStats{
			DayStart: time.Now(),
		}
	}

	rl.prune(model)
	stats := rl.State[model]

	// Check RPD
	if quota.MaxRPD > 0 && stats.DayCount >= quota.MaxRPD {
		return false
	}

	// Check RPM
	if quota.MaxRPM > 0 && len(stats.Requests) >= quota.MaxRPM {
		return false
	}

	// Check TPM
	if quota.MaxTPM > 0 {
		currentTokens := 0
		for _, t := range stats.Tokens {
			currentTokens += t.Count
		}
		if currentTokens+estimatedTokens > quota.MaxTPM {
			return false
		}
	}

	return true
}

// RecordUsage records a successful API call.
func (rl *RateLimiter) RecordUsage(model string, promptTokens, outputTokens int) {
	rl.mu.Lock()
	defer rl.mu.Unlock()

	if _, ok := rl.State[model]; !ok {
		rl.State[model] = &UsageStats{
			DayStart: time.Now(),
		}
	}

	stats := rl.State[model]
	now := time.Now()

	stats.Requests = append(stats.Requests, now)
	stats.Tokens = append(stats.Tokens, TokenEntry{Time: now, Count: promptTokens + outputTokens})
	stats.DayCount++
}

// WaitForCapacity blocks until capacity is available or context is cancelled.
func (rl *RateLimiter) WaitForCapacity(ctx context.Context, model string, tokens int) error {
	ticker := time.NewTicker(1 * time.Second)
	defer ticker.Stop()

	for {
		if rl.CanSend(model, tokens) {
			return nil
		}

		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-ticker.C:
			// check again
		}
	}
}

// Reset clears stats for a model (or all if model is empty).
func (rl *RateLimiter) Reset(model string) {
	rl.mu.Lock()
	defer rl.mu.Unlock()

	if model == "" {
		rl.State = make(map[string]*UsageStats)
	} else {
		delete(rl.State, model)
	}
}

// ModelStats represents a snapshot of usage.
type ModelStats struct {
	RPM      int
	MaxRPM   int
	TPM      int
	MaxTPM   int
	RPD      int
	MaxRPD   int
	DayStart time.Time
}

// Stats returns current stats for a model.
func (rl *RateLimiter) Stats(model string) ModelStats {
	rl.mu.Lock()
	defer rl.mu.Unlock()

	rl.prune(model)

	stats := ModelStats{}
	quota, ok := rl.Quotas[model]
	if ok {
		stats.MaxRPM = quota.MaxRPM
		stats.MaxTPM = quota.MaxTPM
		stats.MaxRPD = quota.MaxRPD
	}

	if s, ok := rl.State[model]; ok {
		stats.RPM = len(s.Requests)
		stats.RPD = s.DayCount
		stats.DayStart = s.DayStart
		for _, t := range s.Tokens {
			stats.TPM += t.Count
		}
	}

	return stats
}

// AllStats returns stats for all tracked models.
func (rl *RateLimiter) AllStats() map[string]ModelStats {
	rl.mu.Lock()
	defer rl.mu.Unlock()

	result := make(map[string]ModelStats)

	// Collect all model names
	for m := range rl.Quotas {
		result[m] = ModelStats{}
	}
	for m := range rl.State {
		result[m] = ModelStats{}
	}

	now := time.Now()
	window := now.Add(-1 * time.Minute)

	for m := range result {
		// Prune inline
		if s, ok := rl.State[m]; ok {
			validReqs := 0
			for _, t := range s.Requests {
				if t.After(window) {
					s.Requests[validReqs] = t
					validReqs++
				}
			}
			s.Requests = s.Requests[:validReqs]

			validTokens := 0
			for _, t := range s.Tokens {
				if t.Time.After(window) {
					s.Tokens[validTokens] = t
					validTokens++
				}
			}
			s.Tokens = s.Tokens[:validTokens]

			if now.Sub(s.DayStart) >= 24*time.Hour {
				s.DayStart = now
				s.DayCount = 0
			}
		}

		ms := ModelStats{}
		if q, ok := rl.Quotas[m]; ok {
			ms.MaxRPM = q.MaxRPM
			ms.MaxTPM = q.MaxTPM
			ms.MaxRPD = q.MaxRPD
		}
		if s, ok := rl.State[m]; ok {
			ms.RPM = len(s.Requests)
			ms.RPD = s.DayCount
			ms.DayStart = s.DayStart
			for _, t := range s.Tokens {
				ms.TPM += t.Count
			}
		}
		result[m] = ms
	}

	return result
}

// CountTokens calls the Google API to count tokens for a prompt.
func CountTokens(apiKey, model, text string) (int, error) {
	url := fmt.Sprintf("https://generativelanguage.googleapis.com/v1beta/models/%s:countTokens?key=%s", model, apiKey)

	reqBody := map[string]any{
		"contents": []any{
			map[string]any{
				"parts": []any{
					map[string]string{"text": text},
				},
			},
		},
	}

	jsonBody, err := json.Marshal(reqBody)
	if err != nil {
		return 0, err
	}

	resp, err := http.Post(url, "application/json", bytes.NewBuffer(jsonBody))
	if err != nil {
		return 0, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return 0, fmt.Errorf("API error %d: %s", resp.StatusCode, string(body))
	}

	var result struct {
		TotalTokens int `json:"totalTokens"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return 0, err
	}

	return result.TotalTokens, nil
}
