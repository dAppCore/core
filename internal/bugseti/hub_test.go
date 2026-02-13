package bugseti

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"

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

// ---- AutoRegister ----

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
