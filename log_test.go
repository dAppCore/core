package core_test

import (
	. "dappco.re/go"
)

type logTestWriteCloser struct {
	w Writer
}

func (w logTestWriteCloser) Write(p []byte) (int, error) {
	return w.w.Write(p)
}

func (w logTestWriteCloser) Close() error {
	return nil
}

// --- Log ---

func TestLog_New_Good(t *T) {
	l := NewLog(LogOptions{Level: LevelInfo})
	AssertNotNil(t, l)
}

func TestLog_AllLevels_Good(t *T) {
	l := NewLog(LogOptions{Level: LevelDebug})
	l.Debug("debug")
	l.Info("info")
	l.Warn("warn")
	l.Error("error")
	l.Security("security event")
}

func TestLog_LevelFiltering_Good(t *T) {
	// At Error level, Debug/Info/Warn should be suppressed (no panic)
	l := NewLog(LogOptions{Level: LevelError})
	l.Debug("suppressed")
	l.Info("suppressed")
	l.Warn("suppressed")
	l.Error("visible")
}

func TestLog_SetLevel_Good(t *T) {
	l := NewLog(LogOptions{Level: LevelInfo})
	l.SetLevel(LevelDebug)
	AssertEqual(t, LevelDebug, l.Level())
}

func TestLog_SetRedactKeys_Good(t *T) {
	l := NewLog(LogOptions{Level: LevelInfo})
	l.SetRedactKeys("password", "token")
	// Redacted keys should mask values in output
	l.Info("login", "password", "secret123", "user", "admin")
}

func TestLog_LevelString_Good(t *T) {
	AssertEqual(t, "debug", LevelDebug.String())
	AssertEqual(t, "info", LevelInfo.String())
	AssertEqual(t, "warn", LevelWarn.String())
	AssertEqual(t, "error", LevelError.String())
}

func TestLog_CoreLog_Good(t *T) {
	c := New()
	AssertNotNil(t, c.Log())
}

func TestLog_ErrorSink_Good(t *T) {
	l := NewLog(LogOptions{Level: LevelInfo})
	var sink ErrorSink = l
	sink.Error("test")
	sink.Warn("test")
}

// --- Default Logger ---

func TestLog_Default_Good(t *T) {
	d := Default()
	AssertNotNil(t, d)
}

func TestLog_SetDefault_Good(t *T) {
	original := Default()
	defer SetDefault(original)

	custom := NewLog(LogOptions{Level: LevelDebug})
	SetDefault(custom)
	AssertEqual(t, custom, Default())
}

func TestLog_PackageLevelFunctions_Good(t *T) {
	// Package-level log functions use the default logger
	Debug("debug msg")
	Info("info msg")
	Warn("warn msg")
	Error("error msg")
	Security("security msg")
}

func TestLog_PackageSetLevel_Good(t *T) {
	original := Default()
	defer SetDefault(original)

	SetLevel(LevelDebug)
	SetRedactKeys("secret")
}

func TestLog_Username_Good(t *T) {
	u := Username()
	AssertNotEmpty(t, u)
}

// --- LogErr ---

func TestLog_LogErr_Good(t *T) {
	l := NewLog(LogOptions{Level: LevelInfo})
	le := NewLogErr(l)
	AssertNotNil(t, le)

	err := E("test.Operation", "something broke", nil)
	le.Log(err)
}

func TestLog_LogErr_Nil_Good(t *T) {
	l := NewLog(LogOptions{Level: LevelInfo})
	le := NewLogErr(l)
	le.Log(nil) // should not panic
}

// --- LogPanic ---

func TestLog_LogPanic_Good(t *T) {
	l := NewLog(LogOptions{Level: LevelInfo})
	lp := NewLogPanic(l)
	AssertNotNil(t, lp)
}

func TestLog_LogPanic_Recover_Good(t *T) {
	l := NewLog(LogOptions{Level: LevelInfo})
	lp := NewLogPanic(l)
	AssertNotPanics(t, func() {
		defer lp.Recover()
		panic("caught")
	})
}

