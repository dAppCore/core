// Package trust implements an agent trust model with tiered access control.
//
// Agents are assigned trust tiers that determine their capabilities:
//
//   - Tier 3 (Full Trust): Internal agents with full access (e.g., Athena, Virgil, Charon)
//   - Tier 2 (Verified): Partner agents with scoped access (e.g., Clotho, Hypnos)
//   - Tier 1 (Untrusted): External/community agents with minimal access
//
// The package provides a Registry for managing agent identities and a PolicyEngine
// for evaluating capability requests against trust policies.
package trust

import (
	"fmt"
	"sync"
	"time"
)

// Tier represents an agent's trust level in the system.
type Tier int

const (
	// TierUntrusted is for external/community agents with minimal access.
	TierUntrusted Tier = 1
	// TierVerified is for partner agents with scoped access.
	TierVerified Tier = 2
	// TierFull is for internal agents with full access.
	TierFull Tier = 3
)

// String returns the human-readable name of the tier.
func (t Tier) String() string {
	switch t {
	case TierUntrusted:
		return "untrusted"
	case TierVerified:
		return "verified"
	case TierFull:
		return "full"
	default:
		return fmt.Sprintf("unknown(%d)", int(t))
	}
}

// Valid returns true if the tier is a recognised trust level.
func (t Tier) Valid() bool {
	return t >= TierUntrusted && t <= TierFull
}

// Capability represents a specific action an agent can perform.
type Capability string

const (
	CapPushRepo      Capability = "repo.push"
	CapMergePR       Capability = "pr.merge"
	CapCreatePR      Capability = "pr.create"
	CapCreateIssue   Capability = "issue.create"
	CapCommentIssue  Capability = "issue.comment"
	CapReadSecrets   Capability = "secrets.read"
	CapRunPrivileged Capability = "cmd.privileged"
	CapAccessWorkspace Capability = "workspace.access"
	CapModifyFlows   Capability = "flows.modify"
)

// Agent represents an agent identity in the trust system.
type Agent struct {
	// Name is the unique identifier for the agent (e.g., "Athena", "Clotho").
	Name string
	// Tier is the agent's trust level.
	Tier Tier
	// ScopedRepos limits repo access for Tier 2 agents. Empty means no repo access.
	// Tier 3 agents ignore this field (they have access to all repos).
	ScopedRepos []string
	// RateLimit is the maximum requests per minute. 0 means unlimited.
	RateLimit int
	// TokenExpiresAt is when the agent's token expires.
	TokenExpiresAt time.Time
	// CreatedAt is when the agent was registered.
	CreatedAt time.Time
}

// Registry manages agent identities and their trust tiers.
type Registry struct {
	mu     sync.RWMutex
	agents map[string]*Agent
}

// NewRegistry creates an empty agent registry.
func NewRegistry() *Registry {
	return &Registry{
		agents: make(map[string]*Agent),
	}
}

// Register adds or updates an agent in the registry.
// Returns an error if the agent name is empty or the tier is invalid.
func (r *Registry) Register(agent Agent) error {
	if agent.Name == "" {
		return fmt.Errorf("trust.Register: agent name is required")
	}
	if !agent.Tier.Valid() {
		return fmt.Errorf("trust.Register: invalid tier %d for agent %q", agent.Tier, agent.Name)
	}
	if agent.CreatedAt.IsZero() {
		agent.CreatedAt = time.Now()
	}
	if agent.RateLimit == 0 {
		agent.RateLimit = defaultRateLimit(agent.Tier)
	}

	r.mu.Lock()
	defer r.mu.Unlock()
	r.agents[agent.Name] = &agent
	return nil
}

// Get returns the agent with the given name, or nil if not found.
func (r *Registry) Get(name string) *Agent {
	r.mu.RLock()
	defer r.mu.RUnlock()
	return r.agents[name]
}

// Remove deletes an agent from the registry.
func (r *Registry) Remove(name string) bool {
	r.mu.Lock()
	defer r.mu.Unlock()
	if _, ok := r.agents[name]; !ok {
		return false
	}
	delete(r.agents, name)
	return true
}

// List returns all registered agents. The returned slice is a snapshot.
func (r *Registry) List() []Agent {
	r.mu.RLock()
	defer r.mu.RUnlock()
	out := make([]Agent, 0, len(r.agents))
	for _, a := range r.agents {
		out = append(out, *a)
	}
	return out
}

// Len returns the number of registered agents.
func (r *Registry) Len() int {
	r.mu.RLock()
	defer r.mu.RUnlock()
	return len(r.agents)
}

// defaultRateLimit returns the default rate limit for a given tier.
func defaultRateLimit(t Tier) int {
	switch t {
	case TierUntrusted:
		return 10
	case TierVerified:
		return 60
	case TierFull:
		return 0 // unlimited
	default:
		return 10
	}
}
