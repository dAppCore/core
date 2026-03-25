// SPDX-License-Identifier: EUPL-1.2

// Named action system for the Core framework.
// Actions are the atomic unit of work — named, registered, invokable,
// and inspectable. The Action registry IS the capability map.
//
// Register a named action:
//
//	c.Action("git.log", func(ctx context.Context, opts core.Options) core.Result {
//	    dir := opts.String("dir")
//	    return c.Process().RunIn(ctx, dir, "git", "log")
//	})
//
// Invoke by name:
//
//	r := c.Action("git.log").Run(ctx, core.NewOptions(
//	    core.Option{Key: "dir", Value: "/path/to/repo"},
//	))
//
// Check capability:
//
//	if c.Action("process.run").Exists() { ... }
//
// List all:
//
//	names := c.Actions()  // ["process.run", "agentic.dispatch", ...]
package core

import "context"

// ActionHandler is the function signature for all named actions.
//
//	func(ctx context.Context, opts core.Options) core.Result
type ActionHandler func(context.Context, Options) Result

// Action is a registered named action.
//
//	action := c.Action("process.run")
//	action.Description  // "Execute a command"
//	action.Schema       // expected input keys
type Action struct {
	Name        string
	Handler     ActionHandler
	Description string
	Schema      Options // declares expected input keys (optional)
	enabled     bool
}

// Run executes the action with panic recovery.
// Returns Result{OK: false} if the action has no handler (not registered).
//
//	r := c.Action("process.run").Run(ctx, opts)
func (a *Action) Run(ctx context.Context, opts Options) (result Result) {
	if a == nil || a.Handler == nil {
		return Result{E("action.Run", Concat("action not registered: ", a.safeName()), nil), false}
	}
	if !a.enabled {
		return Result{E("action.Run", Concat("action disabled: ", a.Name), nil), false}
	}
	defer func() {
		if r := recover(); r != nil {
			result = Result{E("action.Run", Sprint("panic in action ", a.Name, ": ", r), nil), false}
		}
	}()
	return a.Handler(ctx, opts)
}

// Exists returns true if this action has a registered handler.
//
//	if c.Action("process.run").Exists() { ... }
func (a *Action) Exists() bool {
	return a != nil && a.Handler != nil
}

func (a *Action) safeName() string {
	if a == nil {
		return "<nil>"
	}
	return a.Name
}

// --- Core accessor ---

// Action gets or registers a named action.
// With a handler argument: registers the action.
// Without: returns the action for invocation.
//
//	c.Action("process.run", handler)           // register
//	c.Action("process.run").Run(ctx, opts)     // invoke
//	c.Action("process.run").Exists()           // check
func (c *Core) Action(name string, handler ...ActionHandler) *Action {
	if len(handler) > 0 {
		def := &Action{Name: name, Handler: handler[0], enabled: true}
		c.ipc.actions.Set(name, def)
		return def
	}
	r := c.ipc.actions.Get(name)
	if !r.OK {
		return &Action{Name: name} // no handler — Exists() returns false
	}
	return r.Value.(*Action)
}

// Actions returns all registered named action names in registration order.
//
//	names := c.Actions()  // ["process.run", "agentic.dispatch"]
func (c *Core) Actions() []string {
	return c.ipc.actions.Names()
}

// --- Task Composition ---

// Step is a single step in a Task — references an Action by name.
//
//	core.Step{Action: "agentic.qa"}
//	core.Step{Action: "agentic.poke", Async: true}
//	core.Step{Action: "agentic.verify", Input: "previous"}
type Step struct {
	Action string  // name of the Action to invoke
	With   Options // static options (merged with runtime opts)
	Async  bool    // run in background, don't block
	Input  string  // "previous" = output of last step piped as input
}

// TaskDef is a named sequence of Steps.
//
//	c.Task("agent.completion", core.TaskDef{
//	    Steps: []core.Step{
//	        {Action: "agentic.qa"},
//	        {Action: "agentic.auto-pr"},
//	        {Action: "agentic.verify"},
//	        {Action: "agentic.poke", Async: true},
//	    },
//	})
type TaskDef struct {
	Name        string
	Description string
	Steps       []Step
}

// Run executes the task's steps in order. Sync steps run sequentially —
// if any fails, the chain stops. Async steps are dispatched and don't block.
// The "previous" input pipes the last sync step's output to the next step.
//
//	r := c.Task("deploy").Run(ctx, opts)
func (t *TaskDef) Run(ctx context.Context, c *Core, opts Options) Result {
	if t == nil || len(t.Steps) == 0 {
		return Result{E("task.Run", Concat("task has no steps: ", t.safeName()), nil), false}
	}

	var lastResult Result
	for _, step := range t.Steps {
		// Use step's own options, or runtime options if step has none
		stepOpts := stepOptions(step)
		if stepOpts.Len() == 0 {
			stepOpts = opts
		}

		// Pipe previous result as input
		if step.Input == "previous" && lastResult.OK {
			stepOpts.Set("_input", lastResult.Value)
		}

		action := c.Action(step.Action)
		if !action.Exists() {
			return Result{E("task.Run", Concat("action not found: ", step.Action), nil), false}
		}

		if step.Async {
			// Fire and forget — don't block the chain
			go func(a *Action, o Options) {
				defer func() {
					if r := recover(); r != nil {
						Error("async task step panicked", "action", a.Name, "panic", r)
					}
				}()
				a.Run(ctx, o)
			}(action, stepOpts)
			continue
		}

		lastResult = action.Run(ctx, stepOpts)
		if !lastResult.OK {
			return lastResult
		}
	}
	return lastResult
}

func (t *TaskDef) safeName() string {
	if t == nil {
		return "<nil>"
	}
	return t.Name
}

// mergeStepOptions returns the step's With options — runtime opts are passed directly.
// Step.With provides static defaults that the step was registered with.
func stepOptions(step Step) Options {
	return step.With
}

// Task gets or registers a named task.
// With a TaskDef argument: registers the task.
// Without: returns the task for invocation.
//
//	c.Task("deploy", core.TaskDef{Steps: steps})  // register
//	c.Task("deploy").Run(ctx, c, opts)             // invoke
func (c *Core) Task(name string, def ...TaskDef) *TaskDef {
	if len(def) > 0 {
		d := def[0]
		d.Name = name
		c.ipc.tasks.Set(name, &d)
		return &d
	}
	r := c.ipc.tasks.Get(name)
	if !r.OK {
		return &TaskDef{Name: name}
	}
	return r.Value.(*TaskDef)
}

// Tasks returns all registered task names.
func (c *Core) Tasks() []string {
	return c.ipc.tasks.Names()
}
