package core_test

import (
	. "dappco.re/go"
)

func TestResult_Result_Error_Good(t *T) {
	r := Result{Value: NewError("agent dispatch failed"), OK: false}
	AssertEqual(t, "agent dispatch failed", r.Error())
}

func TestResult_Result_Error_Bad(t *T) {
	r := Result{Value: "session token refused", OK: false}
	AssertEqual(t, "session token refused", r.Error())
}

func TestResult_Result_Error_Ugly(t *T) {
	r := Result{OK: false}
	AssertEqual(t, "unknown error", r.Error())
}

func TestResult_Result_Code_Good(t *T) {
	r := Result{Value: NewCode("agent.refused", "dispatch refused"), OK: false}
	AssertEqual(t, "agent.refused", r.Code())
}

func TestResult_Result_Code_Bad(t *T) {
	r := Result{Value: NewError("plain failure"), OK: false}
	AssertEqual(t, "", r.Code())
}

func TestResult_Result_Code_Ugly(t *T) {
	r := Result{Value: NewCode("agent.refused", "dispatch refused"), OK: true}
	AssertEqual(t, "", r.Code())
}

func TestResult_Result_Must_Good(t *T) {
	r := Result{Value: "agent-ready", OK: true}
	AssertEqual(t, "agent-ready", r.Must())
}

func TestResult_Result_Must_Bad(t *T) {
	r := Result{Value: NewError("session token expired"), OK: false}
	AssertPanicsWithError(t, "session token expired", func() {
		_ = r.Must()
	})
}

func TestResult_Result_Must_Ugly(t *T) {
	r := Result{Value: "panic text", OK: false}
	AssertPanicsWithError(t, "panic text", func() {
		_ = r.Must()
	})
}

func TestResult_Result_Or_Good(t *T) {
	r := Result{Value: "primary agent", OK: true}
	AssertEqual(t, "primary agent", r.Or("fallback agent"))
}

func TestResult_Result_Or_Bad(t *T) {
	r := Result{Value: NewError("missing agent"), OK: false}
	AssertEqual(t, "fallback agent", r.Or("fallback agent"))
}

func TestResult_Result_Or_Ugly(t *T) {
	r := Result{Value: nil, OK: true}
	AssertNil(t, r.Or("fallback agent"))
}

func TestResult_Cast_Good(t *T) {
	value, ok := Cast[string](Result{Value: "codex", OK: true})
	AssertTrue(t, ok)
	AssertEqual(t, "codex", value)
}

func TestResult_Cast_Bad(t *T) {
	value, ok := Cast[string](Result{Value: "codex", OK: false})
	AssertFalse(t, ok)
	AssertEqual(t, "", value)
}

func TestResult_Cast_Ugly(t *T) {
	value, ok := Cast[int](Result{Value: "codex", OK: true})
	AssertFalse(t, ok)
	AssertEqual(t, 0, value)
}

func TestResult_Try_Good(t *T) {
	r := Try(func() any { return "dispatch-complete" })
	AssertTrue(t, r.OK)
	AssertEqual(t, "dispatch-complete", r.Value)
}

func TestResult_Try_Bad(t *T) {
	r := Try(func() any { return NewError("dispatch refused") })
	AssertFalse(t, r.OK)
	AssertError(t, r.Value.(error), "dispatch refused")
}

func TestResult_Try_Ugly(t *T) {
	r := Try(func() any { panic("worker panic") })
	AssertFalse(t, r.OK)
	AssertError(t, r.Value.(error), "panic recovered")
}

// --- Ok ---

func TestResult_Ok_Good(t *T) {
	r := Ok(42)
	AssertTrue(t, r.OK)
	AssertEqual(t, 42, r.Value)
}

func TestResult_Ok_Bad(t *T) {
	// Ok with nil value — still OK=true; consumer can opt to treat nil as a sentinel.
	r := Ok(nil)
	AssertTrue(t, r.OK)
	AssertNil(t, r.Value)
}

func TestResult_Ok_Ugly(t *T) {
	// Wrapping an error in Ok still produces OK=true (caller's choice).
	r := Ok(NewError("warning"))
	AssertTrue(t, r.OK)
	AssertNotNil(t, r.Value)
}

// --- Err ---

func TestResult_Fail_Good(t *T) {
	r := Fail(NewError("dispatch failed"))
	AssertFalse(t, r.OK)
	AssertContains(t, r.Error(), "dispatch failed")
}

func TestResult_Fail_Bad(t *T) {
	// Err with nil — produces OK=false, Value nil.
	r := Fail(nil)
	AssertFalse(t, r.OK)
	AssertNil(t, r.Value)
}

func TestResult_Fail_Ugly(t *T) {
	// Err with a coded error — Code() pulls through.
	r := Fail(NewCode("net.timeout", "homelab unreachable"))
	AssertFalse(t, r.OK)
	AssertEqual(t, "net.timeout", r.Code())
}

// --- ResultOf ---

func TestResult_ResultOf_Good(t *T) {
	r := ResultOf("payload", nil)
	AssertTrue(t, r.OK)
	AssertEqual(t, "payload", r.Value)
}

func TestResult_ResultOf_Bad(t *T) {
	r := ResultOf("payload", NewError("network"))
	AssertFalse(t, r.OK)
	AssertContains(t, r.Error(), "network")
}

func TestResult_ResultOf_Ugly(t *T) {
	// (nil, nil) is treated as success.
	r := ResultOf(nil, nil)
	AssertTrue(t, r.OK)
	AssertNil(t, r.Value)
}

// --- MustCast ---

func TestResult_MustCast_Good(t *T) {
	r := Ok("agent.dispatch")
	got := MustCast[string](r)
	AssertEqual(t, "agent.dispatch", got)
}

func TestResult_MustCast_Bad(t *T) {
	// Failed Result panics with the underlying error.
	AssertPanics(t, func() {
		_ = MustCast[string](Fail(NewError("boom")))
	})
}

func TestResult_MustCast_Ugly(t *T) {
	// Type mismatch panics with a descriptive error.
	AssertPanicsWithError(t, "not", func() {
		_ = MustCast[*int](Ok("string-not-pointer"))
	})
}
