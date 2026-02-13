package bugseti

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func testHubService(t *testing.T, serverURL string) *HubService {
	t.Helper()
	cfg := testConfigService(t, nil, nil)
	if serverURL != "" {
		cfg.config.HubURL = serverURL
	}
	return NewHubService(cfg)
}

// ---- NewHubService ----

func TestNewHubService_Good(t *testing.T) {
	h := testHubService(t, "")
	require.NotNil(t, h)
	assert.NotNil(t, h.config)
	assert.NotNil(t, h.client)
	assert.False(t, h.IsConnected())
}

func TestHubServiceName_Good(t *testing.T) {
	h := testHubService(t, "")
	assert.Equal(t, "HubService", h.ServiceName())
}

func TestNewHubService_Good_GeneratesClientID(t *testing.T) {
	h := testHubService(t, "")
	id := h.GetClientID()
	assert.NotEmpty(t, id)
	// 16 bytes = 32 hex characters
	assert.Len(t, id, 32)
}

func TestNewHubService_Good_ReusesClientID(t *testing.T) {
	cfg := testConfigService(t, nil, nil)
	cfg.config.ClientID = "existing-client-id"

	h := NewHubService(cfg)
	assert.Equal(t, "existing-client-id", h.GetClientID())
}

// ---- doRequest ----

