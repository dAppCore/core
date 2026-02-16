package agentci

import (
	"testing"

	"forge.lthn.ai/core/cli/pkg/config"
	"forge.lthn.ai/core/cli/pkg/io"
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
	agents, err := LoadAgents(cfg)
	require.NoError(t, err)
	require.Len(t, agents, 1)

	agent := agents["darbs-claude"]
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
	agents, err := LoadAgents(cfg)
	require.NoError(t, err)
	assert.Len(t, agents, 2)
	assert.Contains(t, agents, "darbs-claude")
	assert.Contains(t, agents, "local-codex")
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
	agents, err := LoadAgents(cfg)
	require.NoError(t, err)
	// Both are returned, but only active-agent has defaults applied.
	assert.Len(t, agents, 2)
	assert.Contains(t, agents, "active-agent")
}

func TestLoadActiveAgents_Good(t *testing.T) {
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
	active, err := LoadActiveAgents(cfg)
	require.NoError(t, err)
	assert.Len(t, active, 1)
	assert.Contains(t, active, "active-agent")
}

func TestLoadAgents_Good_Defaults(t *testing.T) {
	cfg := newTestConfig(t, `
agentci:
  agents:
    minimal:
      host: claude@10.0.0.1
      active: true
`)
	agents, err := LoadAgents(cfg)
	require.NoError(t, err)
	require.Len(t, agents, 1)

	agent := agents["minimal"]
	assert.Equal(t, "/home/claude/ai-work/queue", agent.QueueDir)
	assert.Equal(t, "sonnet", agent.Model)
	assert.Equal(t, "claude", agent.Runner)
}

func TestLoadAgents_Good_NoConfig(t *testing.T) {
	cfg := newTestConfig(t, "")
	agents, err := LoadAgents(cfg)
	require.NoError(t, err)
	assert.Empty(t, agents)
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

func TestLoadAgents_Good_WithDualRun(t *testing.T) {
	cfg := newTestConfig(t, `
agentci:
  agents:
    gemini-agent:
      host: localhost
      runner: gemini
      model: gemini-2.0-flash
      verify_model: gemini-1.5-pro
      dual_run: true
      active: true
`)
	agents, err := LoadAgents(cfg)
	require.NoError(t, err)

	agent := agents["gemini-agent"]
	assert.Equal(t, "gemini", agent.Runner)
	assert.Equal(t, "gemini-2.0-flash", agent.Model)
	assert.Equal(t, "gemini-1.5-pro", agent.VerifyModel)
	assert.True(t, agent.DualRun)
}

func TestLoadClothoConfig_Good(t *testing.T) {
	cfg := newTestConfig(t, `
agentci:
  clotho:
    strategy: clotho-verified
    validation_threshold: 0.9
    signing_key_path: /etc/core/keys/clotho.pub
`)
	cc, err := LoadClothoConfig(cfg)
	require.NoError(t, err)
	assert.Equal(t, "clotho-verified", cc.Strategy)
	assert.Equal(t, 0.9, cc.ValidationThreshold)
	assert.Equal(t, "/etc/core/keys/clotho.pub", cc.SigningKeyPath)
}

func TestLoadClothoConfig_Good_Defaults(t *testing.T) {
	cfg := newTestConfig(t, "")
	cc, err := LoadClothoConfig(cfg)
	require.NoError(t, err)
	assert.Equal(t, "direct", cc.Strategy)
	assert.Equal(t, 0.85, cc.ValidationThreshold)
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

	agents, err := ListAgents(cfg)
	require.NoError(t, err)
	require.Contains(t, agents, "new-agent")
	assert.Equal(t, "claude@10.0.0.5", agents["new-agent"].Host)
	assert.Equal(t, "haiku", agents["new-agent"].Model)
}

func TestSaveAgent_Good_WithDualRun(t *testing.T) {
	cfg := newTestConfig(t, "")

	err := SaveAgent(cfg, "verified-agent", AgentConfig{
		Host:        "claude@10.0.0.5",
		Model:       "gemini-2.0-flash",
		VerifyModel: "gemini-1.5-pro",
		DualRun:     true,
		Active:      true,
	})
	require.NoError(t, err)

	agents, err := ListAgents(cfg)
	require.NoError(t, err)
	require.Contains(t, agents, "verified-agent")
	assert.True(t, agents["verified-agent"].DualRun)
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

	agents, err := LoadActiveAgents(cfg)
	require.NoError(t, err)
	assert.Len(t, agents, 2)
	assert.Equal(t, "claude@alpha", agents["alpha"].Host)
	assert.Equal(t, "opus", agents["alpha"].Model)
	assert.Equal(t, "codex", agents["beta"].Runner)
}
