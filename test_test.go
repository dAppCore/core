package core_test

import (
	. "dappco.re/go"
)

func TestTest_AssertCLI_Bad(t *T) {
	c := New()
	AssertCLI(t, c, CLITest{
		Name:   "process capability refused",
		Cmd:    "agent-dispatch",
		WantOK: false,
	})
}

func TestTest_AssertCLI_Ugly(t *T) {
	c := New()
	var captured Options
	c.Action("process.run", func(_ Context, opts Options) Result {
		captured = opts
		return Result{Value: "homelab health check ok\n", OK: true}
	})

	dir := t.TempDir()
	AssertCLI(t, c, CLITest{
		Name:     "env and directory dispatch",
		Cmd:      "agent-dispatch",
		Args:     []string{"health", "--site=homelab"},
		Dir:      dir,
		Env:      []string{"CORE_AGENT=codex"},
		WantOK:   true,
		Contains: "health check ok",
	})

	AssertEqual(t, dir, captured.String("dir"))
	env, ok := captured.Get("env").Value.([]string)
	AssertTrue(t, ok)
	AssertContains(t, env, "CORE_AGENT=codex")
}

func TestTest_AssertCLIs_Bad(t *T) {
	AssertCLIs(t, New(), []CLITest{
		{Name: "missing process runner", Cmd: "agent-dispatch", WantOK: false},
	})
}

func TestTest_AssertCLIs_Ugly(t *T) {
	AssertCLIs(t, New(), nil)
}

func TestTest_AssertContains_Good(t *T) {
	AssertContains(t, "agent dispatch completed", "dispatch")
}

func TestTest_AssertContains_Bad(t *T) {
	AssertContains(t, map[string]int{"session": 1}, "session")
}

func TestTest_AssertContains_Ugly(t *T) {
	AssertContains(t, "", "")
}

func TestTest_AssertElementsMatch_Good(t *T) {
	AssertElementsMatch(t, []string{"agent", "health", "deploy"}, []string{"deploy", "agent", "health"})
}

func TestTest_AssertElementsMatch_Bad(t *T) {
	AssertElementsMatch(t, []int{1, 1, 2}, []int{1, 2, 1})
}

func TestTest_AssertElementsMatch_Ugly(t *T) {
	AssertElementsMatch(t, []string{}, []string{})
}

func TestTest_AssertEmpty_Good(t *T) {
	AssertEmpty(t, []string{})
}

func TestTest_AssertEmpty_Bad(t *T) {
	AssertEmpty(t, nil)
}

func TestTest_AssertEmpty_Ugly(t *T) {
	AssertEmpty(t, 0)
}

func TestTest_AssertEqual_Good(t *T) {
	AssertEqual(t, "agent dispatch", "agent dispatch")
}

func TestTest_AssertEqual_Bad(t *T) {
	AssertEqual(t, NewError("refused").Error(), "refused")
}

func TestTest_AssertEqual_Ugly(t *T) {
	var left []string
	var right []string
	AssertEqual(t, left, right)
}

func TestTest_AssertError_Good(t *T) {
	AssertError(t, AnError)
}

func TestTest_AssertError_Bad(t *T) {
	AssertError(t, E("agent.Dispatch", "session token refused", nil), "session token refused")
}

func TestTest_AssertError_Ugly(t *T) {
	AssertError(t, Wrap(AnError, "homelab.Health", "nested failure"), "nested failure")
}

func TestTest_AssertErrorIs_Good(t *T) {
	AssertErrorIs(t, Wrap(AnError, "agent.Dispatch", "failed"), AnError)
}

func TestTest_AssertErrorIs_Bad(t *T) {
	joined := ErrorJoin(NewError("first"), AnError)
	AssertErrorIs(t, joined, AnError)
}

func TestTest_AssertErrorIs_Ugly(t *T) {
	AssertErrorIs(t, nil, nil)
}

