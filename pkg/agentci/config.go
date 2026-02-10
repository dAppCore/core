// Package agentci provides configuration, security, and orchestration for AgentCI dispatch targets.
package agentci

import (
	"fmt"

	"github.com/host-uk/core/pkg/config"
)

// AgentConfig represents a single agent machine in the config file.
type AgentConfig struct {
	Host          string   `yaml:"host" mapstructure:"host"`
	QueueDir      string   `yaml:"queue_dir" mapstructure:"queue_dir"`
	ForgejoUser   string   `yaml:"forgejo_user" mapstructure:"forgejo_user"`
	Model         string   `yaml:"model" mapstructure:"model"`                   // primary AI model
	Runner        string   `yaml:"runner" mapstructure:"runner"`                 // runner binary: claude, codex, gemini
	VerifyModel   string   `yaml:"verify_model" mapstructure:"verify_model"`     // secondary model for dual-run
	SecurityLevel string   `yaml:"security_level" mapstructure:"security_level"` // low, high
	Roles         []string `yaml:"roles" mapstructure:"roles"`
	DualRun       bool     `yaml:"dual_run" mapstructure:"dual_run"`
	Active        bool     `yaml:"active" mapstructure:"active"`
}

// ClothoConfig controls the orchestration strategy.
type ClothoConfig struct {
	Strategy            string  `yaml:"strategy" mapstructure:"strategy"`                         // direct, clotho-verified
	ValidationThreshold float64 `yaml:"validation_threshold" mapstructure:"validation_threshold"` // divergence limit (0.0-1.0)
	SigningKeyPath      string  `yaml:"signing_key_path" mapstructure:"signing_key_path"`
}

// LoadAgents reads agent targets from config and returns a map of AgentConfig.
// Returns an empty map (not an error) if no agents are configured.
func LoadAgents(cfg *config.Config) (map[string]AgentConfig, error) {
	var agents map[string]AgentConfig
	if err := cfg.Get("agentci.agents", &agents); err != nil {
		return map[string]AgentConfig{}, nil
	}

	// Validate and apply defaults.
	for name, ac := range agents {
		if !ac.Active {
			continue
		}
		if ac.Host == "" {
			return nil, fmt.Errorf("agent %q: host is required", name)
		}
		if ac.QueueDir == "" {
			ac.QueueDir = "/home/claude/ai-work/queue"
		}
		if ac.Model == "" {
			ac.Model = "sonnet"
		}
		if ac.Runner == "" {
			ac.Runner = "claude"
		}
		agents[name] = ac
	}

	return agents, nil
}

// LoadActiveAgents returns only active agents.
func LoadActiveAgents(cfg *config.Config) (map[string]AgentConfig, error) {
	all, err := LoadAgents(cfg)
	if err != nil {
		return nil, err
	}
	active := make(map[string]AgentConfig)
	for name, ac := range all {
		if ac.Active {
			active[name] = ac
		}
	}
	return active, nil
}

// LoadClothoConfig loads the Clotho orchestrator settings.
// Returns sensible defaults if no config is present.
func LoadClothoConfig(cfg *config.Config) (ClothoConfig, error) {
	var cc ClothoConfig
	if err := cfg.Get("agentci.clotho", &cc); err != nil {
		return ClothoConfig{
			Strategy:            "direct",
			ValidationThreshold: 0.85,
		}, nil
	}
	if cc.Strategy == "" {
		cc.Strategy = "direct"
	}
	if cc.ValidationThreshold == 0 {
		cc.ValidationThreshold = 0.85
	}
	return cc, nil
}

// SaveAgent writes an agent config entry to the config file.
func SaveAgent(cfg *config.Config, name string, ac AgentConfig) error {
	key := fmt.Sprintf("agentci.agents.%s", name)
	data := map[string]any{
		"host":         ac.Host,
		"queue_dir":    ac.QueueDir,
		"forgejo_user": ac.ForgejoUser,
		"active":       ac.Active,
		"dual_run":     ac.DualRun,
	}
	if ac.Model != "" {
		data["model"] = ac.Model
	}
	if ac.Runner != "" {
		data["runner"] = ac.Runner
	}
	if ac.VerifyModel != "" {
		data["verify_model"] = ac.VerifyModel
	}
	if ac.SecurityLevel != "" {
		data["security_level"] = ac.SecurityLevel
	}
	if len(ac.Roles) > 0 {
		data["roles"] = ac.Roles
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
