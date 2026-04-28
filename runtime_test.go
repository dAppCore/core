package core_test

import (
	. "dappco.re/go"
)

// --- ServiceRuntime ---

type testOpts struct {
	URL     string
	Timeout int
}

func TestRuntime_ServiceRuntime_Good(t *T) {
	c := New()
	opts := testOpts{URL: "https://api.lthn.ai", Timeout: 30}
	rt := NewServiceRuntime(c, opts)

	AssertEqual(t, c, rt.Core())
	AssertEqual(t, opts, rt.Options())
	AssertEqual(t, "https://api.lthn.ai", rt.Options().URL)
	AssertNotNil(t, rt.Config())
}

// --- NewWithFactories ---

func TestRuntime_NewWithFactories_Good(t *T) {
	r := NewWithFactories(nil, map[string]ServiceFactory{
		"svc1": func() Result { return Result{Value: Service{}, OK: true} },
		"svc2": func() Result { return Result{Value: Service{}, OK: true} },
	})
	AssertTrue(t, r.OK)
	rt := r.Value.(*Runtime)
	AssertNotNil(t, rt.Core)
}

func TestRuntime_NewWithFactories_NilFactory_Good(t *T) {
	r := NewWithFactories(nil, map[string]ServiceFactory{
		"bad": nil,
	})
	AssertTrue(t, r.OK) // nil factories skipped
}

func TestRuntime_NewRuntime_Good(t *T) {
	r := NewRuntime(nil)
	AssertTrue(t, r.OK)
}

func TestRuntime_ServiceName_Good(t *T) {
	r := NewRuntime(nil)
	rt := r.Value.(*Runtime)
	AssertEqual(t, "Core", rt.ServiceName())
}

// --- Lifecycle via Runtime ---

func TestRuntime_Lifecycle_Good(t *T) {
	started := false
	r := NewWithFactories(nil, map[string]ServiceFactory{
		"test": func() Result {
			return Result{Value: Service{
				OnStart: func() Result { started = true; return Result{OK: true} },
			}, OK: true}
		},
	})
	AssertTrue(t, r.OK)
	rt := r.Value.(*Runtime)

	result := rt.ServiceStartup(Background(), nil)
	AssertTrue(t, result.OK)
	AssertTrue(t, started)
}

func TestRuntime_ServiceShutdown_Good(t *T) {
	stopped := false
	r := NewWithFactories(nil, map[string]ServiceFactory{
		"test": func() Result {
			return Result{Value: Service{
				OnStart: func() Result { return Result{OK: true} },
				OnStop:  func() Result { stopped = true; return Result{OK: true} },
			}, OK: true}
		},
	})
	AssertTrue(t, r.OK)
	rt := r.Value.(*Runtime)

	rt.ServiceStartup(Background(), nil)
	result := rt.ServiceShutdown(Background())
	AssertTrue(t, result.OK)
	AssertTrue(t, stopped)
}

func TestRuntime_ServiceShutdown_NilCore_Good(t *T) {
	rt := &Runtime{}
	result := rt.ServiceShutdown(Background())
	AssertTrue(t, result.OK)
}

func TestCore_ServiceShutdown_Good(t *T) {
	stopped := false
	c := New()
	c.Service("test", Service{
		OnStart: func() Result { return Result{OK: true} },
		OnStop:  func() Result { stopped = true; return Result{OK: true} },
	})
	c.ServiceStartup(Background(), nil)
	result := c.ServiceShutdown(Background())
	AssertTrue(t, result.OK)
	AssertTrue(t, stopped)
}

func TestCore_Context_Good(t *T) {
	c := New()
	c.ServiceStartup(Background(), nil)
	AssertNotNil(t, c.Context())
	c.ServiceShutdown(Background())
}

func TestRuntime_Core_ServiceShutdown_Good(t *T) {
	stopped := false
	c := New()
	c.Service("agent", Service{
		OnStart: func() Result { return Result{OK: true} },
		OnStop:  func() Result { stopped = true; return Result{OK: true} },
	})
	AssertTrue(t, c.ServiceStartup(Background(), nil).OK)

	r := c.ServiceShutdown(Background())

	AssertTrue(t, r.OK)
	AssertTrue(t, stopped)
}

