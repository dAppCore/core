// SPDX-License-Identifier: EUPL-1.2

// OS signal handling — consumer-facing Core surface.
//
// Signal events are emitted as actions. Consumers subscribe via c.Action():
//
//	c.Action("signal.received", func(ctx context.Context, opts core.Options) core.Result {
//	    name := opts.String("name")  // "SIGINT", "SIGTERM", "SIGHUP"
//	    switch name {
//	    case "SIGINT", "SIGTERM":
//	        c.Exit(0)
//	    case "SIGHUP":
//	        c.Config().Reload()
//	    }
//	    return core.Result{OK: true}
//	})
//
// If no signal service is registered (typically by go-process), no actions fire
// and consumers do not observe signals — permission-by-registration, mirroring
// the Process accessor pattern.
//
// Action contract:
//
//	signal.received  service → consumers   {name: string, value: int}
//	signal.start     consumer → service    {signals: []string}
//	signal.stop      consumer → service    {}

package core

// Signal is the Core primitive for OS signal handling.
//
//	if c.Signal().Exists() { /* signals will be observed */ }
type Signal struct {
	core *Core
}

// Signal returns the signal-handling primitive.
//
//	c.Signal().Stop()
func (c *Core) Signal() *Signal { return &Signal{core: c} }

// Stop instructs the signal service to unsubscribe from OS notifications.
// Idempotent. The service shutdown chain calls this automatically.
//
//	c.Signal().Stop()
func (s *Signal) Stop() Result {
	return s.core.Action("signal.stop").Run(s.core.Context(), NewOptions())
}

// Exists reports whether a signal service is registered (and therefore whether
// consumers can expect signal.received broadcasts).
//
//	if !c.Signal().Exists() {
//	    Warn("signal handling unavailable — go-process not registered")
//	}
func (s *Signal) Exists() bool {
	return s.core.Action("signal.received").Exists()
}
