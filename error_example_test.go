package core_test

import (
	. "dappco.re/go"
)

func ExampleErrorSink() {
	var sink ErrorSink = NewLog(LogOptions{Output: NewBuffer()})
	sink.Warn("example")
}

func ExampleErr_Error() {
	err := &Err{Operation: "cache.Get", Message: "missing", Code: "NOT_FOUND"}
	Println(err.Error())
	// Output: cache.Get: missing [NOT_FOUND]
}

func ExampleErr_Unwrap() {
	cause := NewError("root")
	err := &Err{Operation: "cache.Get", Message: "missing", Cause: cause}
	Println(err.Unwrap() == cause)
	// Output: true
}

func ExampleE() {
	err := E("cache.Get", "key not found", nil)
	Println(Operation(err))
	Println(ErrorMessage(err))
	// Output:
	// cache.Get
	// key not found
}

func ExampleWrap() {
	cause := NewError("connection refused")
	err := Wrap(cause, "database.Connect", "failed to reach host")
	Println(Operation(err))
	Println(Is(err, cause))
	// Output:
	// database.Connect
	// true
}

func ExampleWrapCode() {
	err := WrapCode(NewError("bad input"), "VALIDATION", "user.Save", "invalid user")
	Println(ErrorCode(err))
	Println(Operation(err))
	// Output:
	// VALIDATION
	// user.Save
}

func ExampleNewCode() {
	err := NewCode("NOT_FOUND", "resource missing")
	Println(ErrorCode(err))
	Println(ErrorMessage(err))
	// Output:
	// NOT_FOUND
	// resource missing
}

func ExampleIs() {
	cause := NewError("root")
	err := Wrap(cause, "worker.Run", "failed")
	Println(Is(err, cause))
	// Output: true
}

func ExampleAs() {
	err := E("worker.Run", "failed", nil)
	var target *Err
	Println(As(err, &target))
	Println(target.Operation)
	// Output:
	// true
	// worker.Run
}

func ExampleNewError() {
	err := NewError("plain error")
	Println(err.Error())
	// Output: plain error
}

func ExampleErrorJoin() {
	err := ErrorJoin(NewError("first"), NewError("second"))
	Println(Contains(err.Error(), "first"))
	Println(Contains(err.Error(), "second"))
	// Output:
	// true
	// true
}

func ExampleOperation() {
	err := E("cache.Get", "missing", nil)
	Println(Operation(err))
	// Output: cache.Get
}

func ExampleErrorCode() {
	err := NewCode("NOT_FOUND", "missing")
	Println(ErrorCode(err))
	// Output: NOT_FOUND
}

func ExampleErrorMessage() {
	err := E("cache.Get", "missing", nil)
	Println(ErrorMessage(err))
	// Output: missing
}

func ExampleRoot() {
	cause := NewError("original")
	wrapped := Wrap(cause, "op1", "first wrap")
	double := Wrap(wrapped, "op2", "second wrap")
	Println(Root(double))
	// Output: original
}

func ExampleAllOperations() {
	err := Wrap(Wrap(NewError("root"), "db.Query", "failed"), "api.Get", "failed")
	var ops []string
	for op := range AllOperations(err) {
		ops = append(ops, op)
	}
	Println(ops)
	// Output: [api.Get db.Query]
}

func ExampleStackTrace() {
	err := Wrap(Wrap(NewError("root"), "db.Query", "failed"), "api.Get", "failed")
	Println(StackTrace(err))
	// Output: [api.Get db.Query]
}

func ExampleFormatStackTrace() {
	err := Wrap(Wrap(NewError("root"), "db.Query", "failed"), "api.Get", "failed")
	Println(FormatStackTrace(err))
	// Output: api.Get -> db.Query
}

func ExampleErrorLog_Error() {
	var log ErrorLog
	Println(log.Error(nil, "example", "ok").OK)
	// Output: true
}

func ExampleErrorLog_Warn() {
	var log ErrorLog
	Println(log.Warn(nil, "example", "ok").OK)
	// Output: true
}

func ExampleErrorLog_Must() {
	var log ErrorLog
	log.Must(nil, "example", "ok")
	Println("ok")
	// Output: ok
}

func ExampleCrashReport() {
	report := CrashReport{Error: "panic: boom", System: CrashSystem{OperatingSystem: "darwin"}}
	Println(report.Error)
	Println(report.System.OperatingSystem)
	// Output:
	// panic: boom
	// darwin
}

func ExampleCrashSystem() {
	system := CrashSystem{OperatingSystem: "darwin", Architecture: "arm64", Version: "go1.26"}
	Println(system.OperatingSystem)
	Println(system.Architecture)
	// Output:
	// darwin
	// arm64
}

func ExampleErrorPanic_Recover() {
	h := &ErrorPanic{}
	defer h.Recover()
}

func ExampleErrorPanic_SafeGo() {
	h := &ErrorPanic{}
	h.SafeGo(func() {})
}

func ExampleErrorPanic_Reports() {
	h := &ErrorPanic{}
	Println(h.Reports(1).OK)
	// Output: false
}
