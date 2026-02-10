package trust

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// --- Tier ---

func TestTierString_Good(t *testing.T) {
	assert.Equal(t, "untrusted", TierUntrusted.String())
	assert.Equal(t, "verified", TierVerified.String())
	assert.Equal(t, "full", TierFull.String())
}

func TestTierString_Bad_Unknown(t *testing.T) {
	assert.Contains(t, Tier(99).String(), "unknown")
}

func TestTierValid_Good(t *testing.T) {
	assert.True(t, TierUntrusted.Valid())
	assert.True(t, TierVerified.Valid())
	assert.True(t, TierFull.Valid())
}

func TestTierValid_Bad(t *testing.T) {
	assert.False(t, Tier(0).Valid())
	assert.False(t, Tier(4).Valid())
	assert.False(t, Tier(-1).Valid())
}

// --- Registry ---

func TestRegistryRegister_Good(t *testing.T) {
	r := NewRegistry()
	err := r.Register(Agent{Name: "Athena", Tier: TierFull})
	require.NoError(t, err)
	assert.Equal(t, 1, r.Len())
}

func TestRegistryRegister_Good_SetsDefaults(t *testing.T) {
	r := NewRegistry()
	err := r.Register(Agent{Name: "Athena", Tier: TierFull})
	require.NoError(t, err)

	a := r.Get("Athena")
	require.NotNil(t, a)
	assert.Equal(t, 0, a.RateLimit) // full trust = unlimited
	assert.False(t, a.CreatedAt.IsZero())
}

func TestRegistryRegister_Good_TierDefaults(t *testing.T) {
	r := NewRegistry()
	require.NoError(t, r.Register(Agent{Name: "A", Tier: TierUntrusted}))
	require.NoError(t, r.Register(Agent{Name: "B", Tier: TierVerified}))
	require.NoError(t, r.Register(Agent{Name: "C", Tier: TierFull}))

	assert.Equal(t, 10, r.Get("A").RateLimit)
	assert.Equal(t, 60, r.Get("B").RateLimit)
	assert.Equal(t, 0, r.Get("C").RateLimit)
}

func TestRegistryRegister_Good_PreservesExplicitRateLimit(t *testing.T) {
	r := NewRegistry()
	err := r.Register(Agent{Name: "Custom", Tier: TierVerified, RateLimit: 30})
	require.NoError(t, err)
	assert.Equal(t, 30, r.Get("Custom").RateLimit)
}

func TestRegistryRegister_Good_Update(t *testing.T) {
	r := NewRegistry()
	require.NoError(t, r.Register(Agent{Name: "Athena", Tier: TierVerified}))
	require.NoError(t, r.Register(Agent{Name: "Athena", Tier: TierFull}))

	assert.Equal(t, 1, r.Len())
	assert.Equal(t, TierFull, r.Get("Athena").Tier)
}

func TestRegistryRegister_Bad_EmptyName(t *testing.T) {
	r := NewRegistry()
	err := r.Register(Agent{Tier: TierFull})
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "name is required")
}

func TestRegistryRegister_Bad_InvalidTier(t *testing.T) {
	r := NewRegistry()
	err := r.Register(Agent{Name: "Bad", Tier: Tier(0)})
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "invalid tier")
}

func TestRegistryGet_Good(t *testing.T) {
	r := NewRegistry()
	require.NoError(t, r.Register(Agent{Name: "Athena", Tier: TierFull}))
	a := r.Get("Athena")
	require.NotNil(t, a)
	assert.Equal(t, "Athena", a.Name)
}

func TestRegistryGet_Bad_NotFound(t *testing.T) {
	r := NewRegistry()
	assert.Nil(t, r.Get("nonexistent"))
}

func TestRegistryRemove_Good(t *testing.T) {
	r := NewRegistry()
	require.NoError(t, r.Register(Agent{Name: "Athena", Tier: TierFull}))
	assert.True(t, r.Remove("Athena"))
	assert.Equal(t, 0, r.Len())
}

func TestRegistryRemove_Bad_NotFound(t *testing.T) {
	r := NewRegistry()
	assert.False(t, r.Remove("nonexistent"))
}

func TestRegistryList_Good(t *testing.T) {
	r := NewRegistry()
	require.NoError(t, r.Register(Agent{Name: "Athena", Tier: TierFull}))
	require.NoError(t, r.Register(Agent{Name: "Clotho", Tier: TierVerified}))

	agents := r.List()
	assert.Len(t, agents, 2)

	names := make(map[string]bool)
	for _, a := range agents {
		names[a.Name] = true
	}
	assert.True(t, names["Athena"])
	assert.True(t, names["Clotho"])
}

func TestRegistryList_Good_Empty(t *testing.T) {
	r := NewRegistry()
	assert.Empty(t, r.List())
}

func TestRegistryList_Good_Snapshot(t *testing.T) {
	r := NewRegistry()
	require.NoError(t, r.Register(Agent{Name: "Athena", Tier: TierFull}))
	agents := r.List()

	// Modifying the returned slice should not affect the registry.
	agents[0].Tier = TierUntrusted
	assert.Equal(t, TierFull, r.Get("Athena").Tier)
}

// --- Agent ---

func TestAgentTokenExpiry(t *testing.T) {
	agent := Agent{
		Name:           "Test",
		Tier:           TierVerified,
		TokenExpiresAt: time.Now().Add(-1 * time.Hour),
	}
	assert.True(t, time.Now().After(agent.TokenExpiresAt))

	agent.TokenExpiresAt = time.Now().Add(1 * time.Hour)
	assert.True(t, time.Now().Before(agent.TokenExpiresAt))
}
