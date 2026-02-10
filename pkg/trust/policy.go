package trust

import (
	"fmt"
	"strings"
)

// Policy defines the access rules for a given trust tier.
type Policy struct {
	// Tier is the trust level this policy applies to.
	Tier Tier
	// Allowed lists the capabilities granted at this tier.
	Allowed []Capability
	// RequiresApproval lists capabilities that need human/higher-tier approval.
	RequiresApproval []Capability
	// Denied lists explicitly denied capabilities.
	Denied []Capability
}

// PolicyEngine evaluates capability requests against registered policies.
type PolicyEngine struct {
	registry *Registry
	policies map[Tier]*Policy
}

// Decision is the result of a policy evaluation.
type Decision int

const (
	// Deny means the action is not permitted.
	Deny Decision = iota
	// Allow means the action is permitted.
	Allow
	// NeedsApproval means the action requires human or higher-tier approval.
	NeedsApproval
)

// String returns the human-readable name of the decision.
func (d Decision) String() string {
	switch d {
	case Deny:
		return "deny"
	case Allow:
		return "allow"
	case NeedsApproval:
		return "needs_approval"
	default:
		return fmt.Sprintf("unknown(%d)", int(d))
	}
}

// EvalResult contains the outcome of a capability evaluation.
type EvalResult struct {
	Decision Decision
	Agent    string
	Cap      Capability
	Reason   string
}

// NewPolicyEngine creates a policy engine with the given registry and default policies.
func NewPolicyEngine(registry *Registry) *PolicyEngine {
	pe := &PolicyEngine{
		registry: registry,
		policies: make(map[Tier]*Policy),
	}
	pe.loadDefaults()
	return pe
}

// Evaluate checks whether the named agent can perform the given capability.
// If the agent has scoped repos and the capability is repo-scoped, the repo
// parameter is checked against the agent's allowed repos.
func (pe *PolicyEngine) Evaluate(agentName string, cap Capability, repo string) EvalResult {
	agent := pe.registry.Get(agentName)
	if agent == nil {
		return EvalResult{
			Decision: Deny,
			Agent:    agentName,
			Cap:      cap,
			Reason:   "agent not registered",
		}
	}

	policy, ok := pe.policies[agent.Tier]
	if !ok {
		return EvalResult{
			Decision: Deny,
			Agent:    agentName,
			Cap:      cap,
			Reason:   fmt.Sprintf("no policy for tier %s", agent.Tier),
		}
	}

	// Check explicit denials first.
	for _, denied := range policy.Denied {
		if denied == cap {
			return EvalResult{
				Decision: Deny,
				Agent:    agentName,
				Cap:      cap,
				Reason:   fmt.Sprintf("capability %s is denied for tier %s", cap, agent.Tier),
			}
		}
	}

	// Check if capability requires approval.
	for _, approval := range policy.RequiresApproval {
		if approval == cap {
			return EvalResult{
				Decision: NeedsApproval,
				Agent:    agentName,
				Cap:      cap,
				Reason:   fmt.Sprintf("capability %s requires approval for tier %s", cap, agent.Tier),
			}
		}
	}

	// Check if capability is allowed.
	for _, allowed := range policy.Allowed {
		if allowed == cap {
			// For repo-scoped capabilities, verify repo access.
			if isRepoScoped(cap) && len(agent.ScopedRepos) > 0 {
				if !repoAllowed(agent.ScopedRepos, repo) {
					return EvalResult{
						Decision: Deny,
						Agent:    agentName,
						Cap:      cap,
						Reason:   fmt.Sprintf("agent %q does not have access to repo %q", agentName, repo),
					}
				}
			}
			return EvalResult{
				Decision: Allow,
				Agent:    agentName,
				Cap:      cap,
				Reason:   fmt.Sprintf("capability %s allowed for tier %s", cap, agent.Tier),
			}
		}
	}

	return EvalResult{
		Decision: Deny,
		Agent:    agentName,
		Cap:      cap,
		Reason:   fmt.Sprintf("capability %s not granted for tier %s", cap, agent.Tier),
	}
}

// SetPolicy replaces the policy for a given tier.
func (pe *PolicyEngine) SetPolicy(p Policy) error {
	if !p.Tier.Valid() {
		return fmt.Errorf("trust.SetPolicy: invalid tier %d", p.Tier)
	}
	pe.policies[p.Tier] = &p
	return nil
}

// GetPolicy returns the policy for a tier, or nil if none is set.
func (pe *PolicyEngine) GetPolicy(t Tier) *Policy {
	return pe.policies[t]
}

// loadDefaults installs the default trust policies from the issue spec.
func (pe *PolicyEngine) loadDefaults() {
	// Tier 3 — Full Trust
	pe.policies[TierFull] = &Policy{
		Tier: TierFull,
		Allowed: []Capability{
			CapPushRepo,
			CapMergePR,
			CapCreatePR,
			CapCreateIssue,
			CapCommentIssue,
			CapReadSecrets,
			CapRunPrivileged,
			CapAccessWorkspace,
			CapModifyFlows,
		},
	}

	// Tier 2 — Verified
	pe.policies[TierVerified] = &Policy{
		Tier: TierVerified,
		Allowed: []Capability{
			CapPushRepo,     // scoped to assigned repos
			CapCreatePR,     // can create, not merge
			CapCreateIssue,
			CapCommentIssue,
			CapReadSecrets,  // scoped to their repos
		},
		RequiresApproval: []Capability{
			CapMergePR,
		},
		Denied: []Capability{
			CapAccessWorkspace, // cannot access other agents' workspaces
			CapModifyFlows,
			CapRunPrivileged,
		},
	}

	// Tier 1 — Untrusted
	pe.policies[TierUntrusted] = &Policy{
		Tier: TierUntrusted,
		Allowed: []Capability{
			CapCreatePR,    // fork only, checked at enforcement layer
			CapCommentIssue,
		},
		Denied: []Capability{
			CapPushRepo,
			CapMergePR,
			CapCreateIssue,
			CapReadSecrets,
			CapRunPrivileged,
			CapAccessWorkspace,
			CapModifyFlows,
		},
	}
}

// isRepoScoped returns true if the capability is constrained by repo scope.
func isRepoScoped(cap Capability) bool {
	return strings.HasPrefix(string(cap), "repo.") ||
		strings.HasPrefix(string(cap), "pr.") ||
		cap == CapReadSecrets
}

// repoAllowed checks if repo is in the agent's scoped list.
func repoAllowed(scoped []string, repo string) bool {
	if repo == "" {
		return false
	}
	for _, r := range scoped {
		if r == repo {
			return true
		}
	}
	return false
}
