package core_test

import (
	. "dappco.re/go"
)

// fakeProcess registers an in-test process.run handler that echoes a
// composed string so AssertCLI scenarios stay self-contained without a
// real process service. Returns the registered Core.
func fakeProcess(t *T, response string, ok bool) *Core {
	t.Helper()
	c := New()
	c.Action("process.run", func(ctx Context, opts Options) Result {
		if !ok {
			return Result{Value: NewCode("test.process.bad", response), OK: false}
		}
		return Result{Value: response, OK: true}
	})
	return c
}

// --- AssertCLI ---

func TestCLI_AssertCLI_Good(t *T) {
	c := fakeProcess(t, "go version go1.26.0\n", true)
	AssertCLI(t, c, CLITest{
		Cmd:    "go",
		Args:   []string{"version"},
		WantOK: true,
		Output: "go version go1.26.0\n",
	})
}

func TestCLI_AssertCLI_Contains_Good(t *T) {
	c := fakeProcess(t, "go version go1.26.0 darwin/arm64\n", true)
	AssertCLI(t, c, CLITest{
		Cmd:      "go",
		Args:     []string{"version"},
		WantOK:   true,
		Contains: "go1.26",
	})
}

func TestCLI_AssertCLI_WantOK_False_Good(t *T) {
	c := fakeProcess(t, "process not registered", false)
	AssertCLI(t, c, CLITest{
		Cmd:    "missing-binary",
		WantOK: false,
	})
}

func TestCLI_AssertCLI_Dir_Good(t *T) {
	c := fakeProcess(t, "ok\n", true)
	AssertCLI(t, c, CLITest{
		Name:   "run-in-tempdir",
		Cmd:    "true",
		Dir:    t.TempDir(),
		WantOK: true,
		Output: "ok\n",
	})
}

func TestCLI_AssertCLI_Env_Good(t *T) {
	c := fakeProcess(t, "ok\n", true)
	AssertCLI(t, c, CLITest{
		Cmd:    "true",
		Dir:    t.TempDir(),
		Env:    []string{"GOWORK=off"},
		WantOK: true,
		Output: "ok\n",
	})
}

// --- AssertCLIs ---

func TestCLI_AssertCLIs_Good(t *T) {
	c := fakeProcess(t, "fixture-output\n", true)
	AssertCLIs(t, c, []CLITest{
		{Name: "first", Cmd: "go", Args: []string{"version"}, WantOK: true, Contains: "fixture"},
		{Name: "second", Cmd: "go", Args: []string{"env"}, WantOK: true, Output: "fixture-output\n"},
	})
}

// --- AX-7 canonical triplets ---

func TestCli_CliRegister_Good(t *T) {
	c := New()
	r := c.Service("cli")
	AssertTrue(t, r.OK)
	AssertNotNil(t, c.Cli())
}

func TestCli_CliRegister_Bad(t *T) {
	c := New()
	r := CliRegister(c)
	AssertFalse(t, r.OK)
}

func TestCli_CliRegister_Ugly(t *T) {
	AssertPanics(t, func() {
		CliRegister(nil)
	})
}

func TestCli_Cli_Print_Good(t *T) {
	c := New()
	buf := NewBuffer()
	c.Cli().SetOutput(buf)
	c.Cli().Print("agent %s ready", "codex")
	AssertEqual(t, "agent codex ready\n", buf.String())
}

func TestCli_Cli_Print_Bad(t *T) {
	c := New()
	buf := NewBuffer()
	c.Cli().SetOutput(buf)
	c.Cli().Print("agent %s ready")
	AssertContains(t, buf.String(), "%!s(MISSING)")
}

func TestCli_Cli_Print_Ugly(t *T) {
	c := New()
	buf := NewBuffer()
	c.Cli().SetOutput(buf)
	c.Cli().Print("")
	AssertEqual(t, "\n", buf.String())
}

func TestCli_Cli_SetOutput_Good(t *T) {
	c := New()
	buf := NewBuffer()
	c.Cli().SetOutput(buf)
	c.Cli().Print("homelab")
	AssertEqual(t, "homelab\n", buf.String())
}

func TestCli_Cli_SetOutput_Bad(t *T) {
	c := New()
	buf := NewBuffer()
	c.Cli().SetOutput(buf)
	c.Cli().SetOutput(buf)
	c.Cli().Print("agent")
	AssertEqual(t, "agent\n", buf.String())
}

func TestCli_Cli_SetOutput_Ugly(t *T) {
	c := New()
	first := NewBuffer()
	second := NewBuffer()
	c.Cli().SetOutput(first)
	c.Cli().Print("first")
	c.Cli().SetOutput(second)
	c.Cli().Print("second")
	AssertEqual(t, "first\n", first.String())
	AssertEqual(t, "second\n", second.String())
}

