# BugSETI HubService Implementation Plan

> **For Claude:** REQUIRED SUB-SKILL: Use superpowers:executing-plans to implement this plan task-by-task.

**Goal:** Add a HubService to BugSETI that coordinates issue claiming, stats sync, and leaderboard with the agentic portal API.

**Architecture:** Thin HTTP client (`net/http`) in `internal/bugseti/hub.go` talking directly to the portal's `/api/bugseti/*` endpoints. Auto-registers via forge token to get an `ak_` bearer token. Offline-first with pending-ops queue.

**Tech Stack:** Go stdlib (`net/http`, `encoding/json`), Laravel 12 (portal endpoint), httptest (Go tests)

---

### Task 1: Config Fields

Add hub-related fields to the Config struct so HubService can persist its state.

**Files:**
- Modify: `internal/bugseti/config.go`
- Test: `internal/bugseti/fetcher_test.go` (uses `testConfigService`)

**Step 1: Add config fields**

In `internal/bugseti/config.go`, add these fields to the `Config` struct after the `ForgeToken` field:

```go
// Hub coordination (agentic portal)
HubURL     string `json:"hubUrl,omitempty"`     // Portal API base URL (e.g. https://leth.in)
HubToken   string `json:"hubToken,omitempty"`   // Cached ak_ bearer token
ClientID   string `json:"clientId,omitempty"`   // UUID, generated once on first run
ClientName string `json:"clientName,omitempty"` // Display name for leaderboard
```

**Step 2: Add getters/setters**

After the `GetForgeToken()` method, add:

```go
// GetHubURL returns the hub portal URL.
func (c *ConfigService) GetHubURL() string {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.config.HubURL
}

// SetHubURL sets the hub portal URL.
func (c *ConfigService) SetHubURL(url string) error {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.config.HubURL = url
	return c.saveUnsafe()
}

// GetHubToken returns the cached hub API token.
func (c *ConfigService) GetHubToken() string {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.config.HubToken
}

// SetHubToken caches the hub API token.
func (c *ConfigService) SetHubToken(token string) error {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.config.HubToken = token
	return c.saveUnsafe()
}

// GetClientID returns the persistent client UUID.
func (c *ConfigService) GetClientID() string {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.config.ClientID
}

// SetClientID sets the persistent client UUID.
func (c *ConfigService) SetClientID(id string) error {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.config.ClientID = id
	return c.saveUnsafe()
}

// GetClientName returns the display name.
func (c *ConfigService) GetClientName() string {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.config.ClientName
}

// SetClientName sets the display name.
func (c *ConfigService) SetClientName(name string) error {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.config.ClientName = name
	return c.saveUnsafe()
}
```

**Step 3: Run tests**

Run: `go test ./internal/bugseti/... -count=1`
Expected: All existing tests pass (config fields are additive, no breakage).

**Step 4: Commit**

```bash
git add internal/bugseti/config.go
git commit -m "feat(bugseti): add hub config fields (HubURL, HubToken, ClientID, ClientName)"
```

---

### Task 2: HubService Core — Types and Constructor

Create the HubService with data types, constructor, and ServiceName.

**Files:**
- Create: `internal/bugseti/hub.go`
- Create: `internal/bugseti/hub_test.go`

**Step 1: Write failing tests**

Create `internal/bugseti/hub_test.go`:

```go
package bugseti

import (
	"testing"
)

func testHubService(t *testing.T, serverURL string) *HubService {
	t.Helper()
	cfg := testConfigService(t, nil, nil)
	if serverURL != "" {
		cfg.config.HubURL = serverURL
	}
	return NewHubService(cfg)
}

// --- Constructor / ServiceName ---

func TestNewHubService_Good(t *testing.T) {
	h := testHubService(t, "")
	if h == nil {
		t.Fatal("expected non-nil HubService")
	}
	if h.config == nil {
		t.Fatal("expected config to be set")
	}
}

func TestHubServiceName_Good(t *testing.T) {
	h := testHubService(t, "")
	if got := h.ServiceName(); got != "HubService" {
		t.Fatalf("expected HubService, got %s", got)
	}
}

func TestNewHubService_Good_GeneratesClientID(t *testing.T) {
	h := testHubService(t, "")
	id := h.GetClientID()
	if id == "" {
		t.Fatal("expected client ID to be generated")
	}
	if len(id) < 32 {
		t.Fatalf("expected UUID-length client ID, got %d chars", len(id))
	}
}

func TestNewHubService_Good_ReusesClientID(t *testing.T) {
	cfg := testConfigService(t, nil, nil)
	cfg.config.ClientID = "existing-id-12345"
	h := NewHubService(cfg)
	if h.GetClientID() != "existing-id-12345" {
		t.Fatal("expected existing client ID to be preserved")
	}
}
```

**Step 2: Run tests to verify they fail**

Run: `go test ./internal/bugseti/... -run TestNewHubService -count=1`
Expected: FAIL — `NewHubService` not defined.

**Step 3: Write HubService core**

Create `internal/bugseti/hub.go`:

