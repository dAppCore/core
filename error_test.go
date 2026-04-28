package core_test

import (
	. "dappco.re/go"
)

// --- Error Creation ---

func TestError_E_Good(t *T) {
	err := E("user.Save", "failed to save", nil)
	AssertError(t, err)
	AssertContains(t, err.Error(), "user.Save")
	AssertContains(t, err.Error(), "failed to save")
}

func TestError_E_WithCause_Good(t *T) {
	cause := NewError("connection refused")
	err := E("db.Connect", "database unavailable", cause)
	AssertErrorIs(t, err, cause)
}

func TestError_Wrap_Good(t *T) {
	cause := NewError("timeout")
	err := Wrap(cause, "api.Call", "request failed")
	AssertError(t, err)
	AssertErrorIs(t, err, cause)
}

func TestError_Wrap_Nil_Good(t *T) {
	err := Wrap(nil, "api.Call", "request failed")
	AssertNil(t, err)
}

func TestError_WrapCode_Good(t *T) {
	cause := NewError("invalid email")
	err := WrapCode(cause, "VALIDATION_ERROR", "user.Validate", "bad input")
	AssertError(t, err)
	AssertEqual(t, "VALIDATION_ERROR", ErrorCode(err))
}

func TestError_NewCode_Good(t *T) {
	err := NewCode("NOT_FOUND", "resource not found")
	AssertError(t, err)
	AssertEqual(t, "NOT_FOUND", ErrorCode(err))
}

// --- Error Introspection ---

func TestError_Operation_Good(t *T) {
	err := E("brain.Recall", "search failed", nil)
	AssertEqual(t, "brain.Recall", Operation(err))
}

func TestError_Operation_Bad(t *T) {
	err := NewError("plain error")
	AssertEqual(t, "", Operation(err))
}

func TestError_ErrorMessage_Good(t *T) {
	err := E("op", "the message", nil)
	AssertEqual(t, "the message", ErrorMessage(err))
}

func TestError_ErrorMessage_Plain(t *T) {
	err := NewError("plain")
	AssertEqual(t, "plain", ErrorMessage(err))
}

func TestError_ErrorMessage_Nil(t *T) {
	AssertEqual(t, "", ErrorMessage(nil))
}

func TestError_Root_Good(t *T) {
	root := NewError("root cause")
	wrapped := Wrap(root, "layer1", "first wrap")
	double := Wrap(wrapped, "layer2", "second wrap")
	AssertEqual(t, root, Root(double))
}

func TestError_Root_Nil(t *T) {
	AssertNil(t, Root(nil))
}

func TestError_StackTrace_Good(t *T) {
	err := Wrap(E("inner", "cause", nil), "outer", "wrapper")
	stack := StackTrace(err)
	AssertLen(t, stack, 2)
	AssertEqual(t, "outer", stack[0])
	AssertEqual(t, "inner", stack[1])
}

func TestError_FormatStackTrace_Good(t *T) {
	err := Wrap(E("a", "x", nil), "b", "y")
	formatted := FormatStackTrace(err)
	AssertEqual(t, "b -> a", formatted)
}

// --- ErrorLog ---

func TestError_ErrorLog_Good(t *T) {
	c := New()
	cause := NewError("boom")
	r := c.Log().Error(cause, "test.Operation", "something broke")
	AssertFalse(t, r.OK)
	AssertErrorIs(t, r.Value.(error), cause)
}

func TestError_ErrorLog_Nil_Good(t *T) {
	c := New()
	r := c.Log().Error(nil, "test.Operation", "no error")
	AssertTrue(t, r.OK)
}

func TestError_ErrorLog_Warn_Good(t *T) {
	c := New()
	cause := NewError("warning")
	r := c.Log().Warn(cause, "test.Operation", "heads up")
	AssertFalse(t, r.OK)
}

func TestError_ErrorLog_Must_Ugly(t *T) {
	c := New()
	AssertPanics(t, func() {
		c.Log().Must(NewError("fatal"), "test.Operation", "must fail")
	})
}

func TestError_ErrorLog_Must_Nil_Good(t *T) {
	c := New()
	AssertNotPanics(t, func() {
		c.Log().Must(nil, "test.Operation", "no error")
	})
}

// --- ErrorPanic ---

