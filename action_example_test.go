package core_test

import (
	"context"

	. "dappco.re/go"
)

// ExampleAction declares action metadata for an agent dispatch workflow. Consumers copy
// the Result-shaped handler contract for dAppCore actions and tasks.
func ExampleAction() {
	a := &Action{Name: "deploy", Description: "Deploy service"}
	Println(a.Name)
	Println(a.Description)
	// Output:
	// deploy
	// Deploy service
}

// ExampleActionHandler declares an action handler through `ActionHandler` for an agent
// dispatch workflow. Consumers copy the Result-shaped handler contract for dAppCore
// actions and tasks.
func ExampleActionHandler() {
	var handler ActionHandler = func(_ context.Context, opts Options) Result {
		return Result{Value: opts.String("name"), OK: true}
	}
	Println(handler(context.Background(), NewOptions(Option{Key: "name", Value: "deploy"})).Value)
	// Output: deploy
}

// ExampleAction_Run runs `Action.Run` with representative caller inputs for an agent
// dispatch workflow. Consumers copy the Result-shaped handler contract for dAppCore
// actions and tasks.
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

// ExampleCore_Action registers or retrieves an action through `Core.Action` for an agent
// dispatch workflow. Consumers copy the Result-shaped handler contract for dAppCore
// actions and tasks.
func ExampleCore_Action() {
	c := New()
	c.Action("deploy", func(_ context.Context, _ Options) Result { return Result{OK: true} })
	Println(c.Action("deploy").Exists())
	// Output: true
}

// ExampleCore_Actions_action registers the action-oriented path through `Core.Actions` for
// an agent dispatch workflow. Consumers copy the Result-shaped handler contract for
// dAppCore actions and tasks.
func ExampleCore_Actions_action() {
	c := New()
	c.Action("deploy", func(_ context.Context, _ Options) Result { return Result{OK: true} })
	c.Action("test", func(_ context.Context, _ Options) Result { return Result{OK: true} })
	Println(c.Actions())
	// Output: [deploy test]
}

// ExampleStep declares one task step through `Step` for an agent dispatch workflow.
// Consumers copy the Result-shaped handler contract for dAppCore actions and tasks.
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

// ExampleTask declares a named task that points at a deployment planning step. Consumers
// copy the Result-shaped handler contract for dAppCore actions and tasks.
func ExampleTask() {
	task := Task{Name: "deploy", Steps: []Step{{Action: "deploy.plan"}}}
	Println(task.Name)
	Println(task.Steps[0].Action)
	// Output:
	// deploy
	// deploy.plan
}

// ExampleCore_Task_action registers the action-oriented path through `Core.Task` for an
// agent dispatch workflow. Consumers copy the Result-shaped handler contract for dAppCore
// actions and tasks.
func ExampleCore_Task_action() {
	c := New()
	c.Task("deploy", Task{Steps: []Step{{Action: "deploy.plan"}}})
	Println(c.Tasks())
	// Output: [deploy]
}

// ExampleCore_Tasks lists task names through `Core.Tasks` for an agent dispatch workflow.
// Consumers copy the Result-shaped handler contract for dAppCore actions and tasks.
func ExampleCore_Tasks() {
	c := New()
	c.Task("deploy", Task{})
	c.Task("test", Task{})
	Println(c.Tasks())
	// Output: [deploy test]
}

// ExampleAction_Exists checks action availability before and after registering a handler.
// Consumers copy the Result-shaped handler contract for dAppCore actions and tasks.
func ExampleAction_Exists() {
	c := New()
	Println(c.Action("missing").Exists())

	c.Action("present", func(_ context.Context, _ Options) Result { return Result{OK: true} })
	Println(c.Action("present").Exists())
	// Output:
	// false
	// true
}

// ExampleAction_Run_panicRecovery recovers a panicking handler through `Action.Run` for an
// agent dispatch workflow. Consumers copy the Result-shaped handler contract for dAppCore
// actions and tasks.
func ExampleAction_Run_panicRecovery() {
	c := New()
	c.Action("boom", func(_ context.Context, _ Options) Result {
		panic("explosion")
	})

	r := c.Action("boom").Run(context.Background(), NewOptions())
	Println(r.OK)
	// Output: false
}

// ExampleAction_Run_entitlementDenied rejects a gated action through `Action.Run` when
// entitlement denies access for an agent dispatch workflow. Consumers copy the
// Result-shaped handler contract for dAppCore actions and tasks.
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

// ExampleAction_Task_Run runs `Task.Run` with representative caller inputs for background task
// progress. Asynchronous work reports progress through Core task helpers.
func ExampleTask_Run() {
	c := New()
	var order string

	c.Action("step.a", func(_ context.Context, _ Options) Result {
		order += "a"
		return Result{Value: "from-a", OK: true}
	})
	c.Action("step.b", func(_ context.Context, opts Options) Result {
		order += "b"
		input := opts.Get("_input")
		if input.OK {
			return Result{Value: "got:" + input.Value.(string), OK: true}
		}
		return Result{OK: true}
	})

	c.Task("pipe", Task{
		Steps: []Step{
			{Action: "step.a"},
			{Action: "step.b", Input: "previous"},
		},
	})

	r := c.Task("pipe").Run(context.Background(), c, NewOptions())
	Println(order)
	Println(r.Value)
	// Output:
	// ab
	// got:from-a
}

// ExampleCore_PerformAsync starts asynchronous work through `Core.PerformAsync` for
// background task progress. Asynchronous work reports progress through Core task helpers.
func ExampleCore_PerformAsync() {
	c := New()
	c.Action("bg.work", func(_ context.Context, _ Options) Result {
		return Result{Value: "done", OK: true}
	})

	r := c.PerformAsync("bg.work", NewOptions())
	Println(HasPrefix(r.Value.(string), "id-"))
	// Output: true
}

// ExampleCore_Progress reports task progress through `Core.Progress` for background task
// progress. Asynchronous work reports progress through Core task helpers.
func ExampleCore_Progress() {
	c := New()
	var progress float64
	var message string
	c.RegisterAction(func(_ *Core, msg Message) Result {
		if ev, ok := msg.(ActionTaskProgress); ok {
			progress = ev.Progress
			message = ev.Message
		}
		return Result{OK: true}
	})

	c.Progress("task-1", 0.5, "halfway", "deploy")
	Println(progress)
	Println(message)
	// Output:
	// 0.5
	// halfway
}
