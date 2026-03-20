// SPDX-License-Identifier: EUPL-1.2

// Service registry for the Core framework.
//
// Register a service:
//
//	c.Service("auth", core.Service{})
//
// Get a service:
//
//	r := c.Service("auth")
//	if r.OK { svc := r.Value }

package core

// No imports needed — uses package-level string helpers.

// Service is a managed component with optional lifecycle.
type Service struct {
	Name      string
	Options   Options
	OnStart   func() Result
	OnStop    func() Result
	OnReload  func() Result
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
	if c.services == nil {
		c.services = &serviceRegistry{services: make(map[string]*Service)}
	}

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
