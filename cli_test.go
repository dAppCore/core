package core_test

import (
	"bytes"

	. "dappco.re/go/core"
)

// --- Cli Surface ---

func TestCli_Good(t *T) {
	c := New()
	AssertNotNil(t, c.Cli())
}

func TestCli_Banner_Good(t *T) {
	c := New(WithOption("name", "myapp"))
	AssertEqual(t, "myapp", c.Cli().Banner())
}

func TestCli_SetBanner_Good(t *T) {
	c := New()
	c.Cli().SetBanner(func(_ *Cli) string { return "Custom Banner" })
	AssertEqual(t, "Custom Banner", c.Cli().Banner())
}

func TestCli_Run_Good(t *T) {
	c := New()
	executed := false
	c.Command("hello", Command{Action: func(_ Options) Result {
		executed = true
		return Result{Value: "world", OK: true}
	}})
	r := c.Cli().Run("hello")
	AssertTrue(t, r.OK)
	AssertEqual(t, "world", r.Value)
	AssertTrue(t, executed)
}

func TestCli_Run_Nested_Good(t *T) {
	c := New()
	executed := false
	c.Command("deploy/to/homelab", Command{Action: func(_ Options) Result {
		executed = true
		return Result{OK: true}
	}})
	r := c.Cli().Run("deploy", "to", "homelab")
	AssertTrue(t, r.OK)
	AssertTrue(t, executed)
}

func TestCli_Run_WithFlags_Good(t *T) {
	c := New()
	var received Options
	c.Command("serve", Command{Action: func(opts Options) Result {
		received = opts
		return Result{OK: true}
	}})
	c.Cli().Run("serve", "--port=8080", "--debug")
	AssertEqual(t, "8080", received.String("port"))
	AssertTrue(t, received.Bool("debug"))
}

func TestCli_Run_NoCommand_Good(t *T) {
	c := New()
	r := c.Cli().Run()
	AssertFalse(t, r.OK)
}

func TestCli_PrintHelp_Good(t *T) {
	c := New(WithOption("name", "myapp"))
	c.Command("deploy", Command{Action: func(_ Options) Result { return Result{OK: true} }})
	c.Command("serve", Command{Action: func(_ Options) Result { return Result{OK: true} }})
	c.Cli().PrintHelp()
}

func TestCli_SetOutput_Good(t *T) {
	c := New()
	var buf bytes.Buffer
	c.Cli().SetOutput(&buf)
	c.Cli().Print("hello %s", "world")
	AssertContains(t, buf.String(), "hello world")
}
