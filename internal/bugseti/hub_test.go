package bugseti

import (
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