```go
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
	"runtime"
	"sync"
	"time"
)

// HubService coordinates with the agentic portal for issue claiming,
// stats sync, and leaderboard.
type HubService struct {
	config     *ConfigService
	httpClient *http.Client
	mu         sync.Mutex
	connected  bool
	pendingOps []PendingOp
}

// PendingOp represents a failed write operation queued for retry.
type PendingOp struct {
	Method    string    `json:"method"`
	Path      string    `json:"path"`
	Body      []byte    `json:"body"`
	CreatedAt time.Time `json:"createdAt"`
}

// HubClaim represents an issue claim from the portal.
type HubClaim struct {
	IssueID     string    `json:"issue_id"`
	Repo        string    `json:"repo"`
	IssueNumber int       `json:"issue_number"`
	Title       string    `json:"issue_title"`
	URL         string    `json:"issue_url"`
	Status      string    `json:"status"`
	ClaimedAt   time.Time `json:"claimed_at"`
	PRUrl       string    `json:"pr_url,omitempty"`
	PRNumber    int       `json:"pr_number,omitempty"`
}

// LeaderboardEntry represents a single entry on the leaderboard.
type LeaderboardEntry struct {
	Rank            int    `json:"rank"`
	ClientName      string `json:"client_name"`
	ClientVersion   string `json:"client_version,omitempty"`
	IssuesCompleted int    `json:"issues_completed"`
	PRsSubmitted    int    `json:"prs_submitted"`
	PRsMerged       int    `json:"prs_merged"`
	CurrentStreak   int    `json:"current_streak"`
	LongestStreak   int    `json:"longest_streak"`
}

// GlobalStats represents aggregate stats from the portal.
type GlobalStats struct {
	TotalParticipants    int `json:"total_participants"`
	ActiveParticipants   int `json:"active_participants"`
	TotalIssuesAttempted int `json:"total_issues_attempted"`
	TotalIssuesCompleted int `json:"total_issues_completed"`
	TotalPRsSubmitted    int `json:"total_prs_submitted"`
	TotalPRsMerged       int `json:"total_prs_merged"`
	ActiveClaims         int `json:"active_claims"`
	CompletedClaims      int `json:"completed_claims"`
}

// NewHubService creates a new HubService.
func NewHubService(config *ConfigService) *HubService {
	h := &HubService{
		config: config,
		httpClient: &http.Client{
			Timeout: 10 * time.Second,
		},
	}

	// Ensure a persistent client ID exists
	if config.GetClientID() == "" {
		id := generateClientID()
		if err := config.SetClientID(id); err != nil {
			log.Printf("Warning: could not persist client ID: %v", err)
		}
	}

	// Load pending ops from disk
	h.loadPendingOps()

	return h
}

// ServiceName returns the service name for Wails.
func (h *HubService) ServiceName() string {
	return "HubService"
}

// GetClientID returns the persistent client identifier.
func (h *HubService) GetClientID() string {
	return h.config.GetClientID()
}

// IsConnected returns whether the last hub request succeeded.
func (h *HubService) IsConnected() bool {
	h.mu.Lock()
	defer h.mu.Unlock()
	return h.connected
}

// generateClientID creates a random hex client identifier.
func generateClientID() string {
	b := make([]byte, 16)
	if _, err := rand.Read(b); err != nil {
		// Fallback to timestamp-based ID
		return fmt.Sprintf("bugseti-%d", time.Now().UnixNano())
	}
	return hex.EncodeToString(b)
}
```

**Step 4: Run tests**

Run: `go test ./internal/bugseti/... -run TestNewHubService -count=1 && go test ./internal/bugseti/... -run TestHubServiceName -count=1`
Expected: PASS

**Step 5: Commit**

```bash
git add internal/bugseti/hub.go internal/bugseti/hub_test.go
git commit -m "feat(bugseti): add HubService types and constructor"
```

---

### Task 3: HTTP Request Helpers

Add the internal `doRequest` and `doJSON` methods that all API calls use.

**Files:**
- Modify: `internal/bugseti/hub.go`
- Modify: `internal/bugseti/hub_test.go`

**Step 1: Write failing tests**

Add to `hub_test.go`:

```go
import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestDoRequest_Good(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Header.Get("Authorization") != "Bearer test-token" {
			t.Fatal("expected bearer token")
		}
		if r.Header.Get("Content-Type") != "application/json" {
			t.Fatal("expected JSON content type")
		}
		w.WriteHeader(200)
		w.Write([]byte(`{"ok":true}`))
	}))
	defer server.Close()

	h := testHubService(t, server.URL)
	h.config.config.HubToken = "test-token"

	resp, err := h.doRequest("GET", "/test", nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != 200 {
		t.Fatalf("expected 200, got %d", resp.StatusCode)
	}
}

func TestDoRequest_Bad_NoHubURL(t *testing.T) {
	h := testHubService(t, "")
	_, err := h.doRequest("GET", "/test", nil)
	if err == nil {
		t.Fatal("expected error when hub URL is empty")
	}
}

func TestDoRequest_Bad_NetworkError(t *testing.T) {
	h := testHubService(t, "http://127.0.0.1:1") // Nothing listening
	h.config.config.HubToken = "test-token"

	_, err := h.doRequest("GET", "/test", nil)
	if err == nil {
		t.Fatal("expected network error")
	}
}
```

**Step 2: Run to verify failure**

Run: `go test ./internal/bugseti/... -run TestDoRequest -count=1`
Expected: FAIL — `doRequest` not defined.

