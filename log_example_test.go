package core_test

import . "dappco.re/go"

func ExampleLevel_String() {
	Println(LevelInfo.String())
	Println(LevelQuiet.String())
	// Output:
	// info
	// quiet
}

func ExampleNewLog() {
	buf := NewBuffer()
	log := NewLog(LogOptions{Level: LevelInfo, Output: buf})
	log.StyleTimestamp = func(string) string { return "00:00:00" }
	log.Info("server started", "port", 8080)
	Println(Contains(buf.String(), "00:00:00 [INF] server started port=8080"))
	// Output: true
}

func ExampleLog_SetLevel() {
	log := NewLog(LogOptions{Level: LevelWarn, Output: NewBuffer()})
	log.SetLevel(LevelDebug)
	Println(log.Level())
	// Output: debug
}

func ExampleLog_Level() {
	log := NewLog(LogOptions{Level: LevelError, Output: NewBuffer()})
	Println(log.Level())
	// Output: error
}

func ExampleLog_SetOutput() {
	buf := NewBuffer()
	log := NewLog(LogOptions{Level: LevelInfo})
	log.StyleTimestamp = func(string) string { return "00:00:00" }
	log.SetOutput(buf)
	log.Info("ready")
	Println(Contains(buf.String(), "[INF] ready"))
	// Output: true
}

func ExampleLog_SetRedactKeys() {
	buf := NewBuffer()
	log := NewLog(LogOptions{Level: LevelInfo, Output: buf})
	log.StyleTimestamp = func(string) string { return "00:00:00" }
	log.SetRedactKeys("token")
	log.Info("auth", "token", "secret")
	Println(Contains(buf.String(), `token="[REDACTED]"`))
	// Output: true
}

func ExampleLog_Debug() {
	buf := NewBuffer()
	log := NewLog(LogOptions{Level: LevelDebug, Output: buf})
	log.StyleTimestamp = func(string) string { return "00:00:00" }
	log.Debug("trace")
	Println(Contains(buf.String(), "[DBG] trace"))
	// Output: true
}

func ExampleLog_Info() {
	buf := NewBuffer()
	log := NewLog(LogOptions{Level: LevelInfo, Output: buf})
	log.StyleTimestamp = func(string) string { return "00:00:00" }
	log.Info("ready")
	Println(Contains(buf.String(), "[INF] ready"))
	// Output: true
}

func ExampleLog_Warn() {
	buf := NewBuffer()
	log := NewLog(LogOptions{Level: LevelWarn, Output: buf})
	log.StyleTimestamp = func(string) string { return "00:00:00" }
	log.Warn("slow")
	Println(Contains(buf.String(), "[WRN] slow"))
	// Output: true
}

func ExampleLog_Error() {
	buf := NewBuffer()
	log := NewLog(LogOptions{Level: LevelError, Output: buf})
	log.StyleTimestamp = func(string) string { return "00:00:00" }
	log.Error("failed")
	Println(Contains(buf.String(), "[ERR] failed"))
	// Output: true
}

func ExampleLog_Security() {
	buf := NewBuffer()
	log := NewLog(LogOptions{Level: LevelError, Output: buf})
	log.StyleTimestamp = func(string) string { return "00:00:00" }
	log.Security("denied")
	Println(Contains(buf.String(), "[SEC] denied"))
	// Output: true
}

func ExampleUsername() {
	Println(Username() != "")
	// Output: true
}

func ExampleDefault() {
	Println(Default() != nil)
	// Output: true
}

func ExampleSetDefault() {
	old := Default()
	defer SetDefault(old)

	log := NewLog(LogOptions{Level: LevelQuiet, Output: NewBuffer()})
	SetDefault(log)
	Println(Default() == log)
	// Output: true
}

func ExampleSetLevel() {
	old := Default()
	defer SetDefault(old)

	SetDefault(NewLog(LogOptions{Level: LevelQuiet, Output: NewBuffer()}))
	SetLevel(LevelDebug)
	Println(Default().Level())
	// Output: debug
}

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

func ExampleNewLogErr() {
	logErr := NewLogErr(NewLog(LogOptions{Output: NewBuffer()}))
	Println(logErr != nil)
	// Output: true
}

func ExampleLogErr_Log() {
	logErr := NewLogErr(NewLog(LogOptions{Output: NewBuffer()}))
	logErr.Log(nil)
	Println("ok")
	// Output: ok
}

func ExampleNewLogPanic() {
	panicLog := NewLogPanic(NewLog(LogOptions{Output: NewBuffer()}))
	Println(panicLog != nil)
	// Output: true
}

func ExampleLogPanic_Recover() {
	panicLog := NewLogPanic(NewLog(LogOptions{Output: NewBuffer()}))
	defer panicLog.Recover()
}

func ExampleRotationLogOptions() {
	opts := RotationLogOptions{Filename: "app.log", MaxSize: 10, MaxBackups: 3}
	Println(opts.Filename)
	Println(opts.MaxSize)
	// Output:
	// app.log
	// 10
}

func ExampleLogOptions() {
	opts := LogOptions{Level: LevelInfo, Output: NewBuffer(), RedactKeys: []string{"token"}}
	Println(opts.Level)
	Println(opts.RedactKeys)
	// Output:
	// info
	// [token]
}

func ExampleRotationWriterFactory() {
	_ = RotationWriterFactory
}
