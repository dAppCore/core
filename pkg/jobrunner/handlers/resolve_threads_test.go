package handlers

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/host-uk/core/pkg/jobrunner"
)

func TestResolveThreads_Match_Good(t *testing.T) {
	h := NewResolveThreadsHandler(nil, "")
	sig := &jobrunner.PipelineSignal{
		PRState:         "OPEN",
		ThreadsTotal:    4,
		ThreadsResolved: 2,
	}
	assert.True(t, h.Match(sig))
}

func TestResolveThreads_Match_Bad_AllResolved(t *testing.T) {
	h := NewResolveThreadsHandler(nil, "")
	sig := &jobrunner.PipelineSignal{
		PRState:         "OPEN",
		ThreadsTotal:    3,
		ThreadsResolved: 3,
	}
	assert.False(t, h.Match(sig))
}

func TestResolveThreads_Execute_Good(t *testing.T) {
	callCount := 0

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		b, _ := io.ReadAll(r.Body)
		var gqlReq graphqlRequest
		_ = json.Unmarshal(b, &gqlReq)

		callCount++

		if callCount == 1 {
			// First call: fetch threads query.
			resp := threadsResponse{}
			resp.Data.Repository.PullRequest.ReviewThreads.Nodes = []struct {
				ID         string `json:"id"`
				IsResolved bool   `json:"isResolved"`
			}{
				{ID: "thread-1", IsResolved: false},
				{ID: "thread-2", IsResolved: true},
				{ID: "thread-3", IsResolved: false},
			}
			w.Header().Set("Content-Type", "application/json")
			_ = json.NewEncoder(w).Encode(resp)
			return
		}

		// Subsequent calls: resolve mutation.
		resp := resolveResponse{}
		resp.Data.ResolveReviewThread.Thread.ID = gqlReq.Variables["threadId"].(string)
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(resp)
	}))
	defer srv.Close()

	h := NewResolveThreadsHandler(srv.Client(), srv.URL)
	sig := &jobrunner.PipelineSignal{
		RepoOwner:       "host-uk",
		RepoName:        "core-admin",
		PRNumber:        33,
		PRState:         "OPEN",
		ThreadsTotal:    3,
		ThreadsResolved: 1,
	}

	result, err := h.Execute(context.Background(), sig)
	require.NoError(t, err)

	assert.True(t, result.Success)
	assert.Equal(t, "resolve_threads", result.Action)
	assert.Equal(t, "host-uk", result.RepoOwner)
	assert.Equal(t, "core-admin", result.RepoName)
	assert.Equal(t, 33, result.PRNumber)

	// 1 query + 2 mutations (thread-1 and thread-3 are unresolved).
	assert.Equal(t, 3, callCount)
}
