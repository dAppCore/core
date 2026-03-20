package core_test

import (
	"testing"

	. "dappco.re/go/core/pkg/core"
	"github.com/stretchr/testify/assert"
)

// --- IPC: Actions ---

type testMessage struct{ payload string }

func TestAction_Good(t *testing.T) {
	c := New()
	var received Message
	c.RegisterAction(func(_ *Core, msg Message) Result {
		received = msg
		return Result{OK: true}
	})
	r := c.ACTION(testMessage{payload: "hello"})
	assert.True(t, r.OK)
	assert.Equal(t, testMessage{payload: "hello"}, received)
}

func TestAction_Multiple_Good(t *testing.T) {
	c := New()
	count := 0
	handler := func(_ *Core, _ Message) Result { count++; return Result{OK: true} }
	c.RegisterActions(handler, handler, handler)
	c.ACTION(nil)
	assert.Equal(t, 3, count)
}

func TestAction_None_Good(t *testing.T) {
	c := New()
	// No handlers registered — should succeed
	r := c.ACTION(nil)
	assert.True(t, r.OK)
}

// --- IPC: Queries ---

func TestQuery_Good(t *testing.T) {
	c := New()
	c.RegisterQuery(func(_ *Core, q Query) Result {
		if q == "ping" {
			return Result{Value: "pong", OK: true}
		}
		return Result{}
	})
	r := c.QUERY("ping")
	assert.True(t, r.OK)
	assert.Equal(t, "pong", r.Value)
}

func TestQuery_Unhandled_Good(t *testing.T) {
	c := New()
	c.RegisterQuery(func(_ *Core, q Query) Result {
		return Result{}
	})
	r := c.QUERY("unknown")
	assert.False(t, r.OK)
}

func TestQueryAll_Good(t *testing.T) {
	c := New()
	c.RegisterQuery(func(_ *Core, _ Query) Result {
		return Result{Value: "a", OK: true}
	})
	c.RegisterQuery(func(_ *Core, _ Query) Result {
		return Result{Value: "b", OK: true}
	})
	r := c.QUERYALL("anything")
	assert.True(t, r.OK)
	results := r.Value.([]any)
	assert.Len(t, results, 2)
	assert.Contains(t, results, "a")
	assert.Contains(t, results, "b")
}

// --- IPC: Tasks ---

func TestPerform_Good(t *testing.T) {
	c := New()
	c.RegisterTask(func(_ *Core, t Task) Result {
		if t == "compute" {
			return Result{Value: 42, OK: true}
		}
		return Result{}
	})
	r := c.PERFORM("compute")
	assert.True(t, r.OK)
	assert.Equal(t, 42, r.Value)
}
