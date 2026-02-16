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

func TestDismissReviews_Match_Good(t *testing.T) {
	h := NewDismissReviewsHandler(nil)
	sig := &jobrunner.PipelineSignal{
		PRState:         "OPEN",
		ThreadsTotal:    4,
		ThreadsResolved: 2,
	}
	assert.True(t, h.Match(sig))
}

func TestDismissReviews_Match_Bad_AllResolved(t *testing.T) {
	h := NewDismissReviewsHandler(nil)
	sig := &jobrunner.PipelineSignal{
		PRState:         "OPEN",
		ThreadsTotal:    3,
		ThreadsResolved: 3,
	}
	assert.False(t, h.Match(sig))
}

func TestDismissReviews_Execute_Good(t *testing.T) {
	callCount := 0

	srv := httptest.NewServer(withVersion(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		callCount++
		w.Header().Set("Content-Type", "application/json")

		// ListPullReviews (GET)
		if r.Method == http.MethodGet {
			reviews := []map[string]any{
				{
					"id": 1, "state": "REQUEST_CHANGES", "dismissed": false, "stale": true,
					"body": "fix this", "commit_id": "abc123",
				},
				{
					"id": 2, "state": "APPROVED", "dismissed": false, "stale": false,
					"body": "looks good", "commit_id": "abc123",
				},
				{
					"id": 3, "state": "REQUEST_CHANGES", "dismissed": false, "stale": true,
					"body": "needs work", "commit_id": "abc123",
				},
			}
			_ = json.NewEncoder(w).Encode(reviews)
			return
		}

		// DismissPullReview (POST to dismissals endpoint)
		w.WriteHeader(http.StatusOK)
	})))
	defer srv.Close()

	client := newTestForgeClient(t, srv.URL)

	h := NewDismissReviewsHandler(client)
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
	assert.Equal(t, "dismiss_reviews", result.Action)
	assert.Equal(t, "host-uk", result.RepoOwner)
	assert.Equal(t, "core-admin", result.RepoName)
	assert.Equal(t, 33, result.PRNumber)

	// 1 list + 2 dismiss (reviews #1 and #3 are stale REQUEST_CHANGES)
	assert.Equal(t, 3, callCount)
}