**Step 3: Implement helpers**

Add to `hub.go`:

```go
// doRequest performs an HTTP request to the hub API.
// Returns the response (caller must close body) or an error.
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
			return nil, fmt.Errorf("failed to marshal request body: %w", err)
		}
		bodyReader = bytes.NewReader(data)
	}

	req, err := http.NewRequest(method, fullURL, bodyReader)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Accept", "application/json")

	if token := h.config.GetHubToken(); token != "" {
		req.Header.Set("Authorization", "Bearer "+token)
	}

	resp, err := h.httpClient.Do(req)
	if err != nil {
		h.mu.Lock()
		h.connected = false
		h.mu.Unlock()
		return nil, fmt.Errorf("hub request failed: %w", err)
	}

	h.mu.Lock()
	h.connected = true
	h.mu.Unlock()

	return resp, nil
}

// doJSON performs a request and decodes the JSON response into dest.
func (h *HubService) doJSON(method, path string, body interface{}, dest interface{}) error {
	resp, err := h.doRequest(method, path, body)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode == 401 {
		return fmt.Errorf("unauthorised")
	}
	if resp.StatusCode == 409 {
		return &ConflictError{StatusCode: resp.StatusCode}
	}
	if resp.StatusCode == 404 {
		return &NotFoundError{StatusCode: resp.StatusCode}
	}
	if resp.StatusCode >= 400 {
		bodyBytes, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("hub error %d: %s", resp.StatusCode, string(bodyBytes))
	}

	if dest != nil {
		if err := json.NewDecoder(resp.Body).Decode(dest); err != nil {
			return fmt.Errorf("failed to decode response: %w", err)
		}
	}

	return nil
}

// ConflictError indicates a 409 response (e.g. issue already claimed).
type ConflictError struct {
	StatusCode int
}

func (e *ConflictError) Error() string {
	return fmt.Sprintf("conflict (HTTP %d)", e.StatusCode)
}

// NotFoundError indicates a 404 response.
type NotFoundError struct {
	StatusCode int
}

func (e *NotFoundError) Error() string {
	return fmt.Sprintf("not found (HTTP %d)", e.StatusCode)
}
```

**Step 4: Run tests**

Run: `go test ./internal/bugseti/... -run TestDoRequest -count=1`
Expected: PASS

**Step 5: Commit**

```bash
git add internal/bugseti/hub.go internal/bugseti/hub_test.go
git commit -m "feat(bugseti): add hub HTTP request helpers with error types"
```

---

### Task 4: Auto-Register via Forge Token

Implement the auth flow: send forge token to portal, receive `ak_` token.

**Files:**
- Modify: `internal/bugseti/hub.go`
- Modify: `internal/bugseti/hub_test.go`

**Step 1: Write failing tests**

Add to `hub_test.go`:

```go
func TestAutoRegister_Good(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/api/bugseti/auth/forge" {
			t.Fatalf("unexpected path: %s", r.URL.Path)
		}
		if r.Method != "POST" {
			t.Fatalf("expected POST, got %s", r.Method)
		}

		var body map[string]string
		json.NewDecoder(r.Body).Decode(&body)

		if body["forge_url"] == "" || body["forge_token"] == "" {
			w.WriteHeader(400)
			return
		}

		w.WriteHeader(201)
		json.NewEncoder(w).Encode(map[string]string{
			"api_key": "ak_test123456789012345678901234",
		})
	}))
	defer server.Close()

	cfg := testConfigService(t, nil, nil)
	cfg.config.HubURL = server.URL
	cfg.config.ForgeURL = "https://forge.lthn.io"
	cfg.config.ForgeToken = "forge-test-token"
	h := NewHubService(cfg)

	err := h.AutoRegister()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if cfg.GetHubToken() != "ak_test123456789012345678901234" {
		t.Fatalf("expected token to be cached, got %q", cfg.GetHubToken())
	}
}

func TestAutoRegister_Bad_NoForgeToken(t *testing.T) {
	cfg := testConfigService(t, nil, nil)
	cfg.config.HubURL = "http://localhost"
	h := NewHubService(cfg)

	err := h.AutoRegister()
	if err == nil {
		t.Fatal("expected error when forge token is missing")
	}
}

func TestAutoRegister_Good_SkipsIfAlreadyRegistered(t *testing.T) {
	cfg := testConfigService(t, nil, nil)
	cfg.config.HubToken = "ak_existing_token"
	h := NewHubService(cfg)

	err := h.AutoRegister()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	// Token should remain unchanged
	if cfg.GetHubToken() != "ak_existing_token" {
		t.Fatal("existing token should not be overwritten")
	}
}
```

**Step 2: Run to verify failure**

Run: `go test ./internal/bugseti/... -run TestAutoRegister -count=1`
Expected: FAIL — `AutoRegister` not defined.

**Step 3: Implement AutoRegister**

Add to `hub.go`:

