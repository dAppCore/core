package handlers

import (
	"context"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"forge.lthn.ai/core/go/pkg/jobrunner"
)

func TestSendFixCommand_Match_Good_Conflicting(t *testing.T) {
	h := NewSendFixCommandHandler(nil)
	sig := &jobrunner.PipelineSignal{
		PRState:   "OPEN",
		Mergeable: "CONFLICTING",
	}
	assert.True(t, h.Match(sig))
}

func TestSendFixCommand_Match_Good_UnresolvedThreads(t *testing.T) {
	h := NewSendFixCommandHandler(nil)
	sig := &jobrunner.PipelineSignal{
		PRState:         "OPEN",
		Mergeable:       "MERGEABLE",
		CheckStatus:     "FAILURE",
		ThreadsTotal:    3,
		ThreadsResolved: 1,
	}
	assert.True(t, h.Match(sig))
}

func TestSendFixCommand_Match_Bad_Clean(t *testing.T) {
	h := NewSendFixCommandHandler(nil)
	sig := &jobrunner.PipelineSignal{
		PRState:         "OPEN",
		Mergeable:       "MERGEABLE",
		CheckStatus:     "SUCCESS",
		ThreadsTotal:    2,
		ThreadsResolved: 2,
	}
	assert.False(t, h.Match(sig))
}

func TestSendFixCommand_Execute_Good_Conflict(t *testing.T) {
	var capturedMethod string
	var capturedPath string
	var capturedBody string

	srv := httptest.NewServer(withVersion(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		capturedMethod = r.Method
		capturedPath = r.URL.Path
		b, _ := io.ReadAll(r.Body)
		capturedBody = string(b)
		w.WriteHeader(http.StatusCreated)
		_, _ = w.Write([]byte(`{"id":1}`))
	})))
	defer srv.Close()

	client := newTestForgeClient(t, srv.URL)

	h := NewSendFixCommandHandler(client)
	sig := &jobrunner.PipelineSignal{
		RepoOwner: "host-uk",
		RepoName:  "core-tenant",
		PRNumber:  17,
		PRState:   "OPEN",
		Mergeable: "CONFLICTING",
	}

	result, err := h.Execute(context.Background(), sig)
	require.NoError(t, err)

	assert.Equal(t, http.MethodPost, capturedMethod)
	assert.Equal(t, "/api/v1/repos/host-uk/core-tenant/issues/17/comments", capturedPath)
	assert.Contains(t, capturedBody, "fix the merge conflict")

	assert.True(t, result.Success)
	assert.Equal(t, "send_fix_command", result.Action)
	assert.Equal(t, "host-uk", result.RepoOwner)
	assert.Equal(t, "core-tenant", result.RepoName)
	assert.Equal(t, 17, result.PRNumber)
}
