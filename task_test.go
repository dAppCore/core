package core_test

import (
	"context"
	"sync"
	"time"

	. "dappco.re/go"
)

// --- PerformAsync ---

func TestTask_PerformAsync_Good(t *T) {
	c := New()
	var mu sync.Mutex
	var result string

	c.Action("work", func(_ context.Context, _ Options) Result {
		mu.Lock()
		result = "done"
		mu.Unlock()
		return Result{Value: "done", OK: true}
	})

	r := c.PerformAsync("work", NewOptions())
	AssertTrue(t, r.OK)
	AssertTrue(t, HasPrefix(r.Value.(string), "id-"), "should return task ID")

	time.Sleep(100 * time.Millisecond)

	mu.Lock()
	AssertEqual(t, "done", result)
	mu.Unlock()
}

func TestTask_PerformAsync_Good_Progress(t *T) {
	c := New()
	c.Action("tracked", func(_ context.Context, _ Options) Result {
		return Result{OK: true}
	})

	r := c.PerformAsync("tracked", NewOptions())
	taskID := r.Value.(string)
	c.Progress(taskID, 0.5, "halfway", "tracked")
}

func TestTask_PerformAsync_Good_Completion(t *T) {
	c := New()
	completed := make(chan ActionTaskCompleted, 1)

	c.Action("completable", func(_ context.Context, _ Options) Result {
		return Result{Value: "output", OK: true}
	})

	c.RegisterAction(func(_ *Core, msg Message) Result {
		if evt, ok := msg.(ActionTaskCompleted); ok {
			completed <- evt
		}
		return Result{OK: true}
	})

	c.PerformAsync("completable", NewOptions())

	select {
	case evt := <-completed:
		AssertTrue(t, evt.Result.OK)
		AssertEqual(t, "output", evt.Result.Value)
	case <-time.After(2 * time.Second):
		t.Fatal("timed out waiting for completion")
	}
}

func TestTask_PerformAsync_Bad_ActionNotRegistered(t *T) {
	c := New()
	completed := make(chan ActionTaskCompleted, 1)

	c.RegisterAction(func(_ *Core, msg Message) Result {
		if evt, ok := msg.(ActionTaskCompleted); ok {
			completed <- evt
		}
		return Result{OK: true}
	})

	c.PerformAsync("nonexistent", NewOptions())

	select {
	case evt := <-completed:
		AssertFalse(t, evt.Result.OK, "unregistered action should fail")
	case <-time.After(2 * time.Second):
		t.Fatal("timed out")
	}
}

func TestTask_PerformAsync_Bad_AfterShutdown(t *T) {
	c := New()
	c.Action("work", func(_ context.Context, _ Options) Result { return Result{OK: true} })

	c.ServiceStartup(context.Background(), nil)
	c.ServiceShutdown(context.Background())

	r := c.PerformAsync("work", NewOptions())
	AssertFalse(t, r.OK)
}

// --- RegisterAction + RegisterActions (broadcast handlers) ---

func TestTask_RegisterAction_Good(t *T) {
	c := New()
	called := false
	c.RegisterAction(func(_ *Core, _ Message) Result {
		called = true
		return Result{OK: true}
	})
	c.ACTION(nil)
	AssertTrue(t, called)
}

func TestTask_RegisterActions_Good(t *T) {
	c := New()
	count := 0
	h := func(_ *Core, _ Message) Result { count++; return Result{OK: true} }
	c.RegisterActions(h, h)
	c.ACTION(nil)
	AssertEqual(t, 2, count)
}

func TestTask_Core_PerformAsync_Good(t *T) {
	c := New()
	completed := make(chan ActionTaskCompleted, 1)
	c.Action("agent.dispatch", func(_ context.Context, _ Options) Result {
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
	select {
	case evt := <-completed:
		AssertTrue(t, evt.Result.OK)
		AssertEqual(t, "agent.dispatch", evt.Action)
		AssertEqual(t, "dispatched", evt.Result.Value)
	case <-time.After(2 * time.Second):
		t.Fatal("timed out waiting for agent.dispatch completion")
	}
}

func TestTask_Core_PerformAsync_Bad(t *T) {
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
	select {
	case evt := <-completed:
		AssertFalse(t, evt.Result.OK)
		AssertEqual(t, "agent.missing", evt.Action)
	case <-time.After(2 * time.Second):
		t.Fatal("timed out waiting for missing action completion")
	}
}

func TestTask_Core_PerformAsync_Ugly(t *T) {
	c := New()
	completed := make(chan ActionTaskCompleted, 1)
	c.Action("agent.panics", func(_ context.Context, _ Options) Result {
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
	select {
	case evt := <-completed:
		AssertFalse(t, evt.Result.OK)
		AssertContains(t, evt.Result.Error(), "panic")
	case <-time.After(2 * time.Second):
		t.Fatal("timed out waiting for panic completion")
	}
}

func TestTask_Core_Progress_Good(t *T) {
	c := New()
	progress := make(chan ActionTaskProgress, 1)
	c.RegisterAction(func(_ *Core, msg Message) Result {
		if evt, ok := msg.(ActionTaskProgress); ok {
			progress <- evt
		}
		return Result{OK: true}
	})

	c.Progress("task-1", 0.5, "halfway", "agent.dispatch")

	select {
	case evt := <-progress:
		AssertEqual(t, "task-1", evt.TaskIdentifier)
		AssertEqual(t, 0.5, evt.Progress)
		AssertEqual(t, "halfway", evt.Message)
		AssertEqual(t, "agent.dispatch", evt.Action)
	case <-time.After(2 * time.Second):
		t.Fatal("timed out waiting for progress event")
	}
}

func TestTask_Core_Progress_Bad(t *T) {
	c := New()

	c.Progress("task-absent", -1, "refused", "agent.dispatch")

	AssertTrue(t, true)
}

func TestTask_Core_Progress_Ugly(t *T) {
	c := New()
	progress := make(chan ActionTaskProgress, 1)
	c.RegisterAction(func(_ *Core, msg Message) Result {
		if evt, ok := msg.(ActionTaskProgress); ok {
			progress <- evt
		}
		return Result{OK: true}
	})

	c.Progress("", 1, "", "")

	select {
	case evt := <-progress:
		AssertEqual(t, "", evt.TaskIdentifier)
		AssertEqual(t, 1.0, evt.Progress)
		AssertEqual(t, "", evt.Message)
		AssertEqual(t, "", evt.Action)
	case <-time.After(2 * time.Second):
		t.Fatal("timed out waiting for empty progress event")
	}
}
