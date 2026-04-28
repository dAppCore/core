package core_test

import (
	"context"

	. "dappco.re/go"
)

type exampleRegisteredService struct {
	name string
}

// ExampleService retrieves a service through `Service` for service registration. Services
// register by name and can be recovered with typed helpers.
func ExampleService() {
	svc := Service{Name: "cache", Options: NewOptions(Option{Key: "size", Value: 128})}
	Println(svc.Name)
	Println(svc.Options.Int("size"))
	// Output:
	// cache
	// 128
}

// ExampleServiceRegistry declares a service registry through `ServiceRegistry` for service
// registration. Services register by name and can be recovered with typed helpers.
func ExampleServiceRegistry() {
	registry := &ServiceRegistry{Registry: NewRegistry[*Service]()}
	registry.Set("cache", &Service{Name: "cache"})
	Println(registry.Names())
	// Output: [cache]
}

// ExampleCore_Service retrieves a service through `Core.Service` for service registration.
// Services register by name and can be recovered with typed helpers.
func ExampleCore_Service() {
	c := New()
	c.Service("cache", Service{})
	Println(c.Service("cache").OK)
	// Output: true
}

// ExampleCore_RegisterService registers a service through `Core.RegisterService` for
// service registration. Services register by name and can be recovered with typed helpers.
func ExampleCore_RegisterService() {
	c := New()
	r := c.RegisterService("worker", &exampleRegisteredService{name: "worker"})
	Println(r.OK)
	Println(c.Service("worker").Value.(*exampleRegisteredService).name)
	// Output:
	// true
	// worker
}

// ExampleServiceFor retrieves a typed service through `ServiceFor` for service
// registration. Services register by name and can be recovered with typed helpers.
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

// ExampleMustServiceFor retrieves a required typed service through `MustServiceFor` for
// service registration. Services register by name and can be recovered with typed helpers.
func ExampleMustServiceFor() {
	c := New()
	c.RegisterService("worker", &exampleRegisteredService{name: "worker"})
	svc := MustServiceFor[*exampleRegisteredService](c, "worker")
	Println(svc.name)
	// Output: worker
}

// ExampleCore_Services lists registered services through `Core.Services` for service
// registration. Services register by name and can be recovered with typed helpers.
func ExampleCore_Services() {
	c := New()
	c.Service("cache", Service{})
	c.Service("worker", Service{})
	Println(c.Services())
	// Output: [cli cache worker]
}

// ExampleWithService injects a service through `WithService` for service registration.
// Services register by name and can be recovered with typed helpers.
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

// ExampleWithServiceLock injects service locking through `WithServiceLock` for service
// registration. Services register by name and can be recovered with typed helpers.
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