func TestTest_AssertFalse_Good(t *T) {
	AssertFalse(t, false)
}

func TestTest_AssertFalse_Bad(t *T) {
	AssertFalse(t, Result{}.OK)
}

func TestTest_AssertFalse_Ugly(t *T) {
	var token *string
	AssertFalse(t, token != nil)
}

func TestTest_AssertGreater_Good(t *T) {
	AssertGreater(t, 8, 2)
}

func TestTest_AssertGreater_Bad(t *T) {
	AssertGreater(t, "session-token-b", "session-token-a")
}

func TestTest_AssertGreater_Ugly(t *T) {
	AssertGreater(t, uint(1), -1)
}

func TestTest_AssertGreaterOrEqual_Good(t *T) {
	AssertGreaterOrEqual(t, 3, 3)
}

func TestTest_AssertGreaterOrEqual_Bad(t *T) {
	AssertGreaterOrEqual(t, "deploy", "deploy")
}

func TestTest_AssertGreaterOrEqual_Ugly(t *T) {
	AssertGreaterOrEqual(t, 0.5, 0)
}

func TestTest_AssertInDelta_Good(t *T) {
	AssertInDelta(t, 1.00, 1.01, 0.02)
}

func TestTest_AssertInDelta_Bad(t *T) {
	AssertInDelta(t, 10, 10, 0)
}

func TestTest_AssertInDelta_Ugly(t *T) {
	AssertInDelta(t, -0.1, -0.1001, 0.001)
}

func TestTest_AssertLen_Good(t *T) {
	AssertLen(t, []string{"agent", "dispatch"}, 2)
}

func TestTest_AssertLen_Bad(t *T) {
	AssertLen(t, map[string]bool{"refused": true}, 1)
}

func TestTest_AssertLen_Ugly(t *T) {
	AssertLen(t, make(chan string, 3), 0)
}

func TestTest_AssertLess_Good(t *T) {
	AssertLess(t, 2, 8)
}

func TestTest_AssertLess_Bad(t *T) {
	AssertLess(t, "agent-a", "agent-b")
}

func TestTest_AssertLess_Ugly(t *T) {
	AssertLess(t, -1, uint(1))
}

func TestTest_AssertLessOrEqual_Good(t *T) {
	AssertLessOrEqual(t, 3, 3)
}

func TestTest_AssertLessOrEqual_Bad(t *T) {
	AssertLessOrEqual(t, "deploy", "deploy")
}

func TestTest_AssertLessOrEqual_Ugly(t *T) {
	AssertLessOrEqual(t, 0, 0.5)
}

func TestTest_AssertNil_Good(t *T) {
	AssertNil(t, nil)
}

func TestTest_AssertNil_Bad(t *T) {
	var session *string
	AssertNil(t, session)
}

func TestTest_AssertNil_Ugly(t *T) {
	var meta map[string]string
	AssertNil(t, meta)
}

func TestTest_AssertNoError_Good(t *T) {
	AssertNoError(t, nil)
}

func TestTest_AssertNoError_Bad(t *T) {
	AssertNoError(t, nil)
}

func TestTest_AssertNoError_Ugly(t *T) {
	var err error
	AssertNoError(t, err)
}

func TestTest_AssertNotContains_Good(t *T) {
	AssertNotContains(t, []string{"agent", "deploy"}, "rollback")
}

func TestTest_AssertNotContains_Bad(t *T) {
	AssertNotContains(t, map[string]int{"session": 1}, "missing")
}

func TestTest_AssertNotContains_Ugly(t *T) {
	AssertNotContains(t, nil, "agent")
}

func TestTest_AssertNotEmpty_Good(t *T) {
	AssertNotEmpty(t, "agent")
}

func TestTest_AssertNotEmpty_Bad(t *T) {
	AssertNotEmpty(t, []string{"refused"})
}

