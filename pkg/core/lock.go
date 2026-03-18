// SPDX-License-Identifier: EUPL-1.2

// Synchronisation, locking, and lifecycle snapshots for the Core framework.

package core

import (
	"slices"
	"sync"
)

// package-level mutex infrastructure
var (
	lockMu  sync.Mutex
	lockMap = make(map[string]*sync.RWMutex)
)

// Lock is the DTO for a named mutex.
type Lock struct {
	Name string
	Mu   *sync.RWMutex
}

// Lock returns a named Lock, creating the mutex if needed.
func (c *Core) Lock(name string) *Lock {
	lockMu.Lock()
	m, ok := lockMap[name]
	if !ok {
		m = &sync.RWMutex{}
		lockMap[name] = m
	}
	lockMu.Unlock()
	return &Lock{Name: name, Mu: m}
}

// LockEnable marks that the service lock should be applied after initialisation.
func (c *Core) LockEnable(name ...string) {
	n := "srv"
	if len(name) > 0 {
		n = name[0]
	}
	c.Lock(n).Mu.Lock()
	defer c.Lock(n).Mu.Unlock()
	c.srv.lockEnabled = true
}

// LockApply activates the service lock if it was enabled.
func (c *Core) LockApply(name ...string) {
	n := "srv"
	if len(name) > 0 {
		n = name[0]
	}
	c.Lock(n).Mu.Lock()
	defer c.Lock(n).Mu.Unlock()
	if c.srv.lockEnabled {
		c.srv.locked = true
	}
}

// Startables returns a snapshot of services implementing Startable.
func (c *Core) Startables() []Startable {
	c.Lock("srv").Mu.RLock()
	out := slices.Clone(c.srv.startables)
	c.Lock("srv").Mu.RUnlock()
	return out
}

// Stoppables returns a snapshot of services implementing Stoppable.
func (c *Core) Stoppables() []Stoppable {
	c.Lock("srv").Mu.RLock()
	out := slices.Clone(c.srv.stoppables)
	c.Lock("srv").Mu.RUnlock()
	return out
}