func TestDoRequest_Good(t *testing.T) {
	var gotAuth string
	var gotContentType string
	var gotAccept string
	var gotBody map[string]string

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotAuth = r.Header.Get("Authorization")
		gotContentType = r.Header.Get("Content-Type")
		gotAccept = r.Header.Get("Accept")

		if r.Body != nil {
			_ = json.NewDecoder(r.Body).Decode(&gotBody)
		}

		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{"ok":true}`))
	}))
	defer srv.Close()

	cfg := testConfigService(t, nil, nil)
	cfg.config.HubURL = srv.URL
	cfg.config.HubToken = "test-token-123"
	h := NewHubService(cfg)

	body := map[string]string{"key": "value"}
	resp, err := h.doRequest("POST", "/test", body)
	require.NoError(t, err)
	defer resp.Body.Close()

	assert.Equal(t, http.StatusOK, resp.StatusCode)
	assert.Equal(t, "Bearer test-token-123", gotAuth)
	assert.Equal(t, "application/json", gotContentType)
	assert.Equal(t, "application/json", gotAccept)
	assert.Equal(t, "value", gotBody["key"])
	assert.True(t, h.IsConnected())
}

func TestDoRequest_Bad_NoHubURL(t *testing.T) {
	h := testHubService(t, "")

	resp, err := h.doRequest("GET", "/test", nil)
	assert.Nil(t, resp)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "hub URL not configured")
}

func TestDoRequest_Bad_NetworkError(t *testing.T) {
	// Point to a port where nothing is listening.
	h := testHubService(t, "http://127.0.0.1:1")

	resp, err := h.doRequest("GET", "/test", nil)
	assert.Nil(t, resp)
	assert.Error(t, err)
	assert.False(t, h.IsConnected())
}

// ---- Task 4: AutoRegister ----

func TestAutoRegister_Good(t *testing.T) {
	var gotBody map[string]string

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "/api/bugseti/auth/forge", r.URL.Path)
		assert.Equal(t, "POST", r.Method)

		_ = json.NewDecoder(r.Body).Decode(&gotBody)

		w.WriteHeader(http.StatusCreated)
		_, _ = w.Write([]byte(`{"api_key":"ak_test_12345"}`))
	}))
	defer srv.Close()

	cfg := testConfigService(t, nil, nil)
	cfg.config.HubURL = srv.URL
	cfg.config.ForgeURL = "https://forge.example.com"
	cfg.config.ForgeToken = "forge-tok-abc"
	h := NewHubService(cfg)

	err := h.AutoRegister()
	require.NoError(t, err)

	// Verify token was cached.
	assert.Equal(t, "ak_test_12345", h.config.GetHubToken())

	// Verify request body.
	assert.Equal(t, "https://forge.example.com", gotBody["forge_url"])
	assert.Equal(t, "forge-tok-abc", gotBody["forge_token"])
	assert.NotEmpty(t, gotBody["client_id"])
}

func TestAutoRegister_Bad_NoForgeToken(t *testing.T) {
	// Isolate from user's real ~/.core/config.yaml and env vars.
	origHome := os.Getenv("HOME")
	t.Setenv("HOME", t.TempDir())
	t.Setenv("FORGE_TOKEN", "")
	t.Setenv("FORGE_URL", "")
	defer os.Setenv("HOME", origHome)

	cfg := testConfigService(t, nil, nil)
	cfg.config.HubURL = "https://hub.example.com"
	// No forge token set, and env/config are empty in test.
	h := NewHubService(cfg)

	err := h.AutoRegister()
	require.Error(t, err)
	assert.Contains(t, err.Error(), "no forge token available")
}

func TestAutoRegister_Good_SkipsIfAlreadyRegistered(t *testing.T) {
	cfg := testConfigService(t, nil, nil)
	cfg.config.HubURL = "https://hub.example.com"
	cfg.config.HubToken = "existing-token"
	h := NewHubService(cfg)

	err := h.AutoRegister()
	require.NoError(t, err)

	// Token should remain unchanged.
	assert.Equal(t, "existing-token", h.config.GetHubToken())
}

// ---- Task 5: Write Operations ----

func TestRegister_Good(t *testing.T) {
	var gotPath string
	var gotMethod string
	var gotBody map[string]string

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotPath = r.URL.Path
		gotMethod = r.Method
		_ = json.NewDecoder(r.Body).Decode(&gotBody)
		w.WriteHeader(http.StatusOK)
	}))
	defer srv.Close()

	cfg := testConfigService(t, nil, nil)
	cfg.config.HubURL = srv.URL
	cfg.config.HubToken = "tok"
	cfg.config.ClientName = "MyBugSETI"
	h := NewHubService(cfg)

	err := h.Register()
	require.NoError(t, err)
	assert.Equal(t, "/api/bugseti/register", gotPath)
	assert.Equal(t, "POST", gotMethod)
	assert.Equal(t, "MyBugSETI", gotBody["name"])
	assert.NotEmpty(t, gotBody["client_id"])
	assert.NotEmpty(t, gotBody["version"])
	assert.NotEmpty(t, gotBody["os"])
	assert.NotEmpty(t, gotBody["arch"])
}

func TestHeartbeat_Good(t *testing.T) {
	var gotPath string
	var gotMethod string

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotPath = r.URL.Path
		gotMethod = r.Method
		w.WriteHeader(http.StatusOK)
	}))
	defer srv.Close()

	cfg := testConfigService(t, nil, nil)
	cfg.config.HubURL = srv.URL
	cfg.config.HubToken = "tok"
	h := NewHubService(cfg)

	err := h.Heartbeat()
	require.NoError(t, err)
	assert.Equal(t, "/api/bugseti/heartbeat", gotPath)
	assert.Equal(t, "POST", gotMethod)
}

func TestClaimIssue_Good(t *testing.T) {
	now := time.Now().Truncate(time.Second)
	expires := now.Add(30 * time.Minute)

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "/api/bugseti/issues/claim", r.URL.Path)
		assert.Equal(t, "POST", r.Method)

		var body map[string]interface{}
		_ = json.NewDecoder(r.Body).Decode(&body)
		assert.Equal(t, "issue-42", body["issue_id"])
		assert.Equal(t, "org/repo", body["repo"])
		assert.Equal(t, float64(42), body["issue_number"])
		assert.Equal(t, "Fix the bug", body["title"])

		w.WriteHeader(http.StatusOK)
		resp := HubClaim{
			ID:        "claim-1",
			IssueURL:  "https://github.com/org/repo/issues/42",
			ClientID:  "test",
			ClaimedAt: now,
			ExpiresAt: expires,
			Status:    "claimed",
		}
		_ = json.NewEncoder(w).Encode(resp)
	}))
	defer srv.Close()

	cfg := testConfigService(t, nil, nil)
	cfg.config.HubURL = srv.URL
	cfg.config.HubToken = "tok"
	h := NewHubService(cfg)

	issue := &Issue{
		ID:     "issue-42",
		Number: 42,
		Repo:   "org/repo",
		Title:  "Fix the bug",
		URL:    "https://github.com/org/repo/issues/42",
	}

	claim, err := h.ClaimIssue(issue)
	require.NoError(t, err)
	require.NotNil(t, claim)
	assert.Equal(t, "claim-1", claim.ID)
	assert.Equal(t, "claimed", claim.Status)
}

func TestClaimIssue_Bad_Conflict(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusConflict)
	}))
	defer srv.Close()

	cfg := testConfigService(t, nil, nil)
	cfg.config.HubURL = srv.URL
	cfg.config.HubToken = "tok"
	h := NewHubService(cfg)

	issue := &Issue{ID: "issue-99", Number: 99, Repo: "org/repo", Title: "Already claimed"}

	claim, err := h.ClaimIssue(issue)
	assert.Nil(t, claim)
	require.Error(t, err)

	var conflictErr *ConflictError
	assert.ErrorAs(t, err, &conflictErr)
}

func TestUpdateStatus_Good(t *testing.T) {
	var gotPath string
	var gotMethod string
	var gotBody map[string]interface{}

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotPath = r.URL.Path
		gotMethod = r.Method
		_ = json.NewDecoder(r.Body).Decode(&gotBody)
		w.WriteHeader(http.StatusOK)
	}))
	defer srv.Close()

	cfg := testConfigService(t, nil, nil)
	cfg.config.HubURL = srv.URL
	cfg.config.HubToken = "tok"
	h := NewHubService(cfg)

	err := h.UpdateStatus("issue-42", "completed", "https://github.com/org/repo/pull/10", 10)
	require.NoError(t, err)
	assert.Equal(t, "PATCH", gotMethod)
	assert.Equal(t, "/api/bugseti/issues/issue-42/status", gotPath)
	assert.Equal(t, "completed", gotBody["status"])
	assert.Equal(t, "https://github.com/org/repo/pull/10", gotBody["pr_url"])
	assert.Equal(t, float64(10), gotBody["pr_number"])
}

func TestSyncStats_Good(t *testing.T) {
	var gotBody map[string]interface{}

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "/api/bugseti/stats/sync", r.URL.Path)
		assert.Equal(t, "POST", r.Method)
		_ = json.NewDecoder(r.Body).Decode(&gotBody)
		w.WriteHeader(http.StatusOK)
	}))
	defer srv.Close()

	cfg := testConfigService(t, nil, nil)
	cfg.config.HubURL = srv.URL
	cfg.config.HubToken = "tok"
	h := NewHubService(cfg)

	stats := &Stats{
		IssuesAttempted: 10,
		IssuesCompleted: 7,
		IssuesSkipped:   3,
		PRsSubmitted:    6,
		PRsMerged:       5,
		PRsRejected:     1,
		CurrentStreak:   3,
		LongestStreak:   5,
		TotalTimeSpent:  90 * time.Minute,
		ReposContributed: map[string]*RepoStats{
			"org/repo-a": {Name: "org/repo-a"},
			"org/repo-b": {Name: "org/repo-b"},
		},
	}

	err := h.SyncStats(stats)
	require.NoError(t, err)

	assert.NotEmpty(t, gotBody["client_id"])
	statsMap, ok := gotBody["stats"].(map[string]interface{})
	require.True(t, ok)
	assert.Equal(t, float64(10), statsMap["issues_attempted"])
	assert.Equal(t, float64(7), statsMap["issues_completed"])
	assert.Equal(t, float64(3), statsMap["issues_skipped"])
	assert.Equal(t, float64(6), statsMap["prs_submitted"])
	assert.Equal(t, float64(5), statsMap["prs_merged"])
	assert.Equal(t, float64(1), statsMap["prs_rejected"])
	assert.Equal(t, float64(3), statsMap["current_streak"])
	assert.Equal(t, float64(5), statsMap["longest_streak"])
	assert.Equal(t, float64(90), statsMap["total_time_minutes"])

	reposRaw, ok := statsMap["repos_contributed"].([]interface{})
	require.True(t, ok)
	assert.Len(t, reposRaw, 2)
}

// ---- Task 6: Read Operations ----

func TestIsIssueClaimed_Good_Claimed(t *testing.T) {
	now := time.Now().Truncate(time.Second)

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "/api/bugseti/issues/issue-42", r.URL.Path)
		assert.Equal(t, "GET", r.Method)

		w.WriteHeader(http.StatusOK)
		claim := HubClaim{
			ID:        "claim-1",
			IssueURL:  "https://github.com/org/repo/issues/42",
			ClientID:  "client-abc",
			ClaimedAt: now,
			Status:    "claimed",
		}
		_ = json.NewEncoder(w).Encode(claim)
	}))
	defer srv.Close()

	cfg := testConfigService(t, nil, nil)
	cfg.config.HubURL = srv.URL
	cfg.config.HubToken = "tok"
	h := NewHubService(cfg)

	claim, err := h.IsIssueClaimed("issue-42")
	require.NoError(t, err)
	require.NotNil(t, claim)
	assert.Equal(t, "claim-1", claim.ID)
	assert.Equal(t, "claimed", claim.Status)
}

func TestIsIssueClaimed_Good_NotClaimed(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNotFound)
	}))
	defer srv.Close()

	cfg := testConfigService(t, nil, nil)
	cfg.config.HubURL = srv.URL
	cfg.config.HubToken = "tok"
	h := NewHubService(cfg)

	claim, err := h.IsIssueClaimed("issue-999")
	assert.NoError(t, err)
	assert.Nil(t, claim)
}

func TestGetLeaderboard_Good(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "/api/bugseti/leaderboard", r.URL.Path)
		assert.Equal(t, "GET", r.Method)
		assert.Equal(t, "10", r.URL.Query().Get("limit"))

		resp := leaderboardResponse{
			Entries: []LeaderboardEntry{
				{ClientID: "a", ClientName: "Alice", Score: 100, PRsMerged: 10, Rank: 1},
				{ClientID: "b", ClientName: "Bob", Score: 80, PRsMerged: 8, Rank: 2},
			},
			TotalParticipants: 42,
		}
		w.WriteHeader(http.StatusOK)
		_ = json.NewEncoder(w).Encode(resp)
	}))
	defer srv.Close()

	cfg := testConfigService(t, nil, nil)
	cfg.config.HubURL = srv.URL
	cfg.config.HubToken = "tok"
	h := NewHubService(cfg)

	entries, total, err := h.GetLeaderboard(10)
	require.NoError(t, err)
	assert.Equal(t, 42, total)
	require.Len(t, entries, 2)
	assert.Equal(t, "Alice", entries[0].ClientName)
	assert.Equal(t, 1, entries[0].Rank)
	assert.Equal(t, "Bob", entries[1].ClientName)
}

func TestGetGlobalStats_Good(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "/api/bugseti/stats", r.URL.Path)
		assert.Equal(t, "GET", r.Method)

		stats := GlobalStats{
			TotalClients:    100,
			TotalClaims:     500,
			TotalPRsMerged:  300,
			ActiveClaims:    25,
			IssuesAvailable: 150,
		}
		w.WriteHeader(http.StatusOK)
		_ = json.NewEncoder(w).Encode(stats)
	}))
	defer srv.Close()

	cfg := testConfigService(t, nil, nil)
	cfg.config.HubURL = srv.URL
	cfg.config.HubToken = "tok"
	h := NewHubService(cfg)

	stats, err := h.GetGlobalStats()
	require.NoError(t, err)
	require.NotNil(t, stats)
	assert.Equal(t, 100, stats.TotalClients)
	assert.Equal(t, 500, stats.TotalClaims)
	assert.Equal(t, 300, stats.TotalPRsMerged)
	assert.Equal(t, 25, stats.ActiveClaims)
	assert.Equal(t, 150, stats.IssuesAvailable)
}
