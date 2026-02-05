package handlers

import (
	"context"
	"fmt"
	"os/exec"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/host-uk/core/pkg/jobrunner"
)

func TestTickParent_Match_Good(t *testing.T) {
	h := NewTickParentHandler()
	sig := &jobrunner.PipelineSignal{
		PRState: "MERGED",
	}
	assert.True(t, h.Match(sig))
}

func TestTickParent_Match_Bad_Open(t *testing.T) {
	h := NewTickParentHandler()
	sig := &jobrunner.PipelineSignal{
		PRState: "OPEN",
	}
	assert.False(t, h.Match(sig))
}

func TestTickParent_Execute_Good(t *testing.T) {
	// Save and restore the original execCommand.
	original := execCommand
	defer func() { execCommand = original }()

	epicBody := "## Tasks\n- [x] #1\n- [ ] #7\n- [ ] #8\n"
	var callCount int
	var editArgs []string
	var closeArgs []string

	execCommand = func(ctx context.Context, name string, args ...string) *exec.Cmd {
		callCount++
		if callCount == 1 {
			// First call: gh issue view — return the epic body.
			return exec.CommandContext(ctx, "echo", "-n", epicBody)
		}
		if callCount == 2 {
			// Second call: gh issue edit — capture args and succeed.
			editArgs = append([]string{name}, args...)
			return exec.CommandContext(ctx, "echo", "ok")
		}
		// Third call: gh issue close — capture args and succeed.
		closeArgs = append([]string{name}, args...)
		return exec.CommandContext(ctx, "echo", "ok")
	}

	h := NewTickParentHandler()
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
	assert.Equal(t, 3, callCount, "expected three exec calls: view + edit + close")

	// Verify the edit args contain the checked checkbox.
	editJoined := strings.Join(editArgs, " ")
	assert.Contains(t, editJoined, "issue")
	assert.Contains(t, editJoined, "edit")
	assert.Contains(t, editJoined, "42")
	assert.Contains(t, editJoined, fmt.Sprintf("-R %s", sig.RepoFullName()))
	assert.Contains(t, editJoined, "- [x] #7")

	// Verify the close args target the child issue.
	closeJoined := strings.Join(closeArgs, " ")
	assert.Contains(t, closeJoined, "issue")
	assert.Contains(t, closeJoined, "close")
	assert.Contains(t, closeJoined, "7")
	assert.Contains(t, closeJoined, "-R")
	assert.Contains(t, closeJoined, "host-uk/core-php")
}
