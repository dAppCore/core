package handlers

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"forge.lthn.ai/core/cli/pkg/jobrunner"
)

func TestEnableAutoMerge_Match_Good(t *testing.T) {
	h := NewEnableAutoMergeHandler(nil)
	sig := &jobrunner.PipelineSignal{
		PRState:         "OPEN",
		IsDraft:         false,
		Mergeable:       "MERGEABLE",
		CheckStatus:     "SUCCESS",
		ThreadsTotal:    0,
		ThreadsResolved: 0,
	}
	assert.True(t, h.Match(sig))
}

func TestEnableAutoMerge_Match_Bad_Draft(t *testing.T) {
	h := NewEnableAutoMergeHandler(nil)
	sig := &jobrunner.PipelineSignal{
		PRState:         "OPEN",
		IsDraft:         true,
		Mergeable:       "MERGEABLE",
		CheckStatus:     "SUCCESS",
		ThreadsTotal:    0,
		ThreadsResolved: 0,
	}
	assert.False(t, h.Match(sig))
}

func TestEnableAutoMerge_Match_Bad_UnresolvedThreads(t *testing.T) {
	h := NewEnableAutoMergeHandler(nil)
	sig := &jobrunner.PipelineSignal{
		PRState:         "OPEN",
		IsDraft:         false,
		Mergeable:       "MERGEABLE",
		CheckStatus:     "SUCCESS",
		ThreadsTotal:    5,
		ThreadsResolved: 3,
	}
	assert.False(t, h.Match(sig))
}

func TestEnableAutoMerge_Execute_Good(t *testing.T) {
	var capturedPath string
	var capturedMethod string

	srv := httptest.NewServer(withVersion(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		capturedMethod = r.Method
		capturedPath = r.URL.Path
		w.WriteHeader(http.StatusOK)
	})))
	defer srv.Close()

	client := newTestForgeClient(t, srv.URL)

	h := NewEnableAutoMergeHandler(client)
	sig := &jobrunner.PipelineSignal{
		RepoOwner: "host-uk",
		RepoName:  "core-php",
		PRNumber:  55,
	}

	result, err := h.Execute(context.Background(), sig)
	require.NoError(t, err)

	assert.True(t, result.Success)
	assert.Equal(t, "enable_auto_merge", result.Action)
	assert.Equal(t, http.MethodPost, capturedMethod)
	assert.Equal(t, "/api/v1/repos/host-uk/core-php/pulls/55/merge", capturedPath)
}

func TestEnableAutoMerge_Execute_Bad_MergeFailed(t *testing.T) {
	srv := httptest.NewServer(withVersion(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusConflict)
		_ = json.NewEncoder(w).Encode(map[string]string{"message": "merge conflict"})
	})))
	defer srv.Close()

	client := newTestForgeClient(t, srv.URL)

	h := NewEnableAutoMergeHandler(client)
	sig := &jobrunner.PipelineSignal{
		RepoOwner: "host-uk",
		RepoName:  "core-php",
		PRNumber:  55,
	}

	result, err := h.Execute(context.Background(), sig)
	require.NoError(t, err)

	assert.False(t, result.Success)
	assert.Contains(t, result.Error, "merge failed")
}
