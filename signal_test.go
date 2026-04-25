package core_test

import (
	"context"
	"testing"

	. "dappco.re/go/core"
	"github.com/stretchr/testify/assert"
)

func TestSignal_Exists_Good(t *testing.T) {
	// Good: with a registered signal.received action, Exists is true.
	c := New()
	c.Action("signal.received", func(_ context.Context, _ Options) Result {
		return Result{OK: true}
	})
	assert.True(t, c.Signal().Exists())
}

func TestSignal_Exists_Bad(t *testing.T) {
	// Bad: no signal service registered. Exists returns false.
	c := New()
	assert.False(t, c.Signal().Exists())
}

func TestSignal_Exists_Ugly(t *testing.T) {
	// Ugly: a signal.start action is registered but signal.received is not.
	// Exists keys off signal.received specifically — partial registration
	// reports as no service available.
	c := New()
	c.Action("signal.start", func(_ context.Context, _ Options) Result {
		return Result{OK: true}
	})
	assert.False(t, c.Signal().Exists(),
		"Exists must key off signal.received, not just any signal.* action")
}

func TestSignal_Stop_Good(t *testing.T) {
	// Good: signal.stop registered, Stop emits and returns OK.
	c := New()
	called := false
	c.Action("signal.stop", func(_ context.Context, _ Options) Result {
		called = true
		return Result{OK: true}
	})
	r := c.Signal().Stop()
	assert.True(t, r.OK)
	assert.True(t, called)
}

func TestSignal_Stop_Bad(t *testing.T) {
	// Bad: no signal.stop registered. Stop returns Result{OK: false}
	// (permission-by-registration — no handler = no capability).
	c := New()
	r := c.Signal().Stop()
	assert.False(t, r.OK)
}

func TestSignal_Stop_Ugly(t *testing.T) {
	// Ugly: signal.stop registered but handler returns OK: false (refusal).
	// Caller observes the refusal verbatim.
	c := New()
	c.Action("signal.stop", func(_ context.Context, _ Options) Result {
		return Result{OK: false}
	})
	r := c.Signal().Stop()
	assert.False(t, r.OK)
}
