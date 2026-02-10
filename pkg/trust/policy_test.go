package trust

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func newTestEngine(t *testing.T) *PolicyEngine {
	t.Helper()
	r := NewRegistry()
	require.NoError(t, r.Register(Agent{
		Name: "Athena",
		Tier: TierFull,
	}))
	require.NoError(t, r.Register(Agent{
		Name:        "Clotho",
		Tier:        TierVerified,
		ScopedRepos: []string{"host-uk/core", "host-uk/docs"},
	}))
	require.NoError(t, r.Register(Agent{
		Name: "BugSETI-001",
		Tier: TierUntrusted,
	}))
	return NewPolicyEngine(r)
}

// --- Decision ---

func TestDecisionString_Good(t *testing.T) {
	assert.Equal(t, "deny", Deny.String())
	assert.Equal(t, "allow", Allow.String())
	assert.Equal(t, "needs_approval", NeedsApproval.String())
}

func TestDecisionString_Bad_Unknown(t *testing.T) {
	assert.Contains(t, Decision(99).String(), "unknown")
}

// --- Tier 3 (Full Trust) ---

func TestEvaluate_Good_Tier3CanDoAnything(t *testing.T) {
	pe := newTestEngine(t)

	caps := []Capability{
		CapPushRepo, CapMergePR, CapCreatePR, CapCreateIssue,
		CapCommentIssue, CapReadSecrets, CapRunPrivileged,
		CapAccessWorkspace, CapModifyFlows,
	}
	for _, cap := range caps {
		result := pe.Evaluate("Athena", cap, "")
		assert.Equal(t, Allow, result.Decision, "Athena should be allowed %s", cap)
	}
}

// --- Tier 2 (Verified) ---

func TestEvaluate_Good_Tier2CanCreatePR(t *testing.T) {
	pe := newTestEngine(t)
	result := pe.Evaluate("Clotho", CapCreatePR, "host-uk/core")
	assert.Equal(t, Allow, result.Decision)
}

func TestEvaluate_Good_Tier2CanPushToScopedRepo(t *testing.T) {
	pe := newTestEngine(t)
	result := pe.Evaluate("Clotho", CapPushRepo, "host-uk/core")
	assert.Equal(t, Allow, result.Decision)
}

func TestEvaluate_Good_Tier2NeedsApprovalToMerge(t *testing.T) {
	pe := newTestEngine(t)
	result := pe.Evaluate("Clotho", CapMergePR, "host-uk/core")
	assert.Equal(t, NeedsApproval, result.Decision)
}

func TestEvaluate_Good_Tier2CanCreateIssue(t *testing.T) {
	pe := newTestEngine(t)
	result := pe.Evaluate("Clotho", CapCreateIssue, "")
	assert.Equal(t, Allow, result.Decision)
}

func TestEvaluate_Bad_Tier2CannotAccessWorkspace(t *testing.T) {
	pe := newTestEngine(t)
	result := pe.Evaluate("Clotho", CapAccessWorkspace, "")
	assert.Equal(t, Deny, result.Decision)
}

func TestEvaluate_Bad_Tier2CannotModifyFlows(t *testing.T) {
	pe := newTestEngine(t)
	result := pe.Evaluate("Clotho", CapModifyFlows, "")
	assert.Equal(t, Deny, result.Decision)
}

func TestEvaluate_Bad_Tier2CannotRunPrivileged(t *testing.T) {
	pe := newTestEngine(t)
	result := pe.Evaluate("Clotho", CapRunPrivileged, "")
	assert.Equal(t, Deny, result.Decision)
}

func TestEvaluate_Bad_Tier2CannotPushToUnscopedRepo(t *testing.T) {
	pe := newTestEngine(t)
	result := pe.Evaluate("Clotho", CapPushRepo, "host-uk/secret-repo")
	assert.Equal(t, Deny, result.Decision)
	assert.Contains(t, result.Reason, "does not have access")
}

func TestEvaluate_Bad_Tier2RepoScopeEmptyRepo(t *testing.T) {
	pe := newTestEngine(t)
	// Push without specifying a repo should be denied for scoped agents.
	result := pe.Evaluate("Clotho", CapPushRepo, "")
	assert.Equal(t, Deny, result.Decision)
}

// --- Tier 1 (Untrusted) ---

func TestEvaluate_Good_Tier1CanCreatePR(t *testing.T) {
	pe := newTestEngine(t)
	result := pe.Evaluate("BugSETI-001", CapCreatePR, "")
	assert.Equal(t, Allow, result.Decision)
}

func TestEvaluate_Good_Tier1CanCommentIssue(t *testing.T) {
	pe := newTestEngine(t)
	result := pe.Evaluate("BugSETI-001", CapCommentIssue, "")
	assert.Equal(t, Allow, result.Decision)
}

func TestEvaluate_Bad_Tier1CannotPush(t *testing.T) {
	pe := newTestEngine(t)
	result := pe.Evaluate("BugSETI-001", CapPushRepo, "")
	assert.Equal(t, Deny, result.Decision)
}

func TestEvaluate_Bad_Tier1CannotMerge(t *testing.T) {
	pe := newTestEngine(t)
	result := pe.Evaluate("BugSETI-001", CapMergePR, "")
	assert.Equal(t, Deny, result.Decision)
}

