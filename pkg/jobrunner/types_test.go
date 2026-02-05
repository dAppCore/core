package jobrunner

import (
	"encoding/json"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestPipelineSignal_RepoFullName_Good(t *testing.T) {
	sig := &PipelineSignal{
		RepoOwner: "host-uk",
		RepoName:  "core-php",
	}
	assert.Equal(t, "host-uk/core-php", sig.RepoFullName())
}

func TestPipelineSignal_HasUnresolvedThreads_Good(t *testing.T) {
	sig := &PipelineSignal{
		ThreadsTotal:    5,
		ThreadsResolved: 3,
	}
	assert.True(t, sig.HasUnresolvedThreads())
}

func TestPipelineSignal_HasUnresolvedThreads_Bad_AllResolved(t *testing.T) {
	sig := &PipelineSignal{
		ThreadsTotal:    4,
		ThreadsResolved: 4,
	}
	assert.False(t, sig.HasUnresolvedThreads())

	// Also verify zero threads is not unresolved.
	sigZero := &PipelineSignal{
		ThreadsTotal:    0,
		ThreadsResolved: 0,
	}
	assert.False(t, sigZero.HasUnresolvedThreads())
}

func TestActionResult_JSON_Good(t *testing.T) {
	ts := time.Date(2026, 2, 5, 12, 0, 0, 0, time.UTC)
	result := &ActionResult{
		Action:      "merge",
		RepoOwner:   "host-uk",
		RepoName:    "core-tenant",
		EpicNumber:  42,
		ChildNumber: 7,
		PRNumber:    99,
		Success:     true,
		Timestamp:   ts,
		Duration:    1500 * time.Millisecond,
		Cycle:       3,
	}

	data, err := json.Marshal(result)
	require.NoError(t, err)

	var decoded map[string]any
	err = json.Unmarshal(data, &decoded)
	require.NoError(t, err)

	assert.Equal(t, "merge", decoded["action"])
	assert.Equal(t, "host-uk", decoded["repo_owner"])
	assert.Equal(t, "core-tenant", decoded["repo_name"])
	assert.Equal(t, float64(42), decoded["epic"])
	assert.Equal(t, float64(7), decoded["child"])
	assert.Equal(t, float64(99), decoded["pr"])
	assert.Equal(t, true, decoded["success"])
	assert.Equal(t, float64(3), decoded["cycle"])

	// Error field should be omitted when empty.
	_, hasError := decoded["error"]
	assert.False(t, hasError, "error field should be omitted when empty")

	// Verify round-trip with error field present.
	resultWithErr := &ActionResult{
		Action:    "merge",
		RepoOwner: "host-uk",
		RepoName:  "core-tenant",
		Success:   false,
		Error:     "checks failing",
		Timestamp: ts,
		Duration:  200 * time.Millisecond,
		Cycle:     1,
	}
	data2, err := json.Marshal(resultWithErr)
	require.NoError(t, err)

	var decoded2 map[string]any
	err = json.Unmarshal(data2, &decoded2)
	require.NoError(t, err)

	assert.Equal(t, "checks failing", decoded2["error"])
	assert.Equal(t, false, decoded2["success"])
}
