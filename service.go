// SPDX-License-Identifier: EUPL-1.2

// Service registry for the Core framework.
//
// Register a service (DTO with lifecycle hooks):
//
//	c.Service("auth", core.Service{OnStart: startFn})
//
// Register a service instance (auto-discovers Startable/Stoppable/HandleIPCEvents):
//
//	c.RegisterService("display", displayInstance)
//
// Get a service:
//
//	r := c.Service("auth")
//	if r.OK { svc := r.Value }

package core


// Service is a managed component with optional lifecycle.
//
//	svc := core.Service{Name: "agent", OnStart: func() core.Result { return core.Result{OK: true} }}
//	core.New().Service("agent", svc)
type Service struct {
	Name     string
	Instance any // the raw service instance (for interface discovery)
	Options  Options
	OnStart  func() Result
	OnStop   func() Result
	OnReload func() Result
}

// ServiceRegistry holds registered services. Embeds Registry[*Service]
// for thread-safe named storage with insertion order.
//
//	registry := &core.ServiceRegistry{Registry: core.NewRegistry[*core.Service]()}
//	registry.Set("agent", &core.Service{Name: "agent"})
type ServiceRegistry struct {
	*Registry[*Service]
	lockEnabled bool
}

// --- Core service methods ---

// Service gets or registers a service by name.
//
//	c.Service("auth", core.Service{OnStart: startFn})
//	r := c.Service("auth")
func (c *Core) Service(name string, service ...Service) Result {
	if len(service) == 0 {
		r := c.services.Get(name)
		if !r.OK {
			return Result{}
		}
		svc := r.Value.(*Service)
		// Return the instance if available, otherwise the Service DTO
		if svc.Instance != nil {
			return Result{svc.Instance, true}
		}
		return Result{svc, true}
	}

	if name == "" {
		return Result{E("core.Service", "service name cannot be empty", nil), false}
	}

	if c.services.Locked() {
		return Result{E("core.Service", Concat("service \"", name, "\" not permitted — registry locked"), nil), false}
	}
	if c.services.Has(name) {
		return Result{E("core.Service", Join(" ", "service", name, "already registered"), nil), false}
	}

	srv := &service[0]
	srv.Name = name
	return c.services.Set(name, srv)
}

// RegisterService registers a service instance by name.
// Auto-discovers Startable, Stoppable, and HandleIPCEvents interfaces
// on the instance and wires them into the lifecycle and IPC bus.
//
//	c.RegisterService("display", displayInstance)
func (c *Core) RegisterService(name string, instance any) Result {
	if name == "" {
		return Result{E("core.RegisterService", "service name cannot be empty", nil), false}
	}

	if c.services.Locked() {
		return Result{E("core.RegisterService", Concat("service \"", name, "\" not permitted — registry locked"), nil), false}
	}
	if c.services.Has(name) {
		return Result{E("core.RegisterService", Join(" ", "service", name, "already registered"), nil), false}
	}

	srv := &Service{Name: name, Instance: instance}

	// Auto-discover lifecycle interfaces
	if s, ok := instance.(Startable); ok {
		srv.OnStart = func() Result {
			return s.OnStartup(c.context)
		}
	}
	if s, ok := instance.(Stoppable); ok {
		srv.OnStop = func() Result {
			return s.OnShutdown(Background())
		}
	}

	c.services.Set(name, srv)

	// Auto-discover IPC handler
	if handler, ok := instance.(interface {
		HandleIPCEvents(*Core, Message) Result
	}); ok {
		c.ipc.ipcMu.Lock()
		c.ipc.ipcHandlers = append(c.ipc.ipcHandlers, handler.HandleIPCEvents)
		c.ipc.ipcMu.Unlock()
	}

	return Result{OK: true}
}

// ServiceFor retrieves a registered service by name and asserts its type.
//
//	prep, ok := core.ServiceFor[*agentic.PrepSubsystem](c, "agentic")
func ServiceFor[T any](c *Core, name string) (T, bool) {
	var zero T
	r := c.Service(name)
	if !r.OK {
		return zero, false
	}
	typed, ok := r.Value.(T)
	return typed, ok
}

// MustServiceFor retrieves a registered service by name and asserts its type.
// Panics if the service is not found or the type assertion fails.
//
//	cli := core.MustServiceFor[*Cli](c, "cli")
func MustServiceFor[T any](c *Core, name string) T {
	v, ok := ServiceFor[T](c, name)
	if !ok {
		panic(E("core.MustServiceFor", Sprintf("service %q not found or wrong type", name), nil))
	}
	return v
}

// Services returns all registered service names in registration order.
//
//	names := c.Services()
func (c *Core) Services() []string {
	if c.services == nil {
		return nil
	}
	return c.services.Names()
}
