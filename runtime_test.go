package core_test

import (
	"context"

	. "dappco.re/go/core"
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

	result := rt.ServiceStartup(context.Background(), nil)
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

	rt.ServiceStartup(context.Background(), nil)
	result := rt.ServiceShutdown(context.Background())
	AssertTrue(t, result.OK)
	AssertTrue(t, stopped)
}

func TestRuntime_ServiceShutdown_NilCore_Good(t *T) {
	rt := &Runtime{}
	result := rt.ServiceShutdown(context.Background())
	AssertTrue(t, result.OK)
}

func TestCore_ServiceShutdown_Good(t *T) {
	stopped := false
	c := New()
	c.Service("test", Service{
		OnStart: func() Result { return Result{OK: true} },
		OnStop:  func() Result { stopped = true; return Result{OK: true} },
	})
	c.ServiceStartup(context.Background(), nil)
	result := c.ServiceShutdown(context.Background())
	AssertTrue(t, result.OK)
	AssertTrue(t, stopped)
}

func TestCore_Context_Good(t *T) {
	c := New()
	c.ServiceStartup(context.Background(), nil)
	AssertNotNil(t, c.Context())
	c.ServiceShutdown(context.Background())
}
