package core_test

import (
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

	c.ServiceStartup(Background(), nil)
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
	c.Action("process.run", func(_ Context, _ Options) Result { return Result{OK: true} })
	c.Action("brain.recall", func(_ Context, _ Options) Result { return Result{OK: true} })

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

func TestCore_Core_ACTION_Good(t *T) {
	c := New()
	var got Message
	c.RegisterAction(func(_ *Core, msg Message) Result {
		got = msg
		return Result{OK: true}
	})

	r := c.ACTION("agent.dispatch")

	AssertTrue(t, r.OK)
	AssertEqual(t, Message("agent.dispatch"), got)
}

func TestCore_Core_ACTION_Bad(t *T) {
	c := New()
	c.RegisterAction(func(_ *Core, _ Message) Result {
		return Result{Value: NewError("handler refused"), OK: false}
	})

	r := c.ACTION("agent.dispatch")

	AssertTrue(t, r.OK)
}

func TestCore_Core_ACTION_Ugly(t *T) {
	c := New()
	c.RegisterAction(func(_ *Core, _ Message) Result {
		panic("handler panic")
	})

	r := c.ACTION(nil)

	AssertTrue(t, r.OK)
}

func TestCore_Core_App_Good(t *T) {
	c := New(WithOption("name", "homelab"))
	AssertEqual(t, "homelab", c.App().Name)
}

func TestCore_Core_App_Bad(t *T) {
	c := New()
	AssertEqual(t, "", c.App().Name)
}

func TestCore_Core_App_Ugly(t *T) {
	c := New()
	AssertSame(t, c.App(), c.App())
}

func TestCore_Core_Cli_Good(t *T) {
	c := New()
	AssertNotNil(t, c.Cli())
}

func TestCore_Core_Cli_Bad(t *T) {
	c := New(WithServiceLock())
	AssertNotNil(t, c.Cli())
}

func TestCore_Core_Cli_Ugly(t *T) {
	c := New()
	c.Cli().SetOutput(NewBuffer())
	AssertNotPanics(t, func() {
		_ = c.Cli().Run()
	})
}

func TestCore_Core_Config_Good(t *T) {
	c := New()
	c.Config().Set("agent.region", "lab")
	AssertEqual(t, "lab", c.Config().String("agent.region"))
}

func TestCore_Core_Config_Bad(t *T) {
	c := New()
	AssertEqual(t, "", c.Config().String("missing"))
}

func TestCore_Core_Config_Ugly(t *T) {
	c := New()
	AssertSame(t, c.Config(), c.Config())
}

func TestCore_Core_Context_Good(t *T) {
	c := New()
	AssertNotNil(t, c.Context())
}

func TestCore_Core_Context_Bad(t *T) {
	c := New()
	c.ServiceShutdown(Background())
	AssertError(t, c.Context().Err())
}

func TestCore_Core_Context_Ugly(t *T) {
	c := New()
	before := c.Context()
	c.ServiceStartup(Background(), nil)
	AssertNotNil(t, before)
	AssertNotNil(t, c.Context())
}

func TestCore_Core_Core_Good(t *T) {
	c := New()
	AssertSame(t, c, c.Core())
}

func TestCore_Core_Core_Bad(t *T) {
	c := New(WithOption("name", "agent"))
	AssertSame(t, c, c.Core())
}

func TestCore_Core_Core_Ugly(t *T) {
	var c *Core
	AssertNil(t, c.Core())
}

func TestCore_Core_Data_Good(t *T) {
	c := New()
	AssertNotNil(t, c.Data())
}

func TestCore_Core_Data_Bad(t *T) {
	c := New()
	AssertSame(t, c.Data(), c.Data())
}

func TestCore_Core_Data_Ugly(t *T) {
	c := New()
	AssertEqual(t, 0, c.Data().Len())
}

func TestCore_Core_Drive_Good(t *T) {
	c := New()
	AssertNotNil(t, c.Drive())
}

func TestCore_Core_Drive_Bad(t *T) {
	c := New()
	AssertSame(t, c.Drive(), c.Drive())
}

func TestCore_Core_Drive_Ugly(t *T) {
	c := New()
	AssertEqual(t, 0, c.Drive().Len())
}

func TestCore_Core_Env_Good(t *T) {
	c := New()
	AssertEqual(t, OS(), c.Env("OS"))
}

func TestCore_Core_Env_Bad(t *T) {
	c := New()
	AssertEqual(t, "", c.Env("CORE_TEST_MISSING"))
}

func TestCore_Core_Env_Ugly(t *T) {
	t.Setenv("CORE_TEST_SESSION", "token")
	c := New()
	AssertEqual(t, "token", c.Env("CORE_TEST_SESSION"))
}

func TestCore_Core_Error_Good(t *T) {
	c := New()
	AssertNotNil(t, c.Error())
}

func TestCore_Core_Error_Bad(t *T) {
	c := New()
	AssertNotPanics(t, func() {
		c.Error().Recover()
	})
}

func TestCore_Core_Error_Ugly(t *T) {
	c := New()
	AssertSame(t, c.Error(), c.Error())
}

func TestCore_Core_Fs_Good(t *T) {
	c := New()
	AssertNotNil(t, c.Fs())
}

func TestCore_Core_Fs_Bad(t *T) {
	c := New()
	AssertFalse(t, c.Fs().Read(Path(t.TempDir(), "missing")).OK)
}

func TestCore_Core_Fs_Ugly(t *T) {
	c := New()
	AssertSame(t, c.Fs(), c.Fs())
}

func TestCore_Core_I18n_Good(t *T) {
	c := New()
	AssertNotNil(t, c.I18n())
}

func TestCore_Core_I18n_Bad(t *T) {
	c := New()
	AssertSame(t, c.I18n(), c.I18n())
}

func TestCore_Core_I18n_Ugly(t *T) {
	c := New()
	r := c.I18n().Translate("missing.key")
	AssertTrue(t, r.OK)
	AssertEqual(t, "missing.key", r.Value)
}

func TestCore_Core_IPC_Good(t *T) {
	c := New()
	AssertNotNil(t, c.IPC())
}

func TestCore_Core_IPC_Bad(t *T) {
	c := New()
	AssertSame(t, c.IPC(), c.IPC())
}

func TestCore_Core_IPC_Ugly(t *T) {
	c := New()
	AssertTrue(t, c.ACTION(nil).OK)
}

func TestCore_Core_Log_Good(t *T) {
	c := New()
	AssertNotNil(t, c.Log())
}

func TestCore_Core_Log_Bad(t *T) {
	c := New()
	AssertTrue(t, c.Log().Error(nil, "agent.Dispatch", "ok").OK)
}

func TestCore_Core_Log_Ugly(t *T) {
	c := New()
	AssertSame(t, c.Log(), c.Log())
}

func TestCore_Core_LogError_Good(t *T) {
	r := New().LogError(AnError, "agent.Dispatch", "failed")
	AssertFalse(t, r.OK)
	AssertErrorIs(t, r.Value.(error), AnError)
}

func TestCore_Core_LogError_Bad(t *T) {
	r := New().LogError(nil, "agent.Dispatch", "ok")
	AssertTrue(t, r.OK)
}

func TestCore_Core_LogError_Ugly(t *T) {
	r := New().LogError(NewCode("agent.refused", "dispatch refused"), "agent.Dispatch", "failed")
	AssertFalse(t, r.OK)
	AssertEqual(t, "agent.refused", ErrorCode(r.Value.(error)))
}

func TestCore_Core_LogWarn_Good(t *T) {
	r := New().LogWarn(AnError, "agent.Dispatch", "degraded")
	AssertFalse(t, r.OK)
	AssertErrorIs(t, r.Value.(error), AnError)
}

func TestCore_Core_LogWarn_Bad(t *T) {
	r := New().LogWarn(nil, "agent.Dispatch", "ok")
	AssertTrue(t, r.OK)
}

func TestCore_Core_LogWarn_Ugly(t *T) {
	r := New().LogWarn(NewCode("agent.degraded", "dispatch degraded"), "agent.Dispatch", "degraded")
	AssertFalse(t, r.OK)
	AssertEqual(t, "agent.degraded", ErrorCode(r.Value.(error)))
}

func TestCore_Core_Must_Good(t *T) {
	AssertNotPanics(t, func() {
		New().Must(nil, "agent.Dispatch", "ok")
	})
}

func TestCore_Core_Must_Bad(t *T) {
	AssertPanicsWithError(t, "dispatch failed", func() {
		New().Must(AnError, "agent.Dispatch", "dispatch failed")
	})
}

func TestCore_Core_Must_Ugly(t *T) {
	AssertPanicsWithError(t, "agent.refused", func() {
		New().Must(NewCode("agent.refused", "dispatch refused"), "agent.Dispatch", "failed")
	})
}

func TestCore_Core_Options_Good(t *T) {
	c := New(WithOption("name", "agent"))
	AssertEqual(t, "agent", c.Options().String("name"))
}

func TestCore_Core_Options_Bad(t *T) {
	c := New()
	AssertNil(t, c.Options())
}

func TestCore_Core_Options_Ugly(t *T) {
	c := New(WithOptions(NewOptions()))
	AssertEqual(t, 0, c.Options().Len())
}

func TestCore_Core_QUERY_Good(t *T) {
	c := New()
	c.RegisterQuery(func(_ *Core, q Query) Result {
		return Result{Value: Sprintf("ack:%v", q), OK: true}
	})

	r := c.QUERY("agent.status")

	AssertTrue(t, r.OK)
	AssertEqual(t, "ack:agent.status", r.Value)
}

func TestCore_Core_QUERY_Bad(t *T) {
	r := New().QUERY("missing")
	AssertFalse(t, r.OK)
}

func TestCore_Core_QUERY_Ugly(t *T) {
	c := New()
	c.RegisterQuery(func(_ *Core, _ Query) Result {
		return Result{Value: nil, OK: true}
	})

	r := c.QUERY(nil)

	AssertTrue(t, r.OK)
	AssertNil(t, r.Value)
}

func TestCore_Core_QUERYALL_Good(t *T) {
	c := New()
	c.RegisterQuery(func(_ *Core, _ Query) Result { return Result{Value: "agent", OK: true} })
	c.RegisterQuery(func(_ *Core, _ Query) Result { return Result{Value: "health", OK: true} })

	r := c.QUERYALL("status")

	AssertTrue(t, r.OK)
	AssertElementsMatch(t, []any{"agent", "health"}, r.Value.([]any))
}

func TestCore_Core_QUERYALL_Bad(t *T) {
	c := New()
	c.RegisterQuery(func(_ *Core, _ Query) Result { return Result{Value: "ignored", OK: false} })

	r := c.QUERYALL("status")

	AssertTrue(t, r.OK)
	AssertEmpty(t, r.Value.([]any))
}

func TestCore_Core_QUERYALL_Ugly(t *T) {
	c := New()
	c.RegisterQuery(func(_ *Core, _ Query) Result { return Result{Value: nil, OK: true} })

	r := c.QUERYALL(nil)

	AssertTrue(t, r.OK)
	AssertEmpty(t, r.Value.([]any))
}

func TestCore_Core_RegistryOf_Good(t *T) {
	c := New()
	c.Action("agent.dispatch", func(_ Context, _ Options) Result { return Result{OK: true} })

	reg := c.RegistryOf("actions")

	AssertTrue(t, reg.Has("agent.dispatch"))
}

func TestCore_Core_RegistryOf_Bad(t *T) {
	reg := New().RegistryOf("missing")
	AssertEqual(t, 0, reg.Len())
}

func TestCore_Core_RegistryOf_Ugly(t *T) {
	c := New()
	reg := c.RegistryOf("actions")
	c.Action("agent.dispatch", func(_ Context, _ Options) Result { return Result{OK: true} })

	AssertFalse(t, reg.Has("agent.dispatch"))
}

func TestCore_Core_Run_Good(t *T) {
	c := New()
	c.Cli().SetOutput(NewBuffer())
	AssertNotPanics(t, func() {
		c.Run()
	})
}

func TestCore_Core_Run_Bad(t *T) {
	c := New(WithService(func(c *Core) Result {
		return c.Service("agent", Service{OnStart: func() Result { return Result{OK: true} }})
	}))
	c.Cli().SetOutput(NewBuffer())
	AssertNotPanics(t, func() {
		c.Run()
	})
}

func TestCore_Core_Run_Ugly(t *T) {
	c := New()
	c.Cli().SetOutput(NewBuffer())
	AssertNotPanics(t, func() {
		c.Run()
	})
}

func TestCore_Core_RunE_Good(t *T) {
	c := New()
	c.Cli().SetOutput(NewBuffer())
	AssertNoError(t, c.RunE())
}

func TestCore_Core_RunE_Bad(t *T) {
	c := New(WithService(func(c *Core) Result {
		return c.Service("agent", Service{
			OnStart: func() Result { return Result{Value: NewError("startup refused"), OK: false} },
		})
	}))

	err := c.RunE()

	AssertError(t, err, "startup refused")
}

func TestCore_Core_RunE_Ugly(t *T) {
	stopped := false
	c := New(
		WithService(func(c *Core) Result {
			return c.Service("cleanup", Service{
				OnStart: func() Result { return Result{OK: true} },
				OnStop:  func() Result { stopped = true; return Result{OK: true} },
			})
		}),
		WithService(func(c *Core) Result {
			return c.Service("broken", Service{
				OnStart: func() Result { return Result{Value: NewError("boom"), OK: false} },
			})
		}),
	)

	err := c.RunE()

	AssertError(t, err, "boom")
	AssertTrue(t, stopped)
}
