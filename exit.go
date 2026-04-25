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
// All four call the unexported osExit() — the singular boundary in core/go where
// the os.Exit syscall is invoked. Consumers never import "os" for this.

package core

import (
	"context"
	"os"
	"time"
)

// osExit is the singular call to os.Exit in core/go.
// Tests override via the testExitCode hook; production wires straight through.
var osExit = os.Exit

// ExitOptions configures graceful exit behaviour.
//
//	c.ExitWith(core.ExitOptions{Code: 1, Timeout: 5 * time.Second})
type ExitOptions struct {
	// Code is the process exit code passed to os.Exit.
	Code int
	// Timeout bounds how long ServiceShutdown may run before the process
	// terminates anyway. Zero means wait forever (legacy behaviour).
	Timeout time.Duration
}

// Exit terminates the process with the given code, after running shutdown hooks.
//
// Default 30s timeout — long enough for graceful database shutdown / file
// flushes, short enough that ops can SIGKILL after waiting (matches systemd
// TimeoutStopSec).
//
//	// fatal error in a signal handler
//	c.Action("signal.received", func(ctx context.Context, opts core.Options) core.Result {
//	    if opts.String("name") == "SIGINT" { c.Exit(0) }
//	    return core.Result{OK: true}
//	})
func (c *Core) Exit(code int) {
	c.ExitWith(ExitOptions{Code: code, Timeout: 30 * time.Second})
}

// ExitWith runs ServiceShutdown with the given timeout, then exits.
// If shutdown does not complete within Timeout, the process exits anyway and
// a warning is logged.
//
//	// daemon with a tighter shutdown budget
//	c.ExitWith(core.ExitOptions{Code: 0, Timeout: 5 * time.Second})
func (c *Core) ExitWith(opts ExitOptions) {
	ctx := context.Background()
	if opts.Timeout > 0 {
		var cancel context.CancelFunc
		ctx, cancel = context.WithTimeout(ctx, opts.Timeout)
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
			"timeout", opts.Timeout, "code", opts.Code)
	}
	osExit(opts.Code)
}

// ExitNow terminates immediately without running shutdown hooks.
// Use only when shutdown is hung or unsafe (e.g. inside a panic the shutdown
// chain may have caused).
// Also valid when shutdown has already run via a defer chain.
//
//	defer func() {
//	    if r := recover(); r != nil { c.ExitNow(2) }
//	}()
func (c *Core) ExitNow(code int) { osExit(code) }

// Exit (package level) terminates immediately without running shutdown hooks.
// For callsites that do not have a *Core reference (e.g. cli error helpers).
// Equivalent to calling ExitNow on a Core instance.
//
//	// cli/pkg/cli/errors.go — no *Core in scope
//	func Fatal(msg string) {
//	    Error(msg)
//	    core.Exit(1)
//	}
func Exit(code int) { osExit(code) }
