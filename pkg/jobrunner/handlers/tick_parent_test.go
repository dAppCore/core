package handlers

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/host-uk/core/pkg/jobrunner"
)

func TestTickParent_Match_Good(t *testing.T) {
	h := NewTickParentHandler(nil)
	sig := &jobrunner.PipelineSignal{
		PRState: "MERGED",
	}
	assert.True(t, h.Match(sig))
}

func TestTickParent_Match_Bad_Open(t *testing.T) {
	h := NewTickParentHandler(nil)
	sig := &jobrunner.PipelineSignal{
		PRState: "OPEN",
	}
	assert.False(t, h.Match(sig))
}

func TestTickParent_Execute_Good(t *testing.T) {
	epicBody := "## Tasks\n- [x] #1\n- [ ] #7\n- [ ] #8\n"
	var editBody string
	var closeCalled bool

	srv := httptest.NewServer(withVersion(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		path := r.URL.Path
		method := r.Method
		w.Header().Set("Content-Type", "application/json")

		switch {
		// GET issue (fetch epic)
		case method == http.MethodGet && strings.Contains(path, "/issues/42"):
			_ = json.NewEncoder(w).Encode(map[string]any{
				"number": 42,
				"body":   epicBody,
				"title":  "Epic",
			})

		// PATCH issue (edit epic body)
		case method == http.MethodPatch && strings.Contains(path, "/issues/42"):
			b, _ := io.ReadAll(r.Body)
			editBody = string(b)
			_ = json.NewEncoder(w).Encode(map[string]any{
				"number": 42,
				"body":   editBody,
				"title":  "Epic",
			})

		// PATCH issue (close child — state: closed)
		case method == http.MethodPatch && strings.Contains(path, "/issues/7"):
			closeCalled = true
			_ = json.NewEncoder(w).Encode(map[string]any{
				"number": 7,
				"state":  "closed",
			})

		default:
			w.WriteHeader(http.StatusNotFound)
		}
	})))
	defer srv.Close()

	client := newTestForgeClient(t, srv.URL)

	h := NewTickParentHandler(client)
	sig := &jobrunner.PipelineSignal{
		RepoOwner:   "host-uk",
		RepoName:    "core-php",
		EpicNumber:  42,
		ChildNumber: 7,
		PRNumber:    99,
		PRState:     "MERGED",
	}

	result, err := h.Execute(context.Background(), sig)
	require.NoError(t, err)

	assert.True(t, result.Success)
	assert.Equal(t, "tick_parent", result.Action)

	// Verify the edit body contains the checked checkbox.
	assert.Contains(t, editBody, "- [x] #7")
	assert.True(t, closeCalled, "expected child issue to be closed")
}