```go
// AutoRegister exchanges forge credentials for a hub API key.
// Skips if a token is already cached. On 401, clears cached token.
func (h *HubService) AutoRegister() error {
	// Skip if already registered
	if h.config.GetHubToken() != "" {
		return nil
	}

	hubURL := h.config.GetHubURL()
	if hubURL == "" {
		return fmt.Errorf("hub URL not configured")
	}

	forgeURL := h.config.GetForgeURL()
	forgeToken := h.config.GetForgeToken()

	// Fall back to pkg/forge config resolution
	if forgeURL == "" || forgeToken == "" {
		resolvedURL, resolvedToken, err := resolveForgeConfig(forgeURL, forgeToken)
		if err != nil {
			return fmt.Errorf("failed to resolve forge config: %w", err)
		}
		forgeURL = resolvedURL
		forgeToken = resolvedToken
	}

	if forgeToken == "" {
		return fmt.Errorf("forge token not configured — cannot auto-register with hub")
	}

	body := map[string]string{
		"forge_url":   forgeURL,
		"forge_token": forgeToken,
		"client_id":   h.GetClientID(),
	}

	var result struct {
		APIKey string `json:"api_key"`
	}

	data, err := json.Marshal(body)
	if err != nil {
		return fmt.Errorf("failed to marshal register body: %w", err)
	}

	resp, err := h.httpClient.Post(
		hubURL+"/api/bugseti/auth/forge",
		"application/json",
		bytes.NewReader(data),
	)
	if err != nil {
		return fmt.Errorf("hub auto-register failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != 201 && resp.StatusCode != 200 {
		bodyBytes, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("hub auto-register failed (HTTP %d): %s", resp.StatusCode, string(bodyBytes))
	}

	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return fmt.Errorf("failed to decode register response: %w", err)
	}

	if result.APIKey == "" {
		return fmt.Errorf("hub returned empty API key")
	}

	// Cache the token
	if err := h.config.SetHubToken(result.APIKey); err != nil {
		return fmt.Errorf("failed to cache hub token: %w", err)
	}

	log.Printf("Hub: registered with portal, token cached")
	return nil
}

// resolveForgeConfig gets forge URL/token from pkg/forge config chain.
func resolveForgeConfig(flagURL, flagToken string) (string, string, error) {
	// Import forge package for config resolution
	// This uses the same resolution chain: config.yaml → env vars → flags
	forgeURL, forgeToken, err := forgeResolveConfig(flagURL, flagToken)
	if err != nil {
		return "", "", err
	}
	return forgeURL, forgeToken, nil
}
```

Note: `resolveForgeConfig` wraps `forge.ResolveConfig` — we'll use the import directly in the real code. For the plan, this shows the intent.

**Step 4: Run tests**

Run: `go test ./internal/bugseti/... -run TestAutoRegister -count=1`
Expected: PASS

**Step 5: Commit**

```bash
git add internal/bugseti/hub.go internal/bugseti/hub_test.go
git commit -m "feat(bugseti): hub auto-register via forge token"
```

---

### Task 5: Write Operations — Register, Heartbeat, Claim, Update, Release, SyncStats

Implement all write API methods.

**Files:**
- Modify: `internal/bugseti/hub.go`
- Modify: `internal/bugseti/hub_test.go`

**Step 1: Write failing tests**

Add to `hub_test.go`:

```go
func TestRegister_Good(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/api/bugseti/register" {
			t.Fatalf("unexpected path: %s", r.URL.Path)
		}
		var body map[string]string
		json.NewDecoder(r.Body).Decode(&body)
		if body["client_id"] == "" || body["name"] == "" {
			w.WriteHeader(400)
			return
		}
		w.WriteHeader(201)
		json.NewEncoder(w).Encode(map[string]interface{}{"client": body})
	}))
	defer server.Close()

	h := testHubService(t, server.URL)
	h.config.config.HubToken = "ak_test"
	h.config.config.ClientName = "Test Mac"

	err := h.Register()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestHeartbeat_Good(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(200)
		json.NewEncoder(w).Encode(map[string]interface{}{"ok": true})
	}))
	defer server.Close()

	h := testHubService(t, server.URL)
	h.config.config.HubToken = "ak_test"

	err := h.Heartbeat()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestClaimIssue_Good(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(201)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"claim": map[string]interface{}{
				"issue_id": "owner/repo#42",
				"status":   "claimed",
			},
		})
	}))
	defer server.Close()

	h := testHubService(t, server.URL)
	h.config.config.HubToken = "ak_test"

	claim, err := h.ClaimIssue(&Issue{
		ID: "owner/repo#42", Repo: "owner/repo", Number: 42,
		Title: "Fix bug", URL: "https://forge.lthn.io/owner/repo/issues/42",
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if claim == nil || claim.IssueID != "owner/repo#42" {
		t.Fatal("expected claim with correct issue ID")
	}
}

func TestClaimIssue_Bad_Conflict(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(409)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"error": "Issue already claimed",
		})
	}))
	defer server.Close()

	h := testHubService(t, server.URL)
	h.config.config.HubToken = "ak_test"

	_, err := h.ClaimIssue(&Issue{ID: "owner/repo#42", Repo: "owner/repo", Number: 42})
	if err == nil {
		t.Fatal("expected conflict error")
	}
	if _, ok := err.(*ConflictError); !ok {
		t.Fatalf("expected ConflictError, got %T", err)
	}
}

func TestUpdateStatus_Good(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "PATCH" {
			t.Fatalf("expected PATCH, got %s", r.Method)
		}
		w.WriteHeader(200)
		json.NewEncoder(w).Encode(map[string]interface{}{"claim": map[string]string{"status": "completed"}})
	}))
	defer server.Close()

	h := testHubService(t, server.URL)
	h.config.config.HubToken = "ak_test"

	err := h.UpdateStatus("owner/repo#42", "completed", "https://forge.lthn.io/pr/1", 1)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestSyncStats_Good(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(200)
		json.NewEncoder(w).Encode(map[string]interface{}{"synced": true})
	}))
	defer server.Close()

	h := testHubService(t, server.URL)
	h.config.config.HubToken = "ak_test"

	err := h.SyncStats(&Stats{
		IssuesCompleted: 5,
		PRsSubmitted:    3,
		PRsMerged:       2,
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}
```

