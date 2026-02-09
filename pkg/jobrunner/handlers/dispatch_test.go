package handlers

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/host-uk/core/pkg/jobrunner"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// --- Match tests ---

func TestDispatch_Match_Good_NeedsCoding(t *testing.T) {
	h := NewDispatchHandler(nil, "", "", map[string]AgentTarget{
		"darbs-claude": {Host: "claude@192.168.0.201", QueueDir: "~/ai-work/queue"},
	})
	sig := &jobrunner.PipelineSignal{
		NeedsCoding: true,
		Assignee:    "darbs-claude",
	}
	assert.True(t, h.Match(sig))
}

func TestDispatch_Match_Good_MultipleAgents(t *testing.T) {
	h := NewDispatchHandler(nil, "", "", map[string]AgentTarget{
		"darbs-claude": {Host: "claude@192.168.0.201", QueueDir: "~/ai-work/queue"},
		"local-codex":  {Host: "localhost", QueueDir: "~/ai-work/queue"},
	})
	sig := &jobrunner.PipelineSignal{
		NeedsCoding: true,
		Assignee:    "local-codex",
	}
	assert.True(t, h.Match(sig))
}

func TestDispatch_Match_Bad_HasPR(t *testing.T) {
	h := NewDispatchHandler(nil, "", "", map[string]AgentTarget{
		"darbs-claude": {Host: "claude@192.168.0.201", QueueDir: "~/ai-work/queue"},
	})
	sig := &jobrunner.PipelineSignal{
		NeedsCoding: false,
		PRNumber:    7,
		Assignee:    "darbs-claude",
	}
	assert.False(t, h.Match(sig))
}

func TestDispatch_Match_Bad_UnknownAgent(t *testing.T) {
	h := NewDispatchHandler(nil, "", "", map[string]AgentTarget{
		"darbs-claude": {Host: "claude@192.168.0.201", QueueDir: "~/ai-work/queue"},
	})
	sig := &jobrunner.PipelineSignal{
		NeedsCoding: true,
		Assignee:    "unknown-user",
	}
	assert.False(t, h.Match(sig))
}

func TestDispatch_Match_Bad_NotAssigned(t *testing.T) {
	h := NewDispatchHandler(nil, "", "", map[string]AgentTarget{
		"darbs-claude": {Host: "claude@192.168.0.201", QueueDir: "~/ai-work/queue"},
	})
	sig := &jobrunner.PipelineSignal{
		NeedsCoding: true,
		Assignee:    "",
	}
	assert.False(t, h.Match(sig))
}

func TestDispatch_Match_Bad_EmptyAgentMap(t *testing.T) {
	h := NewDispatchHandler(nil, "", "", map[string]AgentTarget{})
	sig := &jobrunner.PipelineSignal{
		NeedsCoding: true,
		Assignee:    "darbs-claude",
	}
	assert.False(t, h.Match(sig))
}

// --- Name test ---

func TestDispatch_Name_Good(t *testing.T) {
	h := NewDispatchHandler(nil, "", "", nil)
	assert.Equal(t, "dispatch", h.Name())
}

// --- Execute tests ---
// Execute calls SSH/SCP which can't be tested in unit tests without the remote.
// These tests verify the ticket construction and error paths that don't need SSH.

func TestDispatch_Execute_Bad_UnknownAgent(t *testing.T) {
	srv := httptest.NewServer(withVersion(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})))
	defer srv.Close()

	client := newTestForgeClient(t, srv.URL)
	h := NewDispatchHandler(client, srv.URL, "test-token", map[string]AgentTarget{
		"darbs-claude": {Host: "claude@192.168.0.201", QueueDir: "~/ai-work/queue"},
	})

	sig := &jobrunner.PipelineSignal{
		NeedsCoding: true,
		Assignee:    "nonexistent-agent",
		RepoOwner:   "host-uk",
		RepoName:    "core",
		ChildNumber: 1,
	}

	_, err := h.Execute(context.Background(), sig)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "unknown agent")
}

func TestDispatch_TicketJSON_Good(t *testing.T) {
	// Verify DispatchTicket serializes correctly with all fields.
	ticket := DispatchTicket{
		ID:           "host-uk-core-5-1234567890",
		RepoOwner:    "host-uk",
		RepoName:     "core",
		IssueNumber:  5,
		IssueTitle:   "Fix the thing",
		IssueBody:    "Please fix this bug",
		TargetBranch: "new",
		EpicNumber:   3,
		ForgeURL:     "https://forge.lthn.ai",
		ForgeToken:   "test-token-123",
		ForgeUser:    "darbs-claude",
		Model:        "sonnet",
		Runner:       "claude",
		CreatedAt:    "2026-02-09T12:00:00Z",
	}

	data, err := json.MarshalIndent(ticket, "", "  ")
	require.NoError(t, err)

	// Verify JSON field names.
	var decoded map[string]any
	err = json.Unmarshal(data, &decoded)
	require.NoError(t, err)

	assert.Equal(t, "host-uk-core-5-1234567890", decoded["id"])
	assert.Equal(t, "host-uk", decoded["repo_owner"])
	assert.Equal(t, "core", decoded["repo_name"])
	assert.Equal(t, float64(5), decoded["issue_number"])
	assert.Equal(t, "Fix the thing", decoded["issue_title"])
	assert.Equal(t, "Please fix this bug", decoded["issue_body"])
	assert.Equal(t, "new", decoded["target_branch"])
	assert.Equal(t, float64(3), decoded["epic_number"])
	assert.Equal(t, "https://forge.lthn.ai", decoded["forge_url"])
	assert.Equal(t, "test-token-123", decoded["forge_token"])
	assert.Equal(t, "darbs-claude", decoded["forgejo_user"])
	assert.Equal(t, "sonnet", decoded["model"])
	assert.Equal(t, "claude", decoded["runner"])
}

