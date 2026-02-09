package agentci

import (
	"testing"

	"github.com/host-uk/core/pkg/config"
	"github.com/host-uk/core/pkg/io"
	"github.com/host-uk/core/pkg/jobrunner/handlers"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func newTestConfig(t *testing.T, yaml string) *config.Config {
	t.Helper()
	m := io.NewMockMedium()
	if yaml != "" {
		m.Files["/tmp/test/config.yaml"] = yaml
	}
	cfg, err := config.New(config.WithMedium(m), config.WithPath("/tmp/test/config.yaml"))
	require.NoError(t, err)
	return cfg
}

func TestLoadAgents_Good(t *testing.T) {
	cfg := newTestConfig(t, `
agentci:
  agents:
    darbs-claude:
      host: claude@192.168.0.201
      queue_dir: /home/claude/ai-work/queue
      forgejo_user: darbs-claude
      model: sonnet
      runner: claude
      active: true
`)
	targets, err := LoadAgents(cfg)
	require.NoError(t, err)
	require.Len(t, targets, 1)

	agent := targets["darbs-claude"]
	assert.Equal(t, "claude@192.168.0.201", agent.Host)
	assert.Equal(t, "/home/claude/ai-work/queue", agent.QueueDir)
	assert.Equal(t, "sonnet", agent.Model)
	assert.Equal(t, "claude", agent.Runner)
}

func TestLoadAgents_Good_MultipleAgents(t *testing.T) {
	cfg := newTestConfig(t, `
agentci:
  agents:
    darbs-claude:
      host: claude@192.168.0.201
      queue_dir: /home/claude/ai-work/queue
      active: true
    local-codex:
      host: localhost
      queue_dir: /home/claude/ai-work/queue
      runner: codex
      active: true
`)
	targets, err := LoadAgents(cfg)
	require.NoError(t, err)
	assert.Len(t, targets, 2)
	assert.Contains(t, targets, "darbs-claude")
	assert.Contains(t, targets, "local-codex")
}

func TestLoadAgents_Good_SkipsInactive(t *testing.T) {
	cfg := newTestConfig(t, `
agentci:
  agents:
    active-agent:
      host: claude@10.0.0.1
      active: true
    offline-agent:
      host: claude@10.0.0.2
      active: false
`)
	targets, err := LoadAgents(cfg)
	require.NoError(t, err)
	assert.Len(t, targets, 1)
	assert.Contains(t, targets, "active-agent")
}

func TestLoadAgents_Good_Defaults(t *testing.T) {
	cfg := newTestConfig(t, `
agentci:
  agents:
    minimal:
      host: claude@10.0.0.1
      active: true
`)
	targets, err := LoadAgents(cfg)
	require.NoError(t, err)
	require.Len(t, targets, 1)

	agent := targets["minimal"]
	assert.Equal(t, "/home/claude/ai-work/queue", agent.QueueDir)
	assert.Equal(t, "sonnet", agent.Model)
	assert.Equal(t, "claude", agent.Runner)
}

func TestLoadAgents_Good_NoConfig(t *testing.T) {
	cfg := newTestConfig(t, "")
	targets, err := LoadAgents(cfg)
	require.NoError(t, err)
	assert.Empty(t, targets)
}

func TestLoadAgents_Bad_MissingHost(t *testing.T) {
	cfg := newTestConfig(t, `
agentci:
  agents:
    broken:
      queue_dir: /tmp
      active: true
`)
	_, err := LoadAgents(cfg)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "host is required")
}

func TestLoadAgents_Good_ReturnsAgentTargets(t *testing.T) {
	cfg := newTestConfig(t, `
agentci:
  agents:
    gemini-agent:
      host: localhost
      runner: gemini
      model: ""
      active: true
`)
	targets, err := LoadAgents(cfg)
	require.NoError(t, err)

	agent := targets["gemini-agent"]
	// Verify it returns the handlers.AgentTarget type.
	var _ handlers.AgentTarget = agent
	assert.Equal(t, "gemini", agent.Runner)
	assert.Equal(t, "sonnet", agent.Model) // default when empty
}