**Step 2: Run to verify failure**

Run: `go test ./internal/bugseti/... -run "TestRegister_Good|TestHeartbeat|TestClaimIssue|TestUpdateStatus|TestSyncStats" -count=1`
Expected: FAIL

**Step 3: Implement write methods**

Add to `hub.go`:

```go
// Register sends client registration to the hub portal.
func (h *HubService) Register() error {
	h.drainPendingOps()

	name := h.config.GetClientName()
	if name == "" {
		name = fmt.Sprintf("BugSETI-%s", h.GetClientID()[:8])
	}

	body := map[string]string{
		"client_id": h.GetClientID(),
		"name":      name,
		"version":   GetVersion(),
		"os":        runtime.GOOS,
		"arch":      runtime.GOARCH,
	}

	return h.doJSON("POST", "/register", body, nil)
}

// Heartbeat sends a heartbeat to the hub portal.
func (h *HubService) Heartbeat() error {
	body := map[string]string{
		"client_id": h.GetClientID(),
	}
	return h.doJSON("POST", "/heartbeat", body, nil)
}

// ClaimIssue claims an issue on the hub portal.
// Returns the claim on success, ConflictError if already claimed.
func (h *HubService) ClaimIssue(issue *Issue) (*HubClaim, error) {
	if issue == nil {
		return nil, fmt.Errorf("issue is nil")
	}

	h.drainPendingOps()

	body := map[string]interface{}{
		"client_id":    h.GetClientID(),
		"issue_id":     issue.ID,
		"repo":         issue.Repo,
		"issue_number": issue.Number,
		"title":        issue.Title,
		"url":          issue.URL,
	}

	var result struct {
		Claim *HubClaim `json:"claim"`
	}

	if err := h.doJSON("POST", "/issues/claim", body, &result); err != nil {
		return nil, err
	}

	return result.Claim, nil
}

// UpdateStatus updates the status of a claimed issue.
func (h *HubService) UpdateStatus(issueID, status, prURL string, prNumber int) error {
	body := map[string]interface{}{
		"client_id": h.GetClientID(),
		"status":    status,
	}
	if prURL != "" {
		body["pr_url"] = prURL
		body["pr_number"] = prNumber
	}

	encodedID := url.PathEscape(issueID)
	return h.doJSON("PATCH", "/issues/"+encodedID+"/status", body, nil)
}

// ReleaseClaim releases a claim on an issue.
func (h *HubService) ReleaseClaim(issueID string) error {
	body := map[string]string{
		"client_id": h.GetClientID(),
	}

	encodedID := url.PathEscape(issueID)
	return h.doJSON("DELETE", "/issues/"+encodedID+"/claim", body, nil)
}

// SyncStats uploads local stats to the hub portal.
func (h *HubService) SyncStats(stats *Stats) error {
	if stats == nil {
		return fmt.Errorf("stats is nil")
	}

	repoNames := make([]string, 0, len(stats.ReposContributed))
	for name := range stats.ReposContributed {
		repoNames = append(repoNames, name)
	}

	body := map[string]interface{}{
		"client_id": h.GetClientID(),
		"stats": map[string]interface{}{
			"issues_attempted":   stats.IssuesAttempted,
			"issues_completed":   stats.IssuesCompleted,
			"issues_skipped":     stats.IssuesSkipped,
			"prs_submitted":      stats.PRsSubmitted,
			"prs_merged":         stats.PRsMerged,
			"prs_rejected":       stats.PRsRejected,
			"current_streak":     stats.CurrentStreak,
			"longest_streak":     stats.LongestStreak,
			"total_time_minutes": int(stats.TotalTimeSpent.Minutes()),
			"repos_contributed":  repoNames,
		},
	}

	return h.doJSON("POST", "/stats/sync", body, nil)
}
```

**Step 4: Run tests**

Run: `go test ./internal/bugseti/... -run "TestRegister_Good|TestHeartbeat|TestClaimIssue|TestUpdateStatus|TestSyncStats" -count=1`
Expected: PASS

**Step 5: Commit**

```bash
git add internal/bugseti/hub.go internal/bugseti/hub_test.go
git commit -m "feat(bugseti): hub write operations (register, heartbeat, claim, update, sync)"
```

---

### Task 6: Read Operations — IsIssueClaimed, ListClaims, GetLeaderboard, GetGlobalStats

**Files:**
- Modify: `internal/bugseti/hub.go`
- Modify: `internal/bugseti/hub_test.go`

**Step 1: Write failing tests**

Add to `hub_test.go`:

