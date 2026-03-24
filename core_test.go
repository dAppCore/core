package core_test

import (
	"context"
	"testing"

	. "dappco.re/go/core"
	"github.com/stretchr/testify/assert"
)

// --- New ---

func TestNew_Good(t *testing.T) {
	c := New().Value.(*Core)
	assert.NotNil(t, c)
}

func TestNew_WithOptions_Good(t *testing.T) {
	c := New(WithOptions(NewOptions(Option{Key: "name", Value: "myapp"}))).Value.(*Core)
	assert.NotNil(t, c)
	assert.Equal(t, "myapp", c.App().Name)
}

func TestNew_WithOptions_Bad(t *testing.T) {
	// Empty options — should still create a valid Core
	c := New(WithOptions(NewOptions())).Value.(*Core)
	assert.NotNil(t, c)
}

func TestNew_WithService_Good(t *testing.T) {
	started := false
	r := New(
		WithOptions(NewOptions(Option{Key: "name", Value: "myapp"})),
		WithService(func(c *Core) Result {
			c.Service("test", Service{
				OnStart: func() Result { started = true; return Result{OK: true} },
			})
			return Result{OK: true}
		}),
	)
	assert.True(t, r.OK)
	c := r.Value.(*Core)

	svc := c.Service("test")
	assert.True(t, svc.OK)

	c.ServiceStartup(context.Background(), nil)
	assert.True(t, started)
}

func TestNew_WithServiceLock_Good(t *testing.T) {
	r := New(
		WithService(func(c *Core) Result {
			c.Service("allowed", Service{})
			return Result{OK: true}
		}),
		WithServiceLock(),
	)
	assert.True(t, r.OK)
	c := r.Value.(*Core)

	// Registration after lock should fail
	reg := c.Service("blocked", Service{})
	assert.False(t, reg.OK)
}

func TestNew_WithService_Bad_FailingOption(t *testing.T) {
	secondCalled := false
	r := New(
		WithService(func(c *Core) Result {
			return Result{Value: E("test", "intentional failure", nil), OK: false}
		}),
		WithService(func(c *Core) Result {
			secondCalled = true
			return Result{OK: true}
		}),
	)
	assert.False(t, r.OK)
	assert.False(t, secondCalled, "second option should not run after first fails")
}

// --- Accessors ---

func TestAccessors_Good(t *testing.T) {
	c := New().Value.(*Core)
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
	c := New(WithOptions(Options{
		{Key: "name", Value: "testapp"},
		{Key: "port", Value: 8080},
		{Key: "debug", Value: true},
	})).Value.(*Core)
	opts := c.Options()
	assert.NotNil(t, opts)
	assert.Equal(t, "testapp", opts.String("name"))
	assert.Equal(t, 8080, opts.Int("port"))
	assert.True(t, opts.Bool("debug"))
}

func TestOptions_Accessor_Nil(t *testing.T) {
	c := New().Value.(*Core)
	// No options passed — Options() returns nil
	assert.Nil(t, c.Options())
}

// --- Core Error/Log Helpers ---

func TestCore_LogError_Good(t *testing.T) {
	c := New().Value.(*Core)
	cause := assert.AnError
	r := c.LogError(cause, "test.Operation", "something broke")
	assert.False(t, r.OK)
	err, ok := r.Value.(error)
	assert.True(t, ok)
	assert.ErrorIs(t, err, cause)
}

func TestCore_LogWarn_Good(t *testing.T) {
	c := New().Value.(*Core)
	r := c.LogWarn(assert.AnError, "test.Operation", "heads up")
	assert.False(t, r.OK)
	_, ok := r.Value.(error)
	assert.True(t, ok)
}

func TestCore_Must_Ugly(t *testing.T) {
	c := New().Value.(*Core)
	assert.Panics(t, func() {
		c.Must(assert.AnError, "test.Operation", "fatal")
	})
}

func TestCore_Must_Nil_Good(t *testing.T) {
	c := New().Value.(*Core)
	assert.NotPanics(t, func() {
		c.Must(nil, "test.Operation", "no error")
	})
}
