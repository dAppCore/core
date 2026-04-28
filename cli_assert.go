// SPDX-License-Identifier: EUPL-1.2

// AX-10 CLI test assertions for the Core framework.
//
// CLITest declares a single binary scenario; AssertCLI dispatches it
// through c.Process() (so the Core must register a "process.run"
// service — typically dappco.re/go-process). AssertCLIs runs a
// table-driven set of cases under sub-tests.
//
// Pair these with `tests/cli/{path}/Taskfile.yaml` per AX-10 — the
// Taskfile drives the `go test` invocation and the CLITest values
// describe the binary behaviour fixtures.
package core

// CLITest is a single CLI scenario. The runner executes Cmd with Args
// in Dir (current working directory if empty), feeds Stdin, and asserts
// the resulting Result matches Want* expectations. Use Output for an
// exact-match check or Contains for a substring assertion.
//
//	tc := core.CLITest{
//	    Cmd: "git", Args: []string{"--version"},
//	    WantOK: true, Contains: "git version",
//	}
type CLITest struct {
	Name     string   // optional scenario name for failure messages
	Cmd      string   // binary to invoke (PATH-resolvable or absolute)
	Args     []string // command-line arguments
	Dir      string   // working directory; "" means current
	Stdin    string   // stdin content; "" means none
	Env      []string // additional env vars (e.g. "GOWORK=off")
	WantOK   bool     // expected Result.OK
	Output   string   // exact stdout match (skipped when "")
	Contains string   // substring stdout match (skipped when "")
}

// AssertCLI executes tc via c.Process() and asserts the outcome. The
// Core must have a service registering "process.run" — typically
// dappco.re/go-process — for the dispatch to reach a real exec.
//
//	c := core.New(core.WithService(process.Register))
//	core.AssertCLI(t, c, core.CLITest{
//	    Cmd: "go", Args: []string{"version"},
//	    WantOK: true, Contains: "go1.",
//	})
func AssertCLI(t TB, c *Core, tc CLITest) {
	t.Helper()

	var r Result
	switch {
	case tc.Dir != "" && len(tc.Env) > 0:
		r = c.Process().RunWithEnv(Background(), tc.Dir, tc.Env, tc.Cmd, tc.Args...)
	case tc.Dir != "":
		r = c.Process().RunIn(Background(), tc.Dir, tc.Cmd, tc.Args...)
	default:
		r = c.Process().Run(Background(), tc.Cmd, tc.Args...)
	}

	label := tc.Name
	if label == "" {
		label = tc.Cmd
	}

	if tc.WantOK != r.OK {
		assertFail(t, false, "AssertCLI", nil, "label", label, "want.OK", tc.WantOK, "got.OK", r.OK, "err", r.Error())
		return
	}
	if !r.OK {
		return
	}

	out, _ := r.Value.(string)
	if tc.Output != "" && out != tc.Output {
		assertFail(t, false, "AssertCLI", nil, "label", label, "stdout-want", tc.Output, "got", out)
	}
	if tc.Contains != "" && !Contains(out, tc.Contains) {
		assertFail(t, false, "AssertCLI", nil, "label", label, "stdout-contains", tc.Contains, "got", out)
	}
}

// AssertCLIs runs each CLITest sequentially under a sub-test named after
// tc.Name (or tc.Cmd if Name is empty). Convenience for table-driven
// CLI scenarios.
//
//	core.AssertCLIs(t, c, []core.CLITest{
//	    {Name: "version", Cmd: "go", Args: []string{"version"}, WantOK: true, Contains: "go1."},
//	    {Name: "vet", Cmd: "go", Args: []string{"vet", "./..."}, WantOK: true},
//	})
func AssertCLIs(t *T, c *Core, cases []CLITest) {
	t.Helper()
	for _, tc := range cases {
		name := tc.Name
		if name == "" {
			name = tc.Cmd
		}
		t.Run(name, func(t *T) {
			AssertCLI(t, c, tc)
		})
	}
}
