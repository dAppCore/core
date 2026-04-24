// SPDX-License-Identifier: EUPL-1.2

// Synchronisation, locking, and lifecycle snapshots for the Core framework.

package core

import (
	"sync"
)

// Lock is the DTO for a named mutex.
//
// Mutex is the backing sync.RWMutex.
//
// Deprecated: direct field access forces consumers to import "sync".
// Use the Lock/Unlock/RLock/RUnlock/TryLock methods instead. Removed in v0.9.0.
type Lock struct {
	Name  string
	Mutex *sync.RWMutex
	locks *Registry[*sync.RWMutex] // per-Core named mutexes
}

// Lock returns a named Lock, creating the mutex if needed.
// Locks are per-Core — separate Core instances do not share mutexes.
//
//	l := c.Lock("drain")
//	l.Lock(); defer l.Unlock()
func (c *Core) Lock(name string) *Lock {
	r := c.lock.locks.Get(name)
	if r.OK {
		return &Lock{Name: name, Mutex: r.Value.(*sync.RWMutex)}
	}
	m := &sync.RWMutex{}
	c.lock.locks.Set(name, m)
	return &Lock{Name: name, Mutex: m}
}

// Lock acquires the named mutex for write.
//
//	c.Lock("drain").Lock()
//	defer c.Lock("drain").Unlock()
func (l *Lock) Lock() { l.Mutex.Lock() }

// Unlock releases the named mutex from write.
//
//	c.Lock("drain").Unlock()
func (l *Lock) Unlock() { l.Mutex.Unlock() }

// RLock acquires the named mutex for read.
//
//	c.Lock("cache").RLock()
//	defer c.Lock("cache").RUnlock()
func (l *Lock) RLock() { l.Mutex.RLock() }

// RUnlock releases the named mutex from read.
//
//	c.Lock("cache").RUnlock()
func (l *Lock) RUnlock() { l.Mutex.RUnlock() }

// TryLock attempts to acquire the write mutex without blocking.
// Returns Result{OK: true} when acquired, Result{OK: false} when held.
//
//	if c.Lock("drain").TryLock().OK {
//	    defer c.Lock("drain").Unlock()
//	    // ...
//	}
func (l *Lock) TryLock() Result {
	if l.Mutex.TryLock() {
		return Result{OK: true}
	}
	return Result{OK: false}
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
