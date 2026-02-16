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
	"net/url"
	"os"
	"path/filepath"
	"runtime"
	"sync"
	"time"

	"forge.lthn.ai/core/cli/pkg/forge"
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
	Body      json.RawMessage `json:"body,omitempty"`
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

// queueOp marshals body to JSON and appends a PendingOp to the queue.
func (h *HubService) queueOp(method, path string, body interface{}) {
	var raw json.RawMessage
	if body != nil {
		data, err := json.Marshal(body)
		if err != nil {
			log.Printf("BugSETI: queueOp marshal error: %v", err)
			return
		}
		raw = data
	}

	h.mu.Lock()
	h.pending = append(h.pending, PendingOp{
		Method:    method,
		Path:      path,
		Body:      raw,
		CreatedAt: time.Now(),
	})
	h.mu.Unlock()

	h.savePendingOps()
}

// drainPendingOps replays queued operations against the hub.
// 5xx/transport errors are kept for retry; 4xx responses are dropped (stale).
func (h *HubService) drainPendingOps() {
	h.mu.Lock()
	ops := h.pending
	h.pending = make([]PendingOp, 0)
	h.mu.Unlock()

	if len(ops) == 0 {
		return
	}

	var failed []PendingOp
	for _, op := range ops {
		var body interface{}
		if len(op.Body) > 0 {
			body = json.RawMessage(op.Body)
		}

		resp, err := h.doRequest(op.Method, op.Path, body)
		if err != nil {
			// Transport error — keep for retry.
			failed = append(failed, op)
			continue
		}
		resp.Body.Close()

		if resp.StatusCode >= 500 {
			// Server error — keep for retry.
			failed = append(failed, op)
		} // 4xx are dropped (stale).
	}

	if len(failed) > 0 {
		h.mu.Lock()
		h.pending = append(failed, h.pending...)
		h.mu.Unlock()
	}

	h.savePendingOps()
}

// savePendingOps persists the pending operations queue to disk.
func (h *HubService) savePendingOps() {
	dataDir := h.config.GetDataDir()
	if dataDir == "" {
		return
	}

	h.mu.RLock()
	data, err := json.Marshal(h.pending)
	h.mu.RUnlock()
	if err != nil {
		log.Printf("BugSETI: savePendingOps marshal error: %v", err)
		return
	}

	path := filepath.Join(dataDir, "hub_pending.json")
	if err := os.WriteFile(path, data, 0600); err != nil {
		log.Printf("BugSETI: savePendingOps write error: %v", err)
	}
}

// loadPendingOps loads the pending operations queue from disk.
// Errors are silently ignored (the file may not exist yet).
func (h *HubService) loadPendingOps() {
	dataDir := h.config.GetDataDir()
	if dataDir == "" {
		return
	}

	path := filepath.Join(dataDir, "hub_pending.json")
	data, err := os.ReadFile(path)
	if err != nil {
		return
	}

	var ops []PendingOp
	if err := json.Unmarshal(data, &ops); err != nil {
		return
	}
	h.pending = ops
}

// PendingCount returns the number of queued pending operations.
func (h *HubService) PendingCount() int {
	h.mu.RLock()
	defer h.mu.RUnlock()
	return len(h.pending)
}

// ---- Task 4: Auto-Register via Forge Token ----

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

// ---- Task 5: Write Operations ----

// Register registers this client with the hub.
func (h *HubService) Register() error {
	h.drainPendingOps()

	name := h.config.GetClientName()
	clientID := h.config.GetClientID()
	if name == "" {
		if len(clientID) >= 8 {
			name = "BugSETI-" + clientID[:8]
		} else {
			name = "BugSETI-" + clientID
		}
	}

	body := map[string]string{
		"client_id": clientID,
		"name":      name,
		"version":   GetVersion(),
		"os":        runtime.GOOS,
		"arch":      runtime.GOARCH,
	}

	return h.doJSON("POST", "/register", body, nil)
}

// Heartbeat sends a heartbeat to the hub.
func (h *HubService) Heartbeat() error {
	body := map[string]string{
		"client_id": h.config.GetClientID(),
	}
	return h.doJSON("POST", "/heartbeat", body, nil)
}

