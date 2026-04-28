// SPDX-License-Identifier: EUPL-1.2

// Context primitives — re-exports of Go's standard context package as
// Core types, so consumers never need to write `import "context"`.
//
//	ctx, cancel := core.WithTimeout(core.Background(), 5*core.Second)
//	defer cancel()
//	r := c.Action("git.log").Run(ctx, opts)
package core

import "context"

// Context is the cancellation/deadline carrier passed to Action handlers,
// Service lifecycle methods, and any I/O-bound Core call. Alias of
// context.Context — interchangeable at call sites.
//
//	ctx := core.Background()
//	r := c.Action("service.startup").Run(ctx, opts)
type Context = context.Context

// CancelFunc cancels a Context derived via WithCancel/WithTimeout/WithDeadline.
// Calling it releases resources associated with the Context.
//
//	ctx, cancel := core.WithCancel(core.Background())
//	defer cancel()
type CancelFunc = context.CancelFunc

// Background returns the root Context — never cancelled, no deadline,
// no values. Use as the top-level Context for the application.
//
//	ctx := core.Background()
//	r := c.Action("agent.dispatch").Run(ctx, opts)
func Background() Context {
	return context.Background()
}

// TODO returns a Context for code paths where the proper Context is
// not yet known — placeholder during refactors. Replace with a real
// Context before production.
//
//	ctx := core.TODO()
//	// TODO: wire a real Context once the caller is plumbed.
func TODO() Context {
	return context.TODO()
}

// WithCancel returns a copy of parent with a new cancellation channel.
// Calling cancel cancels the returned Context but not the parent.
//
//	ctx, cancel := core.WithCancel(core.Background())
//	defer cancel()
//	go worker(ctx)
func WithCancel(parent Context) (Context, CancelFunc) {
	return context.WithCancel(parent)
}

// WithTimeout returns a copy of parent that cancels after d elapses or
// cancel is called, whichever comes first.
//
//	ctx, cancel := core.WithTimeout(core.Background(), 5*core.Second)
//	defer cancel()
//	r := core.HTTPGet("https://lthn.sh/health").WithContext(ctx)
func WithTimeout(parent Context, d Duration) (Context, CancelFunc) {
	return context.WithTimeout(parent, d)
}

// WithDeadline returns a copy of parent that cancels at deadline or
// when cancel is called, whichever comes first.
//
//	deadline := core.Now().Add(2 * core.Minute)
//	ctx, cancel := core.WithDeadline(core.Background(), deadline)
//	defer cancel()
func WithDeadline(parent Context, deadline Time) (Context, CancelFunc) {
	return context.WithDeadline(parent, deadline)
}

// WithValue returns a copy of parent carrying key=val. Use sparingly —
// prefer explicit parameters over Context-carried values for everything
// except request-scoped data (request IDs, auth tokens).
//
//	type requestIDKey struct{}
//	ctx := core.WithValue(core.Background(), requestIDKey{}, "req-12345")
//	id := ctx.Value(requestIDKey{}).(string)
func WithValue(parent Context, key, val any) Context {
	return context.WithValue(parent, key, val)
}
