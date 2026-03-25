// SPDX-License-Identifier: EUPL-1.2

// Synchronisation, locking, and lifecycle snapshots for the Core framework.

package core

import (
	"sync"
)

// Lock is the DTO for a named mutex.
type Lock struct {
	Name  string
	Mutex *sync.RWMutex
	locks *Registry[*sync.RWMutex] // per-Core named mutexes
}

// Lock returns a named Lock, creating the mutex if needed.
// Locks are per-Core — separate Core instances do not share mutexes.
func (c *Core) Lock(name string) *Lock {
	r := c.lock.locks.Get(name)
	if r.OK {
		return &Lock{Name: name, Mutex: r.Value.(*sync.RWMutex)}
	}
	m := &sync.RWMutex{}
	c.lock.locks.Set(name, m)
	return &Lock{Name: name, Mutex: m}
}

// LockEnable marks that the service lock should be applied after initialisation.
func (c *Core) LockEnable(name ...string) {
	c.services.lockEnabled = true
}

// LockApply activates the service lock if it was enabled.
func (c *Core) LockApply(name ...string) {
	if c.services.lockEnabled {
		c.services.Lock()
	}
}

// Startables returns services that have an OnStart function, in registration order.
func (c *Core) Startables() Result {
	if c.services == nil {
		return Result{}
	}
	var out []*Service
	c.services.Each(func(_ string, svc *Service) {
		if svc.OnStart != nil {
			out = append(out, svc)
		}
	})
	return Result{out, true}
}

// Stoppables returns services that have an OnStop function, in registration order.
func (c *Core) Stoppables() Result {
	if c.services == nil {
		return Result{}
	}
	var out []*Service
	c.services.Each(func(_ string, svc *Service) {
		if svc.OnStop != nil {
			out = append(out, svc)
		}
	})
	return Result{out, true}
}
