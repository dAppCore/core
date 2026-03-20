package core_test

import (
	"errors"
	"testing"

	. "forge.lthn.ai/core/go/pkg/core"
	"github.com/stretchr/testify/assert"
)

// --- Error Creation ---

func TestE_Good(t *testing.T) {
	err := E("user.Save", "failed to save", nil)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "user.Save")
	assert.Contains(t, err.Error(), "failed to save")
}

func TestE_WithCause_Good(t *testing.T) {
	cause := errors.New("connection refused")
	err := E("db.Connect", "database unavailable", cause)
	assert.ErrorIs(t, err, cause)
}

func TestWrap_Good(t *testing.T) {
	cause := errors.New("timeout")
	err := Wrap(cause, "api.Call", "request failed")
	assert.Error(t, err)
	assert.ErrorIs(t, err, cause)
}

func TestWrap_Nil_Good(t *testing.T) {
	err := Wrap(nil, "api.Call", "request failed")
	assert.Nil(t, err)
}

func TestWrapCode_Good(t *testing.T) {
	cause := errors.New("invalid email")
	err := WrapCode(cause, "VALIDATION_ERROR", "user.Validate", "bad input")
	assert.Error(t, err)
	assert.Equal(t, "VALIDATION_ERROR", ErrCode(err))
}

func TestNewCode_Good(t *testing.T) {
	err := NewCode("NOT_FOUND", "resource not found")
	assert.Error(t, err)
	assert.Equal(t, "NOT_FOUND", ErrCode(err))
}

// --- Error Introspection ---

func TestOp_Good(t *testing.T) {
	err := E("brain.Recall", "search failed", nil)
	assert.Equal(t, "brain.Recall", Op(err))
}

func TestOp_Bad(t *testing.T) {
	err := errors.New("plain error")
	assert.Equal(t, "", Op(err))
}

func TestErrorMessage_Good(t *testing.T) {
	err := E("op", "the message", nil)
	assert.Equal(t, "the message", ErrorMessage(err))
}

func TestErrorMessage_Plain(t *testing.T) {
	err := errors.New("plain")
	assert.Equal(t, "plain", ErrorMessage(err))
}

func TestErrorMessage_Nil(t *testing.T) {
	assert.Equal(t, "", ErrorMessage(nil))
}

func TestRoot_Good(t *testing.T) {
	root := errors.New("root cause")
	wrapped := Wrap(root, "layer1", "first wrap")
	double := Wrap(wrapped, "layer2", "second wrap")
	assert.Equal(t, root, Root(double))
}

func TestRoot_Nil(t *testing.T) {
	assert.Nil(t, Root(nil))
}

func TestStackTrace_Good(t *testing.T) {
	err := Wrap(E("inner", "cause", nil), "outer", "wrapper")
	stack := StackTrace(err)
	assert.Len(t, stack, 2)
	assert.Equal(t, "outer", stack[0])
	assert.Equal(t, "inner", stack[1])
}

func TestFormatStackTrace_Good(t *testing.T) {
	err := Wrap(E("a", "x", nil), "b", "y")
	formatted := FormatStackTrace(err)
	assert.Equal(t, "b -> a", formatted)
}

// --- ErrorLog ---

func TestErrorLog_Good(t *testing.T) {
	c := New()
	cause := errors.New("boom")
	err := c.Log().Error(cause, "test.Op", "something broke")
	assert.Error(t, err)
	assert.ErrorIs(t, err, cause)
}

func TestErrorLog_Nil_Good(t *testing.T) {
	c := New()
	err := c.Log().Error(nil, "test.Op", "no error")
	assert.Nil(t, err)
}

func TestErrorLog_Warn_Good(t *testing.T) {
	c := New()
	cause := errors.New("warning")
	err := c.Log().Warn(cause, "test.Op", "heads up")
	assert.Error(t, err)
}

func TestErrorLog_Must_Ugly(t *testing.T) {
	c := New()
	assert.Panics(t, func() {
		c.Log().Must(errors.New("fatal"), "test.Op", "must fail")
	})
}

func TestErrorLog_Must_Nil_Good(t *testing.T) {
	c := New()
	assert.NotPanics(t, func() {
		c.Log().Must(nil, "test.Op", "no error")
	})
}

// --- ErrorPanic ---

func TestErrorPanic_Recover_Good(t *testing.T) {
	c := New()
	// Should not panic — Recover catches it
	assert.NotPanics(t, func() {
		defer c.Error().Recover()
		panic("test panic")
	})
}

func TestErrorPanic_SafeGo_Good(t *testing.T) {
	c := New()
	done := make(chan bool, 1)
	c.Error().SafeGo(func() {
		done <- true
	})
	assert.True(t, <-done)
}

func TestErrorPanic_SafeGo_Panic_Good(t *testing.T) {
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

func TestIs_Good(t *testing.T) {
	target := errors.New("target")
	wrapped := Wrap(target, "op", "msg")
	assert.True(t, Is(wrapped, target))
}

func TestAs_Good(t *testing.T) {
	err := E("op", "msg", nil)
	var e *Err
	assert.True(t, As(err, &e))
	assert.Equal(t, "op", e.Op)
}

func TestNewError_Good(t *testing.T) {
	err := NewError("simple error")
	assert.Equal(t, "simple error", err.Error())
}

func TestJoin_Good(t *testing.T) {
	e1 := errors.New("first")
	e2 := errors.New("second")
	joined := Join(e1, e2)
	assert.ErrorIs(t, joined, e1)
	assert.ErrorIs(t, joined, e2)
}
