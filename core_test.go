package core_test

import (
	"context"

	. "dappco.re/go"
)

// --- New ---

func TestCore_New_Good(t *T) {
	c := New()
	AssertNotNil(t, c)
}

func TestCore_New_WithOptions_Good(t *T) {
	c := New(WithOptions(NewOptions(Option{Key: "name", Value: "myapp"})))
	AssertNotNil(t, c)
	AssertEqual(t, "myapp", c.App().Name)
}

func TestCore_New_WithOptions_Bad(t *T) {
	// Empty options — should still create a valid Core
	c := New(WithOptions(NewOptions()))
	AssertNotNil(t, c)
}

func TestCore_New_WithService_Good(t *T) {
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
	AssertTrue(t, svc.OK)

	c.ServiceStartup(context.Background(), nil)
	AssertTrue(t, started)
}

func TestCore_New_WithServiceLock_Good(t *T) {
	c := New(
		WithService(func(c *Core) Result {
			c.Service("allowed", Service{})
			return Result{OK: true}
		}),
		WithServiceLock(),
	)

	// Registration after lock should fail
	reg := c.Service("blocked", Service{})
	AssertFalse(t, reg.OK)
}

func TestCore_New_WithService_Bad_FailingOption(t *T) {
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
	AssertFalse(t, secondCalled, "second option should not run after first fails")
}

// --- Accessors ---

func TestCore_Accessors_Good(t *T) {
	c := New()
	AssertNotNil(t, c.App())
	AssertNotNil(t, c.Data())
	AssertNotNil(t, c.Drive())
	AssertNotNil(t, c.Fs())
	AssertNotNil(t, c.Config())
	AssertNotNil(t, c.Error())
	AssertNotNil(t, c.Log())
	AssertNotNil(t, c.Cli())
	AssertNotNil(t, c.IPC())
	AssertNotNil(t, c.I18n())
	AssertEqual(t, c, c.Core())
}

func TestOptions_Accessor_Good(t *T) {
	c := New(WithOptions(NewOptions(
		Option{Key: "name", Value: "testapp"},
		Option{Key: "port", Value: 8080},
		Option{Key: "debug", Value: true},
	)))
	opts := c.Options()
	AssertNotNil(t, opts)
	AssertEqual(t, "testapp", opts.String("name"))
	AssertEqual(t, 8080, opts.Int("port"))
	AssertTrue(t, opts.Bool("debug"))
}

func TestOptions_Accessor_Nil(t *T) {
	c := New()
	// No options passed — Options() returns nil
	AssertNil(t, c.Options())
}

// --- Core Error/Log Helpers ---

func TestCore_LogError_Good(t *T) {
	c := New()
	cause := AnError
	r := c.LogError(cause, "test.Operation", "something broke")

	err, ok := r.Value.(error)
	AssertTrue(t, ok)
	AssertErrorIs(t, err, cause)
}

func TestCore_LogWarn_Good(t *T) {
	c := New()
	r := c.LogWarn(AnError, "test.Operation", "heads up")

	_, ok := r.Value.(error)
	AssertTrue(t, ok)
}

func TestCore_Must_Ugly(t *T) {
	c := New()
	AssertPanics(t, func() {
		c.Must(AnError, "test.Operation", "fatal")
	})
}

func TestCore_Must_Nil_Good(t *T) {
	c := New()
	AssertNotPanics(t, func() {
		c.Must(nil, "test.Operation", "no error")
	})
}

// --- RegistryOf ---

func TestCore_RegistryOf_Good_Services(t *T) {
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
	AssertTrue(t, reg.Has("alpha"))
	AssertTrue(t, reg.Has("bravo"))
	AssertTrue(t, reg.Has("cli"))
}

func TestCore_RegistryOf_Good_Commands(t *T) {
	c := New()
	c.Command("deploy", Command{Action: func(_ Options) Result { return Result{OK: true} }})
	c.Command("test", Command{Action: func(_ Options) Result { return Result{OK: true} }})

	reg := c.RegistryOf("commands")
	AssertTrue(t, reg.Has("deploy"))
	AssertTrue(t, reg.Has("test"))
}

func TestCore_RegistryOf_Good_Actions(t *T) {
	c := New()
	c.Action("process.run", func(_ context.Context, _ Options) Result { return Result{OK: true} })
	c.Action("brain.recall", func(_ context.Context, _ Options) Result { return Result{OK: true} })

	reg := c.RegistryOf("actions")
	AssertTrue(t, reg.Has("process.run"))
	AssertTrue(t, reg.Has("brain.recall"))
	AssertEqual(t, 2, reg.Len())
}

func TestCore_RegistryOf_Bad_Unknown(t *T) {
	c := New()
	reg := c.RegistryOf("nonexistent")
	AssertEqual(t, 0, reg.Len(), "unknown registry returns empty")
}

// --- RunE ---

func TestCore_RunE_Good(t *T) {
	c := New(
		WithService(func(c *Core) Result {
			return c.Service("healthy", Service{
				OnStart: func() Result { return Result{OK: true} },
				OnStop:  func() Result { return Result{OK: true} },
			})
		}),
	)
	err := c.RunE()
	AssertNoError(t, err)
}

func TestCore_RunE_Bad_StartupFailure(t *T) {
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
	AssertError(t, err)
	AssertContains(t, err.Error(), "startup failed")
}

func TestCore_RunE_Ugly_StartupFailureCallsShutdown(t *T) {
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
	AssertError(t, err)
	AssertTrue(t, shutdownCalled, "ServiceShutdown must be called even when startup fails — cleanup service must get OnStop")
}

// Run() delegates to RunE() — tested via RunE tests above.
// os.Exit behaviour is verified by RunE returning error correctly.
