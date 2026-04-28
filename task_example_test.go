package core_test

import (
	"context"

	. "dappco.re/go"
)

// ExampleTask_Run runs `Task.Run` with representative caller inputs for background task
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
