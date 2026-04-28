package core_test

import (
	"context"

	. "dappco.re/go"
)

// --- NamedAction Register ---

func TestAction_NamedAction_Good_Register(t *T) {
	c := New()
	def := c.Action("process.run", func(_ context.Context, opts Options) Result {
		return Result{Value: "output", OK: true}
	})
	AssertNotNil(t, def)
	AssertEqual(t, "process.run", def.Name)
	AssertTrue(t, def.Exists())
}

func TestAction_NamedAction_Good_Invoke(t *T) {
	c := New()
	c.Action("git.log", func(_ context.Context, opts Options) Result {
		dir := opts.String("dir")
		return Result{Value: Concat("log from ", dir), OK: true}
	})

	r := c.Action("git.log").Run(context.Background(), NewOptions(
		Option{Key: "dir", Value: "/repo"},
	))
	AssertTrue(t, r.OK)
	AssertEqual(t, "log from /repo", r.Value)
}

func TestAction_NamedAction_Bad_NotRegistered(t *T) {
	c := New()
	r := c.Action("missing.action").Run(context.Background(), NewOptions())
	AssertFalse(t, r.OK, "invoking unregistered action must fail")
}

func TestAction_NamedAction_Good_Exists(t *T) {
	c := New()
	c.Action("brain.recall", func(_ context.Context, _ Options) Result {
		return Result{OK: true}
	})
	AssertTrue(t, c.Action("brain.recall").Exists())
	AssertFalse(t, c.Action("brain.forget").Exists())
}

func TestAction_NamedAction_Ugly_PanicRecovery(t *T) {
	c := New()
	c.Action("explode", func(_ context.Context, _ Options) Result {
		panic("boom")
	})
	r := c.Action("explode").Run(context.Background(), NewOptions())
	AssertFalse(t, r.OK, "panicking action must return !OK, not crash")
	err, ok := r.Value.(error)
	AssertTrue(t, ok)
	AssertContains(t, err.Error(), "panic")
}

func TestAction_NamedAction_Ugly_NilAction(t *T) {
	var def *Action
	r := def.Run(context.Background(), NewOptions())
	AssertFalse(t, r.OK)
	AssertFalse(t, def.Exists())
}

// --- Actions listing ---

func TestAction_Actions_Good(t *T) {
	c := New()
	c.Action("process.run", func(_ context.Context, _ Options) Result { return Result{OK: true} })
	c.Action("process.kill", func(_ context.Context, _ Options) Result { return Result{OK: true} })
	c.Action("agentic.dispatch", func(_ context.Context, _ Options) Result { return Result{OK: true} })

	names := c.Actions()
	AssertLen(t, names, 3)
	AssertEqual(t, []string{"process.run", "process.kill", "agentic.dispatch"}, names)
}

func TestAction_Actions_Bad_Empty(t *T) {
	c := New()
	AssertEmpty(t, c.Actions())
}

// --- Action fields ---

func TestAction_NamedAction_Good_DescriptionAndSchema(t *T) {
	c := New()
	def := c.Action("process.run", func(_ context.Context, _ Options) Result { return Result{OK: true} })
	def.Description = "Execute a command synchronously"
	def.Schema = NewOptions(
		Option{Key: "command", Value: "string"},
		Option{Key: "args", Value: "[]string"},
	)

	retrieved := c.Action("process.run")
	AssertEqual(t, "Execute a command synchronously", retrieved.Description)
	AssertTrue(t, retrieved.Schema.Has("command"))
}

// --- Permission by registration ---

func TestAction_NamedAction_Good_PermissionModel(t *T) {
	// Full Core — process registered
	full := New()
	full.Action("process.run", func(_ context.Context, _ Options) Result {
		return Result{Value: "executed", OK: true}
	})

	// Sandboxed Core — no process
	sandboxed := New()

	// Full can execute
	r := full.Action("process.run").Run(context.Background(), NewOptions())
	AssertTrue(t, r.OK)

	// Sandboxed returns not-registered
	r = sandboxed.Action("process.run").Run(context.Background(), NewOptions())
	AssertFalse(t, r.OK, "sandboxed Core must not have process capability")
}

