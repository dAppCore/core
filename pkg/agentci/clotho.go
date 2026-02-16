package agentci

import (
	"context"
	"strings"

	"forge.lthn.ai/core/go/pkg/jobrunner"
)

// RunMode determines the execution strategy for a dispatched task.
type RunMode string

const (
	ModeStandard RunMode = "standard"
	ModeDual     RunMode = "dual" // The Clotho Protocol — dual-run verification
)

// Spinner is the Clotho orchestrator that determines the fate of each task.
type Spinner struct {
	Config ClothoConfig
	Agents map[string]AgentConfig
}

// NewSpinner creates a new Clotho orchestrator.
func NewSpinner(cfg ClothoConfig, agents map[string]AgentConfig) *Spinner {
	return &Spinner{
		Config: cfg,
		Agents: agents,
	}
}

// DeterminePlan decides if a signal requires dual-run verification based on
// the global strategy, agent configuration, and repository criticality.
func (s *Spinner) DeterminePlan(signal *jobrunner.PipelineSignal, agentName string) RunMode {
	if s.Config.Strategy != "clotho-verified" {
		return ModeStandard
	}

	agent, ok := s.Agents[agentName]
	if !ok {
		return ModeStandard
	}
	if agent.DualRun {
		return ModeDual
	}

	// Protect critical repos with dual-run (Axiom 1).
	if signal.RepoName == "core" || strings.Contains(signal.RepoName, "security") {
		return ModeDual
	}

	return ModeStandard
}

// GetVerifierModel returns the model for the secondary "signed" verification run.
func (s *Spinner) GetVerifierModel(agentName string) string {
	agent, ok := s.Agents[agentName]
	if !ok || agent.VerifyModel == "" {
		return "gemini-1.5-pro"
	}
	return agent.VerifyModel
}

// FindByForgejoUser resolves a Forgejo username to the agent config key and config.
// This decouples agent naming (mythological roles) from Forgejo identity.
func (s *Spinner) FindByForgejoUser(forgejoUser string) (string, AgentConfig, bool) {
	if forgejoUser == "" {
		return "", AgentConfig{}, false
	}
	// Direct match on config key first.
	if agent, ok := s.Agents[forgejoUser]; ok {
		return forgejoUser, agent, true
	}
	// Search by ForgejoUser field.
	for name, agent := range s.Agents {
		if agent.ForgejoUser != "" && agent.ForgejoUser == forgejoUser {
			return name, agent, true
		}
	}
	return "", AgentConfig{}, false
}

// Weave compares primary and verifier outputs. Returns true if they converge.
// This is a placeholder for future semantic diff logic.
func (s *Spinner) Weave(ctx context.Context, primaryOutput, signedOutput []byte) (bool, error) {
	return string(primaryOutput) == string(signedOutput), nil
}
