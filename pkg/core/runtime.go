// SPDX-License-Identifier: EUPL-1.2

// Runtime helpers for the Core framework.
// ServiceRuntime is embedded by consumer services.
// Runtime is the GUI binding container (e.g., Wails).

package core

import (
	"context"
	"maps"
	"slices"
)

// --- ServiceRuntime (embedded by consumer services) ---

// ServiceRuntime is embedded in services to provide access to the Core and typed options.
type ServiceRuntime[T any] struct {
	core *Core
	opts T
}

// NewServiceRuntime creates a ServiceRuntime for a service constructor.
func NewServiceRuntime[T any](c *Core, opts T) *ServiceRuntime[T] {
	return &ServiceRuntime[T]{core: c, opts: opts}
}

func (r *ServiceRuntime[T]) Core() *Core     { return r.core }
func (r *ServiceRuntime[T]) Options() T      { return r.opts }
func (r *ServiceRuntime[T]) Config() *Config { return r.core.Config() }

// --- Lifecycle ---

// ServiceStartup runs OnStart for all registered services that have one.
func (c *Core) ServiceStartup(ctx context.Context, options any) Result {
	c.shutdown.Store(false)
	c.context, c.cancel = context.WithCancel(ctx)
	startables := c.Startables()
	if startables.OK {
		for _, s := range startables.Value.([]*Service) {
			if err := ctx.Err(); err != nil {
				return Result{err, false}
			}
			r := s.OnStart()
			if !r.OK {
				return r
			}
		}
	}
	c.ACTION(ActionServiceStartup{})
	return Result{OK: true}
}

// ServiceShutdown drains background tasks, then stops all registered services.
func (c *Core) ServiceShutdown(ctx context.Context) Result {
	c.shutdown.Store(true)
	c.cancel() // signal all context-aware tasks to stop
	c.ACTION(ActionServiceShutdown{})

	// Drain background tasks before stopping services.
	done := make(chan struct{})
	go func() {
		c.waitGroup.Wait()
		close(done)
	}()
	select {
	case <-done:
	case <-ctx.Done():
		return Result{ctx.Err(), false}
	}

	// Stop services
	var firstErr error
	stoppables := c.Stoppables()
	if stoppables.OK {
		for _, s := range stoppables.Value.([]*Service) {
			if err := ctx.Err(); err != nil {
				return Result{err, false}
			}
			r := s.OnStop()
			if !r.OK && firstErr == nil {
				if e, ok := r.Value.(error); ok {
					firstErr = e
				} else {
					firstErr = E("core.ServiceShutdown", Sprint("service OnStop failed: ", r.Value), nil)
				}
			}
		}
	}
	if firstErr != nil {
		return Result{firstErr, false}
	}
	return Result{OK: true}
}

// --- Runtime DTO (GUI binding) ---

// Runtime is the container for GUI runtimes (e.g., Wails).
type Runtime struct {
	app  any
	Core *Core
}

// ServiceFactory defines a function that creates a Service.
type ServiceFactory func() Result

// NewWithFactories creates a Runtime with the provided service factories.
func NewWithFactories(app any, factories map[string]ServiceFactory) Result {
	c := New(Options{{Key: "name", Value: "core"}})
	c.app.Runtime = app

	names := slices.Sorted(maps.Keys(factories))
	for _, name := range names {
		factory := factories[name]
		if factory == nil {
			continue
		}
		r := factory()
		if !r.OK {
			cause, _ := r.Value.(error)
			return Result{E("core.NewWithFactories", Concat("factory \"", name, "\" failed"), cause), false}
		}
		svc, ok := r.Value.(Service)
		if !ok {
			return Result{E("core.NewWithFactories", Concat("factory \"", name, "\" returned non-Service type"), nil), false}
		}
		sr := c.Service(name, svc)
		if !sr.OK {
			return sr
		}
	}
	return Result{&Runtime{app: app, Core: c}, true}
}

// NewRuntime creates a Runtime with no custom services.
func NewRuntime(app any) Result {
	return NewWithFactories(app, map[string]ServiceFactory{})
}

func (r *Runtime) ServiceName() string { return "Core" }
func (r *Runtime) ServiceStartup(ctx context.Context, options any) Result {
	return r.Core.ServiceStartup(ctx, options)
}
func (r *Runtime) ServiceShutdown(ctx context.Context) Result {
	if r.Core != nil {
		return r.Core.ServiceShutdown(ctx)
	}
	return Result{OK: true}
}
