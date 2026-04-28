package core_test

import (
	. "dappco.re/go"
)

func TestContext_Background_Good(t *T) {
	ctx := Background()

	AssertNil(t, ctx.Err())
	AssertNil(t, ctx.Done())
}

func TestContext_Background_Bad(t *T) {
	AssertNil(t, Background().Value("agent"))
}

func TestContext_Background_Ugly(t *T) {
	AssertNotSameContext(t, Background(), TODO())
}

func TestContext_TODO_Good(t *T) {
	ctx := TODO()

	AssertNil(t, ctx.Err())
	AssertNil(t, ctx.Done())
}

func TestContext_TODO_Bad(t *T) {
	AssertNil(t, TODO().Value("agent"))
}

func TestContext_TODO_Ugly(t *T) {
	AssertNotSameContext(t, TODO(), Background())
}

func TestContext_WithCancel_Good(t *T) {
	ctx, cancel := WithCancel(Background())
	cancel()

	assertContextDone(t, ctx)
}

func TestContext_WithCancel_Bad(t *T) {
	AssertPanics(t, func() { WithCancel(nil) })
}

func TestContext_WithCancel_Ugly(t *T) {
	ctx, cancel := WithCancel(Background())

	AssertNotPanics(t, func() {
		cancel()
		cancel()
	})
	assertContextDone(t, ctx)
}

func TestContext_WithDeadline_Good(t *T) {
	ctx, cancel := WithDeadline(Background(), Now().Add(Second))
	defer cancel()
	deadline, ok := ctx.Deadline()

	AssertTrue(t, ok)
	AssertGreater(t, Until(deadline), Duration(0))
}

func TestContext_WithDeadline_Bad(t *T) {
	AssertPanics(t, func() { WithDeadline(nil, Now()) })
}

func TestContext_WithDeadline_Ugly(t *T) {
	ctx, cancel := WithDeadline(Background(), UnixTime(0))
	defer cancel()

	assertContextDone(t, ctx)
}

func TestContext_WithTimeout_Good(t *T) {
	ctx, cancel := WithTimeout(Background(), Millisecond)
	defer cancel()
	Sleep(2 * Millisecond)

	assertContextDone(t, ctx)
}

func TestContext_WithTimeout_Bad(t *T) {
	AssertPanics(t, func() { WithTimeout(nil, Millisecond) })
}

func TestContext_WithTimeout_Ugly(t *T) {
	ctx, cancel := WithTimeout(Background(), 0)
	defer cancel()

	assertContextDone(t, ctx)
}

func TestContext_WithValue_Good(t *T) {
	ctx := WithValue(Background(), "request_id", "req-123")

	AssertEqual(t, "req-123", ctx.Value("request_id"))
}

func TestContext_WithValue_Bad(t *T) {
	AssertPanics(t, func() { WithValue(Background(), nil, "req-123") })
}

func TestContext_WithValue_Ugly(t *T) {
	ctx := WithValue(Background(), "", "empty-key")

	AssertEqual(t, "empty-key", ctx.Value(""))
}

func assertContextDone(t *T, ctx Context) {
	t.Helper()
	select {
	case <-ctx.Done():
		AssertNotNil(t, ctx.Err())
	default:
		t.Fatalf("context not done")
	}
}

func AssertNotSameContext(t *T, a, b Context) {
	t.Helper()
	AssertFalse(t, a == b)
}