func TestTest_AssertNotEmpty_Ugly(t *T) {
	AssertNotEmpty(t, struct{ Agent string }{Agent: "codex"})
}

func TestTest_AssertNotEqual_Good(t *T) {
	AssertNotEqual(t, "agent-a", "agent-b")
}

func TestTest_AssertNotEqual_Bad(t *T) {
	AssertNotEqual(t, 1, "1")
}

func TestTest_AssertNotEqual_Ugly(t *T) {
	AssertNotEqual(t, []string(nil), []string{})
}

func TestTest_AssertNotNil_Good(t *T) {
	AssertNotNil(t, "agent")
}

func TestTest_AssertNotNil_Bad(t *T) {
	AssertNotNil(t, "")
}

func TestTest_AssertNotNil_Ugly(t *T) {
	AssertNotNil(t, struct{}{})
}

func TestTest_AssertNotPanics_Good(t *T) {
	AssertNotPanics(t, func() { /* no-op closure verifies non-panicking behaviour */ })
}

func TestTest_AssertNotPanics_Bad(t *T) {
	AssertNotPanics(t, func() {
		_ = Result{}
	})
}

func TestTest_AssertNotPanics_Ugly(t *T) {
	AssertNotPanics(t, func() {
		var opts Options
		_ = opts.Len()
	})
}

func TestTest_AssertPanics_Good(t *T) {
	AssertPanics(t, func() { panic("agent halted") })
}

func TestTest_AssertPanics_Bad(t *T) {
	AssertPanics(t, func() { panic(AnError) })
}

func TestTest_AssertPanics_Ugly(t *T) {
	AssertPanics(t, func() {
		var values []string
		_ = values[1]
	})
}

func TestTest_AssertPanicsWithError_Good(t *T) {
	AssertPanicsWithError(t, "session token expired", func() {
		panic(NewError("session token expired"))
	})
}

func TestTest_AssertPanicsWithError_Bad(t *T) {
	AssertPanicsWithError(t, "entitlement denied", func() {
		panic("entitlement denied")
	})
}

func TestTest_AssertPanicsWithError_Ugly(t *T) {
	AssertPanicsWithError(t, "", func() {
		panic("")
	})
}

func TestTest_AssertSame_Good(t *T) {
	c := New()
	AssertSame(t, c, c.Core())
}

func TestTest_AssertSame_Bad(t *T) {
	log := Default()
	AssertSame(t, log, log)
}

func TestTest_AssertSame_Ugly(t *T) {
	var left *Core
	var right *Core
	AssertSame(t, left, right)
}

func TestTest_AssertTrue_Good(t *T) {
	AssertTrue(t, true)
}

func TestTest_AssertTrue_Bad(t *T) {
	AssertTrue(t, !Result{}.OK)
}

func TestTest_AssertTrue_Ugly(t *T) {
	AssertTrue(t, len([]string{}) == 0)
}

func TestTest_RequireNoError_Good(t *T) {
	RequireNoError(t, nil)
}

func TestTest_RequireNoError_Bad(t *T) {
	var err error
	RequireNoError(t, err)
}

func TestTest_RequireNoError_Ugly(t *T) {
	RequireNoError(t, nil)
}

func TestTest_RequireNotEmpty_Good(t *T) {
	RequireNotEmpty(t, "agent")
}

func TestTest_RequireNotEmpty_Bad(t *T) {
	RequireNotEmpty(t, []string{"refused"})
}

func TestTest_RequireNotEmpty_Ugly(t *T) {
	RequireNotEmpty(t, struct{ Agent string }{Agent: "codex"})
}

func TestTest_RequireTrue_Good(t *T) {
	RequireTrue(t, true)
}

func TestTest_RequireTrue_Bad(t *T) {
	RequireTrue(t, !Result{}.OK)
}

func TestTest_RequireTrue_Ugly(t *T) {
	RequireTrue(t, len([]string{}) == 0)
}
