package core_test

import (
	"context"

	. "dappco.re/go/core"
)

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
