package core_test

import . "dappco.re/go"

func TestSignal_Exists_Good(t *T) {
	// Good: with a registered signal.received action, Exists is true.
	c := New()
	c.Action("signal.received", func(_ Context, _ Options) Result {
		return Result{OK: true}
	})
	AssertTrue(t, c.Signal().Exists())
}

func TestSignal_Exists_Bad(t *T) {
	// Bad: no signal service registered. Exists returns false.
	c := New()
	AssertFalse(t, c.Signal().Exists())
}

func TestSignal_Exists_Ugly(t *T) {
	// Ugly: a signal.start action is registered but signal.received is not.
	// Exists keys off signal.received specifically — partial registration
	// reports as no service available.
	c := New()
	c.Action("signal.start", func(_ Context, _ Options) Result {
		return Result{OK: true}
	})
	AssertFalse(t, c.Signal().Exists(),
		"Exists must key off signal.received, not just any signal.* action")
}

func TestSignal_Stop_Good(t *T) {
	// Good: signal.stop registered, Stop emits and returns OK.
	c := New()
	called := false
	c.Action("signal.stop", func(_ Context, _ Options) Result {
		called = true
		return Result{OK: true}
	})
	r := c.Signal().Stop()
	AssertTrue(t, r.OK)
	AssertTrue(t, called)
}

func TestSignal_Stop_Bad(t *T) {
	// Bad: no signal.stop registered. Stop returns Result{OK: false}
	// (permission-by-registration — no handler = no capability).
	c := New()
	r := c.Signal().Stop()
	AssertFalse(t, r.OK)
}

func TestSignal_Stop_Ugly(t *T) {
	// Ugly: signal.stop registered but handler returns OK: false (refusal).
	// Caller observes the refusal verbatim.
	c := New()
	c.Action("signal.stop", func(_ Context, _ Options) Result {
		return Result{OK: false}
	})
	r := c.Signal().Stop()
	AssertFalse(t, r.OK)
}

func TestSignal_Core_Signal_Good(t *T) {
	c := New()

	signal := c.Signal()

	AssertNotNil(t, signal)
	AssertFalse(t, signal.Exists())
}

func TestSignal_Core_Signal_Bad(t *T) {
	c := New()

	first := c.Signal()
	second := c.Signal()

	AssertNotNil(t, first)
	AssertNotNil(t, second)
}

func TestSignal_Core_Signal_Ugly(t *T) {
	c := New()
	c.Action("signal.received", func(_ Context, _ Options) Result {
		return Result{OK: true}
	})

	signal := c.Signal()

	AssertTrue(t, signal.Exists())
}

func TestSignal_Signal_Exists_Good(t *T) {
	c := New()
	c.Action("signal.received", func(_ Context, _ Options) Result {
		return Result{OK: true}
	})

	AssertTrue(t, c.Signal().Exists())
}

func TestSignal_Signal_Exists_Bad(t *T) {
	c := New()

	AssertFalse(t, c.Signal().Exists())
}

func TestSignal_Signal_Exists_Ugly(t *T) {
	c := New()
	c.Action("signal.stop", func(_ Context, _ Options) Result {
		return Result{OK: true}
	})

	AssertFalse(t, c.Signal().Exists())
}

func TestSignal_Signal_Stop_Good(t *T) {
	c := New()
	called := false
	c.Action("signal.stop", func(_ Context, _ Options) Result {
		called = true
		return Result{OK: true}
	})

	r := c.Signal().Stop()

	AssertTrue(t, r.OK)
	AssertTrue(t, called)
}

func TestSignal_Signal_Stop_Bad(t *T) {
	c := New()

	r := c.Signal().Stop()

	AssertFalse(t, r.OK)
}

func TestSignal_Signal_Stop_Ugly(t *T) {
	c := New()
	c.Action("signal.stop", func(_ Context, _ Options) Result {
		return Result{Value: E("signal.stop", "agent refused shutdown", nil), OK: false}
	})

	r := c.Signal().Stop()

	AssertFalse(t, r.OK)
	AssertContains(t, r.Error(), "agent refused shutdown")
}
