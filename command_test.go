package core_test

import (
	. "dappco.re/go/core"
)

// --- Command DTO ---

func TestCommand_Register_Good(t *T) {
	c := New()
	r := c.Command("deploy", Command{Action: func(_ Options) Result {
		return Result{Value: "deployed", OK: true}
	}})
	AssertTrue(t, r.OK)
}

func TestCommand_Get_Good(t *T) {
	c := New()
	c.Command("deploy", Command{Action: func(_ Options) Result { return Result{OK: true} }})
	r := c.Command("deploy")
	AssertTrue(t, r.OK)
	AssertNotNil(t, r.Value)
}

func TestCommand_Get_Bad(t *T) {
	c := New()
	r := c.Command("nonexistent")
	AssertFalse(t, r.OK)
}

func TestCommand_Run_Good(t *T) {
	c := New()
	c.Command("greet", Command{Action: func(opts Options) Result {
		return Result{Value: Concat("hello ", opts.String("name")), OK: true}
	}})
	cmd := c.Command("greet").Value.(*Command)
	r := cmd.Run(NewOptions(Option{Key: "name", Value: "world"}))
	AssertTrue(t, r.OK)
	AssertEqual(t, "hello world", r.Value)
}

func TestCommand_Run_NoAction_Good(t *T) {
	c := New()
	c.Command("empty", Command{Description: "no action"})
	cmd := c.Command("empty").Value.(*Command)
	r := cmd.Run(NewOptions())
	AssertFalse(t, r.OK)
}

// --- Nested Commands ---

func TestCommand_Nested_Good(t *T) {
	c := New()
	c.Command("deploy/to/homelab", Command{Action: func(_ Options) Result {
		return Result{Value: "deployed to homelab", OK: true}
	}})

	r := c.Command("deploy/to/homelab")
	AssertTrue(t, r.OK)

	// Parent auto-created
	AssertTrue(t, c.Command("deploy").OK)
	AssertTrue(t, c.Command("deploy/to").OK)
}

func TestCommand_Paths_Good(t *T) {
	c := New()
	c.Command("deploy", Command{Action: func(_ Options) Result { return Result{OK: true} }})
	c.Command("serve", Command{Action: func(_ Options) Result { return Result{OK: true} }})
	c.Command("deploy/to/homelab", Command{Action: func(_ Options) Result { return Result{OK: true} }})

	paths := c.Commands()
	AssertContains(t, paths, "deploy")
	AssertContains(t, paths, "serve")
	AssertContains(t, paths, "deploy/to/homelab")
	AssertContains(t, paths, "deploy/to")
}

// --- I18n Key Derivation ---

func TestCommand_I18nKey_Good(t *T) {
	c := New()
	c.Command("deploy/to/homelab", Command{})
	cmd := c.Command("deploy/to/homelab").Value.(*Command)
	AssertEqual(t, "cmd.deploy.to.homelab.description", cmd.I18nKey())
}

func TestCommand_I18nKey_Custom_Good(t *T) {
	c := New()
	c.Command("deploy", Command{Description: "custom.deploy.key"})
	cmd := c.Command("deploy").Value.(*Command)
	AssertEqual(t, "custom.deploy.key", cmd.I18nKey())
}

func TestCommand_I18nKey_Simple_Good(t *T) {
	c := New()
	c.Command("serve", Command{})
	cmd := c.Command("serve").Value.(*Command)
	AssertEqual(t, "cmd.serve.description", cmd.I18nKey())
}

// --- Managed ---

func TestCommand_IsManaged_Good(t *T) {
	c := New()
	c.Command("serve", Command{
		Action:  func(_ Options) Result { return Result{Value: "running", OK: true} },
		Managed: "process.daemon",
	})
	cmd := c.Command("serve").Value.(*Command)
	AssertTrue(t, cmd.IsManaged())
}

func TestCommand_IsManaged_Bad_NotManaged(t *T) {
	c := New()
	c.Command("deploy", Command{
		Action: func(_ Options) Result { return Result{OK: true} },
	})
	cmd := c.Command("deploy").Value.(*Command)
	AssertFalse(t, cmd.IsManaged())
}

func TestCommand_Duplicate_Bad(t *T) {
	c := New()
	c.Command("deploy", Command{Action: func(_ Options) Result { return Result{OK: true} }})
	r := c.Command("deploy", Command{Action: func(_ Options) Result { return Result{OK: true} }})
	AssertFalse(t, r.OK)
}

func TestCommand_InvalidPath_Bad(t *T) {
	c := New()
	AssertFalse(t, c.Command("/leading", Command{}).OK)
	AssertFalse(t, c.Command("trailing/", Command{}).OK)
	AssertFalse(t, c.Command("double//slash", Command{}).OK)
}

// --- Cli Run with Managed ---

func TestCli_Run_Managed_Good(t *T) {
	c := New()
	ran := false
	c.Command("serve", Command{
		Action:  func(_ Options) Result { ran = true; return Result{OK: true} },
		Managed: "process.daemon",
	})
	r := c.Cli().Run("serve")
	AssertTrue(t, r.OK)
	AssertTrue(t, ran)
}

func TestCli_Run_NoAction_Bad(t *T) {
	c := New()
	c.Command("empty", Command{})
	r := c.Cli().Run("empty")
	AssertFalse(t, r.OK)
}

// --- Empty path ---

func TestCommand_EmptyPath_Bad(t *T) {
	c := New()
	r := c.Command("", Command{})
	AssertFalse(t, r.OK)
}
