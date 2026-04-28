package core_test

import (
	"context"

	. "dappco.re/go"
)

// --- IPC: Actions ---

type testMessage struct{ payload string }

func TestAction_Good(t *T) {
	c := New()
	var received Message
	c.RegisterAction(func(_ *Core, msg Message) Result {
		received = msg
		return Result{OK: true}
	})
	r := c.ACTION(testMessage{payload: "hello"})
	AssertTrue(t, r.OK)
	AssertEqual(t, testMessage{payload: "hello"}, received)
}

func TestAction_Multiple_Good(t *T) {
	c := New()
	count := 0
	handler := func(_ *Core, _ Message) Result { count++; return Result{OK: true} }
	c.RegisterActions(handler, handler, handler)
	c.ACTION(nil)
	AssertEqual(t, 3, count)
}

func TestAction_None_Good(t *T) {
	c := New()
	// No handlers registered — should succeed
	r := c.ACTION(nil)
	AssertTrue(t, r.OK)
}

func TestAction_Bad_HandlerFails(t *T) {
	c := New()
	c.RegisterAction(func(_ *Core, _ Message) Result {
		return Result{Value: NewError("intentional"), OK: false}
	})
	// ACTION is broadcast — even with a failing handler, dispatch succeeds
	r := c.ACTION(testMessage{payload: "test"})
	AssertTrue(t, r.OK)
}

func TestAction_Ugly_HandlerFailsChainContinues(t *T) {
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
	AssertTrue(t, r.OK)
	AssertEqual(t, []int{1, 2, 3}, order, "all 3 handlers must fire even when handler 2 returns !OK")
}

func TestAction_Ugly_HandlerPanicsChainContinues(t *T) {
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
	AssertTrue(t, r.OK)
	AssertEqual(t, []int{1, 3}, order, "handlers 1 and 3 must fire even when handler 2 panics")
}

// --- IPC: Queries ---

func TestIpc_Query_Good(t *T) {
	c := New()
	c.RegisterQuery(func(_ *Core, q Query) Result {
		if q == "ping" {
			return Result{Value: "pong", OK: true}
		}
		return Result{}
	})
	r := c.QUERY("ping")
	AssertTrue(t, r.OK)
	AssertEqual(t, "pong", r.Value)
}

func TestIpc_Query_Unhandled_Good(t *T) {
	c := New()
	c.RegisterQuery(func(_ *Core, q Query) Result {
		return Result{}
	})
	r := c.QUERY("unknown")
	AssertFalse(t, r.OK)
}

func TestIpc_QueryAll_Good(t *T) {
	c := New()
	c.RegisterQuery(func(_ *Core, _ Query) Result {
		return Result{Value: "a", OK: true}
	})
	c.RegisterQuery(func(_ *Core, _ Query) Result {
		return Result{Value: "b", OK: true}
	})
	r := c.QUERYALL("anything")
	AssertTrue(t, r.OK)
	results := r.Value.([]any)
	AssertLen(t, results, 2)
	AssertContains(t, results, "a")
	AssertContains(t, results, "b")
}

// --- IPC: Named Action Invocation ---

func TestIpc_ActionInvoke_Good(t *T) {
	c := New()
	c.Action("compute", func(_ context.Context, opts Options) Result {
		return Result{Value: 42, OK: true}
	})
	r := c.Action("compute").Run(context.Background(), NewOptions())
	AssertTrue(t, r.OK)
	AssertEqual(t, 42, r.Value)
}
