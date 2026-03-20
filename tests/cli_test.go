package core_test

import (
	"testing"

	. "forge.lthn.ai/core/go/pkg/core"
	"github.com/stretchr/testify/assert"
)

// --- Cli ---

func TestCli_Good(t *testing.T) {
	c := New()
	assert.NotNil(t, c.Cli())
	assert.NotNil(t, c.Cli().Command())
}

func TestCli_Named_Good(t *testing.T) {
	c := New(Options{{K: "name", V: "myapp"}})
	assert.NotNil(t, c.Cli().Command())
}

func TestCli_NewChildCommand_Good(t *testing.T) {
	c := New(Options{{K: "name", V: "myapp"}})
	child := c.Cli().NewChildCommand("test", "a test command")
	assert.NotNil(t, child)
}

func TestCli_AddCommand_Good(t *testing.T) {
	c := New()
	cmd := NewCommand("hello", "says hello")
	c.Cli().AddCommand(cmd)
}

func TestCli_Flags_Good(t *testing.T) {
	c := New()
	var name string
	var debug bool
	var port int
	c.Cli().StringFlag("name", "app name", &name)
	c.Cli().BoolFlag("debug", "enable debug", &debug)
	c.Cli().IntFlag("port", "port number", &port)
}

func TestCli_Run_Good(t *testing.T) {
	c := New()
	executed := false
	c.Cli().Command().Action(func() error {
		executed = true
		return nil
	})
	err := c.Cli().Run("")
	assert.NoError(t, err)
	assert.True(t, executed)
}

// --- Command ---

func TestCommand_New_Good(t *testing.T) {
	cmd := NewCommand("test", "a test command")
	assert.NotNil(t, cmd)
}

func TestCommand_Child_Good(t *testing.T) {
	parent := NewCommand("root")
	child := parent.NewChildCommand("sub", "a subcommand")
	assert.NotNil(t, child)
}

func TestCommand_Flags_Good(t *testing.T) {
	cmd := NewCommand("test")
	var name string
	var debug bool
	cmd.StringFlag("name", "app name", &name)
	cmd.BoolFlag("debug", "enable debug", &debug)
}
