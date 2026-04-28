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
