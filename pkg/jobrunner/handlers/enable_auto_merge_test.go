package handlers

import (
	"context"
	"os/exec"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/host-uk/core/pkg/jobrunner"
)

func TestEnableAutoMerge_Match_Good(t *testing.T) {
	h := NewEnableAutoMergeHandler()
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
	h := NewEnableAutoMergeHandler()
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
	h := NewEnableAutoMergeHandler()
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
	// Save and restore the original execCommand.
	original := execCommand
	defer func() { execCommand = original }()

	var capturedArgs []string
	execCommand = func(ctx context.Context, name string, args ...string) *exec.Cmd {
		capturedArgs = append([]string{name}, args...)
		return exec.CommandContext(ctx, "echo", append([]string{name}, args...)...)
	}

	h := NewEnableAutoMergeHandler()
	sig := &jobrunner.PipelineSignal{
		RepoOwner: "host-uk",
		RepoName:  "core-php",
		PRNumber:  55,
	}

	result, err := h.Execute(context.Background(), sig)
	require.NoError(t, err)

	assert.True(t, result.Success)
	assert.Equal(t, "enable_auto_merge", result.Action)

	joined := strings.Join(capturedArgs, " ")
	assert.Contains(t, joined, "--auto")
	assert.Contains(t, joined, "--squash")
	assert.Contains(t, joined, "55")
	assert.Contains(t, joined, "-R")
	assert.Contains(t, joined, "host-uk/core-php")
}
