package core_test

import (
	"sync"
	"testing"
	"time"

	. "forge.lthn.ai/core/go/pkg/core"
	"github.com/stretchr/testify/assert"
)

// --- PerformAsync ---

func TestPerformAsync_Good(t *testing.T) {
	c := New()
	var mu sync.Mutex
	var result string

	c.RegisterTask(func(_ *Core, task Task) (any, bool, error) {
		mu.Lock()
		result = "done"
		mu.Unlock()
		return "completed", true, nil
	})

	taskID := c.PerformAsync("work")
	assert.NotEmpty(t, taskID)

	time.Sleep(100 * time.Millisecond)

	mu.Lock()
	assert.Equal(t, "done", result)
	mu.Unlock()
}

func TestPerformAsync_Progress_Good(t *testing.T) {
	c := New()
	c.RegisterTask(func(_ *Core, task Task) (any, bool, error) {
		return nil, true, nil
	})

	taskID := c.PerformAsync("work")
	c.Progress(taskID, 0.5, "halfway", "work")
}

// --- RegisterAction + RegisterActions ---

func TestRegisterAction_Good(t *testing.T) {
	c := New()
	called := false
	c.RegisterAction(func(_ *Core, _ Message) error {
		called = true
		return nil
	})
	_ = c.Action(nil)
	assert.True(t, called)
}

func TestRegisterActions_Good(t *testing.T) {
	c := New()
	count := 0
	h := func(_ *Core, _ Message) error { count++; return nil }
	c.RegisterActions(h, h)
	_ = c.Action(nil)
	assert.Equal(t, 2, count)
}