func TestRuntime_Core_ServiceShutdown_Bad(t *T) {
	c := New()
	c.Service("agent", Service{
		OnStop: func() Result { return Result{Value: NewError("shutdown refused"), OK: false} },
	})

	r := c.ServiceShutdown(Background())

	AssertFalse(t, r.OK)
	AssertError(t, r.Value.(error), "shutdown refused")
}

func TestRuntime_Core_ServiceShutdown_Ugly(t *T) {
	c := New()
	c.Service("agent", Service{
		OnStop: func() Result { return Result{OK: true} },
	})
	ctx, cancel := WithCancel(Background())
	cancel()

	r := c.ServiceShutdown(ctx)

	AssertFalse(t, r.OK)
	AssertError(t, r.Value.(error))
}

func TestRuntime_Core_ServiceStartup_Good(t *T) {
	started := false
	c := New()
	c.Service("agent", Service{OnStart: func() Result { started = true; return Result{OK: true} }})

	r := c.ServiceStartup(Background(), nil)

	AssertTrue(t, r.OK)
	AssertTrue(t, started)
}

func TestRuntime_Core_ServiceStartup_Bad(t *T) {
	c := New()
	c.Service("agent", Service{OnStart: func() Result { return Result{OK: true} }})
	ctx, cancel := WithCancel(Background())
	cancel()

	r := c.ServiceStartup(ctx, nil)

	AssertFalse(t, r.OK)
	AssertError(t, r.Value.(error))
}

func TestRuntime_Core_ServiceStartup_Ugly(t *T) {
	c := New()
	var sawStartup bool
	c.RegisterAction(func(_ *Core, msg Message) Result {
		_, sawStartup = msg.(ActionServiceStartup)
		return Result{OK: true}
	})

	r := c.ServiceStartup(Background(), nil)

	AssertTrue(t, r.OK)
	AssertTrue(t, sawStartup)
}

func TestRuntime_ServiceRuntime_Config_Good(t *T) {
	c := New()
	rt := NewServiceRuntime(c, testOpts{})

	AssertSame(t, c.Config(), rt.Config())
}

func TestRuntime_ServiceRuntime_Config_Bad(t *T) {
	rt := NewServiceRuntime[testOpts](nil, testOpts{})
	AssertPanics(t, func() {
		_ = rt.Config()
	})
}

func TestRuntime_ServiceRuntime_Config_Ugly(t *T) {
	c := New()
	c.Config().Set("agent.region", "lab")
	rt := NewServiceRuntime(c, testOpts{})

	AssertEqual(t, "lab", rt.Config().String("agent.region"))
}

func TestRuntime_ServiceRuntime_Core_Good(t *T) {
	c := New()
	rt := NewServiceRuntime(c, testOpts{})
	AssertSame(t, c, rt.Core())
}

func TestRuntime_ServiceRuntime_Core_Bad(t *T) {
	rt := NewServiceRuntime[testOpts](nil, testOpts{})
	AssertNil(t, rt.Core())
}

func TestRuntime_ServiceRuntime_Core_Ugly(t *T) {
	c := New()
	rt := NewServiceRuntime(c, testOpts{})
	c.ServiceShutdown(Background())
	AssertSame(t, c, rt.Core())
}

func TestRuntime_ServiceRuntime_Options_Good(t *T) {
	opts := testOpts{URL: "https://api.lthn.ai", Timeout: 30}
	rt := NewServiceRuntime(New(), opts)
	AssertEqual(t, opts, rt.Options())
}

func TestRuntime_ServiceRuntime_Options_Bad(t *T) {
	rt := NewServiceRuntime(New(), testOpts{})
	AssertEqual(t, testOpts{}, rt.Options())
}

func TestRuntime_ServiceRuntime_Options_Ugly(t *T) {
	type queueOptions struct {
		Topics []string
	}
	opts := queueOptions{Topics: []string{}}
	rt := NewServiceRuntime(New(), opts)
	AssertEqual(t, opts, rt.Options())
}

func TestRuntime_NewRuntime_Bad(t *T) {
	r := NewRuntime("gui-runtime")
	AssertTrue(t, r.OK)
	AssertNotNil(t, r.Value.(*Runtime).Core)
}

