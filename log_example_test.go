package core_test

import . "dappco.re/go"

// ExampleLevel_String renders `Level.String` as a stable string for operator logging.
// Loggers support levels, redaction, and default routing for operator output.
func ExampleLevel_String() {
	Println(LevelInfo.String())
	Println(LevelQuiet.String())
	// Output:
	// info
	// quiet
}

// ExampleNewLog creates a logger through `NewLog` for operator logging. Loggers support
// levels, redaction, and default routing for operator output.
func ExampleNewLog() {
	buf := NewBuffer()
	log := NewLog(LogOptions{Level: LevelInfo, Output: buf})
	log.StyleTimestamp = func(string) string { return "00:00:00" }
	log.Info("server started", "port", 8080)
	Println(Contains(buf.String(), "00:00:00 [INF] server started port=8080"))
	// Output: true
}

// ExampleLog_SetLevel changes log level through `Log.SetLevel` for operator logging.
// Loggers support levels, redaction, and default routing for operator output.
func ExampleLog_SetLevel() {
	log := NewLog(LogOptions{Level: LevelWarn, Output: NewBuffer()})
	log.SetLevel(LevelDebug)
	Println(log.Level())
	// Output: debug
}

// ExampleLog_Level reads a logger level through `Log.Level` for operator logging. Loggers
// support levels, redaction, and default routing for operator output.
func ExampleLog_Level() {
	log := NewLog(LogOptions{Level: LevelError, Output: NewBuffer()})
	Println(log.Level())
	// Output: error
}

// ExampleLog_SetOutput redirects output through `Log.SetOutput` for operator logging.
// Loggers support levels, redaction, and default routing for operator output.
func ExampleLog_SetOutput() {
	buf := NewBuffer()
	log := NewLog(LogOptions{Level: LevelInfo})
	log.StyleTimestamp = func(string) string { return "00:00:00" }
	log.SetOutput(buf)
	log.Info("ready")
	Println(Contains(buf.String(), "[INF] ready"))
	// Output: true
}

// ExampleLog_SetRedactKeys configures redaction keys through `Log.SetRedactKeys` for
// operator logging. Loggers support levels, redaction, and default routing for operator
// output.
func ExampleLog_SetRedactKeys() {
	buf := NewBuffer()
	log := NewLog(LogOptions{Level: LevelInfo, Output: buf})
	log.StyleTimestamp = func(string) string { return "00:00:00" }
	log.SetRedactKeys("token")
	log.Info("auth", "token", "secret")
	Println(Contains(buf.String(), `token="[REDACTED]"`))
	// Output: true
}

// ExampleLog_Debug writes a debug event through `Log.Debug` for operator logging. Loggers
// support levels, redaction, and default routing for operator output.
func ExampleLog_Debug() {
	buf := NewBuffer()
	log := NewLog(LogOptions{Level: LevelDebug, Output: buf})
	log.StyleTimestamp = func(string) string { return "00:00:00" }
	log.Debug("trace")
	Println(Contains(buf.String(), "[DBG] trace"))
	// Output: true
}

// ExampleLog_Info writes an info event through `Log.Info` for operator logging. Loggers
// support levels, redaction, and default routing for operator output.
func ExampleLog_Info() {
	buf := NewBuffer()
	log := NewLog(LogOptions{Level: LevelInfo, Output: buf})
	log.StyleTimestamp = func(string) string { return "00:00:00" }
	log.Info("ready")
	Println(Contains(buf.String(), "[INF] ready"))
	// Output: true
}

// ExampleLog_Warn writes a warning event through `Log.Warn` for operator logging. Loggers
// support levels, redaction, and default routing for operator output.
func ExampleLog_Warn() {
	buf := NewBuffer()
	log := NewLog(LogOptions{Level: LevelWarn, Output: buf})
	log.StyleTimestamp = func(string) string { return "00:00:00" }
	log.Warn("slow")
	Println(Contains(buf.String(), "[WRN] slow"))
	// Output: true
}

