package core_test

import (
	"context"

	. "dappco.re/go"
)

type runtimeOptions struct {
	Name string
}

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

func ExampleServiceRuntime_Core() {
	c := New()
	rt := NewServiceRuntime(c, runtimeOptions{})
	Println(rt.Core() == c)
	// Output: true
}

func ExampleServiceRuntime_Options() {
	rt := NewServiceRuntime(New(), runtimeOptions{Name: "worker"})
	Println(rt.Options().Name)
	// Output: worker
}

func ExampleServiceRuntime_Config() {
	c := New()
	c.Config().Set("host", "localhost")
	rt := NewServiceRuntime(c, runtimeOptions{})
	Println(rt.Config().String("host"))
	// Output: localhost
}

func ExampleCore_ServiceStartup() {
	started := false
	c := New()
	c.Service("worker", Service{OnStart: func() Result {
		started = true
		return Result{OK: true}
	}})

	r := c.ServiceStartup(context.Background(), nil)
	Println(r.OK)
	Println(started)
	c.ServiceShutdown(context.Background())
	// Output:
	// true
	// true
}

func ExampleCore_ServiceShutdown() {
	stopped := false
	c := New()
	c.Service("worker", Service{OnStop: func() Result {
		stopped = true
		return Result{OK: true}
	}})

	r := c.ServiceShutdown(context.Background())
	Println(r.OK)
	Println(stopped)
	// Output:
	// true
	// true
}

func ExampleServiceFactory() {
	var factory ServiceFactory = func() Result {
		return Result{Value: Service{OnStart: func() Result { return Result{OK: true} }}, OK: true}
	}
	Println(factory().OK)
	// Output: true
}

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

func ExampleNewRuntime() {
	r := NewRuntime("gui")
	rt := r.Value.(*Runtime)
	Println(r.OK)
	Println(rt.ServiceName())
	// Output:
	// true
	// Core
}

func ExampleRuntime_ServiceName() {
	rt := NewRuntime("gui").Value.(*Runtime)
	Println(rt.ServiceName())
	// Output: Core
}

func ExampleRuntime_ServiceStartup() {
	rt := NewRuntime("gui").Value.(*Runtime)
	Println(rt.ServiceStartup(context.Background(), nil).OK)
	rt.ServiceShutdown(context.Background())
	// Output: true
}

func ExampleRuntime_ServiceShutdown() {
	rt := NewRuntime("gui").Value.(*Runtime)
	Println(rt.ServiceShutdown(context.Background()).OK)
	// Output: true
}
