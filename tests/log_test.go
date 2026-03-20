package core_test

import (
	"os"
	"testing"

	. "forge.lthn.ai/core/go/pkg/core"
	"github.com/stretchr/testify/assert"
)

// --- Log ---

func TestLog_New_Good(t *testing.T) {
	l := NewLog(LogOptions{Level: LevelInfo})
	assert.NotNil(t, l)
}

func TestLog_AllLevels_Good(t *testing.T) {
	l := NewLog(LogOptions{Level: LevelDebug})
	l.Debug("debug")
	l.Info("info")
	l.Warn("warn")
	l.Error("error")
	l.Security("security event")
}

func TestLog_LevelFiltering_Good(t *testing.T) {
	// At Error level, Debug/Info/Warn should be suppressed (no panic)
	l := NewLog(LogOptions{Level: LevelError})
	l.Debug("suppressed")
	l.Info("suppressed")
	l.Warn("suppressed")
	l.Error("visible")
}

func TestLog_SetLevel_Good(t *testing.T) {
	l := NewLog(LogOptions{Level: LevelInfo})
	l.SetLevel(LevelDebug)
	assert.Equal(t, LevelDebug, l.Level())
}

func TestLog_SetRedactKeys_Good(t *testing.T) {
	l := NewLog(LogOptions{Level: LevelInfo})
	l.SetRedactKeys("password", "token")
	// Redacted keys should mask values in output
	l.Info("login", "password", "secret123", "user", "admin")
}

func TestLog_LevelString_Good(t *testing.T) {
	assert.Equal(t, "debug", LevelDebug.String())
	assert.Equal(t, "info", LevelInfo.String())
	assert.Equal(t, "warn", LevelWarn.String())
	assert.Equal(t, "error", LevelError.String())
}

func TestLog_CoreLog_Good(t *testing.T) {
	c := New()
	assert.NotNil(t, c.Log())
}

func TestLog_ErrorSink_Good(t *testing.T) {
	l := NewLog(LogOptions{Level: LevelInfo})
	var sink ErrorSink = l
	sink.Error("test")
	sink.Warn("test")
}

// --- Default Logger ---

func TestLog_Default_Good(t *testing.T) {
	d := Default()
	assert.NotNil(t, d)
}

func TestLog_SetDefault_Good(t *testing.T) {
	original := Default()
	defer SetDefault(original)

	custom := NewLog(LogOptions{Level: LevelDebug})
	SetDefault(custom)
	assert.Equal(t, custom, Default())
}

func TestLog_PackageLevelFunctions_Good(t *testing.T) {
	// Package-level log functions use the default logger
	Debug("debug msg")
	Info("info msg")
	Warn("warn msg")
	Error("error msg")
	Security("security msg")
}

func TestLog_PackageSetLevel_Good(t *testing.T) {
	original := Default()
	defer SetDefault(original)

	SetLevel(LevelDebug)
	SetRedactKeys("secret")
}

func TestLog_Username_Good(t *testing.T) {
	u := Username()
	assert.NotEmpty(t, u)
}

// --- LogErr ---

func TestLogErr_Good(t *testing.T) {
	l := NewLog(LogOptions{Level: LevelInfo})
	le := NewLogErr(l)
	assert.NotNil(t, le)

	err := E("test.Op", "something broke", nil)
	le.Log(err)
}

func TestLogErr_Nil_Good(t *testing.T) {
	l := NewLog(LogOptions{Level: LevelInfo})
	le := NewLogErr(l)
	le.Log(nil) // should not panic
}

// --- LogPan ---

func TestLogPan_Good(t *testing.T) {
	l := NewLog(LogOptions{Level: LevelInfo})
	lp := NewLogPan(l)
	assert.NotNil(t, lp)
}

func TestLogPan_Recover_Good(t *testing.T) {
	l := NewLog(LogOptions{Level: LevelInfo})
	lp := NewLogPan(l)
	assert.NotPanics(t, func() {
		defer lp.Recover()
		panic("caught")
	})
}

// --- SetOutput ---

func TestLog_SetOutput_Good(t *testing.T) {
	l := NewLog(LogOptions{Level: LevelInfo})
	l.SetOutput(os.Stderr)
	// Should not panic — just changes where logs go
	l.Info("redirected")
}