// ExampleLog_Error writes or renders an error through `Log.Error` for operator logging.
// Loggers support levels, redaction, and default routing for operator output.
func ExampleLog_Error() {
	buf := NewBuffer()
	log := NewLog(LogOptions{Level: LevelError, Output: buf})
	log.StyleTimestamp = func(string) string { return "00:00:00" }
	log.Error("failed")
	Println(Contains(buf.String(), "[ERR] failed"))
	// Output: true
}

// ExampleLog_Security writes a security event through `Log.Security` for operator logging.
// Loggers support levels, redaction, and default routing for operator output.
func ExampleLog_Security() {
	buf := NewBuffer()
	log := NewLog(LogOptions{Level: LevelError, Output: buf})
	log.StyleTimestamp = func(string) string { return "00:00:00" }
	log.Security("denied")
	Println(Contains(buf.String(), "[SEC] denied"))
	// Output: true
}

// ExampleUsername reads the current username through `Username` for operator logging.
// Loggers support levels, redaction, and default routing for operator output.
func ExampleUsername() {
	Println(Username() != "")
	// Output: true
}

// ExampleDefault reads the default logger through `Default` for operator logging. Loggers
// support levels, redaction, and default routing for operator output.
func ExampleDefault() {
	Println(Default() != nil)
	// Output: true
}

// ExampleSetDefault replaces the default logger through `SetDefault` for operator logging.
// Loggers support levels, redaction, and default routing for operator output.
func ExampleSetDefault() {
	old := Default()
	defer SetDefault(old)

	log := NewLog(LogOptions{Level: LevelQuiet, Output: NewBuffer()})
	SetDefault(log)
	Println(Default() == log)
	// Output: true
}

// ExampleSetLevel changes log level through `SetLevel` for operator logging. Loggers
// support levels, redaction, and default routing for operator output.
func ExampleSetLevel() {
	old := Default()
	defer SetDefault(old)

	SetDefault(NewLog(LogOptions{Level: LevelQuiet, Output: NewBuffer()}))
	SetLevel(LevelDebug)
	Println(Default().Level())
	// Output: debug
}

// ExampleSetRedactKeys configures redaction keys through `SetRedactKeys` for operator
// logging. Loggers support levels, redaction, and default routing for operator output.
func ExampleSetRedactKeys() {
	old := Default()
	defer SetDefault(old)

	buf := NewBuffer()
	log := NewLog(LogOptions{Level: LevelInfo, Output: buf})
	log.StyleTimestamp = func(string) string { return "00:00:00" }
	SetDefault(log)
	SetRedactKeys("token")
	Info("auth", "token", "secret")
	Println(Contains(buf.String(), `token="[REDACTED]"`))
	// Output: true
}

// ExampleDebug writes a debug event through `Debug` for operator logging. Loggers support
// levels, redaction, and default routing for operator output.
func ExampleDebug() {
	old := Default()
	defer SetDefault(old)

	buf := NewBuffer()
	log := NewLog(LogOptions{Level: LevelDebug, Output: buf})
	log.StyleTimestamp = func(string) string { return "00:00:00" }
	SetDefault(log)
	Debug("trace")
	Println(Contains(buf.String(), "[DBG] trace"))
	// Output: true
}

// ExampleInfo writes an info event through `Info` for operator logging. Loggers support
// levels, redaction, and default routing for operator output.
func ExampleInfo() {
	old := Default()
	defer SetDefault(old)

	buf := NewBuffer()
	log := NewLog(LogOptions{Level: LevelInfo, Output: buf})
	log.StyleTimestamp = func(string) string { return "00:00:00" }
	SetDefault(log)
	Info("server started", "port", 8080)
	Println(Contains(buf.String(), "[INF] server started port=8080"))
	// Output: true
}

