package core_test

import . "dappco.re/go"

// --- NamedAction Register ---

func TestAction_NamedAction_Good_Register(t *T) {
	c := New()
	def := c.Action("process.run", func(_ Context, opts Options) Result {
		return Result{Value: "output", OK: true}
	})
	AssertNotNil(t, def)
	AssertEqual(t, "process.run", def.Name)
	AssertTrue(t, def.Exists())
}

func TestAction_NamedAction_Good_Invoke(t *T) {
	c := New()
	c.Action("git.log", func(_ Context, opts Options) Result {
		dir := opts.String("dir")
		return Result{Value: Concat("log from ", dir), OK: true}
	})

	r := c.Action("git.log").Run(Background(), NewOptions(
		Option{Key: "dir", Value: "/repo"},
	))
	AssertTrue(t, r.OK)
	AssertEqual(t, "log from /repo", r.Value)
}

func TestAction_NamedAction_Bad_NotRegistered(t *T) {
	c := New()
	r := c.Action("missing.action").Run(Background(), NewOptions())
	AssertFalse(t, r.OK, "invoking unregistered action must fail")
}

func TestAction_NamedAction_Good_Exists(t *T) {
	c := New()
	c.Action("brain.recall", func(_ Context, _ Options) Result {
		return Result{OK: true}
	})
	AssertTrue(t, c.Action("brain.recall").Exists())
	AssertFalse(t, c.Action("brain.forget").Exists())
}

func TestAction_NamedAction_Ugly_PanicRecovery(t *T) {
	c := New()
	c.Action("explode", func(_ Context, _ Options) Result {
		panic("boom")
	})
	r := c.Action("explode").Run(Background(), NewOptions())
	AssertFalse(t, r.OK, "panicking action must return !OK, not crash")
	err, ok := r.Value.(error)
	AssertTrue(t, ok)
	AssertContains(t, err.Error(), "panic")
}

func TestAction_NamedAction_Ugly_NilAction(t *T) {
	var def *Action
	r := def.Run(Background(), NewOptions())
	AssertFalse(t, r.OK)
	AssertFalse(t, def.Exists())
}

// --- Actions listing ---

func TestAction_Actions_Good(t *T) {
	c := New()
	c.Action("process.run", func(_ Context, _ Options) Result { return Result{OK: true} })
	c.Action("process.kill", func(_ Context, _ Options) Result { return Result{OK: true} })
	c.Action("agentic.dispatch", func(_ Context, _ Options) Result { return Result{OK: true} })

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
	def := c.Action("process.run", func(_ Context, _ Options) Result { return Result{OK: true} })
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
	full.Action("process.run", func(_ Context, _ Options) Result {
		return Result{Value: "executed", OK: true}
	})

	// Sandboxed Core — no process
	sandboxed := New()

	// Full can execute
	r := full.Action("process.run").Run(Background(), NewOptions())
	AssertTrue(t, r.OK)

	// Sandboxed returns not-registered
	r = sandboxed.Action("process.run").Run(Background(), NewOptions())
	AssertFalse(t, r.OK, "sandboxed Core must not have process capability")
}

// --- Action overwrite ---

func TestAction_NamedAction_Good_Overwrite(t *T) {
	c := New()
	c.Action("hot.reload", func(_ Context, _ Options) Result {
		return Result{Value: "v1", OK: true}
	})
	c.Action("hot.reload", func(_ Context, _ Options) Result {
		return Result{Value: "v2", OK: true}
	})

	r := c.Action("hot.reload").Run(Background(), NewOptions())
	AssertTrue(t, r.OK)
	AssertEqual(t, "v2", r.Value, "latest handler wins")
}

// --- Task Composition ---

func TestAction_Task_Good_Sequential(t *T) {
	c := New()
	var order []string
	c.Action("step.a", func(_ Context, _ Options) Result {
		order = append(order, "a")
		return Result{Value: "output-a", OK: true}
	})
	c.Action("step.b", func(_ Context, _ Options) Result {
		order = append(order, "b")
		return Result{Value: "output-b", OK: true}
	})

	c.Task("pipeline", Task{
		Steps: []Step{
			{Action: "step.a"},
			{Action: "step.b"},
		},
	})

	r := c.Task("pipeline").Run(Background(), c, NewOptions())
	AssertTrue(t, r.OK)
	AssertEqual(t, []string{"a", "b"}, order, "steps must run in order")
	AssertEqual(t, "output-b", r.Value, "last step's result is returned")
}

