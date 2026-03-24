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
	mu    sync.Mutex            // protects locks map
	locks map[string]*sync.RWMutex // per-Core named mutexes
}

// Lock returns a named Lock, creating the mutex if needed.
// Locks are per-Core — separate Core instances do not share mutexes.
func (c *Core) Lock(name string) *Lock {
	c.lock.mu.Lock()
	if c.lock.locks == nil {
		c.lock.locks = make(map[string]*sync.RWMutex)
	}
	m, ok := c.lock.locks[name]
	if !ok {
		m = &sync.RWMutex{}
		c.lock.locks[name] = m
	}
	c.lock.mu.Unlock()
	return &Lock{Name: name, Mutex: m}
}

// LockEnable marks that the service lock should be applied after initialisation.
func (c *Core) LockEnable(name ...string) {
	n := "srv"
	if len(name) > 0 {
		n = name[0]
	}
	c.Lock(n).Mutex.Lock()
	defer c.Lock(n).Mutex.Unlock()
	c.services.lockEnabled = true
}

// LockApply activates the service lock if it was enabled.
func (c *Core) LockApply(name ...string) {
	n := "srv"
	if len(name) > 0 {
		n = name[0]
	}
	c.Lock(n).Mutex.Lock()
	defer c.Lock(n).Mutex.Unlock()
	if c.services.lockEnabled {
		c.services.locked = true
	}
}

// Startables returns services that have an OnStart function.
func (c *Core) Startables() Result {
	if c.services == nil {
		return Result{}
	}
	c.Lock("srv").Mutex.RLock()
	defer c.Lock("srv").Mutex.RUnlock()
	var out []*Service
	for _, svc := range c.services.services {
		if svc.OnStart != nil {
			out = append(out, svc)
		}
	}
	return Result{out, true}
}

// Stoppables returns services that have an OnStop function.
func (c *Core) Stoppables() Result {
	if c.services == nil {
		return Result{}
	}
	c.Lock("srv").Mutex.RLock()
	defer c.Lock("srv").Mutex.RUnlock()
	var out []*Service
	for _, svc := range c.services.services {
		if svc.OnStop != nil {
			out = append(out, svc)
		}
	}
	return Result{out, true}
}
