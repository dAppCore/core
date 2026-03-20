package core_test

import (
	"testing"

	. "forge.lthn.ai/core/go/pkg/core"
	"github.com/stretchr/testify/assert"
)

// --- Log (Structured Logger) ---

func TestLog_New_Good(t *testing.T) {
	l := NewLog(LogOpts{Level: LevelInfo})
	assert.NotNil(t, l)
}

func TestLog_Levels_Good(t *testing.T) {
	for _, level := range []Level{LevelDebug, LevelInfo, LevelWarn, LevelError} {
		l := NewLog(LogOpts{Level: level})
		l.Debug("debug msg")
		l.Info("info msg")
		l.Warn("warn msg")
		l.Error("error msg")
	}
}

func TestLog_CoreLog_Good(t *testing.T) {
	c := New()
	assert.NotNil(t, c.Log())
}

func TestLog_ErrorSink_Interface(t *testing.T) {
	l := NewLog(LogOpts{Level: LevelInfo})
	var sink ErrorSink = l
	sink.Error("test", "key", "val")
	sink.Warn("test", "key", "val")
}