func TestAction_Task_Bad_StepFails(t *T) {
	c := New()
	var order []string
	c.Action("step.ok", func(_ Context, _ Options) Result {
		order = append(order, "ok")
		return Result{OK: true}
	})
	c.Action("step.fail", func(_ Context, _ Options) Result {
		order = append(order, "fail")
		return Result{Value: NewError("broke"), OK: false}
	})
	c.Action("step.never", func(_ Context, _ Options) Result {
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

	r := c.Task("broken").Run(Background(), c, NewOptions())
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
	r := c.Task("missing").Run(Background(), c, NewOptions())
	AssertFalse(t, r.OK)
}

func TestAction_Task_Good_PreviousInput(t *T) {
	c := New()
	c.Action("produce", func(_ Context, _ Options) Result {
		return Result{Value: "data-from-step-1", OK: true}
	})
	c.Action("consume", func(_ Context, opts Options) Result {
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

	r := c.Task("pipe").Run(Background(), c, NewOptions())
	AssertTrue(t, r.OK)
	AssertEqual(t, "got: data-from-step-1", r.Value)
}

func TestAction_Task_Ugly_EmptySteps(t *T) {
	c := New()
	c.Task("empty", Task{})
	r := c.Task("empty").Run(Background(), c, NewOptions())
	AssertFalse(t, r.OK)
}

func TestAction_Tasks_Good(t *T) {
	c := New()
	c.Task("deploy", Task{Steps: []Step{{Action: "x"}}})
	c.Task("review", Task{Steps: []Step{{Action: "y"}}})
	AssertEqual(t, []string{"deploy", "review"}, c.Tasks())
}

// --- AX-7 canonical triplets ---

func TestAction_Action_Run_Good(t *T) {
	c := New()
	a := c.Action("agent.dispatch", func(_ Context, opts Options) Result {
		return Result{Value: opts.String("agent"), OK: true}
	})
	r := a.Run(Background(), NewOptions(Option{Key: "agent", Value: "codex"}))
	AssertTrue(t, r.OK)
	AssertEqual(t, "codex", r.Value)
}

func TestAction_Action_Run_Bad(t *T) {
	a := &Action{Name: "agent.missing"}
	r := a.Run(Background(), NewOptions())
	AssertFalse(t, r.OK)
	AssertContains(t, r.Error(), "not registered")
}

func TestAction_Action_Run_Ugly(t *T) {
	c := New()
	a := c.Action("agent.panic", func(_ Context, _ Options) Result {
		panic("dispatch crashed")
	})
	r := a.Run(Background(), NewOptions())
	AssertFalse(t, r.OK)
	AssertContains(t, r.Error(), "panic")
}

func TestAction_Action_Exists_Good(t *T) {
	a := &Action{Name: "agent.dispatch", Handler: func(_ Context, _ Options) Result { return Result{OK: true} }}
	AssertTrue(t, a.Exists())
}

func TestAction_Action_Exists_Bad(t *T) {
	a := &Action{Name: "agent.dispatch"}
	AssertFalse(t, a.Exists())
}

func TestAction_Action_Exists_Ugly(t *T) {
	var a *Action
	AssertFalse(t, a.Exists())
}

func TestAction_Core_Action_Good(t *T) {
	c := New()
	a := c.Action("agent.dispatch", func(_ Context, _ Options) Result {
		return Result{Value: "queued", OK: true}
	})
	AssertTrue(t, a.Exists())
	AssertEqual(t, "queued", c.Action("agent.dispatch").Run(Background(), NewOptions()).Value)
}

func TestAction_Core_Action_Bad(t *T) {
	c := New()
	a := c.Action("agent.missing")
	AssertFalse(t, a.Exists())
	AssertFalse(t, a.Run(Background(), NewOptions()).OK)
}

func TestAction_Core_Action_Ugly(t *T) {
	c := New()
	c.Action("agent.dispatch", func(_ Context, _ Options) Result { return Result{Value: "v1", OK: true} })
	c.Action("agent.dispatch", func(_ Context, _ Options) Result { return Result{Value: "v2", OK: true} })
	r := c.Action("agent.dispatch").Run(Background(), NewOptions())
	AssertTrue(t, r.OK)
	AssertEqual(t, "v2", r.Value)
}

func TestAction_Core_Actions_Good(t *T) {
	c := New()
	c.Action("agent.prepare", func(_ Context, _ Options) Result { return Result{OK: true} })
	c.Action("agent.dispatch", func(_ Context, _ Options) Result { return Result{OK: true} })
	AssertEqual(t, []string{"agent.prepare", "agent.dispatch"}, c.Actions())
}

func TestAction_Core_Actions_Bad(t *T) {
	c := New()
	AssertEmpty(t, c.Actions())
}

func TestAction_Core_Actions_Ugly(t *T) {
	c := New()
	c.Action("agent.prepare", func(_ Context, _ Options) Result { return Result{OK: true} })
	names := c.Actions()
	names[0] = "mutated"
	AssertEqual(t, []string{"agent.prepare"}, c.Actions())
}

func TestAction_Task_Run_Good(t *T) {
	c := New()
	c.Action("agent.prepare", func(_ Context, _ Options) Result {
		return Result{Value: "prepared", OK: true}
	})
	task := &Task{Name: "agent.dispatch", Steps: []Step{{Action: "agent.prepare"}}}
	r := task.Run(Background(), c, NewOptions())
	AssertTrue(t, r.OK)
	AssertEqual(t, "prepared", r.Value)
}

func TestAction_Task_Run_Bad(t *T) {
	c := New()
	task := &Task{Name: "agent.dispatch", Steps: []Step{{Action: "agent.missing"}}}
	r := task.Run(Background(), c, NewOptions())
	AssertFalse(t, r.OK)
}

func TestAction_Task_Run_Ugly(t *T) {
	c := New()
	c.Action("agent.prepare", func(_ Context, _ Options) Result {
		return Result{Value: "session-token", OK: true}
	})
	c.Action("agent.dispatch", func(_ Context, opts Options) Result {
		return Result{Value: opts.String("_input"), OK: true}
	})
	task := &Task{Name: "agent.pipeline", Steps: []Step{
		{Action: "agent.prepare"},
		{Action: "agent.dispatch", Input: "previous"},
	}}
	r := task.Run(Background(), c, NewOptions())
	AssertTrue(t, r.OK)
	AssertEqual(t, "session-token", r.Value)
}

func TestAction_Core_Task_Good(t *T) {
	c := New()
	task := c.Task("agent.pipeline", Task{Steps: []Step{{Action: "agent.prepare"}}})
	AssertEqual(t, "agent.pipeline", task.Name)
	AssertLen(t, task.Steps, 1)
}

func TestAction_Core_Task_Bad(t *T) {
	c := New()
	task := c.Task("agent.missing")
	r := task.Run(Background(), c, NewOptions())
	AssertFalse(t, r.OK)
}

func TestAction_Core_Task_Ugly(t *T) {
	c := New()
	c.Task("agent.pipeline", Task{Steps: []Step{{Action: "agent.prepare"}}})
	c.Task("agent.pipeline", Task{Steps: []Step{{Action: "agent.dispatch"}}})
	task := c.Task("agent.pipeline")
	AssertEqual(t, "agent.dispatch", task.Steps[0].Action)
}

func TestAction_Core_Tasks_Good(t *T) {
	c := New()
	c.Task("agent.prepare", Task{Steps: []Step{{Action: "agent.prepare"}}})
	c.Task("agent.dispatch", Task{Steps: []Step{{Action: "agent.dispatch"}}})
	AssertEqual(t, []string{"agent.prepare", "agent.dispatch"}, c.Tasks())
}

func TestAction_Core_Tasks_Bad(t *T) {
	c := New()
	AssertEmpty(t, c.Tasks())
}

func TestAction_Core_Tasks_Ugly(t *T) {
	c := New()
	c.Task("agent.prepare", Task{Steps: []Step{{Action: "agent.prepare"}}})
	tasks := c.Tasks()
	tasks[0] = "mutated"
	AssertEqual(t, []string{"agent.prepare"}, c.Tasks())
}

// --- PerformAsync ---

func TestAction_PerformAsync_Good(t *T) {
	c := New()
	var mu Mutex
	var result string

	c.Action("work", func(_ Context, _ Options) Result {
		mu.Lock()
		result = "done"
		mu.Unlock()
		return Result{Value: "done", OK: true}
	})

	r := c.PerformAsync("work", NewOptions())
	AssertTrue(t, r.OK)
	AssertTrue(t, HasPrefix(r.Value.(string), "id-"), "should return task ID")

	Sleep(100 * Millisecond)

	mu.Lock()
	AssertEqual(t, "done", result)
	mu.Unlock()
}

func TestAction_PerformAsync_Good_Progress(t *T) {
	c := New()
	c.Action("tracked", func(_ Context, _ Options) Result {
		return Result{OK: true}
	})

	r := c.PerformAsync("tracked", NewOptions())
	taskID := r.Value.(string)
	c.Progress(taskID, 0.5, "halfway", "tracked")
}

func TestAction_PerformAsync_Good_Completion(t *T) {
	c := New()
	completed := make(chan ActionTaskCompleted, 1)

	c.Action("completable", func(_ Context, _ Options) Result {
		return Result{Value: "output", OK: true}
	})

	c.RegisterAction(func(_ *Core, msg Message) Result {
		if evt, ok := msg.(ActionTaskCompleted); ok {
			completed <- evt
		}
		return Result{OK: true}
	})

	c.PerformAsync("completable", NewOptions())

	timeout, cancel := WithTimeout(Background(), 2*Second)
	defer cancel()
	select {
	case evt := <-completed:
		AssertTrue(t, evt.Result.OK)
		AssertEqual(t, "output", evt.Result.Value)
	case <-timeout.Done():
		t.Fatal("timed out waiting for completion")
	}
}

func TestAction_PerformAsync_Bad_ActionNotRegistered(t *T) {
	c := New()
	completed := make(chan ActionTaskCompleted, 1)

	c.RegisterAction(func(_ *Core, msg Message) Result {
		if evt, ok := msg.(ActionTaskCompleted); ok {
			completed <- evt
		}
		return Result{OK: true}
	})

	c.PerformAsync("nonexistent", NewOptions())

	timeout, cancel := WithTimeout(Background(), 2*Second)
	defer cancel()
	select {
	case evt := <-completed:
		AssertFalse(t, evt.Result.OK, "unregistered action should fail")
	case <-timeout.Done():
		t.Fatal("timed out")
	}
}

func TestAction_PerformAsync_Bad_AfterShutdown(t *T) {
	c := New()
	c.Action("work", func(_ Context, _ Options) Result { return Result{OK: true} })

	c.ServiceStartup(Background(), nil)
	c.ServiceShutdown(Background())

	r := c.PerformAsync("work", NewOptions())
	AssertFalse(t, r.OK)
}

// --- RegisterAction + RegisterActions (broadcast handlers) ---

func TestAction_RegisterAction_Good(t *T) {
	c := New()
	called := false
	c.RegisterAction(func(_ *Core, _ Message) Result {
		called = true
		return Result{OK: true}
	})
	c.ACTION(nil)
	AssertTrue(t, called)
}

func TestAction_RegisterActions_Good(t *T) {
	c := New()
	count := 0
	h := func(_ *Core, _ Message) Result { count++; return Result{OK: true} }
	c.RegisterActions(h, h)
	c.ACTION(nil)
	AssertEqual(t, 2, count)
}

func TestAction_Core_PerformAsync_Good(t *T) {
	c := New()
	completed := make(chan ActionTaskCompleted, 1)
	c.Action("agent.dispatch", func(_ Context, _ Options) Result {
		return Result{Value: "dispatched", OK: true}
	})
	c.RegisterAction(func(_ *Core, msg Message) Result {
		if evt, ok := msg.(ActionTaskCompleted); ok {
			completed <- evt
		}
		return Result{OK: true}
	})

	r := c.PerformAsync("agent.dispatch", NewOptions())

	AssertTrue(t, r.OK)
	AssertTrue(t, HasPrefix(r.Value.(string), "id-"))
	timeout, cancel := WithTimeout(Background(), 2*Second)
	defer cancel()
	select {
	case evt := <-completed:
		AssertTrue(t, evt.Result.OK)
		AssertEqual(t, "agent.dispatch", evt.Action)
		AssertEqual(t, "dispatched", evt.Result.Value)
	case <-timeout.Done():
		t.Fatal("timed out waiting for agent.dispatch completion")
	}
}

func TestAction_Core_PerformAsync_Bad(t *T) {
	c := New()
	completed := make(chan ActionTaskCompleted, 1)
	c.RegisterAction(func(_ *Core, msg Message) Result {
		if evt, ok := msg.(ActionTaskCompleted); ok {
			completed <- evt
		}
		return Result{OK: true}
	})

	r := c.PerformAsync("agent.missing", NewOptions())

	AssertTrue(t, r.OK)
	timeout, cancel := WithTimeout(Background(), 2*Second)
	defer cancel()
	select {
	case evt := <-completed:
		AssertFalse(t, evt.Result.OK)
		AssertEqual(t, "agent.missing", evt.Action)
	case <-timeout.Done():
		t.Fatal("timed out waiting for missing action completion")
	}
}

func TestAction_Core_PerformAsync_Ugly(t *T) {
	c := New()
	completed := make(chan ActionTaskCompleted, 1)
	c.Action("agent.panics", func(_ Context, _ Options) Result {
		panic("dispatch blew up")
	})
	c.RegisterAction(func(_ *Core, msg Message) Result {
		if evt, ok := msg.(ActionTaskCompleted); ok {
			completed <- evt
		}
		return Result{OK: true}
	})

	r := c.PerformAsync("agent.panics", NewOptions())

	AssertTrue(t, r.OK)
	timeout, cancel := WithTimeout(Background(), 2*Second)
	defer cancel()
	select {
	case evt := <-completed:
		AssertFalse(t, evt.Result.OK)
		AssertContains(t, evt.Result.Error(), "panic")
	case <-timeout.Done():
		t.Fatal("timed out waiting for panic completion")
	}
}

func TestAction_Core_Progress_Good(t *T) {
	c := New()
	progress := make(chan ActionTaskProgress, 1)
	c.RegisterAction(func(_ *Core, msg Message) Result {
		if evt, ok := msg.(ActionTaskProgress); ok {
			progress <- evt
		}
		return Result{OK: true}
	})

	c.Progress("task-1", 0.5, "halfway", "agent.dispatch")

	timeout, cancel := WithTimeout(Background(), 2*Second)
	defer cancel()
	select {
	case evt := <-progress:
		AssertEqual(t, "task-1", evt.TaskIdentifier)
		AssertEqual(t, 0.5, evt.Progress)
		AssertEqual(t, "halfway", evt.Message)
		AssertEqual(t, "agent.dispatch", evt.Action)
	case <-timeout.Done():
		t.Fatal("timed out waiting for progress event")
	}
}

func TestAction_Core_Progress_Bad(t *T) {
	c := New()

	c.Progress("task-absent", -1, "refused", "agent.dispatch")

	AssertTrue(t, true)
}

func TestAction_Core_Progress_Ugly(t *T) {
	c := New()
	progress := make(chan ActionTaskProgress, 1)
	c.RegisterAction(func(_ *Core, msg Message) Result {
		if evt, ok := msg.(ActionTaskProgress); ok {
			progress <- evt
		}
		return Result{OK: true}
	})

	c.Progress("", 1, "", "")

	timeout, cancel := WithTimeout(Background(), 2*Second)
	defer cancel()
	select {
	case evt := <-progress:
		AssertEqual(t, "", evt.TaskIdentifier)
		AssertEqual(t, 1.0, evt.Progress)
		AssertEqual(t, "", evt.Message)
		AssertEqual(t, "", evt.Action)
	case <-timeout.Done():
		t.Fatal("timed out waiting for empty progress event")
	}
}
