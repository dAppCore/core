package core_test

import (
	"context"
	"testing"

	. "dappco.re/go/core"
	"github.com/stretchr/testify/assert"
)

// --- NamedAction Register ---

func TestAction_NamedAction_Good_Register(t *testing.T) {
	c := New()
	def := c.Action("process.run", func(_ context.Context, opts Options) Result {
		return Result{Value: "output", OK: true}
	})
	assert.NotNil(t, def)
	assert.Equal(t, "process.run", def.Name)
	assert.True(t, def.Exists())
}

func TestAction_NamedAction_Good_Invoke(t *testing.T) {
	c := New()
	c.Action("git.log", func(_ context.Context, opts Options) Result {
		dir := opts.String("dir")
		return Result{Value: "log from " + dir, OK: true}
	})

	r := c.Action("git.log").Run(context.Background(), NewOptions(
		Option{Key: "dir", Value: "/repo"},
	))
	assert.True(t, r.OK)
	assert.Equal(t, "log from /repo", r.Value)
}

func TestAction_NamedAction_Bad_NotRegistered(t *testing.T) {
	c := New()
	r := c.Action("missing.action").Run(context.Background(), NewOptions())
	assert.False(t, r.OK, "invoking unregistered action must fail")
}

func TestAction_NamedAction_Good_Exists(t *testing.T) {
	c := New()
	c.Action("brain.recall", func(_ context.Context, _ Options) Result {
		return Result{OK: true}
	})
	assert.True(t, c.Action("brain.recall").Exists())
	assert.False(t, c.Action("brain.forget").Exists())
}

func TestAction_NamedAction_Ugly_PanicRecovery(t *testing.T) {
	c := New()
	c.Action("explode", func(_ context.Context, _ Options) Result {
		panic("boom")
	})
	r := c.Action("explode").Run(context.Background(), NewOptions())
	assert.False(t, r.OK, "panicking action must return !OK, not crash")
	err, ok := r.Value.(error)
	assert.True(t, ok)
	assert.Contains(t, err.Error(), "panic")
}

func TestAction_NamedAction_Ugly_NilAction(t *testing.T) {
	var def *Action
	r := def.Run(context.Background(), NewOptions())
	assert.False(t, r.OK)
	assert.False(t, def.Exists())
}

// --- Actions listing ---

func TestAction_Actions_Good(t *testing.T) {
	c := New()
	c.Action("process.run", func(_ context.Context, _ Options) Result { return Result{OK: true} })
	c.Action("process.kill", func(_ context.Context, _ Options) Result { return Result{OK: true} })
	c.Action("agentic.dispatch", func(_ context.Context, _ Options) Result { return Result{OK: true} })

	names := c.Actions()
	assert.Len(t, names, 3)
	assert.Equal(t, []string{"process.run", "process.kill", "agentic.dispatch"}, names)
}

func TestAction_Actions_Bad_Empty(t *testing.T) {
	c := New()
	assert.Empty(t, c.Actions())
}

// --- Action fields ---

func TestAction_NamedAction_Good_DescriptionAndSchema(t *testing.T) {
	c := New()
	def := c.Action("process.run", func(_ context.Context, _ Options) Result { return Result{OK: true} })
	def.Description = "Execute a command synchronously"
	def.Schema = NewOptions(
		Option{Key: "command", Value: "string"},
		Option{Key: "args", Value: "[]string"},
	)

	retrieved := c.Action("process.run")
	assert.Equal(t, "Execute a command synchronously", retrieved.Description)
	assert.True(t, retrieved.Schema.Has("command"))
}

// --- Permission by registration ---

func TestAction_NamedAction_Good_PermissionModel(t *testing.T) {
	// Full Core — process registered
	full := New()
	full.Action("process.run", func(_ context.Context, _ Options) Result {
		return Result{Value: "executed", OK: true}
	})

	// Sandboxed Core — no process
	sandboxed := New()

	// Full can execute
	r := full.Action("process.run").Run(context.Background(), NewOptions())
	assert.True(t, r.OK)

	// Sandboxed returns not-registered
	r = sandboxed.Action("process.run").Run(context.Background(), NewOptions())
	assert.False(t, r.OK, "sandboxed Core must not have process capability")
}

// --- Action overwrite ---

func TestAction_NamedAction_Good_Overwrite(t *testing.T) {
	c := New()
	c.Action("hot.reload", func(_ context.Context, _ Options) Result {
		return Result{Value: "v1", OK: true}
	})
	c.Action("hot.reload", func(_ context.Context, _ Options) Result {
		return Result{Value: "v2", OK: true}
	})

	r := c.Action("hot.reload").Run(context.Background(), NewOptions())
	assert.True(t, r.OK)
	assert.Equal(t, "v2", r.Value, "latest handler wins")
}

// --- Task Composition ---

func TestAction_Task_Good_Sequential(t *testing.T) {
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

	c.Task("pipeline", TaskDef{
		Steps: []Step{
			{Action: "step.a"},
			{Action: "step.b"},
		},
	})

	r := c.Task("pipeline").Run(context.Background(), c, NewOptions())
	assert.True(t, r.OK)
	assert.Equal(t, []string{"a", "b"}, order, "steps must run in order")
	assert.Equal(t, "output-b", r.Value, "last step's result is returned")
}

func TestAction_Task_Bad_StepFails(t *testing.T) {
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

	c.Task("broken", TaskDef{
		Steps: []Step{
			{Action: "step.ok"},
			{Action: "step.fail"},
			{Action: "step.never"},
		},
	})

	r := c.Task("broken").Run(context.Background(), c, NewOptions())
	assert.False(t, r.OK)
	assert.Equal(t, []string{"ok", "fail"}, order, "chain stops on failure, step.never skipped")
}

func TestAction_Task_Bad_MissingAction(t *testing.T) {
	c := New()
	c.Task("missing", TaskDef{
		Steps: []Step{
			{Action: "nonexistent"},
		},
	})
	r := c.Task("missing").Run(context.Background(), c, NewOptions())
	assert.False(t, r.OK)
}

func TestAction_Task_Good_PreviousInput(t *testing.T) {
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

	c.Task("pipe", TaskDef{
		Steps: []Step{
			{Action: "produce"},
			{Action: "consume", Input: "previous"},
		},
	})

	r := c.Task("pipe").Run(context.Background(), c, NewOptions())
	assert.True(t, r.OK)
	assert.Equal(t, "got: data-from-step-1", r.Value)
}

func TestAction_Task_Ugly_EmptySteps(t *testing.T) {
	c := New()
	c.Task("empty", TaskDef{})
	r := c.Task("empty").Run(context.Background(), c, NewOptions())
	assert.False(t, r.OK)
}

func TestAction_Tasks_Good(t *testing.T) {
	c := New()
	c.Task("deploy", TaskDef{Steps: []Step{{Action: "x"}}})
	c.Task("review", TaskDef{Steps: []Step{{Action: "y"}}})
	assert.Equal(t, []string{"deploy", "review"}, c.Tasks())
}
