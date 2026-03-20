// SPDX-License-Identifier: EUPL-1.2

// Service registry, lifecycle tracking, and runtime helpers for the Core framework.

package core

import "fmt"

// --- Service Registry DTO ---

// Service holds service registry data.
type Service struct {
	Services    map[string]any
	startables  []Startable
	stoppables  []Stoppable
	lockEnabled bool
	locked      bool
}


// --- Core service methods ---

// Service gets or registers a service.
//
//	c.Service()                  // returns *Service
//	c.Service("auth")            // returns the "auth" service
//	c.Service("auth", myService) // registers "auth"
func (c *Core) Service(args ...any) any {
	switch len(args) {
	case 0:
		return c.service
	case 1:
		name, _ := args[0].(string)
		c.Lock("srv").Mu.RLock()
		v, ok := c.service.Services[name]
		c.Lock("srv").Mu.RUnlock()
		if !ok {
			return nil
		}
		return v
	default:
		name, _ := args[0].(string)
		if name == "" {
			return E("core.Service", "service name cannot be empty", nil)
		}
		c.Lock("srv").Mu.Lock()
		defer c.Lock("srv").Mu.Unlock()
		if c.service.locked {
			return E("core.Service", fmt.Sprintf("service %q is not permitted by the serviceLock setting", name), nil)
		}
		if _, exists := c.service.Services[name]; exists {
			return E("core.Service", fmt.Sprintf("service %q already registered", name), nil)
		}
		svc := args[1]
		if c.service.Services == nil {
			c.service.Services = make(map[string]any)
		}
		c.service.Services[name] = svc
		if st, ok := svc.(Startable); ok {
			c.service.startables = append(c.service.startables, st)
		}
		if st, ok := svc.(Stoppable); ok {
			c.service.stoppables = append(c.service.stoppables, st)
		}
		if lp, ok := svc.(LocaleProvider); ok {
			c.i18n.AddLocales(lp.Locales())
		}
		return nil
	}
}