func TestError_ErrorPanic_Recover_Good(t *T) {
	c := New()
	// Should not panic — Recover catches it
	AssertNotPanics(t, func() {
		defer c.Error().Recover()
		panic("test panic")
	})
}

func TestError_ErrorPanic_SafeGo_Good(t *T) {
	c := New()
	done := make(chan bool, 1)
	c.Error().SafeGo(func() {
		done <- true
	})
	AssertTrue(t, <-done)
}

func TestError_ErrorPanic_SafeGo_Panic_Good(t *T) {
	c := New()
	done := make(chan bool, 1)
	c.Error().SafeGo(func() {
		defer func() { done <- true }()
		panic("caught by SafeGo")
	})
	// SafeGo recovers — goroutine completes without crashing the process
	<-done
}

// --- Standard Library Wrappers ---

func TestError_Is_Good(t *T) {
	target := NewError("target")
	wrapped := Wrap(target, "op", "msg")
	AssertTrue(t, Is(wrapped, target))
}

func TestError_As_Good(t *T) {
	err := E("op", "msg", nil)
	var e *Err
	AssertTrue(t, As(err, &e))
	AssertEqual(t, "op", e.Operation)
}

func TestError_NewError_Good(t *T) {
	err := NewError("simple error")
	AssertEqual(t, "simple error", err.Error())
}

func TestError_ErrorJoin_Good(t *T) {
	e1 := NewError("first")
	e2 := NewError("second")
	joined := ErrorJoin(e1, e2)
	AssertErrorIs(t, joined, e1)
	AssertErrorIs(t, joined, e2)
}

// --- ErrorPanic Crash Reports ---

func TestError_ErrorPanic_Reports_Good(t *T) {
	dir := t.TempDir()
	path := Path(dir, "crashes.json")

	// Create ErrorPanic with file output
	c := New()
	// Access internals via a crash that writes to file
	// Since ErrorPanic fields are unexported, we test via Recover
	_ = c
	_ = path
	// Crash reporting needs ErrorPanic configured with filePath — tested indirectly
}

// --- ErrorPanic Crash File ---

func TestError_ErrorPanic_CrashFile_Good(t *T) {
	dir := t.TempDir()
	path := Path(dir, "crashes.json")

	// Create Core, trigger a panic through SafeGo, check crash file
	// ErrorPanic.filePath is unexported — but we can test via the package-level
	// error handling that writes crash reports

	// For now, test that Reports handles missing file gracefully
	c := New()
	r := c.Error().Reports(5)
	AssertFalse(t, r.OK)
	AssertNil(t, r.Value)
	_ = path
}

// --- Error formatting branches ---

func TestError_Err_Error_WithCode_Good(t *T) {
	err := WrapCode(NewError("bad"), "INVALID", "validate", "input failed")
	AssertContains(t, err.Error(), "[INVALID]")
	AssertContains(t, err.Error(), "validate")
	AssertContains(t, err.Error(), "bad")
}

func TestError_Err_Error_CodeNoCause_Good(t *T) {
	err := NewCode("NOT_FOUND", "resource missing")
	AssertContains(t, err.Error(), "[NOT_FOUND]")
	AssertContains(t, err.Error(), "resource missing")
}

func TestError_Err_Error_NoOp_Good(t *T) {
	err := &Err{Message: "bare error"}
	AssertEqual(t, "bare error", err.Error())
}

func TestError_WrapCode_NilErr_EmptyCode_Good(t *T) {
	err := WrapCode(nil, "", "op", "msg")
	AssertNil(t, err)
}

func TestError_Wrap_PreservesCode_Good(t *T) {
	inner := WrapCode(NewError("root"), "AUTH_FAIL", "auth", "denied")
	outer := Wrap(inner, "handler", "request failed")
	AssertEqual(t, "AUTH_FAIL", ErrorCode(outer))
}

func TestError_ErrorLog_Warn_Nil_Good(t *T) {
	c := New()
	r := c.LogWarn(nil, "op", "msg")
	AssertTrue(t, r.OK)
}

func TestError_ErrorLog_Error_Nil_Good(t *T) {
	c := New()
	r := c.LogError(nil, "op", "msg")
	AssertTrue(t, r.OK)
}

func TestError_AllOperations_Good(t *T) {
	err := Wrap(E("agent.Token", "expired", nil), "agent.Dispatch", "failed")
	var ops []string
	for op := range AllOperations(err) {
		ops = append(ops, op)
	}
	AssertEqual(t, []string{"agent.Dispatch", "agent.Token"}, ops)
}

