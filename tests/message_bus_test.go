package core_test

import (
	. "forge.lthn.ai/core/go/pkg/core"
	"errors"
	"sync"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestBus_Action_Good(t *testing.T) {
	c, _ := New()

	var received []Message
	c.IPC().RegisterAction(func(_ *Core, msg Message) error {
		received = append(received, msg)
		return nil
	})
	c.IPC().RegisterAction(func(_ *Core, msg Message) error {
		received = append(received, msg)
		return nil
	})

	err := c.IPC().Action("hello")
	assert.NoError(t, err)
	assert.Len(t, received, 2)
}

func TestBus_Action_Bad(t *testing.T) {
	c, _ := New()

	err1 := errors.New("handler1 failed")
	err2 := errors.New("handler2 failed")

	c.IPC().RegisterAction(func(_ *Core, msg Message) error { return err1 })
	c.IPC().RegisterAction(func(_ *Core, msg Message) error { return nil })
	c.IPC().RegisterAction(func(_ *Core, msg Message) error { return err2 })

	err := c.IPC().Action("test")
	assert.Error(t, err)
	assert.ErrorIs(t, err, err1)
	assert.ErrorIs(t, err, err2)
}

func TestBus_RegisterAction_Good(t *testing.T) {
	c, _ := New()

	var coreRef *Core
	c.IPC().RegisterAction(func(core *Core, msg Message) error {
		coreRef = core
		return nil
	})

	_ = c.IPC().Action(nil)
	assert.Same(t, c, coreRef, "handler should receive the Core reference")
}

func TestBus_Query_Good(t *testing.T) {
	c, _ := New()

	c.IPC().RegisterQuery(func(_ *Core, q Query) (any, bool, error) {
		return "first", true, nil
	})

	result, handled, err := c.IPC().Query(TestQuery{Value: "test"})
	assert.NoError(t, err)
	assert.True(t, handled)
	assert.Equal(t, "first", result)
}

func TestBus_QueryAll_Good(t *testing.T) {
	c, _ := New()

	c.IPC().RegisterQuery(func(_ *Core, q Query) (any, bool, error) {
		return "a", true, nil
	})
	c.IPC().RegisterQuery(func(_ *Core, q Query) (any, bool, error) {
		return nil, false, nil // skips
	})
	c.IPC().RegisterQuery(func(_ *Core, q Query) (any, bool, error) {
		return "b", true, nil
	})

	results, err := c.IPC().QueryAll(TestQuery{})
	assert.NoError(t, err)
	assert.Equal(t, []any{"a", "b"}, results)
}

func TestBus_Perform_Good(t *testing.T) {
	c, _ := New()

	c.IPC().RegisterTask(func(_ *Core, t Task) (any, bool, error) {
		return "done", true, nil
	})

	result, handled, err := c.IPC().Perform(TestTask{})
	assert.NoError(t, err)
	assert.True(t, handled)
	assert.Equal(t, "done", result)
}

func TestBus_ConcurrentAccess_Good(t *testing.T) {
	c, _ := New()

	var wg sync.WaitGroup
	const goroutines = 20

	// Concurrent register + dispatch
	for i := 0; i < goroutines; i++ {
		wg.Add(2)
		go func() {
			defer wg.Done()
			c.IPC().RegisterAction(func(_ *Core, msg Message) error { return nil })
		}()
		go func() {
			defer wg.Done()
			_ = c.IPC().Action("ping")
		}()
	}

	for i := 0; i < goroutines; i++ {
		wg.Add(2)
		go func() {
			defer wg.Done()
			c.IPC().RegisterQuery(func(_ *Core, q Query) (any, bool, error) { return nil, false, nil })
		}()
		go func() {
			defer wg.Done()
			_, _ = c.IPC().QueryAll(TestQuery{})
		}()
	}

	for i := 0; i < goroutines; i++ {
		wg.Add(2)
		go func() {
			defer wg.Done()
			c.IPC().RegisterTask(func(_ *Core, t Task) (any, bool, error) { return nil, false, nil })
		}()
		go func() {
			defer wg.Done()
			_, _, _ = c.IPC().Perform(TestTask{})
		}()
	}

	wg.Wait()
}

func TestBus_Action_NoHandlers(t *testing.T) {
	c, _ := New()
	err := c.IPC().Action("no one listening")
	assert.NoError(t, err)
}

func TestBus_Query_NoHandlers(t *testing.T) {
	c, _ := New()
	result, handled, err := c.IPC().Query(TestQuery{})
	assert.NoError(t, err)
	assert.False(t, handled)
	assert.Nil(t, result)
}

func TestBus_QueryAll_NoHandlers(t *testing.T) {
	c, _ := New()
	results, err := c.IPC().QueryAll(TestQuery{})
	assert.NoError(t, err)
	assert.Empty(t, results)
}

func TestBus_Perform_NoHandlers(t *testing.T) {
	c, _ := New()
	result, handled, err := c.IPC().Perform(TestTask{})
	assert.NoError(t, err)
	assert.False(t, handled)
	assert.Nil(t, result)
}
