package core

import (
	"errors"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

// captureExit swaps the package-level osExit hook for the duration of the test.
// Returns (captured-code, restore-func). The captured code defaults to -1 so
// tests can distinguish "not called" from "called with 0".
func captureExit(t *testing.T) (codePtr *int, restore func()) {
	t.Helper()
	captured := -1
	codePtr = &captured
	prev := osExit
	osExit = func(code int) { captured = code }
	return codePtr, func() { osExit = prev }
}

func TestExit_Exit_Good(t *testing.T) {
	got, restore := captureExit(t)
	defer restore()

	c := New()
	c.Exit(0)

	assert.Equal(t, 0, *got)
}

func TestExit_Exit_Bad(t *testing.T) {
	// Bad: caller passes a non-zero code via a fatal error path.
	// Recoverable boundary: we observe the captured code, no process death.
	got, restore := captureExit(t)
	defer restore()

	c := New()
	c.Exit(127)

	assert.Equal(t, 127, *got)
}

func TestExit_Exit_Ugly(t *testing.T) {
	// Ugly: Exit called twice (e.g. signal handler races user-triggered exit).
	// Both calls land; second wins. ServiceShutdown is idempotent.
	got, restore := captureExit(t)
	defer restore()

	c := New()
	c.Exit(1)
	c.Exit(2)

	assert.Equal(t, 2, *got)
}

func TestExit_ExitWith_Good(t *testing.T) {
	got, restore := captureExit(t)
	defer restore()

	c := New()
	c.ExitWith(ExitOptions{Code: 5, Timeout: 100 * time.Millisecond})

	assert.Equal(t, 5, *got)
}

func TestExit_ExitWith_Bad(t *testing.T) {
	// Bad: zero timeout = wait forever. With a registered service whose OnStop
	// returns immediately, ServiceShutdown completes; Exit lands.
	got, restore := captureExit(t)
	defer restore()

	c := New()
	c.ExitWith(ExitOptions{Code: 9, Timeout: 0})

	assert.Equal(t, 9, *got)
}

func TestExit_ExitWithNegativeTimeout_Bad(t *testing.T) {
	got, restore := captureExit(t)
	defer restore()

	// Service whose OnStop returns immediately so we don't wait the full 30s.
	c := New()
	c.Service("fast", Service{
		OnStop: func() Result {
			return Result{OK: true}
		},
	})

	start := time.Now()
	c.ExitWith(ExitOptions{Code: 7, Timeout: -1})
	elapsed := time.Since(start)

	assert.Equal(t, 7, *got, "ExitWith should still call osExit with the requested code")
	assert.Less(t, elapsed, 5*time.Second,
		"negative timeout must NOT fall through to wait-forever; default fallback should bound shutdown")
}

func TestExit_ExitWithZeroTimeout_Good(t *testing.T) {
	got, restore := captureExit(t)
	defer restore()

	c := New()
	c.Service("instant", Service{OnStop: func() Result { return Result{OK: true} }})

	c.ExitWith(ExitOptions{Code: 4, Timeout: 0})

	// Timeout==0 → wait forever — but the service's OnStop returns
	// immediately so we shouldn't actually wait. The test asserts the
	// documented zero-means-wait-forever-but-returns-when-ready semantics.
	assert.Equal(t, 4, *got)
}

func TestExit_ExitWith_Ugly(t *testing.T) {
	// Ugly: shutdown takes longer than the timeout. Service blocks for 200ms,
	// timeout is 10ms — process exits with the warning logged, no panic.
	got, restore := captureExit(t)
	defer restore()

	c := New()
	c.Service("slow", Service{OnStop: func() Result {
		time.Sleep(200 * time.Millisecond)
		return Result{OK: true}
	}})
	start := time.Now()
	c.ExitWith(ExitOptions{Code: 3, Timeout: 10 * time.Millisecond})
	elapsed := time.Since(start)

	assert.Equal(t, 3, *got)
	assert.Less(t, elapsed, 200*time.Millisecond,
		"ExitWith must respect the timeout, not wait for slow shutdown")
}

func TestExit_ExitNow_Good(t *testing.T) {
	got, restore := captureExit(t)
	defer restore()

	c := New()
	c.ExitNow(0)

	assert.Equal(t, 0, *got)
}

func TestExit_ExitNow_Bad(t *testing.T) {
	// Bad: ExitNow called from a panic recovery path with non-zero code.
	got, restore := captureExit(t)
	defer restore()

	c := New()
	defer func() {
		if r := recover(); r != nil {
			c.ExitNow(2)
		}
	}()
	func() { panic(errors.New("boom")) }()

	assert.Equal(t, 2, *got)
}

func TestExit_ExitNow_Ugly(t *testing.T) {
	// Ugly: ExitNow does NOT run shutdown — verify the OnStop hook is NOT called.
	got, restore := captureExit(t)
	defer restore()

	stopped := false
	c := New()
	c.Service("hook", Service{OnStop: func() Result {
		stopped = true
		return Result{OK: true}
	}})
	c.ExitNow(4)

	assert.Equal(t, 4, *got)
	assert.False(t, stopped,
		"ExitNow must skip the shutdown chain — OnStop must not run")
}

func TestRun_FailurePath_ShutdownFiresOnce_Bad(t *testing.T) {
	got, restore := captureExit(t)
	defer restore()

	shutdownCount := 0
	c := New()
	c.Service("failing", Service{
		OnStart: func() Result {
			return Result{Value: errors.New("simulated startup failure"), OK: false}
		},
		OnStop: func() Result {
			shutdownCount++
			return Result{OK: true}
		},
	})
	c.Run()

	assert.Equal(t, 1, *got, "Run should call osExit(1) on failure")
	assert.Equal(t, 1, shutdownCount, "OnStop should fire exactly once, not twice")
}

func TestExit_PackageExit_Good(t *testing.T) {
	got, restore := captureExit(t)
	defer restore()

	Exit(0)

	assert.Equal(t, 0, *got)
}

func TestExit_PackageExit_Bad(t *testing.T) {
	// Bad: package-level Exit called with non-zero code from cli error helper.
	got, restore := captureExit(t)
	defer restore()

	Exit(1)

	assert.Equal(t, 1, *got)
}

func TestExit_PackageExit_Ugly(t *testing.T) {
	// Ugly: package-level Exit called repeatedly. Each call lands.
	got, restore := captureExit(t)
	defer restore()

	Exit(1)
	Exit(2)
	Exit(3)

	assert.Equal(t, 3, *got)
}
