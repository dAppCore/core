package core_test

import (
	"testing"

	. "forge.lthn.ai/core/go/pkg/core"
	"github.com/stretchr/testify/assert"
)

// --- Command DTO ---

func TestCommand_Register_Good(t *testing.T) {
	c := New()
	result := c.Command("deploy", func(_ Options) Result[any] {
		return Result[any]{Value: "deployed", OK: true}
	})
	assert.Nil(t, result) // nil = success
}

func TestCommand_Get_Good(t *testing.T) {
	c := New()
	c.Command("deploy", func(_ Options) Result[any] {
		return Result[any]{OK: true}
	})
	cmd := c.Command("deploy")
	assert.NotNil(t, cmd)
}

func TestCommand_Get_Bad(t *testing.T) {
	c := New()
	cmd := c.Command("nonexistent")
	assert.Nil(t, cmd)
}

func TestCommand_Run_Good(t *testing.T) {
	c := New()
	c.Command("greet", func(opts Options) Result[any] {
		return Result[any]{Value: "hello " + opts.String("name"), OK: true}
	})
	cmd := c.Command("greet").(*Command)
	r := cmd.Run(Options{{K: "name", V: "world"}})
	assert.True(t, r.OK)
	assert.Equal(t, "hello world", r.Value)
}

func TestCommand_Run_NoAction_Good(t *testing.T) {
	c := New()
	c.Command("empty", Options{{K: "description", V: "no action"}})
	cmd := c.Command("empty").(*Command)
	r := cmd.Run(Options{})
	assert.False(t, r.OK)
}

// --- Nested Commands ---

func TestCommand_Nested_Good(t *testing.T) {
	c := New()
	c.Command("deploy/to/homelab", func(_ Options) Result[any] {
		return Result[any]{Value: "deployed to homelab", OK: true}
	})

	// Direct path lookup
	cmd := c.Command("deploy/to/homelab")
	assert.NotNil(t, cmd)

	// Parent auto-created
	parent := c.Command("deploy")
	assert.NotNil(t, parent)

	mid := c.Command("deploy/to")
	assert.NotNil(t, mid)
}

func TestCommand_Paths_Good(t *testing.T) {
	c := New()
	c.Command("deploy", func(_ Options) Result[any] { return Result[any]{OK: true} })
	c.Command("serve", func(_ Options) Result[any] { return Result[any]{OK: true} })
	c.Command("deploy/to/homelab", func(_ Options) Result[any] { return Result[any]{OK: true} })

	paths := c.Commands()
	assert.Contains(t, paths, "deploy")
	assert.Contains(t, paths, "serve")
	assert.Contains(t, paths, "deploy/to/homelab")
	assert.Contains(t, paths, "deploy/to") // auto-created parent
}

// --- I18n Key Derivation ---

func TestCommand_I18nKey_Good(t *testing.T) {
	c := New()
	c.Command("deploy/to/homelab", func(_ Options) Result[any] { return Result[any]{OK: true} })
	cmd := c.Command("deploy/to/homelab").(*Command)
	assert.Equal(t, "cmd.deploy.to.homelab.description", cmd.I18nKey())
}

func TestCommand_I18nKey_Custom_Good(t *testing.T) {
	c := New()
	c.Command("deploy", func(_ Options) Result[any] { return Result[any]{OK: true} }, Options{{K: "description", V: "custom.deploy.key"}})
	cmd := c.Command("deploy").(*Command)
	assert.Equal(t, "custom.deploy.key", cmd.I18nKey())
}

func TestCommand_I18nKey_Simple_Good(t *testing.T) {
	c := New()
	c.Command("serve", func(_ Options) Result[any] { return Result[any]{OK: true} })
	cmd := c.Command("serve").(*Command)
	assert.Equal(t, "cmd.serve.description", cmd.I18nKey())
}

// --- Lifecycle ---

func TestCommand_Lifecycle_NoImpl_Good(t *testing.T) {
	c := New()
	c.Command("serve", func(_ Options) Result[any] {
		return Result[any]{Value: "running", OK: true}
	})
	cmd := c.Command("serve").(*Command)

	// Start falls back to Run when no lifecycle impl
	r := cmd.Start(Options{})
	assert.True(t, r.OK)
	assert.Equal(t, "running", r.Value)

	// Stop/Restart/Reload/Signal return empty Result without lifecycle
	assert.False(t, cmd.Stop().OK)
	assert.False(t, cmd.Restart().OK)
	assert.False(t, cmd.Reload().OK)
	assert.False(t, cmd.Signal("HUP").OK)
}

// --- Empty path ---

func TestCommand_EmptyPath_Bad(t *testing.T) {
	c := New()
	result := c.Command("", func(_ Options) Result[any] { return Result[any]{OK: true} })
	assert.NotNil(t, result) // error
}