func TestRuntime_NewRuntime_Ugly(t *T) {
	app := struct{ Name string }{Name: "agent-ui"}
	r := NewRuntime(app)
	AssertTrue(t, r.OK)
	AssertEqual(t, "core", r.Value.(*Runtime).Core.App().Name)
}

func TestRuntime_NewServiceRuntime_Good(t *T) {
	c := New()
	rt := NewServiceRuntime(c, testOpts{URL: "https://api.lthn.ai", Timeout: 30})
	AssertSame(t, c, rt.Core())
	AssertEqual(t, 30, rt.Options().Timeout)
}

func TestRuntime_NewServiceRuntime_Bad(t *T) {
	rt := NewServiceRuntime[testOpts](nil, testOpts{URL: "offline"})
	AssertNil(t, rt.Core())
	AssertEqual(t, "offline", rt.Options().URL)
}

func TestRuntime_NewServiceRuntime_Ugly(t *T) {
	rt := NewServiceRuntime(New(), testOpts{})
	AssertEqual(t, testOpts{}, rt.Options())
}

func TestRuntime_NewWithFactories_Bad(t *T) {
	r := NewWithFactories(nil, map[string]ServiceFactory{
		"agent": func() Result { return Result{Value: NewError("factory refused"), OK: false} },
	})
	AssertFalse(t, r.OK)
	AssertError(t, r.Value.(error), "factory \"agent\" failed")
}

func TestRuntime_NewWithFactories_Ugly(t *T) {
	r := NewWithFactories(nil, map[string]ServiceFactory{
		"agent": func() Result { return Result{Value: "not a service", OK: true} },
	})
	AssertFalse(t, r.OK)
	AssertError(t, r.Value.(error), "non-Service")
}

func TestRuntime_Runtime_ServiceName_Good(t *T) {
	AssertEqual(t, "Core", (&Runtime{}).ServiceName())
}

func TestRuntime_Runtime_ServiceName_Bad(t *T) {
	var r *Runtime
	AssertEqual(t, "Core", r.ServiceName())
}

func TestRuntime_Runtime_ServiceName_Ugly(t *T) {
	r := &Runtime{Core: New()}
	AssertEqual(t, "Core", r.ServiceName())
}

func TestRuntime_Runtime_ServiceShutdown_Good(t *T) {
	stopped := false
	rt := &Runtime{Core: New()}
	rt.Core.Service("agent", Service{OnStop: func() Result { stopped = true; return Result{OK: true} }})

	r := rt.ServiceShutdown(Background())

	AssertTrue(t, r.OK)
	AssertTrue(t, stopped)
}

func TestRuntime_Runtime_ServiceShutdown_Bad(t *T) {
	rt := &Runtime{Core: New()}
	rt.Core.Service("agent", Service{OnStop: func() Result { return Result{Value: NewError("stop refused"), OK: false} }})

	r := rt.ServiceShutdown(Background())

	AssertFalse(t, r.OK)
	AssertError(t, r.Value.(error), "stop refused")
}

func TestRuntime_Runtime_ServiceShutdown_Ugly(t *T) {
	rt := &Runtime{}
	r := rt.ServiceShutdown(Background())
	AssertTrue(t, r.OK)
}

func TestRuntime_Runtime_ServiceStartup_Good(t *T) {
	started := false
	rt := &Runtime{Core: New()}
	rt.Core.Service("agent", Service{OnStart: func() Result { started = true; return Result{OK: true} }})

	r := rt.ServiceStartup(Background(), nil)

	AssertTrue(t, r.OK)
	AssertTrue(t, started)
}

func TestRuntime_Runtime_ServiceStartup_Bad(t *T) {
	rt := &Runtime{Core: New()}
	rt.Core.Service("agent", Service{OnStart: func() Result { return Result{Value: NewError("start refused"), OK: false} }})

	r := rt.ServiceStartup(Background(), nil)

	AssertFalse(t, r.OK)
	AssertError(t, r.Value.(error), "start refused")
}

func TestRuntime_Runtime_ServiceStartup_Ugly(t *T) {
	rt := &Runtime{}
	AssertPanics(t, func() {
		_ = rt.ServiceStartup(Background(), nil)
	})
}