// ClaimIssue claims an issue on the hub, returning the claim details.
// Returns a ConflictError if the issue is already claimed by another client.
func (h *HubService) ClaimIssue(issue *Issue) (*HubClaim, error) {
	h.drainPendingOps()

	body := map[string]interface{}{
		"client_id":    h.config.GetClientID(),
		"issue_id":     issue.ID,
		"repo":         issue.Repo,
		"issue_number": issue.Number,
		"title":        issue.Title,
		"url":          issue.URL,
	}

	var claim HubClaim
	if err := h.doJSON("POST", "/issues/claim", body, &claim); err != nil {
		return nil, err
	}
	return &claim, nil
}

// UpdateStatus updates the status of a claimed issue on the hub.
func (h *HubService) UpdateStatus(issueID, status, prURL string, prNumber int) error {
	body := map[string]interface{}{
		"client_id": h.config.GetClientID(),
		"status":    status,
	}
	if prURL != "" {
		body["pr_url"] = prURL
	}
	if prNumber > 0 {
		body["pr_number"] = prNumber
	}

	path := "/issues/" + url.PathEscape(issueID) + "/status"
	return h.doJSON("PATCH", path, body, nil)
}

// ReleaseClaim releases a previously claimed issue back to the pool.
func (h *HubService) ReleaseClaim(issueID string) error {
	body := map[string]string{
		"client_id": h.config.GetClientID(),
	}

	path := "/issues/" + url.PathEscape(issueID) + "/claim"
	return h.doJSON("DELETE", path, body, nil)
}

// SyncStats uploads local statistics to the hub.
func (h *HubService) SyncStats(stats *Stats) error {
	// Build repos_contributed as a flat string slice from the map keys.
	repos := make([]string, 0, len(stats.ReposContributed))
	for k := range stats.ReposContributed {
		repos = append(repos, k)
	}

	body := map[string]interface{}{
		"client_id": h.config.GetClientID(),
		"stats": map[string]interface{}{
			"issues_attempted":  stats.IssuesAttempted,
			"issues_completed":  stats.IssuesCompleted,
			"issues_skipped":    stats.IssuesSkipped,
			"prs_submitted":     stats.PRsSubmitted,
			"prs_merged":        stats.PRsMerged,
			"prs_rejected":      stats.PRsRejected,
			"current_streak":    stats.CurrentStreak,
			"longest_streak":    stats.LongestStreak,
			"total_time_minutes": int(stats.TotalTimeSpent.Minutes()),
			"repos_contributed": repos,
		},
	}

	return h.doJSON("POST", "/stats/sync", body, nil)
}

// ---- Task 6: Read Operations ----

// IsIssueClaimed checks whether an issue is currently claimed on the hub.
// Returns the claim if it exists, or (nil, nil) if the issue is not claimed (404).
func (h *HubService) IsIssueClaimed(issueID string) (*HubClaim, error) {
	path := "/issues/" + url.PathEscape(issueID)

	var claim HubClaim
	if err := h.doJSON("GET", path, nil, &claim); err != nil {
		if _, ok := err.(*NotFoundError); ok {
			return nil, nil
		}
		return nil, err
	}
	return &claim, nil
}

// ListClaims returns claimed issues, optionally filtered by status and/or repo.
func (h *HubService) ListClaims(status, repo string) ([]*HubClaim, error) {
	params := url.Values{}
	if status != "" {
		params.Set("status", status)
	}
	if repo != "" {
		params.Set("repo", repo)
	}

	path := "/issues/claimed"
	if encoded := params.Encode(); encoded != "" {
		path += "?" + encoded
	}

	var claims []*HubClaim
	if err := h.doJSON("GET", path, nil, &claims); err != nil {
		return nil, err
	}
	return claims, nil
}

// leaderboardResponse wraps the hub leaderboard JSON envelope.
type leaderboardResponse struct {
	Entries           []LeaderboardEntry `json:"entries"`
	TotalParticipants int                `json:"totalParticipants"`
}

// GetLeaderboard fetches the top N leaderboard entries from the hub.
func (h *HubService) GetLeaderboard(limit int) ([]LeaderboardEntry, int, error) {
	path := fmt.Sprintf("/leaderboard?limit=%d", limit)

	var resp leaderboardResponse
	if err := h.doJSON("GET", path, nil, &resp); err != nil {
		return nil, 0, err
	}
	return resp.Entries, resp.TotalParticipants, nil
}

// GetGlobalStats fetches aggregate statistics from the hub.
func (h *HubService) GetGlobalStats() (*GlobalStats, error) {
	var stats GlobalStats
	if err := h.doJSON("GET", "/stats", nil, &stats); err != nil {
		return nil, err
	}
	return &stats, nil
}
