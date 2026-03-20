package core_test

import (
	"testing"

	. "forge.lthn.ai/core/go/pkg/core"
	"github.com/stretchr/testify/assert"
)

// --- Cli Surface ---

func TestCli_Good(t *testing.T) {
	c := New()
	assert.NotNil(t, c.Cli())
}

func TestCli_Banner_Good(t *testing.T) {
	c := New(Options{{K: "name", V: "myapp"}})
	assert.Equal(t, "myapp", c.Cli().Banner())
}

func TestCli_SetBanner_Good(t *testing.T) {
	c := New()
	c.Cli().SetBanner(func(_ *Cli) string { return "Custom Banner" })
	assert.Equal(t, "Custom Banner", c.Cli().Banner())
}

func TestCli_Run_Good(t *testing.T) {
	c := New()
	executed := false
	c.Command("hello", func(_ Options) Result {
		executed = true
		return Result{Value: "world", OK: true}
	})
	r := c.Cli().Run("hello")
	assert.True(t, r.OK)
	assert.Equal(t, "world", r.Value)
	assert.True(t, executed)
}

func TestCli_Run_Nested_Good(t *testing.T) {
	c := New()
	executed := false
	c.Command("deploy/to/homelab", func(_ Options) Result {
		executed = true
		return Result{OK: true}
	})
	r := c.Cli().Run("deploy", "to", "homelab")
	assert.True(t, r.OK)
	assert.True(t, executed)
}

func TestCli_Run_WithFlags_Good(t *testing.T) {
	c := New()
	var received Options
	c.Command("serve", func(opts Options) Result {
		received = opts
		return Result{OK: true}
	})
	c.Cli().Run("serve", "--port=8080", "--debug")
	assert.Equal(t, "8080", received.String("port"))
	assert.True(t, received.Bool("debug"))
}

func TestCli_Run_NoCommand_Good(t *testing.T) {
	c := New()
	// No commands registered — should not panic
	r := c.Cli().Run()
	assert.False(t, r.OK)
}

func TestCli_PrintHelp_Good(t *testing.T) {
	c := New(Options{{K: "name", V: "myapp"}})
	c.Command("deploy", func(_ Options) Result { return Result{OK: true} })
	c.Command("serve", func(_ Options) Result { return Result{OK: true} })
	// Should not panic
	c.Cli().PrintHelp()
}