// --- SetOutput ---

func TestLog_SetOutput_Good(t *T) {
	l := NewLog(LogOptions{Level: LevelInfo})
	l.SetOutput(NewBuilder())
	l.Info("redirected")
}

// --- Log suppression by level ---

func TestLog_Quiet_Suppresses_Ugly(t *T) {
	l := NewLog(LogOptions{Level: LevelQuiet})
	// These should not panic even though nothing is logged
	l.Debug("suppressed")
	l.Info("suppressed")
	l.Warn("suppressed")
	l.Error("suppressed")
}

func TestLog_ErrorLevel_Suppresses_Ugly(t *T) {
	l := NewLog(LogOptions{Level: LevelError})
	l.Debug("suppressed") // below threshold
	l.Info("suppressed")  // below threshold
	l.Warn("suppressed")  // below threshold
	l.Error("visible")    // at threshold
}

func TestLog_Debug_Good(t *T) {
	original := Default()
	defer SetDefault(original)
	out := NewBuffer()
	SetDefault(NewLog(LogOptions{Level: LevelDebug, Output: out}))

	Debug("agent trace", "task", "dispatch")

	AssertContains(t, out.String(), "[DBG]")
	AssertContains(t, out.String(), "agent trace")
}

func TestLog_Debug_Bad(t *T) {
	original := Default()
	defer SetDefault(original)
	out := NewBuffer()
	SetDefault(NewLog(LogOptions{Level: LevelInfo, Output: out}))

	Debug("suppressed trace")

	AssertEqual(t, "", out.String())
}

func TestLog_Debug_Ugly(t *T) {
	original := Default()
	defer SetDefault(original)
	out := NewBuffer()
	SetDefault(NewLog(LogOptions{Level: LevelDebug, Output: out}))

	Debug("dangling key", "session")

	AssertContains(t, out.String(), "session=<nil>")
}

func TestLog_Default_Bad(t *T) {
	original := Default()
	defer SetDefault(original)

	SetDefault(nil)

	AssertNil(t, Default())
}

func TestLog_Default_Ugly(t *T) {
	original := Default()
	defer SetDefault(original)
	custom := NewLog(LogOptions{Level: LevelQuiet, Output: NewBuffer()})

	SetDefault(custom)

	AssertSame(t, custom, Default())
}

func TestLog_Error_Good(t *T) {
	original := Default()
	defer SetDefault(original)
	out := NewBuffer()
	SetDefault(NewLog(LogOptions{Level: LevelError, Output: out}))

	Error("agent failed", "err", AnError)

	AssertContains(t, out.String(), "[ERR]")
	AssertContains(t, out.String(), "agent failed")
}

func TestLog_Error_Bad(t *T) {
	original := Default()
	defer SetDefault(original)
	out := NewBuffer()
	SetDefault(NewLog(LogOptions{Level: LevelQuiet, Output: out}))

	Error("suppressed failure")

	AssertEqual(t, "", out.String())
}

func TestLog_Error_Ugly(t *T) {
	original := Default()
	defer SetDefault(original)
	out := NewBuffer()
	SetDefault(NewLog(LogOptions{Level: LevelError, Output: out}))

	Error("odd keyvals", "session")

	AssertContains(t, out.String(), "session=<nil>")
}

func TestLog_Info_Good(t *T) {
	original := Default()
	defer SetDefault(original)
	out := NewBuffer()
	SetDefault(NewLog(LogOptions{Level: LevelInfo, Output: out}))

	Info("agent ready", "service", "homelab")

	AssertContains(t, out.String(), "[INF]")
	AssertContains(t, out.String(), "service=\"homelab\"")
}

func TestLog_Info_Bad(t *T) {
	original := Default()
	defer SetDefault(original)
	out := NewBuffer()
	SetDefault(NewLog(LogOptions{Level: LevelWarn, Output: out}))

	Info("suppressed info")

	AssertEqual(t, "", out.String())
}

