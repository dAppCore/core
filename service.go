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
type Service struct {
	Name     string
	Instance any // the raw service instance (for interface discovery)
	Options  Options
	OnStart  func() Result
	OnStop   func() Result
	OnReload func() Result
}

// serviceRegistry holds registered services.
type serviceRegistry struct {
	services    map[string]*Service
	lockEnabled bool
	locked      bool
}

// --- Core service methods ---

// Service gets or registers a service by name.
//
//	c.Service("auth", core.Service{OnStart: startFn})
//	r := c.Service("auth")
func (c *Core) Service(name string, service ...Service) Result {
	if len(service) == 0 {
		c.Lock("srv").Mutex.RLock()
		v, ok := c.services.services[name]
		c.Lock("srv").Mutex.RUnlock()
		return Result{v, ok}
	}

	if name == "" {
		return Result{E("core.Service", "service name cannot be empty", nil), false}
	}

	c.Lock("srv").Mutex.Lock()
	defer c.Lock("srv").Mutex.Unlock()

	if c.services.locked {
		return Result{E("core.Service", Concat("service \"", name, "\" not permitted — registry locked"), nil), false}
	}
	if _, exists := c.services.services[name]; exists {
		return Result{E("core.Service", Join(" ", "service", name, "already registered"), nil), false}
	}

	srv := &service[0]
	srv.Name = name
	c.services.services[name] = srv

	return Result{OK: true}
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

	c.Lock("srv").Mutex.Lock()
	defer c.Lock("srv").Mutex.Unlock()

	if c.services.locked {
		return Result{E("core.RegisterService", Concat("service \"", name, "\" not permitted — registry locked"), nil), false}
	}
	if _, exists := c.services.services[name]; exists {
		return Result{E("core.RegisterService", Join(" ", "service", name, "already registered"), nil), false}
	}

	srv := &Service{Name: name, Instance: instance}

	// Auto-discover lifecycle interfaces
	if s, ok := instance.(Startable); ok {
		srv.OnStart = func() Result {
			if err := s.OnStartup(c.context); err != nil {
				return Result{err, false}
			}
			return Result{OK: true}
		}
	}
	if s, ok := instance.(Stoppable); ok {
		srv.OnStop = func() Result {
			if err := s.OnShutdown(c.context); err != nil {
				return Result{err, false}
			}
			return Result{OK: true}
		}
	}

	c.services.services[name] = srv

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
	svc := r.Value.(*Service)
	if svc.Instance == nil {
		return zero, false
	}
	typed, ok := svc.Instance.(T)
	return typed, ok
}

// Services returns all registered service names.
//
//	names := c.Services()
func (c *Core) Services() []string {
	if c.services == nil {
		return nil
	}
	c.Lock("srv").Mutex.RLock()
	defer c.Lock("srv").Mutex.RUnlock()
	var names []string
	for k := range c.services.services {
		names = append(names, k)
	}
	return names
}