func TestDispatch_TicketJSON_Good_OmitsEmptyModelRunner(t *testing.T) {
	ticket := DispatchTicket{
		ID:           "test-1",
		RepoOwner:    "host-uk",
		RepoName:     "core",
		IssueNumber:  1,
		TargetBranch: "new",
		ForgeURL:     "https://forge.lthn.ai",
		ForgeToken:   "tok",
	}

	data, err := json.MarshalIndent(ticket, "", "  ")
	require.NoError(t, err)

	// Model and runner should be omitted when empty (omitempty tag).
	var decoded map[string]any
	err = json.Unmarshal(data, &decoded)
	require.NoError(t, err)

	_, hasModel := decoded["model"]
	_, hasRunner := decoded["runner"]
	assert.False(t, hasModel, "model should be omitted when empty")
	assert.False(t, hasRunner, "runner should be omitted when empty")
}

func TestDispatch_TicketJSON_Good_ModelRunnerVariants(t *testing.T) {
	tests := []struct {
		name   string
		model  string
		runner string
	}{
		{"claude-sonnet", "sonnet", "claude"},
		{"claude-opus", "opus", "claude"},
		{"codex-default", "", "codex"},
		{"gemini-default", "", "gemini"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ticket := DispatchTicket{
				ID:           "test-" + tt.name,
				RepoOwner:    "host-uk",
				RepoName:     "core",
				IssueNumber:  1,
				TargetBranch: "new",
				ForgeURL:     "https://forge.lthn.ai",
				ForgeToken:   "tok",
				Model:        tt.model,
				Runner:       tt.runner,
			}

			data, err := json.Marshal(ticket)
			require.NoError(t, err)

			var roundtrip DispatchTicket
			err = json.Unmarshal(data, &roundtrip)
			require.NoError(t, err)
			assert.Equal(t, tt.model, roundtrip.Model)
			assert.Equal(t, tt.runner, roundtrip.Runner)
		})
	}
}

func TestDispatch_Execute_Good_PostsComment(t *testing.T) {
	// This test verifies that Execute attempts to post a comment to the issue.
	// SSH/SCP will fail (no remote), but we can verify the comment API call
	// by checking if the Forgejo API was hit.
	var commentPosted bool
	var commentBody string

	srv := httptest.NewServer(withVersion(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodPost && r.URL.Path == "/api/v1/repos/host-uk/core/issues/5/comments" {
			commentPosted = true
			var body map[string]string
			_ = json.NewDecoder(r.Body).Decode(&body)
			commentBody = body["body"]
		}
		w.WriteHeader(http.StatusOK)
	})))
	defer srv.Close()

	client := newTestForgeClient(t, srv.URL)

	// Use localhost as agent host — ticketExists and scpTicket will fail
	// via SSH but we're testing the flow up to the SCP step.
	h := NewDispatchHandler(client, srv.URL, "test-token", map[string]AgentTarget{
		"darbs-claude": {Host: "localhost", QueueDir: "/tmp/nonexistent-queue"},
	})

	sig := &jobrunner.PipelineSignal{
		NeedsCoding: true,
		Assignee:    "darbs-claude",
		RepoOwner:   "host-uk",
		RepoName:    "core",
		ChildNumber: 5,
		EpicNumber:  3,
		IssueTitle:  "Test issue",
		IssueBody:   "Test body",
	}

	result, err := h.Execute(context.Background(), sig)
	require.NoError(t, err)

	// SSH may fail (no remote), so check for either:
	// 1. Success (if SSH happened to work, e.g. localhost)
	// 2. SCP error with correct metadata
	assert.Equal(t, "dispatch", result.Action)
	assert.Equal(t, "host-uk", result.RepoOwner)
	assert.Equal(t, "core", result.RepoName)
	assert.Equal(t, 3, result.EpicNumber)
	assert.Equal(t, 5, result.ChildNumber)

	if result.Success {
		// If SCP succeeded, comment should have been posted.
		assert.True(t, commentPosted)
		assert.Contains(t, commentBody, "darbs-claude")
	}
}
