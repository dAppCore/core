package core_test

import (
	"testing"

	. "dappco.re/go/core"
	"github.com/stretchr/testify/assert"
)

// --- Command DTO ---

func TestCommand_Register_Good(t *testing.T) {
	c := New().Value.(*Core)
	r := c.Command("deploy", Command{Action: func(_ Options) Result {
		return Result{Value: "deployed", OK: true}
	}})
	assert.True(t, r.OK)
}

func TestCommand_Get_Good(t *testing.T) {
	c := New().Value.(*Core)
	c.Command("deploy", Command{Action: func(_ Options) Result { return Result{OK: true} }})
	r := c.Command("deploy")
	assert.True(t, r.OK)
	assert.NotNil(t, r.Value)
}

func TestCommand_Get_Bad(t *testing.T) {
	c := New().Value.(*Core)
	r := c.Command("nonexistent")
	assert.False(t, r.OK)
}

func TestCommand_Run_Good(t *testing.T) {
	c := New().Value.(*Core)
	c.Command("greet", Command{Action: func(opts Options) Result {
		return Result{Value: Concat("hello ", opts.String("name")), OK: true}
	}})
	cmd := c.Command("greet").Value.(*Command)
	r := cmd.Run(NewOptions(Option{Key: "name", Value: "world"}))
	assert.True(t, r.OK)
	assert.Equal(t, "hello world", r.Value)
}

func TestCommand_Run_NoAction_Good(t *testing.T) {
	c := New().Value.(*Core)
	c.Command("empty", Command{Description: "no action"})
	cmd := c.Command("empty").Value.(*Command)
	r := cmd.Run(NewOptions())
	assert.False(t, r.OK)
}

// --- Nested Commands ---

func TestCommand_Nested_Good(t *testing.T) {
	c := New().Value.(*Core)
	c.Command("deploy/to/homelab", Command{Action: func(_ Options) Result {
		return Result{Value: "deployed to homelab", OK: true}
	}})

	r := c.Command("deploy/to/homelab")
	assert.True(t, r.OK)

	// Parent auto-created
	assert.True(t, c.Command("deploy").OK)
	assert.True(t, c.Command("deploy/to").OK)
}

func TestCommand_Paths_Good(t *testing.T) {
	c := New().Value.(*Core)
	c.Command("deploy", Command{Action: func(_ Options) Result { return Result{OK: true} }})
	c.Command("serve", Command{Action: func(_ Options) Result { return Result{OK: true} }})
	c.Command("deploy/to/homelab", Command{Action: func(_ Options) Result { return Result{OK: true} }})

	paths := c.Commands()
	assert.Contains(t, paths, "deploy")
	assert.Contains(t, paths, "serve")
	assert.Contains(t, paths, "deploy/to/homelab")
	assert.Contains(t, paths, "deploy/to")
}

// --- I18n Key Derivation ---

func TestCommand_I18nKey_Good(t *testing.T) {
	c := New().Value.(*Core)
	c.Command("deploy/to/homelab", Command{})
	cmd := c.Command("deploy/to/homelab").Value.(*Command)
	assert.Equal(t, "cmd.deploy.to.homelab.description", cmd.I18nKey())
}

func TestCommand_I18nKey_Custom_Good(t *testing.T) {
	c := New().Value.(*Core)
	c.Command("deploy", Command{Description: "custom.deploy.key"})
	cmd := c.Command("deploy").Value.(*Command)
	assert.Equal(t, "custom.deploy.key", cmd.I18nKey())
}

func TestCommand_I18nKey_Simple_Good(t *testing.T) {
	c := New().Value.(*Core)
	c.Command("serve", Command{})
	cmd := c.Command("serve").Value.(*Command)
	assert.Equal(t, "cmd.serve.description", cmd.I18nKey())
}

// --- Lifecycle ---

func TestCommand_Lifecycle_NoImpl_Good(t *testing.T) {
	c := New().Value.(*Core)
	c.Command("serve", Command{Action: func(_ Options) Result {
		return Result{Value: "running", OK: true}
	}})
	cmd := c.Command("serve").Value.(*Command)

	r := cmd.Start(NewOptions())
	assert.True(t, r.OK)
	assert.Equal(t, "running", r.Value)

	assert.False(t, cmd.Stop().OK)
	assert.False(t, cmd.Restart().OK)
	assert.False(t, cmd.Reload().OK)
	assert.False(t, cmd.Signal("HUP").OK)
}

// --- Lifecycle with Implementation ---

type testLifecycle struct {
	started   bool
	stopped   bool
	restarted bool
	reloaded  bool
	signalled string
}

func (l *testLifecycle) Start(opts Options) Result {
	l.started = true
	return Result{Value: "started", OK: true}
}
func (l *testLifecycle) Stop() Result {
	l.stopped = true
	return Result{OK: true}
}
func (l *testLifecycle) Restart() Result {
	l.restarted = true
	return Result{OK: true}
}
func (l *testLifecycle) Reload() Result {
	l.reloaded = true
	return Result{OK: true}
}
func (l *testLifecycle) Signal(sig string) Result {
	l.signalled = sig
	return Result{Value: sig, OK: true}
}

func TestCommand_Lifecycle_WithImpl_Good(t *testing.T) {
	c := New().Value.(*Core)
	lc := &testLifecycle{}
	c.Command("daemon", Command{Lifecycle: lc})
	cmd := c.Command("daemon").Value.(*Command)

	r := cmd.Start(NewOptions())
	assert.True(t, r.OK)
	assert.True(t, lc.started)

	assert.True(t, cmd.Stop().OK)
	assert.True(t, lc.stopped)

	assert.True(t, cmd.Restart().OK)
	assert.True(t, lc.restarted)

	assert.True(t, cmd.Reload().OK)
	assert.True(t, lc.reloaded)

	r = cmd.Signal("HUP")
	assert.True(t, r.OK)
	assert.Equal(t, "HUP", lc.signalled)
}

func TestCommand_Duplicate_Bad(t *testing.T) {
	c := New().Value.(*Core)
	c.Command("deploy", Command{Action: func(_ Options) Result { return Result{OK: true} }})
	r := c.Command("deploy", Command{Action: func(_ Options) Result { return Result{OK: true} }})
	assert.False(t, r.OK)
}

func TestCommand_InvalidPath_Bad(t *testing.T) {
	c := New().Value.(*Core)
	assert.False(t, c.Command("/leading", Command{}).OK)
	assert.False(t, c.Command("trailing/", Command{}).OK)
	assert.False(t, c.Command("double//slash", Command{}).OK)
}

// --- Cli Run with Lifecycle ---

func TestCli_Run_Lifecycle_Good(t *testing.T) {
	c := New().Value.(*Core)
	lc := &testLifecycle{}
	c.Command("serve", Command{Lifecycle: lc})
	r := c.Cli().Run("serve")
	assert.True(t, r.OK)
	assert.True(t, lc.started)
}

func TestCli_Run_NoActionNoLifecycle_Bad(t *testing.T) {
	c := New().Value.(*Core)
	c.Command("empty", Command{})
	r := c.Cli().Run("empty")
	assert.False(t, r.OK)
}

// --- Empty path ---

func TestCommand_EmptyPath_Bad(t *testing.T) {
	c := New().Value.(*Core)
	r := c.Command("", Command{})
	assert.False(t, r.OK)
}
