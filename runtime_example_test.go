package core_test

import . "dappco.re/go"

type runtimeOptions struct {
	Name string
}

// ExampleNewServiceRuntime constructs a service runtime through `NewServiceRuntime` for
// service runtime lifecycle. Service runtime setup joins Core, Config, Options, and
// factories in one lifecycle path.
func ExampleNewServiceRuntime() {
	c := New()
	rt := NewServiceRuntime(c, runtimeOptions{Name: "worker"})

	Println(rt.Core() == c)
	Println(rt.Options().Name)
	Println(rt.Config() != nil)
	// Output:
	// true
	// worker
	// true
}

// ExampleServiceRuntime_Core returns the Core instance through `ServiceRuntime.Core` for
// service runtime lifecycle. Service runtime setup joins Core, Config, Options, and
// factories in one lifecycle path.
func ExampleServiceRuntime_Core() {
	c := New()
	rt := NewServiceRuntime(c, runtimeOptions{})
	Println(rt.Core() == c)
	// Output: true
}

// ExampleServiceRuntime_Options reads runtime options through `ServiceRuntime.Options` for
// service runtime lifecycle. Service runtime setup joins Core, Config, Options, and
// factories in one lifecycle path.
func ExampleServiceRuntime_Options() {
	rt := NewServiceRuntime(New(), runtimeOptions{Name: "worker"})
	Println(rt.Options().Name)
	// Output: worker
}

// ExampleServiceRuntime_Config reads runtime configuration through `ServiceRuntime.Config`
// for service runtime lifecycle. Service runtime setup joins Core, Config, Options, and
// factories in one lifecycle path.
func ExampleServiceRuntime_Config() {
	c := New()
	c.Config().Set("host", "localhost")
	rt := NewServiceRuntime(c, runtimeOptions{})
	Println(rt.Config().String("host"))
	// Output: localhost
}

// ExampleCore_ServiceStartup runs service startup through `Core.ServiceStartup` for
// service runtime lifecycle. Service runtime setup joins Core, Config, Options, and
// factories in one lifecycle path.
func ExampleCore_ServiceStartup() {
	started := false
	c := New()
	c.Service("worker", Service{OnStart: func() Result {
		started = true
		return Result{OK: true}
	}})

	r := c.ServiceStartup(Background(), nil)
	Println(r.OK)
	Println(started)
	c.ServiceShutdown(Background())
	// Output:
	// true
	// true
}

// ExampleCore_ServiceShutdown runs service shutdown through `Core.ServiceShutdown` for
// service runtime lifecycle. Service runtime setup joins Core, Config, Options, and
// factories in one lifecycle path.
func ExampleCore_ServiceShutdown() {
	stopped := false
	c := New()
	c.Service("worker", Service{OnStop: func() Result {
		stopped = true
		return Result{OK: true}
	}})

	r := c.ServiceShutdown(Background())
	Println(r.OK)
	Println(stopped)
	// Output:
	// true
	// true
}

// ExampleServiceFactory declares a service factory through `ServiceFactory` for service
// runtime lifecycle. Service runtime setup joins Core, Config, Options, and factories in
// one lifecycle path.
func ExampleServiceFactory() {
	var factory ServiceFactory = func() Result {
		return Result{Value: Service{OnStart: func() Result { return Result{OK: true} }}, OK: true}
	}
	Println(factory().OK)
	// Output: true
}

// ExampleNewWithFactories constructs Core with service factories through
// `NewWithFactories` for service runtime lifecycle. Service runtime setup joins Core,
// Config, Options, and factories in one lifecycle path.
func ExampleNewWithFactories() {
	r := NewWithFactories("gui", map[string]ServiceFactory{
		"beta":  func() Result { return Result{Value: Service{}, OK: true} },
		"alpha": func() Result { return Result{Value: Service{}, OK: true} },
	})
	rt := r.Value.(*Runtime)

	Println(r.OK)
	Println(rt.Core.App().Runtime)
	Println(rt.Core.Services())
	// Output:
	// true
	// gui
	// [cli alpha beta]
}

// ExampleNewRuntime constructs a runtime through `NewRuntime` for service runtime
// lifecycle. Service runtime setup joins Core, Config, Options, and factories in one
// lifecycle path.
func ExampleNewRuntime() {
	r := NewRuntime("gui")
	rt := r.Value.(*Runtime)
	Println(r.OK)
	Println(rt.ServiceName())
	// Output:
	// true
	// Core
}

// ExampleRuntime_ServiceName reads a service name through `Runtime.ServiceName` for
// service runtime lifecycle. Service runtime setup joins Core, Config, Options, and
// factories in one lifecycle path.
func ExampleRuntime_ServiceName() {
	rt := NewRuntime("gui").Value.(*Runtime)
	Println(rt.ServiceName())
	// Output: Core
}

// ExampleRuntime_ServiceStartup runs service startup through `Runtime.ServiceStartup` for
// service runtime lifecycle. Service runtime setup joins Core, Config, Options, and
// factories in one lifecycle path.
func ExampleRuntime_ServiceStartup() {
	rt := NewRuntime("gui").Value.(*Runtime)
	Println(rt.ServiceStartup(Background(), nil).OK)
	rt.ServiceShutdown(Background())
	// Output: true
}

// ExampleRuntime_ServiceShutdown runs service shutdown through `Runtime.ServiceShutdown`
// for service runtime lifecycle. Service runtime setup joins Core, Config, Options, and
// factories in one lifecycle path.
func ExampleRuntime_ServiceShutdown() {
	rt := NewRuntime("gui").Value.(*Runtime)
	Println(rt.ServiceShutdown(Background()).OK)
	// Output: true
}
