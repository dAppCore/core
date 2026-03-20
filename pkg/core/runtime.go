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
func (r *ServiceRuntime[T]) Opts() T         { return r.opts }
func (r *ServiceRuntime[T]) Config() *Config { return r.core.Config() }

// --- Lifecycle ---

// ServiceStartup runs OnStart for all registered services that have one.
func (c *Core) ServiceStartup(ctx context.Context, options any) error {
	for _, s := range c.Startables() {
		if err := ctx.Err(); err != nil {
			return err
		}
		r := s.OnStart()
		if !r.OK {
			if err, ok := r.Value.(error); ok {
				return err
			}
		}
	}
	_ = c.ACTION(ActionServiceStartup{})
	return nil
}

// ServiceShutdown runs OnStop for all registered services that have one.
func (c *Core) ServiceShutdown(ctx context.Context) error {
	c.shutdown.Store(true)
	_ = c.ACTION(ActionServiceShutdown{})
	for _, s := range c.Stoppables() {
		if err := ctx.Err(); err != nil {
			return err
		}
		s.OnStop()
	}
	done := make(chan struct{})
	go func() {
		c.wg.Wait()
		close(done)
	}()
	select {
	case <-done:
	case <-ctx.Done():
		return ctx.Err()
	}
	return nil
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
	c := New(Options{{K: "name", V: "core"}})
	c.app.Runtime = app

	names := slices.Sorted(maps.Keys(factories))
	for _, name := range names {
		factory := factories[name]
		if factory == nil {
			continue
		}
		r := factory()
		if !r.OK {
			continue
		}
		if svc, ok := r.Value.(Service); ok {
			c.Service(name, svc)
		}
	}
	return Result{Value: &Runtime{app: app, Core: c}, OK: true}
}

// NewRuntime creates a Runtime with no custom services.
func NewRuntime(app any) Result {
	return NewWithFactories(app, map[string]ServiceFactory{})
}

func (r *Runtime) ServiceName() string { return "Core" }
func (r *Runtime) ServiceStartup(ctx context.Context, options any) error {
	return r.Core.ServiceStartup(ctx, options)
}
func (r *Runtime) ServiceShutdown(ctx context.Context) error {
	if r.Core != nil {
		return r.Core.ServiceShutdown(ctx)
	}
	return nil
}