```go
func TestIsIssueClaimed_Good_Claimed(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(200)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"claim": map[string]interface{}{"issue_id": "o/r#1", "status": "claimed"},
		})
	}))
	defer server.Close()

	h := testHubService(t, server.URL)
	h.config.config.HubToken = "ak_test"

	claim, err := h.IsIssueClaimed("o/r#1")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if claim == nil {
		t.Fatal("expected claim")
	}
}

func TestIsIssueClaimed_Good_NotClaimed(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(404)
	}))
	defer server.Close()

	h := testHubService(t, server.URL)
	h.config.config.HubToken = "ak_test"

	claim, err := h.IsIssueClaimed("o/r#1")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if claim != nil {
		t.Fatal("expected nil claim for unclaimed issue")
	}
}

func TestGetLeaderboard_Good(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Query().Get("limit") != "10" {
			t.Fatalf("expected limit=10, got %s", r.URL.Query().Get("limit"))
		}
		w.WriteHeader(200)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"leaderboard":        []map[string]interface{}{{"rank": 1, "client_name": "Alice"}},
			"total_participants": 5,
		})
	}))
	defer server.Close()

	h := testHubService(t, server.URL)
	h.config.config.HubToken = "ak_test"

	entries, total, err := h.GetLeaderboard(10)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(entries) != 1 || total != 5 {
		t.Fatalf("expected 1 entry, 5 total; got %d, %d", len(entries), total)
	}
}

func TestGetGlobalStats_Good(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(200)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"global": map[string]interface{}{
				"total_participants": 11,
				"active_claims":      3,
			},
		})
	}))
	defer server.Close()

	h := testHubService(t, server.URL)
	h.config.config.HubToken = "ak_test"

	stats, err := h.GetGlobalStats()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if stats.TotalParticipants != 11 {
		t.Fatalf("expected 11 participants, got %d", stats.TotalParticipants)
	}
}
```

**Step 2: Run to verify failure, then implement**

Add to `hub.go`:

```go
// IsIssueClaimed checks if an issue is claimed on the hub.
// Returns the claim if found, nil if not claimed.
func (h *HubService) IsIssueClaimed(issueID string) (*HubClaim, error) {
	var result struct {
		Claim *HubClaim `json:"claim"`
	}

	encodedID := url.PathEscape(issueID)
	err := h.doJSON("GET", "/issues/"+encodedID, nil, &result)
	if err != nil {
		if _, ok := err.(*NotFoundError); ok {
			return nil, nil // Not claimed
		}
		return nil, err
	}

	return result.Claim, nil
}

// ListClaims returns active claims from the hub, with optional filters.
func (h *HubService) ListClaims(status, repo string) ([]*HubClaim, error) {
	path := "/issues/claimed"
	params := url.Values{}
	if status != "" {
		params.Set("status", status)
	}
	if repo != "" {
		params.Set("repo", repo)
	}
	if len(params) > 0 {
		path += "?" + params.Encode()
	}

	var result struct {
		Claims []*HubClaim `json:"claims"`
	}

	if err := h.doJSON("GET", path, nil, &result); err != nil {
		return nil, err
	}

	return result.Claims, nil
}

// GetLeaderboard returns the leaderboard from the hub portal.
func (h *HubService) GetLeaderboard(limit int) ([]LeaderboardEntry, int, error) {
	if limit <= 0 {
		limit = 20
	}

	path := fmt.Sprintf("/leaderboard?limit=%d", limit)

	var result struct {
		Leaderboard       []LeaderboardEntry `json:"leaderboard"`
		TotalParticipants int                `json:"total_participants"`
	}

	if err := h.doJSON("GET", path, nil, &result); err != nil {
		return nil, 0, err
	}

	return result.Leaderboard, result.TotalParticipants, nil
}

// GetGlobalStats returns aggregate stats from the hub portal.
func (h *HubService) GetGlobalStats() (*GlobalStats, error) {
	var result struct {
		Global *GlobalStats `json:"global"`
	}

	if err := h.doJSON("GET", "/stats", nil, &result); err != nil {
		return nil, err
	}

	return result.Global, nil
}
```

**Step 3: Run tests**

Run: `go test ./internal/bugseti/... -run "TestIsIssueClaimed|TestGetLeaderboard|TestGetGlobalStats" -count=1`
Expected: PASS

**Step 4: Commit**

```bash
git add internal/bugseti/hub.go internal/bugseti/hub_test.go
git commit -m "feat(bugseti): hub read operations (claims, leaderboard, global stats)"
```

---

### Task 7: Pending Operations Queue

Implement offline-first: queue failed writes, persist to disk, drain on reconnect.

**Files:**
- Modify: `internal/bugseti/hub.go`
- Modify: `internal/bugseti/hub_test.go`

**Step 1: Write failing tests**

Add to `hub_test.go`:

