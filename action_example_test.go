package core_test

import (
	"context"

	. "dappco.re/go"
)

func ExampleAction() {
	a := &Action{Name: "deploy", Description: "Deploy service"}
	Println(a.Name)
	Println(a.Description)
	// Output:
	// deploy
	// Deploy service
}

func ExampleActionHandler() {
	var handler ActionHandler = func(_ context.Context, opts Options) Result {
		return Result{Value: opts.String("name"), OK: true}
	}
	Println(handler(context.Background(), NewOptions(Option{Key: "name", Value: "deploy"})).Value)
	// Output: deploy
}

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

func ExampleCore_Action() {
	c := New()
	c.Action("deploy", func(_ context.Context, _ Options) Result { return Result{OK: true} })
	Println(c.Action("deploy").Exists())
	// Output: true
}

func ExampleCore_Actions_action() {
	c := New()
	c.Action("deploy", func(_ context.Context, _ Options) Result { return Result{OK: true} })
	c.Action("test", func(_ context.Context, _ Options) Result { return Result{OK: true} })
	Println(c.Actions())
	// Output: [deploy test]
}

func ExampleStep() {
	step := Step{
		Action: "deploy",
		With:   NewOptions(Option{Key: "target", Value: "homelab"}),
		Input:  "previous",
	}
	Println(step.Action)
	Println(step.With.String("target"))
	Println(step.Input)
	// Output:
	// deploy
	// homelab
	// previous
}

func ExampleTask() {
	task := Task{Name: "deploy", Steps: []Step{{Action: "deploy.plan"}}}
	Println(task.Name)
	Println(task.Steps[0].Action)
	// Output:
	// deploy
	// deploy.plan
}

func ExampleCore_Task_action() {
	c := New()
	c.Task("deploy", Task{Steps: []Step{{Action: "deploy.plan"}}})
	Println(c.Tasks())
	// Output: [deploy]
}

func ExampleCore_Tasks() {
	c := New()
	c.Task("deploy", Task{})
	c.Task("test", Task{})
	Println(c.Tasks())
	// Output: [deploy test]
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
