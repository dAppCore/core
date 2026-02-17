package marketplace

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestParseIndex_Good(t *testing.T) {
	raw := `{
		"version": 1,
		"modules": [
			{"code": "mining-xmrig", "name": "XMRig Miner", "repo": "https://forge.lthn.io/host-uk/mod-xmrig.git", "sign_key": "abc123", "category": "miner"},
			{"code": "utils-cyberchef", "name": "CyberChef", "repo": "https://forge.lthn.io/host-uk/mod-cyberchef.git", "sign_key": "def456", "category": "utils"}
		],
		"categories": ["miner", "utils"]
	}`
	idx, err := ParseIndex([]byte(raw))
	require.NoError(t, err)
	assert.Equal(t, 1, idx.Version)
	assert.Len(t, idx.Modules, 2)
	assert.Equal(t, "mining-xmrig", idx.Modules[0].Code)
}

func TestSearch_Good(t *testing.T) {
	idx := &Index{
		Modules: []Module{
			{Code: "mining-xmrig", Name: "XMRig Miner", Category: "miner"},
			{Code: "utils-cyberchef", Name: "CyberChef", Category: "utils"},
		},
	}
	results := idx.Search("miner")
	assert.Len(t, results, 1)
	assert.Equal(t, "mining-xmrig", results[0].Code)
}

func TestByCategory_Good(t *testing.T) {
	idx := &Index{
		Modules: []Module{
			{Code: "a", Category: "miner"},
			{Code: "b", Category: "utils"},
			{Code: "c", Category: "miner"},
		},
	}
	miners := idx.ByCategory("miner")
	assert.Len(t, miners, 2)
}

func TestFind_Good(t *testing.T) {
	idx := &Index{
		Modules: []Module{
			{Code: "mining-xmrig", Name: "XMRig"},
		},
	}
	m, ok := idx.Find("mining-xmrig")
	assert.True(t, ok)
	assert.Equal(t, "XMRig", m.Name)
}

func TestFind_Bad_NotFound(t *testing.T) {
	idx := &Index{}
	_, ok := idx.Find("nope")
	assert.False(t, ok)
}
