// SPDX-License-Identifier: EUPL-1.2

// Runtime helpers for the Core framework.
// ServiceRuntime is embedded by consumer services.
// Runtime is the GUI binding container (e.g., Wails).

package core

import (
	"context"
	"errors"
	"fmt"
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

func (r *ServiceRuntime[T]) Core() *Core  { return r.core }
func (r *ServiceRuntime[T]) Opts() T      { return r.opts }
func (r *ServiceRuntime[T]) Config() *Config { return r.core.Config() }

// --- Lifecycle ---

// ServiceStartup runs the startup lifecycle for all registered services.
func (c *Core) ServiceStartup(ctx context.Context, options any) error {
	startables := c.Startables()
	var agg error
	for _, s := range startables {
		if err := ctx.Err(); err != nil {
			return errors.Join(agg, err)
		}
		if err := s.OnStartup(ctx); err != nil {
			agg = errors.Join(agg, err)
		}
	}
	if err := c.ACTION(ActionServiceStartup{}); err != nil {
		agg = errors.Join(agg, err)
	}
	return agg
}

// ServiceShutdown runs the shutdown lifecycle for all registered services.
func (c *Core) ServiceShutdown(ctx context.Context) error {
	c.shutdown.Store(true)
	var agg error
	if err := c.ACTION(ActionServiceShutdown{}); err != nil {
		agg = errors.Join(agg, err)
	}
	stoppables := c.Stoppables()
	for _, s := range slices.Backward(stoppables) {
		if err := ctx.Err(); err != nil {
			agg = errors.Join(agg, err)
			break
		}
		if err := s.OnShutdown(ctx); err != nil {
			agg = errors.Join(agg, err)
		}
	}
	done := make(chan struct{})
	go func() {
		c.wg.Wait()
		close(done)
	}()
	select {
	case <-done:
	case <-ctx.Done():
		agg = errors.Join(agg, ctx.Err())
	}
	return agg
}

// --- Runtime DTO (GUI binding) ---

// Runtime is the container for GUI runtimes (e.g., Wails).
type Runtime struct {
	app  any
	Core *Core
}

// ServiceFactory defines a function that creates a service instance.
type ServiceFactory func() (any, error)

// NewWithFactories creates a Runtime with the provided service factories.
func NewWithFactories(app any, factories map[string]ServiceFactory) (*Runtime, error) {
	coreOpts := []Option{WithApp(app)}
	names := slices.Sorted(maps.Keys(factories))
	for _, name := range names {
		factory := factories[name]
		if factory == nil {
			return nil, E("core.NewWithFactories", fmt.Sprintf("factory is nil for service %q", name), nil)
		}
		svc, err := factory()
		if err != nil {
			return nil, E("core.NewWithFactories", fmt.Sprintf("failed to create service %q", name), err)
		}
		svcCopy := svc
		coreOpts = append(coreOpts, WithName(name, func(c *Core) (any, error) { return svcCopy, nil }))
	}
	coreInstance, err := New(coreOpts...)
	if err != nil {
		return nil, err
	}
	return &Runtime{app: app, Core: coreInstance}, nil
}

// NewRuntime creates a Runtime with no custom services.
func NewRuntime(app any) (*Runtime, error) {
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