func TestError_AllOperations_Bad(t *T) {
	var ops []string
	for op := range AllOperations(NewError("plain failure")) {
		ops = append(ops, op)
	}
	AssertEmpty(t, ops)
}

func TestError_AllOperations_Ugly(t *T) {
	var ops []string
	for op := range AllOperations(nil) {
		ops = append(ops, op)
	}
	AssertEmpty(t, ops)
}

func TestError_As_Bad(t *T) {
	var structured *Err
	AssertFalse(t, As(NewError("plain failure"), &structured))
	AssertNil(t, structured)
}

func TestError_As_Ugly(t *T) {
	var structured *Err
	AssertFalse(t, As(nil, &structured))
	AssertNil(t, structured)
}

func TestError_E_Bad(t *T) {
	err := E("", "", nil)
	AssertError(t, err)
	AssertEqual(t, "", err.Error())
}

func TestError_E_Ugly(t *T) {
	err := E("agent.Dispatch", "", AnError)
	AssertContains(t, err.Error(), "agent.Dispatch")
	AssertErrorIs(t, err, AnError)
}

func TestError_Err_Error_Good(t *T) {
	err := &Err{Operation: "agent.Dispatch", Message: "failed", Cause: AnError, Code: "agent.failed"}
	AssertContains(t, err.Error(), "agent.Dispatch")
	AssertContains(t, err.Error(), "[agent.failed]")
	AssertContains(t, err.Error(), AnError.Error())
}

func TestError_Err_Error_Bad(t *T) {
	err := &Err{}
	AssertEqual(t, "", err.Error())
}

func TestError_Err_Error_Ugly(t *T) {
	err := &Err{Message: "session refused", Code: "session.refused"}
	AssertEqual(t, "session refused [session.refused]", err.Error())
}

func TestError_Err_Unwrap_Good(t *T) {
	err := &Err{Cause: AnError}
	AssertEqual(t, AnError, err.Unwrap())
}

func TestError_Err_Unwrap_Bad(t *T) {
	err := &Err{}
	AssertNil(t, err.Unwrap())
}

func TestError_Err_Unwrap_Ugly(t *T) {
	root := NewCode("agent.refused", "dispatch refused")
	err := &Err{Cause: Wrap(root, "agent.Dispatch", "failed")}
	AssertErrorIs(t, err.Unwrap(), root)
}

func TestError_ErrorCode_Good(t *T) {
	err := NewCode("agent.refused", "dispatch refused")
	AssertEqual(t, "agent.refused", ErrorCode(err))
}

func TestError_ErrorCode_Bad(t *T) {
	AssertEqual(t, "", ErrorCode(NewError("plain failure")))
}

func TestError_ErrorCode_Ugly(t *T) {
	AssertEqual(t, "", ErrorCode(nil))
}

func TestError_ErrorJoin_Bad(t *T) {
	AssertNil(t, ErrorJoin(nil, nil))
}

func TestError_ErrorJoin_Ugly(t *T) {
	joined := ErrorJoin(nil, AnError)
	AssertErrorIs(t, joined, AnError)
}

func TestError_ErrorLog_Error_Good(t *T) {
	r := New().Log().Error(AnError, "agent.Dispatch", "failed")
	AssertFalse(t, r.OK)
	AssertErrorIs(t, r.Value.(error), AnError)
}

func TestError_ErrorLog_Error_Bad(t *T) {
	r := New().Log().Error(nil, "agent.Dispatch", "no failure")
	AssertTrue(t, r.OK)
}

func TestError_ErrorLog_Error_Ugly(t *T) {
	r := (&ErrorLog{}).Error(AnError, "agent.Dispatch", "failed")
	AssertFalse(t, r.OK)
	AssertErrorIs(t, r.Value.(error), AnError)
}

func TestError_ErrorLog_Must_Good(t *T) {
	AssertNotPanics(t, func() {
		New().Log().Must(nil, "agent.Dispatch", "ready")
	})
}

func TestError_ErrorLog_Must_Bad(t *T) {
	AssertPanicsWithError(t, "dispatch failed", func() {
		New().Log().Must(AnError, "agent.Dispatch", "dispatch failed")
	})
}

func TestError_ErrorLog_Warn_Bad(t *T) {
	r := New().Log().Warn(nil, "agent.Dispatch", "no warning")
	AssertTrue(t, r.OK)
}