// --- Action overwrite ---

func TestAction_NamedAction_Good_Overwrite(t *T) {
	c := New()
	c.Action("hot.reload", func(_ context.Context, _ Options) Result {
		return Result{Value: "v1", OK: true}
	})
	c.Action("hot.reload", func(_ context.Context, _ Options) Result {
		return Result{Value: "v2", OK: true}
	})

	r := c.Action("hot.reload").Run(context.Background(), NewOptions())
	AssertTrue(t, r.OK)
	AssertEqual(t, "v2", r.Value, "latest handler wins")
}

// --- Task Composition ---

func TestAction_Task_Good_Sequential(t *T) {
	c := New()
	var order []string
	c.Action("step.a", func(_ context.Context, _ Options) Result {
		order = append(order, "a")
		return Result{Value: "output-a", OK: true}
	})
	c.Action("step.b", func(_ context.Context, _ Options) Result {
		order = append(order, "b")
		return Result{Value: "output-b", OK: true}
	})

	c.Task("pipeline", Task{
		Steps: []Step{
			{Action: "step.a"},
			{Action: "step.b"},
		},
	})

	r := c.Task("pipeline").Run(context.Background(), c, NewOptions())
	AssertTrue(t, r.OK)
	AssertEqual(t, []string{"a", "b"}, order, "steps must run in order")
	AssertEqual(t, "output-b", r.Value, "last step's result is returned")
}

func TestAction_Task_Bad_StepFails(t *T) {
	c := New()
	var order []string
	c.Action("step.ok", func(_ context.Context, _ Options) Result {
		order = append(order, "ok")
		return Result{OK: true}
	})
	c.Action("step.fail", func(_ context.Context, _ Options) Result {
		order = append(order, "fail")
		return Result{Value: NewError("broke"), OK: false}
	})
	c.Action("step.never", func(_ context.Context, _ Options) Result {
		order = append(order, "never")
		return Result{OK: true}
	})

	c.Task("broken", Task{
		Steps: []Step{
			{Action: "step.ok"},
			{Action: "step.fail"},
			{Action: "step.never"},
		},
	})

	r := c.Task("broken").Run(context.Background(), c, NewOptions())
	AssertFalse(t, r.OK)
	AssertEqual(t, []string{"ok", "fail"}, order, "chain stops on failure, step.never skipped")
}

func TestAction_Task_Bad_MissingAction(t *T) {
	c := New()
	c.Task("missing", Task{
		Steps: []Step{
			{Action: "nonexistent"},
		},
	})
	r := c.Task("missing").Run(context.Background(), c, NewOptions())
	AssertFalse(t, r.OK)
}

func TestAction_Task_Good_PreviousInput(t *T) {
	c := New()
	c.Action("produce", func(_ context.Context, _ Options) Result {
		return Result{Value: "data-from-step-1", OK: true}
	})
	c.Action("consume", func(_ context.Context, opts Options) Result {
		input := opts.Get("_input")
		if !input.OK {
			return Result{Value: "no input", OK: true}
		}
		return Result{Value: "got: " + input.Value.(string), OK: true}
	})

	c.Task("pipe", Task{
		Steps: []Step{
			{Action: "produce"},
			{Action: "consume", Input: "previous"},
		},
	})

	r := c.Task("pipe").Run(context.Background(), c, NewOptions())
	AssertTrue(t, r.OK)
	AssertEqual(t, "got: data-from-step-1", r.Value)
}

func TestAction_Task_Ugly_EmptySteps(t *T) {
	c := New()
	c.Task("empty", Task{})
	r := c.Task("empty").Run(context.Background(), c, NewOptions())
	AssertFalse(t, r.OK)
}

func TestAction_Tasks_Good(t *T) {
	c := New()
	c.Task("deploy", Task{Steps: []Step{{Action: "x"}}})
	c.Task("review", Task{Steps: []Step{{Action: "y"}}})
	AssertEqual(t, []string{"deploy", "review"}, c.Tasks())
}
