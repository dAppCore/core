package core_test

import (
	"context"

	. "dappco.re/go"
)

type exampleRegisteredService struct {
	name string
}

func ExampleService() {
	svc := Service{Name: "cache", Options: NewOptions(Option{Key: "size", Value: 128})}
	Println(svc.Name)
	Println(svc.Options.Int("size"))
	// Output:
	// cache
	// 128
}

func ExampleServiceRegistry() {
	registry := &ServiceRegistry{Registry: NewRegistry[*Service]()}
	registry.Set("cache", &Service{Name: "cache"})
	Println(registry.Names())
	// Output: [cache]
}

func ExampleCore_Service() {
	c := New()
	c.Service("cache", Service{})
	Println(c.Service("cache").OK)
	// Output: true
}

func ExampleCore_RegisterService() {
	c := New()
	r := c.RegisterService("worker", &exampleRegisteredService{name: "worker"})
	Println(r.OK)
	Println(c.Service("worker").Value.(*exampleRegisteredService).name)
	// Output:
	// true
	// worker
}

func ExampleServiceFor() {
	c := New(
		WithService(func(c *Core) Result {
			return c.Service("cache", Service{
				OnStart: func() Result { return Result{OK: true} },
			})
		}),
	)

	svc := c.Service("cache")
	Println(svc.OK)
	// Output: true
}

func ExampleMustServiceFor() {
	c := New()
	c.RegisterService("worker", &exampleRegisteredService{name: "worker"})
	svc := MustServiceFor[*exampleRegisteredService](c, "worker")
	Println(svc.name)
	// Output: worker
}

func ExampleCore_Services() {
	c := New()
	c.Service("cache", Service{})
	c.Service("worker", Service{})
	Println(c.Services())
	// Output: [cli cache worker]
}

func ExampleWithService() {
	started := false
	c := New(
		WithService(func(c *Core) Result {
			return c.Service("worker", Service{
				OnStart: func() Result { started = true; return Result{OK: true} },
			})
		}),
	)
	c.ServiceStartup(context.Background(), nil)
	Println(started)
	c.ServiceShutdown(context.Background())
	// Output: true
}

func ExampleWithServiceLock() {
	c := New(
		WithService(func(c *Core) Result {
			return c.Service("allowed", Service{})
		}),
		WithServiceLock(),
	)

	// Can't register after lock
	r := c.Service("blocked", Service{})
	Println(r.OK)
	// Output: false
}