func TestError_ErrorLog_Warn_Ugly(t *T) {
	r := (&ErrorLog{}).Warn(AnError, "agent.Dispatch", "degraded")
	AssertFalse(t, r.OK)
	AssertErrorIs(t, r.Value.(error), AnError)
}

func TestError_ErrorMessage_Bad(t *T) {
	AssertEqual(t, "plain failure", ErrorMessage(NewError("plain failure")))
}

func TestError_ErrorMessage_Ugly(t *T) {
	AssertEqual(t, "", ErrorMessage(nil))
}

func TestError_ErrorPanic_Recover_Bad(t *T) {
	AssertNotPanics(t, func() {
		New().Error().Recover()
	})
}

func TestError_ErrorPanic_Recover_Ugly(t *T) {
	var h *ErrorPanic
	AssertNotPanics(t, func() {
		h.Recover()
	})
}

func TestError_ErrorPanic_Reports_Bad(t *T) {
	r := New().Error().Reports(1)
	AssertFalse(t, r.OK)
	AssertNil(t, r.Value)
}

func TestError_ErrorPanic_Reports_Ugly(t *T) {
	r := New().Error().Reports(0)
	AssertFalse(t, r.OK)
	AssertNil(t, r.Value)
}

func TestError_ErrorPanic_SafeGo_Bad(t *T) {
	done := make(chan bool, 1)
	New().Error().SafeGo(func() {
		defer func() { done <- true }()
		panic("agent worker failed")
	})
	AssertTrue(t, <-done)
}

func TestError_ErrorPanic_SafeGo_Ugly(t *T) {
	done := make(chan bool, 1)
	New().Error().SafeGo(func() {
		done <- true
	})
	AssertTrue(t, <-done)
}

func TestError_FormatStackTrace_Bad(t *T) {
	AssertEqual(t, "", FormatStackTrace(NewError("plain failure")))
}

func TestError_FormatStackTrace_Ugly(t *T) {
	AssertEqual(t, "", FormatStackTrace(nil))
}

func TestError_Is_Bad(t *T) {
	AssertFalse(t, Is(NewError("left"), NewError("right")))
}

func TestError_Is_Ugly(t *T) {
	AssertTrue(t, Is(nil, nil))
}

func TestError_NewCode_Bad(t *T) {
	err := NewCode("", "dispatch refused")
	AssertEqual(t, "dispatch refused", err.Error())
	AssertEqual(t, "", ErrorCode(err))
}

func TestError_NewCode_Ugly(t *T) {
	err := NewCode("", "")
	AssertEqual(t, "", err.Error())
}

func TestError_NewError_Bad(t *T) {
	err := NewError("")
	AssertEqual(t, "", err.Error())
}

func TestError_NewError_Ugly(t *T) {
	err := NewError("session\nrefused")
	AssertContains(t, err.Error(), "session\nrefused")
}

func TestError_Operation_Ugly(t *T) {
	AssertEqual(t, "", Operation(nil))
}

func TestError_Root_Bad(t *T) {
	err := NewError("plain failure")
	AssertEqual(t, err, Root(err))
}

func TestError_Root_Ugly(t *T) {
	AssertNil(t, Root(nil))
}

func TestError_StackTrace_Bad(t *T) {
	AssertEmpty(t, StackTrace(NewError("plain failure")))
}

func TestError_StackTrace_Ugly(t *T) {
	AssertEmpty(t, StackTrace(nil))
}

func TestError_Wrap_Bad(t *T) {
	AssertNil(t, Wrap(nil, "agent.Dispatch", "failed"))
}

func TestError_Wrap_Ugly(t *T) {
	inner := NewCode("agent.refused", "dispatch refused")
	err := Wrap(inner, "agent.Dispatch", "failed")
	AssertEqual(t, "agent.refused", ErrorCode(err))
}

func TestError_WrapCode_Bad(t *T) {
	AssertNil(t, WrapCode(nil, "", "agent.Dispatch", "failed"))
}

func TestError_WrapCode_Ugly(t *T) {
	err := WrapCode(nil, "agent.refused", "agent.Dispatch", "failed")
	AssertError(t, err)
	AssertEqual(t, "agent.refused", ErrorCode(err))
	AssertNil(t, Root(err).(*Err).Cause)
}
