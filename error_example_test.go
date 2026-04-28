package core_test

import (
	. "dappco.re/go"
)

// ExampleErrorSink declares an error sink through `ErrorSink` for dAppCore error handling.
// Operations, codes, roots, and crash reports remain inspectable through core errors.
func ExampleErrorSink() {
	var sink ErrorSink = NewLog(LogOptions{Output: NewBuffer()})
	sink.Warn("example")
}

// ExampleErr_Error writes or renders an error through `Err.Error` for dAppCore error
// handling. Operations, codes, roots, and crash reports remain inspectable through core
// errors.
func ExampleErr_Error() {
	err := &Err{Operation: "cache.Get", Message: "missing", Code: "NOT_FOUND"}
	Println(err.Error())
	// Output: cache.Get: missing [NOT_FOUND]
}

// ExampleErr_Unwrap unwraps an error through `Err.Unwrap` for dAppCore error handling.
// Operations, codes, roots, and crash reports remain inspectable through core errors.
func ExampleErr_Unwrap() {
	cause := NewError("root")
	err := &Err{Operation: "cache.Get", Message: "missing", Cause: cause}
	Println(err.Unwrap() == cause)
	// Output: true
}

// ExampleE builds a scoped error through `E` for dAppCore error handling. Operations,
// codes, roots, and crash reports remain inspectable through core errors.
func ExampleE() {
	err := E("cache.Get", "key not found", nil)
	Println(Operation(err))
	Println(ErrorMessage(err))
	// Output:
	// cache.Get
	// key not found
}

// ExampleWrap wraps an error through `Wrap` for dAppCore error handling. Operations,
// codes, roots, and crash reports remain inspectable through core errors.
func ExampleWrap() {
	cause := NewError("connection refused")
	err := Wrap(cause, "database.Connect", "failed to reach host")
	Println(Operation(err))
	Println(Is(err, cause))
	// Output:
	// database.Connect
	// true
}

// ExampleWrapCode wraps an error with a code through `WrapCode` for dAppCore error
// handling. Operations, codes, roots, and crash reports remain inspectable through core
// errors.
func ExampleWrapCode() {
	err := WrapCode(NewError("bad input"), "VALIDATION", "user.Save", "invalid user")
	Println(ErrorCode(err))
	Println(Operation(err))
	// Output:
	// VALIDATION
	// user.Save
}

// ExampleNewCode creates a coded error through `NewCode` for dAppCore error handling.
// Operations, codes, roots, and crash reports remain inspectable through core errors.
func ExampleNewCode() {
	err := NewCode("NOT_FOUND", "resource missing")
	Println(ErrorCode(err))
	Println(ErrorMessage(err))
	// Output:
	// NOT_FOUND
	// resource missing
}

// ExampleIs matches an error through `Is` for dAppCore error handling. Operations, codes,
// roots, and crash reports remain inspectable through core errors.
func ExampleIs() {
	cause := NewError("root")
	err := Wrap(cause, "worker.Run", "failed")
	Println(Is(err, cause))
	// Output: true
}

// ExampleAs extracts a typed error through `As` for dAppCore error handling. Operations,
// codes, roots, and crash reports remain inspectable through core errors.
func ExampleAs() {
	err := E("worker.Run", "failed", nil)
	var target *Err
	Println(As(err, &target))
	Println(target.Operation)
	// Output:
	// true
	// worker.Run
}

// ExampleNewError creates a simple error through `NewError` for dAppCore error handling.
// Operations, codes, roots, and crash reports remain inspectable through core errors.
func ExampleNewError() {
	err := NewError("plain error")
	Println(err.Error())
	// Output: plain error
}

// ExampleErrorJoin joins errors through `ErrorJoin` for dAppCore error handling.
// Operations, codes, roots, and crash reports remain inspectable through core errors.
func ExampleErrorJoin() {
	err := ErrorJoin(NewError("first"), NewError("second"))
	Println(Contains(err.Error(), "first"))
	Println(Contains(err.Error(), "second"))
	// Output:
	// true
	// true
}

// ExampleOperation reads the operation from an error through `Operation` for dAppCore
// error handling. Operations, codes, roots, and crash reports remain inspectable through
// core errors.
func ExampleOperation() {
	err := E("cache.Get", "missing", nil)
	Println(Operation(err))
	// Output: cache.Get
}

// ExampleErrorCode reads the code from an error through `ErrorCode` for dAppCore error
// handling. Operations, codes, roots, and crash reports remain inspectable through core
// errors.
func ExampleErrorCode() {
	err := NewCode("NOT_FOUND", "missing")
	Println(ErrorCode(err))
	// Output: NOT_FOUND
}