func TestLog_Info_Ugly(t *T) {
	original := Default()
	defer SetDefault(original)
	out := NewBuffer()
	SetDefault(NewLog(LogOptions{Level: LevelInfo, Output: out}))

	Info("", "agent", "")

	AssertContains(t, out.String(), "agent=\"\"")
}

func TestLog_Level_String_Good(t *T) {
	AssertEqual(t, "info", LevelInfo.String())
}

func TestLog_Level_String_Bad(t *T) {
	AssertEqual(t, "unknown", Level(99).String())
}

func TestLog_Level_String_Ugly(t *T) {
	AssertEqual(t, "quiet", LevelQuiet.String())
}

func TestLog_Log_Debug_Good(t *T) {
	out := NewBuffer()
	l := NewLog(LogOptions{Level: LevelDebug, Output: out})

	l.Debug("agent trace")

	AssertContains(t, out.String(), "[DBG]")
}

func TestLog_Log_Debug_Bad(t *T) {
	out := NewBuffer()
	l := NewLog(LogOptions{Level: LevelInfo, Output: out})

	l.Debug("suppressed")

	AssertEqual(t, "", out.String())
}

func TestLog_Log_Debug_Ugly(t *T) {
	out := NewBuffer()
	l := NewLog(LogOptions{Level: LevelDebug, Output: out})

	l.Debug("odd", "key")

	AssertContains(t, out.String(), "key=<nil>")
}

func TestLog_Log_Error_Good(t *T) {
	out := NewBuffer()
	l := NewLog(LogOptions{Level: LevelError, Output: out})

	l.Error("agent failed")

	AssertContains(t, out.String(), "[ERR]")
}

func TestLog_Log_Error_Bad(t *T) {
	out := NewBuffer()
	l := NewLog(LogOptions{Level: LevelQuiet, Output: out})

	l.Error("suppressed")

	AssertEqual(t, "", out.String())
}

func TestLog_Log_Error_Ugly(t *T) {
	out := NewBuffer()
	l := NewLog(LogOptions{Level: LevelError, Output: out})

	l.Error("with error", "err", E("agent.Dispatch", "failed", AnError))

	AssertContains(t, out.String(), "stack=\"agent.Dispatch\"")
}

func TestLog_Log_Info_Good(t *T) {
	out := NewBuffer()
	l := NewLog(LogOptions{Level: LevelInfo, Output: out})

	l.Info("agent ready")

	AssertContains(t, out.String(), "[INF]")
}

func TestLog_Log_Info_Bad(t *T) {
	out := NewBuffer()
	l := NewLog(LogOptions{Level: LevelWarn, Output: out})

	l.Info("suppressed")

	AssertEqual(t, "", out.String())
}

func TestLog_Log_Info_Ugly(t *T) {
	out := NewBuffer()
	l := NewLog(LogOptions{Level: LevelInfo, Output: out})

	l.Info("empty value", "agent", "")

	AssertContains(t, out.String(), "agent=\"\"")
}

func TestLog_Log_Level_Good(t *T) {
	l := NewLog(LogOptions{Level: LevelWarn})
	AssertEqual(t, LevelWarn, l.Level())
}

func TestLog_Log_Level_Bad(t *T) {
	l := NewLog(LogOptions{Level: LevelQuiet})
	AssertEqual(t, LevelQuiet, l.Level())
}

func TestLog_Log_Level_Ugly(t *T) {
	l := NewLog(LogOptions{Level: Level(99)})
	AssertEqual(t, Level(99), l.Level())
}

func TestLog_Log_Security_Good(t *T) {
	out := NewBuffer()
	l := NewLog(LogOptions{Level: LevelError, Output: out})

	l.Security("entitlement denied", "action", "process.run")

	AssertContains(t, out.String(), "[SEC]")
	AssertContains(t, out.String(), "process.run")
}