// ExampleWarn writes a warning event through `Warn` for operator logging. Loggers support
// levels, redaction, and default routing for operator output.
func ExampleWarn() {
	old := Default()
	defer SetDefault(old)

	buf := NewBuffer()
	log := NewLog(LogOptions{Level: LevelWarn, Output: buf})
	log.StyleTimestamp = func(string) string { return "00:00:00" }
	SetDefault(log)
	Warn("deprecated", "feature", "old-api")
	Println(Contains(buf.String(), `[WRN] deprecated feature="old-api"`))
	// Output: true
}

// ExampleError writes or renders an error through `Error` for operator logging. Loggers
// support levels, redaction, and default routing for operator output.
func ExampleError() {
	old := Default()
	defer SetDefault(old)

	buf := NewBuffer()
	log := NewLog(LogOptions{Level: LevelError, Output: buf})
	log.StyleTimestamp = func(string) string { return "00:00:00" }
	SetDefault(log)
	Error("failed", "op", "deploy")
	Println(Contains(buf.String(), `[ERR] failed op="deploy"`))
	// Output: true
}

// ExampleSecurity writes a security event through `Security` for operator logging. Loggers
// support levels, redaction, and default routing for operator output.
func ExampleSecurity() {
	old := Default()
	defer SetDefault(old)

	buf := NewBuffer()
	log := NewLog(LogOptions{Level: LevelError, Output: buf})
	log.StyleTimestamp = func(string) string { return "00:00:00" }
	SetDefault(log)
	Security("access denied", "user", "unknown", "action", "admin.nuke")
	Println(Contains(buf.String(), `[SEC] access denied user="unknown" action="admin.nuke"`))
	// Output: true
}

// ExampleNewLogErr creates an error logging sink through `NewLogErr` for operator logging.
// Loggers support levels, redaction, and default routing for operator output.
func ExampleNewLogErr() {
	logErr := NewLogErr(NewLog(LogOptions{Output: NewBuffer()}))
	Println(logErr != nil)
	// Output: true
}

// ExampleLogErr_Log logs through `LogErr.Log` for operator logging. Loggers support
// levels, redaction, and default routing for operator output.
func ExampleLogErr_Log() {
	logErr := NewLogErr(NewLog(LogOptions{Output: NewBuffer()}))
	logErr.Log(nil)
	Println("ok")
	// Output: ok
}

// ExampleNewLogPanic creates a panic logging helper through `NewLogPanic` for operator
// logging. Loggers support levels, redaction, and default routing for operator output.
func ExampleNewLogPanic() {
	panicLog := NewLogPanic(NewLog(LogOptions{Output: NewBuffer()}))
	Println(panicLog != nil)
	// Output: true
}

// ExampleLogPanic_Recover documents panic logging recovery when no panic is active.
// Loggers support levels, redaction, and default routing for operator output.
func ExampleLogPanic_Recover() {
	panicLog := NewLogPanic(NewLog(LogOptions{Output: NewBuffer()}))
	defer panicLog.Recover()
}

// ExampleRotationLogOptions declares rotation settings through `RotationLogOptions` for
// operator logging. Loggers support levels, redaction, and default routing for operator
// output.
func ExampleRotationLogOptions() {
	opts := RotationLogOptions{Filename: "app.log", MaxSize: 10, MaxBackups: 3}
	Println(opts.Filename)
	Println(opts.MaxSize)
	// Output:
	// app.log
	// 10
}

// ExampleLogOptions declares logger settings through `LogOptions` for operator logging.
// Loggers support levels, redaction, and default routing for operator output.
func ExampleLogOptions() {
	opts := LogOptions{Level: LevelInfo, Output: NewBuffer(), RedactKeys: []string{"token"}}
	Println(opts.Level)
	Println(opts.RedactKeys)
	// Output:
	// info
	// [token]
}

// ExampleRotationWriterFactory documents the log rotation extension point without creating
// a writer. Loggers support levels, redaction, and default routing for operator output.
func ExampleRotationWriterFactory() {
	_ = RotationWriterFactory
}
