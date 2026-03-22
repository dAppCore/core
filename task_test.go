package core_test

import (
	"context"
	"sync"
	"testing"
	"time"

	. "dappco.re/go/core"
	"github.com/stretchr/testify/assert"
)

// --- PerformAsync ---

func TestPerformAsync_Good(t *testing.T) {
	c := New()
	var mu sync.Mutex
	var result string

	c.RegisterTask(func(_ *Core, task Task) Result {
		mu.Lock()
		result = "done"
		mu.Unlock()
		return Result{"completed", true}
	})

	r := c.PerformAsync("work")
	assert.True(t, r.OK)
	taskID := r.Value.(string)
	assert.NotEmpty(t, taskID)

	time.Sleep(100 * time.Millisecond)

	mu.Lock()
	assert.Equal(t, "done", result)
	mu.Unlock()
}

func TestPerformAsync_Progress_Good(t *testing.T) {
	c := New()
	c.RegisterTask(func(_ *Core, task Task) Result {
		return Result{OK: true}
	})

	r := c.PerformAsync("work")
	taskID := r.Value.(string)
	c.Progress(taskID, 0.5, "halfway", "work")
}

func TestPerformAsync_Completion_Good(t *testing.T) {
	c := New()
	completed := make(chan ActionTaskCompleted, 1)

	c.RegisterTask(func(_ *Core, task Task) Result {
		return Result{Value: "result", OK: true}
	})
	c.RegisterAction(func(_ *Core, msg Message) Result {
		if evt, ok := msg.(ActionTaskCompleted); ok {
			completed <- evt
		}
		return Result{OK: true}
	})

	c.PerformAsync("work")

	select {
	case evt := <-completed:
		assert.Nil(t, evt.Error)
		assert.Equal(t, "result", evt.Result)
	case <-time.After(2 * time.Second):
		t.Fatal("timed out waiting for completion")
	}
}

func TestPerformAsync_NoHandler_Good(t *testing.T) {
	c := New()
	completed := make(chan ActionTaskCompleted, 1)

	c.RegisterAction(func(_ *Core, msg Message) Result {
		if evt, ok := msg.(ActionTaskCompleted); ok {
			completed <- evt
		}
		return Result{OK: true}
	})

	c.PerformAsync("unhandled")

	select {
	case evt := <-completed:
		assert.NotNil(t, evt.Error)
	case <-time.After(2 * time.Second):
		t.Fatal("timed out")
	}
}

func TestPerformAsync_AfterShutdown_Bad(t *testing.T) {
	c := New()
	c.ServiceStartup(context.Background(), nil)
	c.ServiceShutdown(context.Background())

	r := c.PerformAsync("should not run")
	assert.False(t, r.OK)
}

// --- RegisterAction + RegisterActions ---

func TestRegisterAction_Good(t *testing.T) {
	c := New()
	called := false
	c.RegisterAction(func(_ *Core, _ Message) Result {
		called = true
		return Result{OK: true}
	})
	c.Action(nil)
	assert.True(t, called)
}

func TestRegisterActions_Good(t *testing.T) {
	c := New()
	count := 0
	h := func(_ *Core, _ Message) Result { count++; return Result{OK: true} }
	c.RegisterActions(h, h)
	c.Action(nil)
	assert.Equal(t, 2, count)
}
