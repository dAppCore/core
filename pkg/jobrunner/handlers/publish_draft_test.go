package handlers

import (
	"context"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/host-uk/core/pkg/jobrunner"
)

func TestPublishDraft_Match_Good(t *testing.T) {
	h := NewPublishDraftHandler(nil, "")
	sig := &jobrunner.PipelineSignal{
		IsDraft:     true,
		PRState:     "OPEN",
		CheckStatus: "SUCCESS",
	}
	assert.True(t, h.Match(sig))
}

func TestPublishDraft_Match_Bad_NotDraft(t *testing.T) {
	h := NewPublishDraftHandler(nil, "")
	sig := &jobrunner.PipelineSignal{
		IsDraft:     false,
		PRState:     "OPEN",
		CheckStatus: "SUCCESS",
	}
	assert.False(t, h.Match(sig))
}

func TestPublishDraft_Match_Bad_ChecksFailing(t *testing.T) {
	h := NewPublishDraftHandler(nil, "")
	sig := &jobrunner.PipelineSignal{
		IsDraft:     true,
		PRState:     "OPEN",
		CheckStatus: "FAILURE",
	}
	assert.False(t, h.Match(sig))
}

func TestPublishDraft_Execute_Good(t *testing.T) {
	var capturedMethod string
	var capturedPath string
	var capturedBody string

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		capturedMethod = r.Method
		capturedPath = r.URL.Path
		b, _ := io.ReadAll(r.Body)
		capturedBody = string(b)
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{"draft":false}`))
	}))
	defer srv.Close()

	h := NewPublishDraftHandler(srv.Client(), srv.URL)
	sig := &jobrunner.PipelineSignal{
		RepoOwner: "host-uk",
		RepoName:  "core-php",
		PRNumber:  42,
		IsDraft:   true,
		PRState:   "OPEN",
	}

	result, err := h.Execute(context.Background(), sig)
	require.NoError(t, err)

	assert.Equal(t, http.MethodPatch, capturedMethod)
	assert.Equal(t, "/repos/host-uk/core-php/pulls/42", capturedPath)
	assert.Contains(t, capturedBody, `"draft":false`)

	assert.True(t, result.Success)
	assert.Equal(t, "publish_draft", result.Action)
	assert.Equal(t, "host-uk", result.RepoOwner)
	assert.Equal(t, "core-php", result.RepoName)
	assert.Equal(t, 42, result.PRNumber)
}