func TestLog_Log_Security_Bad(t *T) {
	out := NewBuffer()
	l := NewLog(LogOptions{Level: LevelQuiet, Output: out})

	l.Security("suppressed security")

	AssertEqual(t, "", out.String())
}

func TestLog_Log_Security_Ugly(t *T) {
	out := NewBuffer()
	l := NewLog(LogOptions{Level: LevelError, Output: out})

	l.Security("newline attempt", "token", "a\nb")

	AssertContains(t, out.String(), "token=\"a\\nb\"")
}

func TestLog_Log_SetLevel_Good(t *T) {
	l := NewLog(LogOptions{Level: LevelInfo})
	l.SetLevel(LevelDebug)
	AssertEqual(t, LevelDebug, l.Level())
}

func TestLog_Log_SetLevel_Bad(t *T) {
	out := NewBuffer()
	l := NewLog(LogOptions{Level: LevelDebug, Output: out})
	l.SetLevel(LevelQuiet)

	l.Error("suppressed")

	AssertEqual(t, "", out.String())
}

func TestLog_Log_SetLevel_Ugly(t *T) {
	l := NewLog(LogOptions{Level: LevelInfo})
	l.SetLevel(Level(99))
	AssertEqual(t, Level(99), l.Level())
}

func TestLog_Log_SetOutput_Good(t *T) {
	out := NewBuffer()
	l := NewLog(LogOptions{Level: LevelInfo})
	l.SetOutput(out)

	l.Info("redirected")

	AssertContains(t, out.String(), "redirected")
}

func TestLog_Log_SetOutput_Bad(t *T) {
	l := NewLog(LogOptions{Level: LevelInfo})
	AssertNotPanics(t, func() {
		l.SetOutput(nil)
	})
}

func TestLog_Log_SetOutput_Ugly(t *T) {
	first := NewBuffer()
	second := NewBuffer()
	l := NewLog(LogOptions{Level: LevelInfo, Output: first})

	l.Info("first")
	l.SetOutput(second)
	l.Info("second")

	AssertContains(t, first.String(), "first")
	AssertNotContains(t, first.String(), "second")
	AssertContains(t, second.String(), "second")
}

func TestLog_Log_SetRedactKeys_Good(t *T) {
	out := NewBuffer()
	l := NewLog(LogOptions{Level: LevelInfo, Output: out})
	l.SetRedactKeys("token")

	l.Info("login", "token", "secret")

	AssertContains(t, out.String(), "[REDACTED]")
	AssertNotContains(t, out.String(), "secret")
}

func TestLog_Log_SetRedactKeys_Bad(t *T) {
	out := NewBuffer()
	l := NewLog(LogOptions{Level: LevelInfo, Output: out})
	l.SetRedactKeys()

	l.Info("login", "token", "visible")

	AssertContains(t, out.String(), "visible")
}

func TestLog_Log_SetRedactKeys_Ugly(t *T) {
	out := NewBuffer()
	l := NewLog(LogOptions{Level: LevelInfo, Output: out})
	l.SetRedactKeys("token", "authorization")

	l.Info("login", "token", "secret", "authorization", "bearer")

	AssertNotContains(t, out.String(), "secret")
	AssertNotContains(t, out.String(), "bearer")
}

func TestLog_Log_Warn_Good(t *T) {
	out := NewBuffer()
	l := NewLog(LogOptions{Level: LevelWarn, Output: out})

	l.Warn("agent degraded")

	AssertContains(t, out.String(), "[WRN]")
}

func TestLog_Log_Warn_Bad(t *T) {
	out := NewBuffer()
	l := NewLog(LogOptions{Level: LevelError, Output: out})

	l.Warn("suppressed")

	AssertEqual(t, "", out.String())
}