```go
func TestPendingOps_Good_QueueAndDrain(t *testing.T) {
	callCount := 0
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		callCount++
		if r.URL.Path == "/api/bugseti/register" {
			// First register drains pending ops — the heartbeat will come first
			w.WriteHeader(200)
			json.NewEncoder(w).Encode(map[string]interface{}{"client": nil})
			return
		}
		w.WriteHeader(200)
		json.NewEncoder(w).Encode(map[string]interface{}{"ok": true})
	}))
	defer server.Close()

	h := testHubService(t, server.URL)
	h.config.config.HubToken = "ak_test"

	// Manually add a pending op
	h.mu.Lock()
	h.pendingOps = append(h.pendingOps, PendingOp{
		Method:    "POST",
		Path:      "/heartbeat",
		Body:      []byte(`{"client_id":"test"}`),
		CreatedAt: time.Now(),
	})
	h.mu.Unlock()

	// Register should drain the pending heartbeat first
	err := h.Register()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if callCount < 2 {
		t.Fatalf("expected at least 2 calls (drain + register), got %d", callCount)
	}
}

func TestPendingOps_Good_PersistAndLoad(t *testing.T) {
	cfg := testConfigService(t, nil, nil)
	h1 := NewHubService(cfg)

	// Add pending op
	h1.mu.Lock()
	h1.pendingOps = append(h1.pendingOps, PendingOp{
		Method:    "POST",
		Path:      "/heartbeat",
		Body:      []byte(`{"test":true}`),
		CreatedAt: time.Now(),
	})
	h1.mu.Unlock()
	h1.savePendingOps()

	// Create new service — should load persisted ops
	h2 := NewHubService(cfg)
	h2.mu.Lock()
	count := len(h2.pendingOps)
	h2.mu.Unlock()

	if count != 1 {
		t.Fatalf("expected 1 pending op after reload, got %d", count)
	}
}
```

**Step 2: Implement pending ops**

Add to `hub.go`:

```go
// queueOp adds a failed write to the pending queue.
func (h *HubService) queueOp(method, path string, body interface{}) {
	data, _ := json.Marshal(body)
	h.mu.Lock()
	h.pendingOps = append(h.pendingOps, PendingOp{
		Method:    method,
		Path:      path,
		Body:      data,
		CreatedAt: time.Now(),
	})
	h.mu.Unlock()
	h.savePendingOps()
}

// drainPendingOps replays queued operations. Called before write methods.
func (h *HubService) drainPendingOps() {
	h.mu.Lock()
	ops := h.pendingOps
	h.pendingOps = nil
	h.mu.Unlock()

	if len(ops) == 0 {
		return
	}

	log.Printf("Hub: draining %d pending operations", len(ops))
	var failed []PendingOp

	for _, op := range ops {
		resp, err := h.doRequest(op.Method, op.Path, json.RawMessage(op.Body))
		if err != nil {
			failed = append(failed, op)
			continue
		}
		resp.Body.Close()
		if resp.StatusCode >= 500 {
			failed = append(failed, op)
		}
		// 4xx errors are dropped (stale data)
	}

	if len(failed) > 0 {
		h.mu.Lock()
		h.pendingOps = append(failed, h.pendingOps...)
		h.mu.Unlock()
	}

	h.savePendingOps()
}

// savePendingOps persists the pending queue to disk.
func (h *HubService) savePendingOps() {
	dataDir := h.config.GetDataDir()
	if dataDir == "" {
		return
	}

	h.mu.Lock()
	ops := h.pendingOps
	h.mu.Unlock()

	data, err := json.Marshal(ops)
	if err != nil {
		return
	}

	path := filepath.Join(dataDir, "hub_pending.json")
	os.WriteFile(path, data, 0600)
}

// loadPendingOps loads persisted pending operations from disk.
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

	h.mu.Lock()
	h.pendingOps = ops
	h.mu.Unlock()
}

// PendingCount returns the number of queued operations.
func (h *HubService) PendingCount() int {
	h.mu.Lock()
	defer h.mu.Unlock()
	return len(h.pendingOps)
}
```

Also add `"os"` and `"path/filepath"` to the imports in `hub.go`.

**Step 3: Run tests**

Run: `go test ./internal/bugseti/... -run TestPendingOps -count=1`
Expected: PASS

**Step 4: Commit**

```bash
git add internal/bugseti/hub.go internal/bugseti/hub_test.go
git commit -m "feat(bugseti): hub pending operations queue with disk persistence"
```

---

### Task 8: Integration — main.go and Wails Registration

Wire HubService into the app lifecycle.

**Files:**
- Modify: `cmd/bugseti/main.go`

**Step 1: Create HubService in main.go**

After the `submitService` creation, add:

```go
hubService := bugseti.NewHubService(configService)
```

Add to the services slice:

```go
application.NewService(hubService),
```

After `log.Println("Starting BugSETI...")`, add:

```go
// Attempt hub registration (non-blocking, logs warnings on failure)
if hubURL := configService.GetHubURL(); hubURL != "" {
	if err := hubService.AutoRegister(); err != nil {
		log.Printf("Hub: auto-register skipped: %v", err)
	} else if err := hubService.Register(); err != nil {
		log.Printf("Hub: registration failed: %v", err)
	}
}
```

**Step 2: Build and verify**

Run: `task bugseti:build`
Expected: Builds successfully.

Run: `go test ./internal/bugseti/... -count=1`
Expected: All tests pass.

**Step 3: Commit**

```bash
git add cmd/bugseti/main.go
git commit -m "feat(bugseti): wire HubService into app lifecycle"
```

---

### Task 9: Laravel Auth/Forge Endpoint

Create the portal-side endpoint that exchanges a forge token for an `ak_` API key.

**Files:**
- Create: `agentic/app/Mod/BugSeti/Controllers/AuthController.php`
- Modify: `agentic/app/Mod/BugSeti/Routes/api.php`

