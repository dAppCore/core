package core_test

import (
	"sync"
	"testing"
	"time"

	. "dappco.re/go/core/pkg/core"
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
