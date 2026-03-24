package core_test

import (
	"bytes"
	"testing"

	. "dappco.re/go/core"
	"github.com/stretchr/testify/assert"
)

// --- Cli Surface ---

func TestCli_Good(t *testing.T) {
	c := New()
	assert.NotNil(t, c.Cli())
}

func TestCli_Banner_Good(t *testing.T) {
	c := New(WithOption("name", "myapp"))
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
	c.Command("hello", Command{Action: func(_ Options) Result {
		executed = true
		return Result{Value: "world", OK: true}
	}})
	r := c.Cli().Run("hello")
	assert.True(t, r.OK)
	assert.Equal(t, "world", r.Value)
	assert.True(t, executed)
}

func TestCli_Run_Nested_Good(t *testing.T) {
	c := New()
	executed := false
	c.Command("deploy/to/homelab", Command{Action: func(_ Options) Result {
		executed = true
		return Result{OK: true}
	}})
	r := c.Cli().Run("deploy", "to", "homelab")
	assert.True(t, r.OK)
	assert.True(t, executed)
}

func TestCli_Run_WithFlags_Good(t *testing.T) {
	c := New()
	var received Options
	c.Command("serve", Command{Action: func(opts Options) Result {
		received = opts
		return Result{OK: true}
	}})
	c.Cli().Run("serve", "--port=8080", "--debug")
	assert.Equal(t, "8080", received.String("port"))
	assert.True(t, received.Bool("debug"))
}

func TestCli_Run_NoCommand_Good(t *testing.T) {
	c := New()
	r := c.Cli().Run()
	assert.False(t, r.OK)
}

func TestCli_PrintHelp_Good(t *testing.T) {
	c := New(WithOption("name", "myapp"))
	c.Command("deploy", Command{Action: func(_ Options) Result { return Result{OK: true} }})
	c.Command("serve", Command{Action: func(_ Options) Result { return Result{OK: true} }})
	c.Cli().PrintHelp()
}

func TestCli_SetOutput_Good(t *testing.T) {
	c := New()
	var buf bytes.Buffer
	c.Cli().SetOutput(&buf)
	c.Cli().Print("hello %s", "world")
	assert.Contains(t, buf.String(), "hello world")
}
