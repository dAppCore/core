// Package bugseti provides services for the BugSETI distributed bug fixing application.
package bugseti

import (
	"bytes"
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"sync"
	"time"

	"github.com/host-uk/core/pkg/forge"
)

// HubService coordinates with the agentic portal for issue assignment and leaderboard.
type HubService struct {
	config    *ConfigService
	client    *http.Client
	connected bool
	pending   []PendingOp
	mu        sync.RWMutex
}

// PendingOp represents an operation queued for retry when the hub is unreachable.
type PendingOp struct {
	Method    string      `json:"method"`
	Path      string      `json:"path"`
	Body      interface{} `json:"body,omitempty"`
	CreatedAt time.Time   `json:"createdAt"`
}

// HubClaim represents a claimed issue from the hub.
type HubClaim struct {
	ID        string    `json:"id"`
	IssueURL  string    `json:"issueUrl"`
	ClientID  string    `json:"clientId"`
	ClaimedAt time.Time `json:"claimedAt"`
	ExpiresAt time.Time `json:"expiresAt"`
	Status    string    `json:"status"`
}

// LeaderboardEntry represents a single entry on the leaderboard.
type LeaderboardEntry struct {
	ClientID   string `json:"clientId"`
	ClientName string `json:"clientName"`
	Score      int    `json:"score"`
	PRsMerged  int    `json:"prsMerged"`
	Rank       int    `json:"rank"`
}

// GlobalStats holds aggregate statistics from the hub.
type GlobalStats struct {
	TotalClients    int `json:"totalClients"`
	TotalClaims     int `json:"totalClaims"`
	TotalPRsMerged  int `json:"totalPrsMerged"`
	ActiveClaims    int `json:"activeClaims"`
	IssuesAvailable int `json:"issuesAvailable"`
}

// ConflictError indicates a 409 response from the hub (e.g. issue already claimed).
type ConflictError struct {
	StatusCode int
}

func (e *ConflictError) Error() string {
	return fmt.Sprintf("conflict: status %d", e.StatusCode)
}

// NotFoundError indicates a 404 response from the hub.
type NotFoundError struct {
	StatusCode int
}

func (e *NotFoundError) Error() string {
	return fmt.Sprintf("not found: status %d", e.StatusCode)
}

// NewHubService creates a new HubService with the given config.
// If the config has no ClientID, one is generated and persisted.
func NewHubService(config *ConfigService) *HubService {
	h := &HubService{
		config: config,
		client: &http.Client{
			Timeout: 10 * time.Second,
		},
		pending: make([]PendingOp, 0),
	}

	// Generate client ID if not set.
	if config.GetClientID() == "" {
		id := generateClientID()
		_ = config.SetClientID(id)
	}

	h.loadPendingOps()

	return h
}

// ServiceName returns the service name for Wails.
func (h *HubService) ServiceName() string {
	return "HubService"
}

// GetClientID returns the client ID from config.
func (h *HubService) GetClientID() string {
	return h.config.GetClientID()
}

// IsConnected returns whether the hub was reachable on the last request.
func (h *HubService) IsConnected() bool {
	h.mu.RLock()
	defer h.mu.RUnlock()
	return h.connected
}

// generateClientID creates a random hex string (16 bytes = 32 hex chars).
func generateClientID() string {
	b := make([]byte, 16)
	if _, err := rand.Read(b); err != nil {
		// Fallback: this should never happen with crypto/rand.
		return fmt.Sprintf("fallback-%d", time.Now().UnixNano())
	}
	return hex.EncodeToString(b)
}

