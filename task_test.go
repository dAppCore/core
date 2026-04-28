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
