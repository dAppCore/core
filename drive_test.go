package core_test

import (
	"testing"

	. "dappco.re/go/core"
	"github.com/stretchr/testify/assert"
)

// --- Drive (Transport Handles) ---

func TestDrive_New_Good(t *testing.T) {
	c := New().Value.(*Core)
	r := c.Drive().New(Options{
		{Key: "name", Value: "api"},
		{Key: "transport", Value: "https://api.lthn.ai"},
	})
	assert.True(t, r.OK)
	assert.Equal(t, "api", r.Value.(*DriveHandle).Name)
	assert.Equal(t, "https://api.lthn.ai", r.Value.(*DriveHandle).Transport)
}

func TestDrive_New_Bad(t *testing.T) {
	c := New().Value.(*Core)
	// Missing name
	r := c.Drive().New(Options{
		{Key: "transport", Value: "https://api.lthn.ai"},
	})
	assert.False(t, r.OK)
}

func TestDrive_Get_Good(t *testing.T) {
	c := New().Value.(*Core)
	c.Drive().New(Options{
		{Key: "name", Value: "ssh"},
		{Key: "transport", Value: "ssh://claude@10.69.69.165"},
	})
	r := c.Drive().Get("ssh")
	assert.True(t, r.OK)
	handle := r.Value.(*DriveHandle)
	assert.Equal(t, "ssh://claude@10.69.69.165", handle.Transport)
}

func TestDrive_Get_Bad(t *testing.T) {
	c := New().Value.(*Core)
	r := c.Drive().Get("nonexistent")
	assert.False(t, r.OK)
}

func TestDrive_Has_Good(t *testing.T) {
	c := New().Value.(*Core)
	c.Drive().New(Options{{Key: "name", Value: "mcp"}, {Key: "transport", Value: "mcp://mcp.lthn.sh"}})
	assert.True(t, c.Drive().Has("mcp"))
	assert.False(t, c.Drive().Has("missing"))
}

func TestDrive_Names_Good(t *testing.T) {
	c := New().Value.(*Core)
	c.Drive().New(Options{{Key: "name", Value: "api"}, {Key: "transport", Value: "https://api.lthn.ai"}})
	c.Drive().New(Options{{Key: "name", Value: "ssh"}, {Key: "transport", Value: "ssh://claude@10.69.69.165"}})
	c.Drive().New(Options{{Key: "name", Value: "mcp"}, {Key: "transport", Value: "mcp://mcp.lthn.sh"}})
	names := c.Drive().Names()
	assert.Len(t, names, 3)
	assert.Contains(t, names, "api")
	assert.Contains(t, names, "ssh")
	assert.Contains(t, names, "mcp")
}

func TestDrive_OptionsPreserved_Good(t *testing.T) {
	c := New().Value.(*Core)
	c.Drive().New(Options{
		{Key: "name", Value: "api"},
		{Key: "transport", Value: "https://api.lthn.ai"},
		{Key: "timeout", Value: 30},
	})
	r := c.Drive().Get("api")
	assert.True(t, r.OK)
	handle := r.Value.(*DriveHandle)
	assert.Equal(t, 30, handle.Options.Int("timeout"))
}
