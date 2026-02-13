// Package bugseti provides services for the BugSETI distributed bug fixing application.
package bugseti

import (
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"net/http"
	"sync"
	"time"
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

// loadPendingOps is a no-op placeholder (disk persistence comes in Task 7).
func (h *HubService) loadPendingOps() {}

// savePendingOps is a no-op placeholder (disk persistence comes in Task 7).
func (h *HubService) savePendingOps() {}
