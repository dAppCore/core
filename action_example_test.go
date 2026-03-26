package core_test

import (
	"context"

	. "dappco.re/go/core"
)

func ExampleAction_Run() {
	c := New()
	c.Action("double", func(_ context.Context, opts Options) Result {
		return Result{Value: opts.Int("n") * 2, OK: true}
	})

	r := c.Action("double").Run(context.Background(), NewOptions(
		Option{Key: "n", Value: 21},
	))
	Println(r.Value)
	// Output: 42
}

func ExampleAction_Exists() {
	c := New()
	Println(c.Action("missing").Exists())

	c.Action("present", func(_ context.Context, _ Options) Result { return Result{OK: true} })
	Println(c.Action("present").Exists())
	// Output:
	// false
	// true
}

func ExampleAction_Run_panicRecovery() {
	c := New()
	c.Action("boom", func(_ context.Context, _ Options) Result {
		panic("explosion")
	})

	r := c.Action("boom").Run(context.Background(), NewOptions())
	Println(r.OK)
	// Output: false
}

func ExampleAction_Run_entitlementDenied() {
	c := New()
	c.Action("premium", func(_ context.Context, _ Options) Result {
		return Result{Value: "secret", OK: true}
	})
	c.SetEntitlementChecker(func(action string, _ int, _ context.Context) Entitlement {
		if action == "premium" {
			return Entitlement{Allowed: false, Reason: "upgrade"}
		}
		return Entitlement{Allowed: true, Unlimited: true}
	})

	r := c.Action("premium").Run(context.Background(), NewOptions())
	Println(r.OK)
	// Output: false
}