func TestCli_Cli_Run_Good(t *T) {
	c := New()
	var target string
	c.Command("deploy/to/homelab", Command{Action: func(opts Options) Result {
		target = opts.String("target")
		return Result{Value: "deployed", OK: true}
	}})
	r := c.Cli().Run("deploy", "to", "homelab", "--target=lethean")
	AssertTrue(t, r.OK)
	AssertEqual(t, "lethean", target)
}

func TestCli_Cli_Run_Bad(t *T) {
	c := New()
	c.Command("agent/status", Command{})
	r := c.Cli().Run("agent", "status")
	AssertFalse(t, r.OK)
}

func TestCli_Cli_Run_Ugly(t *T) {
	c := New(WithOption("name", "homelab"))
	buf := NewBuffer()
	c.Cli().SetOutput(buf)
	c.Cli().SetBanner(func(_ *Cli) string { return "homelab ops" })
	r := c.Cli().Run("missing")
	AssertFalse(t, r.OK)
	AssertContains(t, buf.String(), "homelab ops")
}

func TestCli_Cli_RunMissingCommand_Bad(t *T) {
	c := New(WithOption("name", "homelab"))
	buf := NewBuffer()
	c.Cli().SetOutput(buf)
	c.Cli().SetBanner(func(_ *Cli) string { return "homelab ops" })
	c.Command("system/status", Command{Action: func(_ Options) Result { return Result{OK: true} }})

	r := c.Cli().Run("agent", "missing")

	AssertFalse(t, r.OK)
	AssertContains(t, buf.String(), "homelab ops")
	AssertContains(t, buf.String(), "system/status")
}

func TestCli_Cli_RunFlagForms_Ugly(t *T) {
	c := New()
	var dryRun bool
	var name string
	c.Command("agent/run", Command{Action: func(opts Options) Result {
		dryRun = opts.Bool("dry-run")
		name = opts.String("_arg")
		return Result{OK: true}
	}})

	r := c.Cli().Run("agent", "run", "--dry-run", "codex")

	AssertTrue(t, r.OK)
	AssertTrue(t, dryRun)
	AssertEqual(t, "codex", name)
}

func TestCli_Cli_PrintHelp_Good(t *T) {
	c := New(WithOption("name", "homelab"))
	buf := NewBuffer()
	c.Cli().SetOutput(buf)
	c.Command("agent/status", Command{Action: func(_ Options) Result { return Result{OK: true} }})
	c.Cli().PrintHelp()
	AssertContains(t, buf.String(), "homelab commands:")
	AssertContains(t, buf.String(), "agent/status")
}

func TestCli_Cli_PrintHelp_Bad(t *T) {
	c := New()
	buf := NewBuffer()
	c.Cli().SetOutput(buf)
	c.Command("agent/hidden", Command{Hidden: true, Action: func(_ Options) Result { return Result{OK: true} }})
	c.Cli().PrintHelp()
	AssertNotContains(t, buf.String(), "agent/hidden")
}

func TestCli_Cli_PrintHelp_Ugly(t *T) {
	c := New()
	buf := NewBuffer()
	c.Cli().SetOutput(buf)
	c.Cli().PrintHelp()
	AssertEqual(t, "Commands:\n", buf.String())
}

func TestCli_Cli_PrintHelpTranslated_Good(t *T) {
	c := New()
	buf := NewBuffer()
	c.Cli().SetOutput(buf)
	c.I18n().SetTranslator(&mockTranslator{})
	c.Command("agent/status", Command{Action: func(_ Options) Result { return Result{OK: true} }})

	c.Cli().PrintHelp()

	AssertContains(t, buf.String(), "agent/status")
	AssertContains(t, buf.String(), "translated:cmd.agent.status.description")
}

func TestCli_Cli_SetBanner_Good(t *T) {
	c := New()
	c.Cli().SetBanner(func(_ *Cli) string { return "dAppCore agent" })
	AssertEqual(t, "dAppCore agent", c.Cli().Banner())
}

func TestCli_Cli_SetBanner_Bad(t *T) {
	c := New(WithOption("name", "homelab"))
	c.Cli().SetBanner(nil)
	AssertEqual(t, "homelab", c.Cli().Banner())
}

func TestCli_Cli_SetBanner_Ugly(t *T) {
	c := New(WithOption("name", "homelab"))
	c.Cli().SetBanner(func(cl *Cli) string {
		return Concat(cl.Core().App().Name, " banner")
	})
	AssertEqual(t, "homelab banner", c.Cli().Banner())
}

func TestCli_Cli_Banner_Good(t *T) {
	c := New()
	c.Cli().SetBanner(func(_ *Cli) string { return "agent dispatch" })
	AssertEqual(t, "agent dispatch", c.Cli().Banner())
}

func TestCli_Cli_Banner_Bad(t *T) {
	c := New(WithOption("name", "homelab"))
	AssertEqual(t, "homelab", c.Cli().Banner())
}

func TestCli_Cli_Banner_Ugly(t *T) {
	c := New()
	AssertEqual(t, "", c.Cli().Banner())
}
