package core_test

import (
	"testing"

	. "forge.lthn.ai/core/go/pkg/core"
	"github.com/stretchr/testify/assert"
)

// --- Drive (Transport Handles) ---

func TestDrive_New_Good(t *testing.T) {
	c := New()
	r := c.Drive().New(Options{
		{K: "name", V: "api"},
		{K: "transport", V: "https://api.lthn.ai"},
	})
	assert.True(t, r.OK)
	assert.Equal(t, "api", r.Value.Name)
	assert.Equal(t, "https://api.lthn.ai", r.Value.Transport)
}

func TestDrive_New_Bad(t *testing.T) {
	c := New()
	// Missing name
	r := c.Drive().New(Options{
		{K: "transport", V: "https://api.lthn.ai"},
	})
	assert.False(t, r.OK)
}

func TestDrive_Get_Good(t *testing.T) {
	c := New()
	c.Drive().New(Options{
		{K: "name", V: "ssh"},
		{K: "transport", V: "ssh://claude@10.69.69.165"},
	})
	handle := c.Drive().Get("ssh")
	assert.NotNil(t, handle)
	assert.Equal(t, "ssh://claude@10.69.69.165", handle.Transport)
}

func TestDrive_Get_Bad(t *testing.T) {
	c := New()
	handle := c.Drive().Get("nonexistent")
	assert.Nil(t, handle)
}

func TestDrive_Has_Good(t *testing.T) {
	c := New()
	c.Drive().New(Options{{K: "name", V: "mcp"}, {K: "transport", V: "mcp://mcp.lthn.sh"}})
	assert.True(t, c.Drive().Has("mcp"))
	assert.False(t, c.Drive().Has("missing"))
}

func TestDrive_Names_Good(t *testing.T) {
	c := New()
	c.Drive().New(Options{{K: "name", V: "api"}, {K: "transport", V: "https://api.lthn.ai"}})
	c.Drive().New(Options{{K: "name", V: "ssh"}, {K: "transport", V: "ssh://claude@10.69.69.165"}})
	c.Drive().New(Options{{K: "name", V: "mcp"}, {K: "transport", V: "mcp://mcp.lthn.sh"}})
	names := c.Drive().Names()
	assert.Len(t, names, 3)
	assert.Contains(t, names, "api")
	assert.Contains(t, names, "ssh")
	assert.Contains(t, names, "mcp")
}

func TestDrive_OptionsPreserved_Good(t *testing.T) {
	c := New()
	c.Drive().New(Options{
		{K: "name", V: "api"},
		{K: "transport", V: "https://api.lthn.ai"},
		{K: "timeout", V: 30},
	})
	handle := c.Drive().Get("api")
	assert.Equal(t, 30, handle.Options.Int("timeout"))
}