func TestSaveAgent_Good(t *testing.T) {
	cfg := newTestConfig(t, "")

	err := SaveAgent(cfg, "new-agent", AgentConfig{
		Host:        "claude@10.0.0.5",
		QueueDir:    "/home/claude/ai-work/queue",
		ForgejoUser: "new-agent",
		Model:       "haiku",
		Runner:      "claude",
		Active:      true,
	})
	require.NoError(t, err)

	// Verify we can load it back.
	agents, err := ListAgents(cfg)
	require.NoError(t, err)
	require.Contains(t, agents, "new-agent")
	assert.Equal(t, "claude@10.0.0.5", agents["new-agent"].Host)
	assert.Equal(t, "haiku", agents["new-agent"].Model)
}

func TestSaveAgent_Good_OmitsEmptyOptionals(t *testing.T) {
	cfg := newTestConfig(t, "")

	err := SaveAgent(cfg, "minimal", AgentConfig{
		Host:   "claude@10.0.0.1",
		Active: true,
	})
	require.NoError(t, err)

	agents, err := ListAgents(cfg)
	require.NoError(t, err)
	assert.Contains(t, agents, "minimal")
}

func TestRemoveAgent_Good(t *testing.T) {
	cfg := newTestConfig(t, `
agentci:
  agents:
    to-remove:
      host: claude@10.0.0.1
      active: true
    to-keep:
      host: claude@10.0.0.2
      active: true
`)
	err := RemoveAgent(cfg, "to-remove")
	require.NoError(t, err)

	agents, err := ListAgents(cfg)
	require.NoError(t, err)
	assert.NotContains(t, agents, "to-remove")
	assert.Contains(t, agents, "to-keep")
}

func TestRemoveAgent_Bad_NotFound(t *testing.T) {
	cfg := newTestConfig(t, `
agentci:
  agents:
    existing:
      host: claude@10.0.0.1
      active: true
`)
	err := RemoveAgent(cfg, "nonexistent")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "not found")
}

func TestRemoveAgent_Bad_NoAgents(t *testing.T) {
	cfg := newTestConfig(t, "")
	err := RemoveAgent(cfg, "anything")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "no agents configured")
}

func TestListAgents_Good(t *testing.T) {
	cfg := newTestConfig(t, `
agentci:
  agents:
    agent-a:
      host: claude@10.0.0.1
      active: true
    agent-b:
      host: claude@10.0.0.2
      active: false
`)
	agents, err := ListAgents(cfg)
	require.NoError(t, err)
	assert.Len(t, agents, 2)
	assert.True(t, agents["agent-a"].Active)
	assert.False(t, agents["agent-b"].Active)
}

func TestListAgents_Good_Empty(t *testing.T) {
	cfg := newTestConfig(t, "")
	agents, err := ListAgents(cfg)
	require.NoError(t, err)
	assert.Empty(t, agents)
}

func TestRoundTrip_SaveThenLoad(t *testing.T) {
	cfg := newTestConfig(t, "")

	// Save two agents.
	err := SaveAgent(cfg, "alpha", AgentConfig{
		Host:        "claude@alpha",
		QueueDir:    "/home/claude/work/queue",
		ForgejoUser: "alpha-bot",
		Model:       "opus",
		Runner:      "claude",
		Active:      true,
	})
	require.NoError(t, err)

	err = SaveAgent(cfg, "beta", AgentConfig{
		Host:        "claude@beta",
		ForgejoUser: "beta-bot",
		Runner:      "codex",
		Active:      true,
	})
	require.NoError(t, err)

	// Load as AgentTargets (what the dispatch handler uses).
	targets, err := LoadAgents(cfg)
	require.NoError(t, err)
	assert.Len(t, targets, 2)
	assert.Equal(t, "claude@alpha", targets["alpha"].Host)
	assert.Equal(t, "opus", targets["alpha"].Model)
	assert.Equal(t, "codex", targets["beta"].Runner)
}