// ExampleErrorMessage reads the message from an error through `ErrorMessage` for dAppCore
// error handling. Operations, codes, roots, and crash reports remain inspectable through
// core errors.
func ExampleErrorMessage() {
	err := E("cache.Get", "missing", nil)
	Println(ErrorMessage(err))
	// Output: missing
}

// ExampleRoot reports the root value through `Root` for dAppCore error handling.
// Operations, codes, roots, and crash reports remain inspectable through core errors.
func ExampleRoot() {
	cause := NewError("original")
	wrapped := Wrap(cause, "op1", "first wrap")
	double := Wrap(wrapped, "op2", "second wrap")
	Println(Root(double))
	// Output: original
}

// ExampleAllOperations lists wrapped operations through `AllOperations` for dAppCore error
// handling. Operations, codes, roots, and crash reports remain inspectable through core
// errors.
func ExampleAllOperations() {
	err := Wrap(Wrap(NewError("root"), "db.Query", "failed"), "api.Get", "failed")
	var ops []string
	for op := range AllOperations(err) {
		ops = append(ops, op)
	}
	Println(ops)
	// Output: [api.Get db.Query]
}

// ExampleStackTrace captures a stack trace through `StackTrace` for dAppCore error
// handling. Operations, codes, roots, and crash reports remain inspectable through core
// errors.
func ExampleStackTrace() {
	err := Wrap(Wrap(NewError("root"), "db.Query", "failed"), "api.Get", "failed")
	Println(StackTrace(err))
	// Output: [api.Get db.Query]
}

// ExampleFormatStackTrace formats a stack trace through `FormatStackTrace` for dAppCore
// error handling. Operations, codes, roots, and crash reports remain inspectable through
// core errors.
func ExampleFormatStackTrace() {
	err := Wrap(Wrap(NewError("root"), "db.Query", "failed"), "api.Get", "failed")
	Println(FormatStackTrace(err))
	// Output: api.Get -> db.Query
}

// ExampleErrorLog_Error writes or renders an error through `ErrorLog.Error` for dAppCore
// error handling. Operations, codes, roots, and crash reports remain inspectable through
// core errors.
func ExampleErrorLog_Error() {
	var log ErrorLog
	Println(log.Error(nil, "example", "ok").OK)
	// Output: true
}

// ExampleErrorLog_Warn writes a warning event through `ErrorLog.Warn` for dAppCore error
// handling. Operations, codes, roots, and crash reports remain inspectable through core
// errors.
func ExampleErrorLog_Warn() {
	var log ErrorLog
	Println(log.Warn(nil, "example", "ok").OK)
	// Output: true
}

// ExampleErrorLog_Must unwraps a successful Result through `ErrorLog.Must` for dAppCore
// error handling. Operations, codes, roots, and crash reports remain inspectable through
// core errors.
func ExampleErrorLog_Must() {
	var log ErrorLog
	log.Must(nil, "example", "ok")
	Println("ok")
	// Output: ok
}

// ExampleCrashReport builds a crash report through `CrashReport` for dAppCore error
// handling. Operations, codes, roots, and crash reports remain inspectable through core
// errors.
func ExampleCrashReport() {
	report := CrashReport{Error: "panic: boom", System: CrashSystem{OperatingSystem: "darwin"}}
	Println(report.Error)
	Println(report.System.OperatingSystem)
	// Output:
	// panic: boom
	// darwin
}

// ExampleCrashSystem collects crash reports through `CrashSystem` for dAppCore error
// handling. Operations, codes, roots, and crash reports remain inspectable through core
// errors.
func ExampleCrashSystem() {
	system := CrashSystem{OperatingSystem: "darwin", Architecture: "arm64", Version: "go1.26"}
	Println(system.OperatingSystem)
	Println(system.Architecture)
	// Output:
	// darwin
	// arm64
}

// ExampleErrorPanic_Recover documents panic recovery when no panic is active. Operations,
// codes, roots, and crash reports remain inspectable through core errors.
func ExampleErrorPanic_Recover() {
	h := &ErrorPanic{}
	defer h.Recover()
}

// ExampleErrorPanic_SafeGo launches protected work that reports panics instead of crashing
// the caller. Operations, codes, roots, and crash reports remain inspectable through core
// errors.
func ExampleErrorPanic_SafeGo() {
	h := &ErrorPanic{}
	h.SafeGo(func() {})
}

// ExampleErrorPanic_Reports lists captured reports through `ErrorPanic.Reports` for
// dAppCore error handling. Operations, codes, roots, and crash reports remain inspectable
// through core errors.
func ExampleErrorPanic_Reports() {
	h := &ErrorPanic{}
	Println(h.Reports(1).OK)
	// Output: false
}
