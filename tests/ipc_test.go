package core_test

import (
	"testing"

	. "forge.lthn.ai/core/go/pkg/core"
	"github.com/stretchr/testify/assert"
)

// --- IPC: Actions ---

type testMessage struct{ payload string }

func TestAction_Good(t *testing.T) {
	c := New()
	var received Message
	c.RegisterAction(func(_ *Core, msg Message) error {
		received = msg
		return nil
	})
	err := c.ACTION(testMessage{payload: "hello"})
	assert.NoError(t, err)
	assert.Equal(t, testMessage{payload: "hello"}, received)
}

func TestAction_Multiple_Good(t *testing.T) {
	c := New()
	count := 0
	handler := func(_ *Core, _ Message) error { count++; return nil }
	c.RegisterActions(handler, handler, handler)
	_ = c.ACTION(nil)
	assert.Equal(t, 3, count)
}

func TestAction_None_Good(t *testing.T) {
	c := New()
	// No handlers registered — should not error
	err := c.ACTION(nil)
	assert.NoError(t, err)
}

// --- IPC: Queries ---

func TestQuery_Good(t *testing.T) {
	c := New()
	c.RegisterQuery(func(_ *Core, q Query) (any, bool, error) {
		if q == "ping" {
			return "pong", true, nil
		}
		return nil, false, nil
	})
	result, handled, err := c.QUERY("ping")
	assert.NoError(t, err)
	assert.True(t, handled)
	assert.Equal(t, "pong", result)
}

func TestQuery_Unhandled_Good(t *testing.T) {
	c := New()
	c.RegisterQuery(func(_ *Core, q Query) (any, bool, error) {
		return nil, false, nil
	})
	_, handled, err := c.QUERY("unknown")
	assert.NoError(t, err)
	assert.False(t, handled)
}

func TestQueryAll_Good(t *testing.T) {
	c := New()
	c.RegisterQuery(func(_ *Core, _ Query) (any, bool, error) {
		return "a", true, nil
	})
	c.RegisterQuery(func(_ *Core, _ Query) (any, bool, error) {
		return "b", true, nil
	})
	results, err := c.QUERYALL("anything")
	assert.NoError(t, err)
	assert.Len(t, results, 2)
	assert.Contains(t, results, "a")
	assert.Contains(t, results, "b")
}

// --- IPC: Tasks ---

func TestPerform_Good(t *testing.T) {
	c := New()
	c.RegisterTask(func(_ *Core, t Task) (any, bool, error) {
		if t == "compute" {
			return 42, true, nil
		}
		return nil, false, nil
	})
	result, handled, err := c.PERFORM("compute")
	assert.NoError(t, err)
	assert.True(t, handled)
	assert.Equal(t, 42, result)
}
