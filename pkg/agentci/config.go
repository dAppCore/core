// Package agentci provides configuration and management for AgentCI dispatch targets.
package agentci

import (
	"fmt"

	"github.com/host-uk/core/pkg/config"
	"github.com/host-uk/core/pkg/jobrunner/handlers"
)

// AgentConfig represents a single agent machine in the config file.
type AgentConfig struct {
	Host        string `yaml:"host" mapstructure:"host"`
	QueueDir    string `yaml:"queue_dir" mapstructure:"queue_dir"`
	ForgejoUser string `yaml:"forgejo_user" mapstructure:"forgejo_user"`
	Model       string `yaml:"model" mapstructure:"model"`   // claude model: sonnet, haiku, opus (default: sonnet)
	Runner      string `yaml:"runner" mapstructure:"runner"` // runner binary: claude, codex (default: claude)
	Active      bool   `yaml:"active" mapstructure:"active"`
}

// LoadAgents reads agent targets from config and returns a map suitable for the dispatch handler.
// Returns an empty map (not an error) if no agents are configured.
func LoadAgents(cfg *config.Config) (map[string]handlers.AgentTarget, error) {
	var agents map[string]AgentConfig
	if err := cfg.Get("agentci.agents", &agents); err != nil {
		// No config is fine — just no agents.
		return map[string]handlers.AgentTarget{}, nil
	}

	targets := make(map[string]handlers.AgentTarget)
	for name, ac := range agents {
		if !ac.Active {
			continue
		}
		if ac.Host == "" {
			return nil, fmt.Errorf("agent %q: host is required", name)
		}
		queueDir := ac.QueueDir
		if queueDir == "" {
			queueDir = "/home/claude/ai-work/queue"
		}
		model := ac.Model
		if model == "" {
			model = "sonnet"
		}
		runner := ac.Runner
		if runner == "" {
			runner = "claude"
		}
		targets[name] = handlers.AgentTarget{
			Host:     ac.Host,
			QueueDir: queueDir,
			Model:    model,
			Runner:   runner,
		}
	}

	return targets, nil
}

// SaveAgent writes an agent config entry to the config file.
func SaveAgent(cfg *config.Config, name string, ac AgentConfig) error {
	key := fmt.Sprintf("agentci.agents.%s", name)
	data := map[string]any{
		"host":         ac.Host,
		"queue_dir":    ac.QueueDir,
		"forgejo_user": ac.ForgejoUser,
		"active":       ac.Active,
	}
	if ac.Model != "" {
		data["model"] = ac.Model
	}
	if ac.Runner != "" {
		data["runner"] = ac.Runner
	}
	return cfg.Set(key, data)
}

// RemoveAgent removes an agent from the config file.
func RemoveAgent(cfg *config.Config, name string) error {
	var agents map[string]AgentConfig
	if err := cfg.Get("agentci.agents", &agents); err != nil {
		return fmt.Errorf("no agents configured")
	}
	if _, ok := agents[name]; !ok {
		return fmt.Errorf("agent %q not found", name)
	}
	delete(agents, name)
	return cfg.Set("agentci.agents", agents)
}

// ListAgents returns all configured agents (active and inactive).
func ListAgents(cfg *config.Config) (map[string]AgentConfig, error) {
	var agents map[string]AgentConfig
	if err := cfg.Get("agentci.agents", &agents); err != nil {
		return map[string]AgentConfig{}, nil
	}
	return agents, nil
}
