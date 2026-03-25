package core_test

import (
	"context"
	"testing"

	. "dappco.re/go/core"
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

func TestAction_Bad_HandlerFails(t *testing.T) {
	c := New()
	c.RegisterAction(func(_ *Core, _ Message) Result {
		return Result{Value: NewError("intentional"), OK: false}
	})
	// ACTION is broadcast — even with a failing handler, dispatch succeeds
	r := c.ACTION(testMessage{payload: "test"})
	assert.True(t, r.OK)
}

func TestAction_Ugly_HandlerFailsChainContinues(t *testing.T) {
	c := New()
	var order []int
	c.RegisterAction(func(_ *Core, _ Message) Result {
		order = append(order, 1)
		return Result{OK: true}
	})
	c.RegisterAction(func(_ *Core, _ Message) Result {
		order = append(order, 2)
		return Result{Value: NewError("handler 2 fails"), OK: false}
	})
	c.RegisterAction(func(_ *Core, _ Message) Result {
		order = append(order, 3)
		return Result{OK: true}
	})
	r := c.ACTION(testMessage{payload: "test"})
	assert.True(t, r.OK)
	assert.Equal(t, []int{1, 2, 3}, order, "all 3 handlers must fire even when handler 2 returns !OK")
}

func TestAction_Ugly_HandlerPanicsChainContinues(t *testing.T) {
	c := New()
	var order []int
	c.RegisterAction(func(_ *Core, _ Message) Result {
		order = append(order, 1)
		return Result{OK: true}
	})
	c.RegisterAction(func(_ *Core, _ Message) Result {
		panic("handler 2 explodes")
	})
	c.RegisterAction(func(_ *Core, _ Message) Result {
		order = append(order, 3)
		return Result{OK: true}
	})
	r := c.ACTION(testMessage{payload: "test"})
	assert.True(t, r.OK)
	assert.Equal(t, []int{1, 3}, order, "handlers 1 and 3 must fire even when handler 2 panics")
}

// --- IPC: Queries ---

func TestIpc_Query_Good(t *testing.T) {
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

func TestIpc_Query_Unhandled_Good(t *testing.T) {
	c := New()
	c.RegisterQuery(func(_ *Core, q Query) Result {
		return Result{}
	})
	r := c.QUERY("unknown")
	assert.False(t, r.OK)
}

func TestIpc_QueryAll_Good(t *testing.T) {
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

// --- IPC: Named Action Invocation ---

func TestIpc_ActionInvoke_Good(t *testing.T) {
	c := New()
	c.Action("compute", func(_ context.Context, opts Options) Result {
		return Result{Value: 42, OK: true}
	})
	r := c.Action("compute").Run(context.Background(), NewOptions())
	assert.True(t, r.OK)
	assert.Equal(t, 42, r.Value)
}