func TestLog_Log_Warn_Ugly(t *T) {
	out := NewBuffer()
	l := NewLog(LogOptions{Level: LevelWarn, Output: out})

	l.Warn("odd", "key")

	AssertContains(t, out.String(), "key=<nil>")
}

func TestLog_LogErr_Log_Good(t *T) {
	out := NewBuffer()
	le := NewLogErr(NewLog(LogOptions{Level: LevelError, Output: out}))

	le.Log(E("agent.Dispatch", "failed", AnError))

	AssertContains(t, out.String(), "agent.Dispatch")
	AssertContains(t, out.String(), "failed")
}

func TestLog_LogErr_Log_Bad(t *T) {
	out := NewBuffer()
	le := NewLogErr(NewLog(LogOptions{Level: LevelError, Output: out}))

	le.Log(nil)

	AssertEqual(t, "", out.String())
}

func TestLog_LogErr_Log_Ugly(t *T) {
	out := NewBuffer()
	le := NewLogErr(NewLog(LogOptions{Level: LevelError, Output: out}))

	le.Log(Wrap(E("agent.Token", "expired", nil), "agent.Dispatch", "failed"))

	AssertContains(t, out.String(), "agent.Dispatch -> agent.Token")
}

func TestLog_LogPanic_Recover_Bad(t *T) {
	out := NewBuffer()
	lp := NewLogPanic(NewLog(LogOptions{Level: LevelError, Output: out}))

	AssertNotPanics(t, func() {
		lp.Recover()
	})
	AssertEqual(t, "", out.String())
}

func TestLog_LogPanic_Recover_Ugly(t *T) {
	out := NewBuffer()
	lp := NewLogPanic(NewLog(LogOptions{Level: LevelError, Output: out}))

	AssertNotPanics(t, func() {
		defer lp.Recover()
		panic(NewError("session token expired"))
	})

	AssertContains(t, out.String(), "panic recovered")
	AssertContains(t, out.String(), "session token expired")
}

func TestLog_NewLog_Good(t *T) {
	out := NewBuffer()
	l := NewLog(LogOptions{Level: LevelInfo, Output: out})

	l.Info("agent ready")

	AssertContains(t, out.String(), "agent ready")
}

func TestLog_NewLog_Bad(t *T) {
	l := NewLog(LogOptions{})
	AssertNotNil(t, l)
	AssertEqual(t, LevelQuiet, l.Level())
}

func TestLog_NewLog_Ugly(t *T) {
	previous := RotationWriterFactory
	defer func() { RotationWriterFactory = previous }()
	out := NewBuffer()
	RotationWriterFactory = func(_ RotationLogOptions) WriteCloser {
		return logTestWriteCloser{w: out}
	}

	l := NewLog(LogOptions{
		Level:    LevelInfo,
		Rotation: &RotationLogOptions{Filename: "agent.log"},
	})
	l.Info("rotated")

	AssertContains(t, out.String(), "rotated")
}

func TestLog_NewLogErr_Good(t *T) {
	le := NewLogErr(NewLog(LogOptions{Level: LevelError, Output: NewBuffer()}))
	AssertNotNil(t, le)
}

func TestLog_NewLogErr_Bad(t *T) {
	le := NewLogErr(nil)
	AssertNotNil(t, le)
}

func TestLog_NewLogErr_Ugly(t *T) {
	out := NewBuffer()
	le := NewLogErr(NewLog(LogOptions{Level: LevelQuiet, Output: out}))
	le.Log(AnError)
	AssertEqual(t, "", out.String())
}

func TestLog_NewLogPanic_Good(t *T) {
	lp := NewLogPanic(NewLog(LogOptions{Level: LevelError, Output: NewBuffer()}))
	AssertNotNil(t, lp)
}

func TestLog_NewLogPanic_Bad(t *T) {
	lp := NewLogPanic(nil)
	AssertNotNil(t, lp)
}

