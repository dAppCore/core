package core_test

import (
	"context"
	"fmt"

	. "dappco.re/go/core"
)

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
	fmt.Println(order)
	fmt.Println(r.Value)
	// Output:
	// ab
	// got:from-a
}

func ExampleCore_PerformAsync() {
	c := New()
	c.Action("bg.work", func(_ context.Context, _ Options) Result {
		return Result{Value: "done", OK: true}
	})

	r := c.PerformAsync("bg.work", NewOptions())
	fmt.Println(HasPrefix(r.Value.(string), "id-"))
	// Output: true
}
