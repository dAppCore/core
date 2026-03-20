package core_test

import (
	"testing"

	. "forge.lthn.ai/core/go/pkg/core"
	"github.com/stretchr/testify/assert"
)

// --- New ---

func TestNew_Good(t *testing.T) {
	c := New()
	assert.NotNil(t, c)
}

func TestNew_WithOptions_Good(t *testing.T) {
	c := New(Options{{Key: "name", Value: "myapp"}})
	assert.NotNil(t, c)
	assert.Equal(t, "myapp", c.App().Name)
}

func TestNew_WithOptions_Bad(t *testing.T) {
	// Empty options — should still create a valid Core
	c := New(Options{})
	assert.NotNil(t, c)
}

// --- Accessors ---

func TestAccessors_Good(t *testing.T) {
	c := New()
	assert.NotNil(t, c.App())
	assert.NotNil(t, c.Data())
	assert.NotNil(t, c.Drive())
	assert.NotNil(t, c.Fs())
	assert.NotNil(t, c.Config())
	assert.NotNil(t, c.Error())
	assert.NotNil(t, c.Log())
	assert.NotNil(t, c.Cli())
	assert.NotNil(t, c.IPC())
	assert.NotNil(t, c.I18n())
	assert.Equal(t, c, c.Core())
}

func TestOptions_Accessor_Good(t *testing.T) {
	c := New(Options{
		{Key: "name", Value: "testapp"},
		{Key: "port", Value: 8080},
		{Key: "debug", Value: true},
	})
	opts := c.Options()
	assert.NotNil(t, opts)
	assert.Equal(t, "testapp", opts.String("name"))
	assert.Equal(t, 8080, opts.Int("port"))
	assert.True(t, opts.Bool("debug"))
}

func TestOptions_Accessor_Nil(t *testing.T) {
	c := New()
	// No options passed — Options() returns nil
	assert.Nil(t, c.Options())
}

// --- Core Error/Log Helpers ---

func TestCore_LogError_Good(t *testing.T) {
	c := New()
	cause := assert.AnError
	r := c.LogError(cause, "test.Operation", "something broke")
	assert.False(t, r.OK)
	err, ok := r.Value.(error)
	assert.True(t, ok)
	assert.ErrorIs(t, err, cause)
}

func TestCore_LogWarn_Good(t *testing.T) {
	c := New()
	r := c.LogWarn(assert.AnError, "test.Operation", "heads up")
	assert.False(t, r.OK)
	_, ok := r.Value.(error)
	assert.True(t, ok)
}

func TestCore_Must_Ugly(t *testing.T) {
	c := New()
	assert.Panics(t, func() {
		c.Must(assert.AnError, "test.Operation", "fatal")
	})
}

func TestCore_Must_Nil_Good(t *testing.T) {
	c := New()
	assert.NotPanics(t, func() {
		c.Must(nil, "test.Operation", "no error")
	})
}
