package core_test

import (
	"context"

	. "dappco.re/go"
)

func ExampleCore_Process_accessor() {
	c := New()
	Println(c.Process() != nil)
	// Output: true
}

func ExampleProcess_Run() {
	c := New()
	c.Action("process.run", func(_ context.Context, opts Options) Result {
		return Result{Value: Join(" ", append([]string{opts.String("command")}, opts.Get("args").Value.([]string)...)...), OK: true}
	})

	r := c.Process().Run(c.Context(), "go", "test", "./...")
	Println(r.Value)
	// Output: go test ./...
}

func ExampleProcess_RunIn() {
	c := New()
	c.Action("process.run", func(_ context.Context, opts Options) Result {
		return Result{Value: opts.String("dir"), OK: true}
	})

	r := c.Process().RunIn(c.Context(), "/repo", "go", "test")
	Println(r.Value)
	// Output: /repo
}

func ExampleProcess_RunWithEnv() {
	c := New()
	c.Action("process.run", func(_ context.Context, opts Options) Result {
		return Result{Value: opts.Get("env").Value.([]string)[0], OK: true}
	})

	r := c.Process().RunWithEnv(c.Context(), "/repo", []string{"GOWORK=off"}, "go", "test")
	Println(r.Value)
	// Output: GOWORK=off
}

func ExampleProcess_Start() {
	c := New()
	c.Action("process.start", func(_ context.Context, opts Options) Result {
		return Result{Value: opts.String("id"), OK: true}
	})

	r := c.Process().Start(c.Context(), NewOptions(Option{Key: "id", Value: "worker"}))
	Println(r.Value)
	// Output: worker
}

func ExampleProcess_Kill() {
	c := New()
	c.Action("process.kill", func(_ context.Context, opts Options) Result {
		return Result{Value: Concat("stopped:", opts.String("id")), OK: true}
	})

	r := c.Process().Kill(c.Context(), NewOptions(Option{Key: "id", Value: "worker"}))
	Println(r.Value)
	// Output: stopped:worker
}

func ExampleProcess_Exists() {
	c := New()
	Println(c.Process().Exists())
	c.Action("process.run", func(_ context.Context, _ Options) Result { return Result{OK: true} })
	Println(c.Process().Exists())
	// Output:
	// false
	// true
}
