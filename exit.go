// SPDX-License-Identifier: EUPL-1.2

// Process termination with graceful shutdown.
//
// Always prefer returning errors from RunE() over calling Exit. Use Exit only
// when you cannot return: signal handlers, panic recovery, or fatal errors deep
// in callbacks where the caller chain has no place for an error.
//
// Surface (in order of preference):
//
//	c.Exit(0)               // graceful, runs ServiceShutdown, 30s timeout
//	c.ExitWith(opts)        // graceful, custom timeout
//	c.ExitNow(2)            // skip shutdown, immediate (panic recovery only)
//	core.Exit(1)            // package-level, no *Core in scope (cli error helpers)
//
// All four call core.Exit after graceful shutdown when needed. Consumers never
// import "os" for this.

package core

// ExitOptions configures graceful exit behaviour.
//
//	c.ExitWith(core.ExitOptions{Code: 1, Timeout: 5 * core.Second})
type ExitOptions struct {
	// Code is the process exit code passed to core.Exit.
	Code int
	// Timeout bounds how long ServiceShutdown may run before the process
	// terminates anyway. Zero means wait forever (legacy behaviour).
	Timeout Duration
}

// Exit terminates the process with the given code, after running shutdown hooks.
//
// Default 30s timeout — long enough for graceful database shutdown / file
// flushes, short enough that ops can SIGKILL after waiting (matches systemd
// TimeoutStopSec).
//
//	// fatal error in a signal handler
//	c.Action("signal.received", func(ctx Context, opts core.Options) core.Result {
//	    if opts.String("name") == "SIGINT" { c.Exit(0) }
//	    return core.Result{OK: true}
//	})
func (c *Core) Exit(code int) {
	c.ExitWith(ExitOptions{Code: code, Timeout: 30 * Second})
}

// ExitWith runs ServiceShutdown with the given timeout, then exits.
// If shutdown does not complete within Timeout, the process exits anyway and
// a warning is logged.
//
//	// daemon with a tighter shutdown budget
//	c.ExitWith(core.ExitOptions{Code: 0, Timeout: 5 * core.Second})
func (c *Core) ExitWith(opts ExitOptions) {
	ctx := Background()
	timeout := opts.Timeout
	if timeout != 0 {
		var cancel CancelFunc
		ctx, cancel = WithTimeout(ctx, timeout)
		defer cancel()
	}
	done := make(chan struct{})
	go func() {
		_ = c.ServiceShutdown(ctx)
		close(done)
	}()
	select {
	case <-done:
		// shutdown completed
	case <-ctx.Done():
		Warn("exit timeout, forcing immediate termination",
			"timeout", timeout, "code", opts.Code)
	}
	Exit(opts.Code)
}

// ExitNow terminates immediately without running shutdown hooks.
// Use only when shutdown is hung or unsafe (e.g. inside a panic the shutdown
// chain may have caused).
//
//	defer func() {
//	    if r := recover(); r != nil { c.ExitNow(2) }
//	}()
func (c *Core) ExitNow(code int) { Exit(code) }
