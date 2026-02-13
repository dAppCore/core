package bugseti

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
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
