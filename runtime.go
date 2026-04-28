// SPDX-License-Identifier: EUPL-1.2

// Runtime helpers for the Core framework.
// ServiceRuntime is embedded by consumer services.
// Runtime is the GUI binding container (e.g., Wails).

package core

// --- ServiceRuntime (embedded by consumer services) ---

// ServiceRuntime is embedded in services to provide access to the Core and typed options.
//
//	c := core.New()
//	runtime := core.NewServiceRuntime(c, core.CliOptions{})
//	_ = runtime.Core()
type ServiceRuntime[T any] struct {
	core *Core
	opts T
}

// NewServiceRuntime creates a ServiceRuntime for a service constructor.
//
//	c := core.New()
//	runtime := core.NewServiceRuntime(c, core.CliOptions{})
//	_ = runtime.Options()
func NewServiceRuntime[T any](c *Core, opts T) *ServiceRuntime[T] {
	return &ServiceRuntime[T]{core: c, opts: opts}
}

// Core returns the Core instance this service is registered with.
//
//	c := s.Core()
func (r *ServiceRuntime[T]) Core() *Core { return r.core }

// Options returns the typed options this service was created with.
//
//	opts := s.Options()  // MyOptions{BufferSize: 1024, ...}
func (r *ServiceRuntime[T]) Options() T { return r.opts }

// Config is a shortcut to s.Core().Config().
//
//	host := s.Config().String("database.host")
func (r *ServiceRuntime[T]) Config() *Config { return r.core.Config() }

// --- Lifecycle ---

// ServiceStartup runs OnStart for all registered services that have one.
//
//	c := core.New()
//	r := c.ServiceStartup(Background(), nil)
//	if !r.OK { return r }
func (c *Core) ServiceStartup(ctx Context, options any) Result {
	c.shutdown.Store(false)
	c.context, c.cancel = WithCancel(ctx)
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
//
//	c := core.New()
//	r := c.ServiceShutdown(Background())
//	if !r.OK { return r }
func (c *Core) ServiceShutdown(ctx Context) Result {
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
//
//	r := core.Runtime{Core: core.New()}
//	core.Println(r.ServiceName())
type Runtime struct {
	app  any
	Core *Core
}

// ServiceFactory defines a function that creates a Service.
//
//	factory := func() core.Result {
//	    return core.Result{Value: core.Service{Name: "agent"}, OK: true}
//	}
//	_ = core.ServiceFactory(factory)
type ServiceFactory func() Result

// NewWithFactories creates a Runtime with the provided service factories.
//
//	factories := map[string]core.ServiceFactory{
//	    "agent": func() core.Result { return core.Result{Value: core.Service{}, OK: true} },
//	}
//	r := core.NewWithFactories(nil, factories)
//	if !r.OK { return r }
func NewWithFactories(app any, factories map[string]ServiceFactory) Result {
	c := New(WithOptions(NewOptions(Option{Key: "name", Value: "core"})))
	c.app.Runtime = app

	names := MapKeys(factories)
	SliceSort(names)
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
//
//	r := core.NewRuntime(nil)
//	if !r.OK { return r }
//	runtime := r.Value.(*core.Runtime)
//	_ = runtime.Core
func NewRuntime(app any) Result {
	return NewWithFactories(app, map[string]ServiceFactory{})
}

// ServiceName returns "Core" — the Runtime's service identity.
//
//	runtime := &core.Runtime{Core: core.New()}
//	name := runtime.ServiceName()
//	core.Println(name)
func (r *Runtime) ServiceName() string { return "Core" }

// ServiceStartup starts all services via the embedded Core.
//
//	r := core.Runtime{Core: core.New()}
//	result := r.ServiceStartup(Background(), nil)
//	if !result.OK { return result }
func (r *Runtime) ServiceStartup(ctx Context, options any) Result {
	return r.Core.ServiceStartup(ctx, options)
}

// ServiceShutdown stops all services via the embedded Core.
//
//	r := core.Runtime{Core: core.New()}
//	result := r.ServiceShutdown(Background())
//	if !result.OK { return result }
func (r *Runtime) ServiceShutdown(ctx Context) Result {
	if r.Core != nil {
		return r.Core.ServiceShutdown(ctx)
	}
	return Result{OK: true}
}