**Step 1: Create AuthController**

Create `agentic/app/Mod/BugSeti/Controllers/AuthController.php`:

```php
<?php

declare(strict_types=1);

namespace Mod\BugSeti\Controllers;

use Core\Agentic\Models\AgentApiKey;
use Core\Agentic\Models\Workspace;
use Illuminate\Http\JsonResponse;
use Illuminate\Http\Request;
use Illuminate\Support\Facades\Http;

class AuthController
{
    /**
     * Exchange a Forgejo token for a BugSETI API key.
     *
     * POST /api/bugseti/auth/forge
     * No authentication required (this IS the bootstrap).
     */
    public function forge(Request $request): JsonResponse
    {
        $validated = $request->validate([
            'forge_url' => 'required|url|max:500',
            'forge_token' => 'required|string|max:255',
            'client_id' => 'required|string|max:64',
        ]);

        // Validate the forge token against the Forgejo API
        $response = Http::withToken($validated['forge_token'])
            ->timeout(10)
            ->get(rtrim($validated['forge_url'], '/') . '/api/v1/user');

        if (! $response->ok()) {
            return response()->json([
                'error' => 'Invalid Forgejo token — could not verify identity.',
            ], 401);
        }

        $forgeUser = $response->json();
        $forgeName = $forgeUser['full_name'] ?: $forgeUser['login'] ?? 'Unknown';

        // Find or create workspace for BugSETI clients
        $workspace = Workspace::firstOrCreate(
            ['slug' => 'bugseti-community'],
            ['name' => 'BugSETI Community', 'owner_id' => null]
        );

        // Check if this client already has a key
        $existingKey = AgentApiKey::where('workspace_id', $workspace->id)
            ->where('name', 'like', '%' . $validated['client_id'] . '%')
            ->whereNull('revoked_at')
            ->first();

        if ($existingKey) {
            // Revoke old key and issue new one
            $existingKey->update(['revoked_at' => now()]);
        }

        $apiKey = AgentApiKey::generate(
            workspace: $workspace->id,
            name: "BugSETI — {$forgeName} ({$validated['client_id']})",
            permissions: ['bugseti.read', 'bugseti.write'],
            rateLimit: 100,
            expiresAt: null,
        );

        return response()->json([
            'api_key' => $apiKey->plainTextKey,
            'forge_user' => $forgeName,
        ], 201);
    }
}
```

**Step 2: Add route**

In `agentic/app/Mod/BugSeti/Routes/api.php`, add **outside** the authenticated groups:

```php
// Unauthenticated bootstrap — exchanges forge token for API key
Route::post('/auth/forge', [AuthController::class, 'forge']);
```

Add the use statement at top of file:

```php
use Mod\BugSeti\Controllers\AuthController;
```

**Step 3: Test manually**

```bash
cd /Users/snider/Code/host-uk/agentic
php artisan migrate
curl -X POST http://leth.test/api/bugseti/auth/forge \
  -H "Content-Type: application/json" \
  -d '{"forge_url":"https://forge.lthn.io","forge_token":"500ecb79c79da940205f37580438575dbf7a82be","client_id":"test-client-1"}'
```

Expected: 201 with `{"api_key":"ak_...","forge_user":"..."}`.

**Step 4: Commit**

```bash
cd /Users/snider/Code/host-uk/agentic
git add app/Mod/BugSeti/Controllers/AuthController.php app/Mod/BugSeti/Routes/api.php
git commit -m "feat(bugseti): add /auth/forge endpoint for token exchange"
```

---

### Task 10: Full Integration Test

Build the binary, configure hub URL, and verify end-to-end.

**Files:** None (verification only)

**Step 1: Run all Go tests**

```bash
cd /Users/snider/Code/host-uk/core
go test ./internal/bugseti/... -count=1 -v
```

Expected: All tests pass.

**Step 2: Build binary**

```bash
task bugseti:build
```

Expected: Binary builds at `bin/bugseti`.

**Step 3: Configure hub URL and test launch**

```bash
# Set hub URL to devnet
cat ~/.config/bugseti/config.json | python3 -c "
import json,sys
c = json.load(sys.stdin)
c['hubUrl'] = 'https://leth.in'
json.dump(c, sys.stdout, indent=2)
" > /tmp/bugseti-config.json && mv /tmp/bugseti-config.json ~/.config/bugseti/config.json
```

Launch `./bin/bugseti` — should start without errors, attempt hub registration.

**Step 4: Final commit if needed**

```bash
git add -A && git commit -m "feat(bugseti): HubService integration complete"
```

---

### Summary

| Task | Description | Files |
|------|-------------|-------|
| 1 | Config fields | config.go |
| 2 | HubService types + constructor | hub.go, hub_test.go |
| 3 | HTTP request helpers | hub.go, hub_test.go |
| 4 | Auto-register via forge | hub.go, hub_test.go |
| 5 | Write operations | hub.go, hub_test.go |
| 6 | Read operations | hub.go, hub_test.go |
| 7 | Pending ops queue | hub.go, hub_test.go |
| 8 | main.go integration | main.go |
| 9 | Laravel auth/forge endpoint | AuthController.php, api.php |
| 10 | Full integration test | (verification) |