func TestEvaluate_Bad_Tier1CannotCreateIssue(t *testing.T) {
	pe := newTestEngine(t)
	result := pe.Evaluate("BugSETI-001", CapCreateIssue, "")
	assert.Equal(t, Deny, result.Decision)
}

func TestEvaluate_Bad_Tier1CannotReadSecrets(t *testing.T) {
	pe := newTestEngine(t)
	result := pe.Evaluate("BugSETI-001", CapReadSecrets, "")
	assert.Equal(t, Deny, result.Decision)
}

func TestEvaluate_Bad_Tier1CannotRunPrivileged(t *testing.T) {
	pe := newTestEngine(t)
	result := pe.Evaluate("BugSETI-001", CapRunPrivileged, "")
	assert.Equal(t, Deny, result.Decision)
}

// --- Edge cases ---

func TestEvaluate_Bad_UnknownAgent(t *testing.T) {
	pe := newTestEngine(t)
	result := pe.Evaluate("Unknown", CapCreatePR, "")
	assert.Equal(t, Deny, result.Decision)
	assert.Contains(t, result.Reason, "not registered")
}

func TestEvaluate_Good_EvalResultFields(t *testing.T) {
	pe := newTestEngine(t)
	result := pe.Evaluate("Athena", CapPushRepo, "")
	assert.Equal(t, "Athena", result.Agent)
	assert.Equal(t, CapPushRepo, result.Cap)
	assert.NotEmpty(t, result.Reason)
}

// --- SetPolicy ---

func TestSetPolicy_Good(t *testing.T) {
	pe := newTestEngine(t)
	err := pe.SetPolicy(Policy{
		Tier:    TierVerified,
		Allowed: []Capability{CapPushRepo, CapMergePR},
	})
	require.NoError(t, err)

	// Verify the new policy is in effect.
	result := pe.Evaluate("Clotho", CapMergePR, "host-uk/core")
	assert.Equal(t, Allow, result.Decision)
}

func TestSetPolicy_Bad_InvalidTier(t *testing.T) {
	pe := newTestEngine(t)
	err := pe.SetPolicy(Policy{Tier: Tier(0)})
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "invalid tier")
}

func TestGetPolicy_Good(t *testing.T) {
	pe := newTestEngine(t)
	p := pe.GetPolicy(TierFull)
	require.NotNil(t, p)
	assert.Equal(t, TierFull, p.Tier)
}

func TestGetPolicy_Bad_NotFound(t *testing.T) {
	pe := newTestEngine(t)
	assert.Nil(t, pe.GetPolicy(Tier(99)))
}

// --- isRepoScoped / repoAllowed helpers ---

func TestIsRepoScoped_Good(t *testing.T) {
	assert.True(t, isRepoScoped(CapPushRepo))
	assert.True(t, isRepoScoped(CapCreatePR))
	assert.True(t, isRepoScoped(CapMergePR))
	assert.True(t, isRepoScoped(CapReadSecrets))
}

func TestIsRepoScoped_Bad_NotScoped(t *testing.T) {
	assert.False(t, isRepoScoped(CapRunPrivileged))
	assert.False(t, isRepoScoped(CapAccessWorkspace))
	assert.False(t, isRepoScoped(CapModifyFlows))
}

func TestRepoAllowed_Good(t *testing.T) {
	scoped := []string{"host-uk/core", "host-uk/docs"}
	assert.True(t, repoAllowed(scoped, "host-uk/core"))
	assert.True(t, repoAllowed(scoped, "host-uk/docs"))
}

func TestRepoAllowed_Bad_NotInScope(t *testing.T) {
	scoped := []string{"host-uk/core"}
	assert.False(t, repoAllowed(scoped, "host-uk/secret"))
}

func TestRepoAllowed_Bad_EmptyRepo(t *testing.T) {
	scoped := []string{"host-uk/core"}
	assert.False(t, repoAllowed(scoped, ""))
}

func TestRepoAllowed_Bad_EmptyScope(t *testing.T) {
	assert.False(t, repoAllowed(nil, "host-uk/core"))
	assert.False(t, repoAllowed([]string{}, "host-uk/core"))
}

// --- Tier 3 ignores repo scoping ---

func TestEvaluate_Good_Tier3IgnoresRepoScope(t *testing.T) {
	r := NewRegistry()
	require.NoError(t, r.Register(Agent{
		Name:        "Virgil",
		Tier:        TierFull,
		ScopedRepos: []string{}, // empty scope should not restrict Tier 3
	}))
	pe := NewPolicyEngine(r)

	result := pe.Evaluate("Virgil", CapPushRepo, "any-repo")
	assert.Equal(t, Allow, result.Decision)
}

// --- Default rate limits ---

func TestDefaultRateLimit(t *testing.T) {
	assert.Equal(t, 10, defaultRateLimit(TierUntrusted))
	assert.Equal(t, 60, defaultRateLimit(TierVerified))
	assert.Equal(t, 0, defaultRateLimit(TierFull))
	assert.Equal(t, 10, defaultRateLimit(Tier(99))) // unknown defaults to 10
}