// doRequest builds and executes an HTTP request against the hub API.
// It returns the raw *http.Response and any transport-level error.
func (h *HubService) doRequest(method, path string, body interface{}) (*http.Response, error) {
	hubURL := h.config.GetHubURL()
	if hubURL == "" {
		return nil, fmt.Errorf("hub URL not configured")
	}

	fullURL := hubURL + "/api/bugseti" + path

	var bodyReader io.Reader
	if body != nil {
		data, err := json.Marshal(body)
		if err != nil {
			return nil, fmt.Errorf("marshal request body: %w", err)
		}
		bodyReader = bytes.NewReader(data)
	}

	req, err := http.NewRequest(method, fullURL, bodyReader)
	if err != nil {
		return nil, fmt.Errorf("build request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Accept", "application/json")

	token := h.config.GetHubToken()
	if token != "" {
		req.Header.Set("Authorization", "Bearer "+token)
	}

	resp, err := h.client.Do(req)
	if err != nil {
		h.mu.Lock()
		h.connected = false
		h.mu.Unlock()
		return nil, err
	}

	h.mu.Lock()
	h.connected = true
	h.mu.Unlock()

	return resp, nil
}

// doJSON executes an HTTP request and decodes the JSON response into dest.
// It handles common error status codes with typed errors.
func (h *HubService) doJSON(method, path string, body, dest interface{}) error {
	resp, err := h.doRequest(method, path, body)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	switch {
	case resp.StatusCode == http.StatusUnauthorized:
		return fmt.Errorf("unauthorised")
	case resp.StatusCode == http.StatusConflict:
		return &ConflictError{StatusCode: resp.StatusCode}
	case resp.StatusCode == http.StatusNotFound:
		return &NotFoundError{StatusCode: resp.StatusCode}
	case resp.StatusCode >= 400:
		respBody, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("hub error %d: %s", resp.StatusCode, string(respBody))
	}

	if dest != nil {
		if err := json.NewDecoder(resp.Body).Decode(dest); err != nil {
			return fmt.Errorf("decode response: %w", err)
		}
	}

	return nil
}

// loadPendingOps is a no-op placeholder (disk persistence comes in Task 7).
func (h *HubService) loadPendingOps() {}

// savePendingOps is a no-op placeholder (disk persistence comes in Task 7).
func (h *HubService) savePendingOps() {}

// drainPendingOps replays queued operations (no-op until Task 7).
func (h *HubService) drainPendingOps() {}

// AutoRegister exchanges a Forge API token for a hub API key.
// If a hub token is already configured, this is a no-op.
func (h *HubService) AutoRegister() error {
	// Skip if already registered.
	if h.config.GetHubToken() != "" {
		return nil
	}

	hubURL := h.config.GetHubURL()
	if hubURL == "" {
		return fmt.Errorf("hub URL not configured")
	}

	// Resolve forge credentials from config/env.
	forgeURL := h.config.GetForgeURL()
	forgeToken := h.config.GetForgeToken()
	if forgeToken == "" {
		resolvedURL, resolvedToken, err := forge.ResolveConfig(forgeURL, "")
		if err != nil {
			return fmt.Errorf("resolve forge config: %w", err)
		}
		forgeURL = resolvedURL
		forgeToken = resolvedToken
	}

	if forgeToken == "" {
		return fmt.Errorf("no forge token available (set FORGE_TOKEN or run: core forge config --token TOKEN)")
	}

	// Build request body.
	payload := map[string]string{
		"forge_url":   forgeURL,
		"forge_token": forgeToken,
		"client_id":   h.config.GetClientID(),
	}
	data, err := json.Marshal(payload)
	if err != nil {
		return fmt.Errorf("marshal auto-register body: %w", err)
	}

	// POST directly (no bearer token yet).
	resp, err := h.client.Post(hubURL+"/api/bugseti/auth/forge", "application/json", bytes.NewReader(data))
	if err != nil {
		h.mu.Lock()
		h.connected = false
		h.mu.Unlock()
		return fmt.Errorf("auto-register request: %w", err)
	}
	defer resp.Body.Close()

	h.mu.Lock()
	h.connected = true
	h.mu.Unlock()

	if resp.StatusCode >= 400 {
		respBody, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("auto-register failed %d: %s", resp.StatusCode, string(respBody))
	}

	var result struct {
		APIKey string `json:"api_key"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return fmt.Errorf("decode auto-register response: %w", err)
	}

	if err := h.config.SetHubToken(result.APIKey); err != nil {
		return fmt.Errorf("cache hub token: %w", err)
	}

	log.Printf("BugSETI: auto-registered with hub, token cached")
	return nil
}