func TestLog_NewLogPanic_Ugly(t *T) {
	out := NewBuffer()
	lp := NewLogPanic(NewLog(LogOptions{Level: LevelQuiet, Output: out}))
	AssertNotPanics(t, func() {
		defer lp.Recover()
		panic("suppressed")
	})
	AssertEqual(t, "", out.String())
}

func TestLog_Security_Good(t *T) {
	original := Default()
	defer SetDefault(original)
	out := NewBuffer()
	SetDefault(NewLog(LogOptions{Level: LevelError, Output: out}))

	Security("entitlement denied")

	AssertContains(t, out.String(), "[SEC]")
}

func TestLog_Security_Bad(t *T) {
	original := Default()
	defer SetDefault(original)
	out := NewBuffer()
	SetDefault(NewLog(LogOptions{Level: LevelQuiet, Output: out}))

	Security("suppressed")

	AssertEqual(t, "", out.String())
}

func TestLog_Security_Ugly(t *T) {
	original := Default()
	defer SetDefault(original)
	out := NewBuffer()
	SetDefault(NewLog(LogOptions{Level: LevelError, Output: out}))

	Security("token", "value", "a\nb")

	AssertContains(t, out.String(), "value=\"a\\nb\"")
}

func TestLog_SetDefault_Bad(t *T) {
	original := Default()
	defer SetDefault(original)

	SetDefault(nil)

	AssertNil(t, Default())
}

func TestLog_SetDefault_Ugly(t *T) {
	original := Default()
	defer SetDefault(original)
	first := NewLog(LogOptions{Level: LevelWarn, Output: NewBuffer()})
	second := NewLog(LogOptions{Level: LevelDebug, Output: NewBuffer()})

	SetDefault(first)
	SetDefault(second)

	AssertSame(t, second, Default())
}

func TestLog_SetLevel_Bad(t *T) {
	original := Default()
	defer SetDefault(original)
	out := NewBuffer()
	SetDefault(NewLog(LogOptions{Level: LevelDebug, Output: out}))

	SetLevel(LevelQuiet)
	Error("suppressed")

	AssertEqual(t, "", out.String())
}

func TestLog_SetLevel_Ugly(t *T) {
	original := Default()
	defer SetDefault(original)
	l := NewLog(LogOptions{Level: LevelInfo, Output: NewBuffer()})
	SetDefault(l)

	SetLevel(Level(99))

	AssertEqual(t, Level(99), l.Level())
}

func TestLog_SetRedactKeys_Bad(t *T) {
	original := Default()
	defer SetDefault(original)
	out := NewBuffer()
	SetDefault(NewLog(LogOptions{Level: LevelInfo, Output: out}))

	SetRedactKeys()
	Info("login", "token", "visible")

	AssertContains(t, out.String(), "visible")
}

func TestLog_SetRedactKeys_Ugly(t *T) {
	original := Default()
	defer SetDefault(original)
	out := NewBuffer()
	SetDefault(NewLog(LogOptions{Level: LevelInfo, Output: out}))

	SetRedactKeys("token", "authorization")
	Info("login", "token", "secret", "authorization", "bearer")

	AssertNotContains(t, out.String(), "secret")
	AssertNotContains(t, out.String(), "bearer")
}

func TestLog_Username_Bad(t *T) {
	AssertNotPanics(t, func() {
		_ = Username()
	})
}

func TestLog_Username_Ugly(t *T) {
	AssertEqual(t, Username(), Username())
}

func TestLog_Warn_Bad(t *T) {
	original := Default()
	defer SetDefault(original)
	out := NewBuffer()
	SetDefault(NewLog(LogOptions{Level: LevelError, Output: out}))

	Warn("suppressed")

	AssertEqual(t, "", out.String())
}

func TestLog_Warn_Ugly(t *T) {
	original := Default()
	defer SetDefault(original)
	out := NewBuffer()
	SetDefault(NewLog(LogOptions{Level: LevelWarn, Output: out}))

	Warn("odd", "key")

	AssertContains(t, out.String(), "key=<nil>")
}
