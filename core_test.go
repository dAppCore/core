package core_test

import (
	"context"
	"testing"

	. "dappco.re/go/core"
	"github.com/stretchr/testify/assert"
)

// --- New ---

func TestCore_New_Good(t *testing.T) {
	c := New()
	assert.NotNil(t, c)
}

func TestCore_New_WithOptions_Good(t *testing.T) {
	c := New(WithOptions(NewOptions(Option{Key: "name", Value: "myapp"})))
	assert.NotNil(t, c)
	assert.Equal(t, "myapp", c.App().Name)
}

func TestCore_New_WithOptions_Bad(t *testing.T) {
	// Empty options — should still create a valid Core
	c := New(WithOptions(NewOptions()))
	assert.NotNil(t, c)
}

func TestCore_New_WithService_Good(t *testing.T) {
	started := false
	c := New(
		WithOptions(NewOptions(Option{Key: "name", Value: "myapp"})),
		WithService(func(c *Core) Result {
			c.Service("test", Service{
				OnStart: func() Result { started = true; return Result{OK: true} },
			})
			return Result{OK: true}
		}),
	)

	svc := c.Service("test")
	assert.True(t, svc.OK)

	c.ServiceStartup(context.Background(), nil)
	assert.True(t, started)
}

func TestCore_New_WithServiceLock_Good(t *testing.T) {
	c := New(
		WithService(func(c *Core) Result {
			c.Service("allowed", Service{})
			return Result{OK: true}
		}),
		WithServiceLock(),
	)

	// Registration after lock should fail
	reg := c.Service("blocked", Service{})
	assert.False(t, reg.OK)
}

func TestCore_New_WithService_Bad_FailingOption(t *testing.T) {
	secondCalled := false
	_ = New(
		WithService(func(c *Core) Result {
			return Result{Value: E("test", "intentional failure", nil), OK: false}
		}),
		WithService(func(c *Core) Result {
			secondCalled = true
			return Result{OK: true}
		}),
	)
	assert.False(t, secondCalled, "second option should not run after first fails")
}

// --- Accessors ---

func TestCore_Accessors_Good(t *testing.T) {
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
	c := New(WithOptions(NewOptions(
		Option{Key: "name", Value: "testapp"},
		Option{Key: "port", Value: 8080},
		Option{Key: "debug", Value: true},
	)))
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
	
	err, ok := r.Value.(error)
	assert.True(t, ok)
	assert.ErrorIs(t, err, cause)
}

func TestCore_LogWarn_Good(t *testing.T) {
	c := New()
	r := c.LogWarn(assert.AnError, "test.Operation", "heads up")
	
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

// --- RegistryOf ---

func TestCore_RegistryOf_Good_Services(t *testing.T) {
	c := New(
		WithService(func(c *Core) Result {
			return c.Service("alpha", Service{})
		}),
		WithService(func(c *Core) Result {
			return c.Service("bravo", Service{})
		}),
	)
	reg := c.RegistryOf("services")
	// cli is auto-registered + our 2
	assert.True(t, reg.Has("alpha"))
	assert.True(t, reg.Has("bravo"))
	assert.True(t, reg.Has("cli"))
}

func TestCore_RegistryOf_Good_Commands(t *testing.T) {
	c := New()
	c.Command("deploy", Command{Action: func(_ Options) Result { return Result{OK: true} }})
	c.Command("test", Command{Action: func(_ Options) Result { return Result{OK: true} }})

	reg := c.RegistryOf("commands")
	assert.True(t, reg.Has("deploy"))
	assert.True(t, reg.Has("test"))
}

func TestCore_RegistryOf_Good_Actions(t *testing.T) {
	c := New()
	c.Action("process.run", func(_ context.Context, _ Options) Result { return Result{OK: true} })
	c.Action("brain.recall", func(_ context.Context, _ Options) Result { return Result{OK: true} })

	reg := c.RegistryOf("actions")
	assert.True(t, reg.Has("process.run"))
	assert.True(t, reg.Has("brain.recall"))
	assert.Equal(t, 2, reg.Len())
}

func TestCore_RegistryOf_Bad_Unknown(t *testing.T) {
	c := New()
	reg := c.RegistryOf("nonexistent")
	assert.Equal(t, 0, reg.Len(), "unknown registry returns empty")
}

// --- RunE ---

func TestCore_RunE_Good(t *testing.T) {
	c := New(
		WithService(func(c *Core) Result {
			return c.Service("healthy", Service{
				OnStart: func() Result { return Result{OK: true} },
				OnStop:  func() Result { return Result{OK: true} },
			})
		}),
	)
	err := c.RunE()
	assert.NoError(t, err)
}

func TestCore_RunE_Bad_StartupFailure(t *testing.T) {
	c := New(
		WithService(func(c *Core) Result {
			return c.Service("broken", Service{
				OnStart: func() Result {
					return Result{Value: NewError("startup failed"), OK: false}
				},
			})
		}),
	)
	err := c.RunE()
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "startup failed")
}

func TestCore_RunE_Ugly_StartupFailureCallsShutdown(t *testing.T) {
	shutdownCalled := false
	c := New(
		WithService(func(c *Core) Result {
			return c.Service("cleanup", Service{
				OnStart: func() Result { return Result{OK: true} },
				OnStop:  func() Result { shutdownCalled = true; return Result{OK: true} },
			})
		}),
		WithService(func(c *Core) Result {
			return c.Service("broken", Service{
				OnStart: func() Result {
					return Result{Value: NewError("boom"), OK: false}
				},
			})
		}),
	)
	err := c.RunE()
	assert.Error(t, err)
	assert.True(t, shutdownCalled, "ServiceShutdown must be called even when startup fails — cleanup service must get OnStop")
}

// Run() delegates to RunE() — tested via RunE tests above.
// os.Exit behaviour is verified by RunE returning error correctly.
