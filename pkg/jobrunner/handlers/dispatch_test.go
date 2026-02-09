package handlers

import (
	"testing"

	"github.com/host-uk/core/pkg/jobrunner"
	"github.com/stretchr/testify/assert"
)

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
