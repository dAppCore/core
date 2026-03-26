// SPDX-License-Identifier: EUPL-1.2

// Message bus for the Core framework.
// Dispatches actions (fire-and-forget), queries (first responder),
// and tasks (first executor) between registered handlers.

package core

import (
	"slices"
	"sync"
)

// Ipc holds IPC dispatch data and the named action registry.
//
//	ipc := (&core.Ipc{}).New()
type Ipc struct {
	ipcMu       sync.RWMutex
	ipcHandlers []func(*Core, Message) Result

	queryMu       sync.RWMutex
	queryHandlers []QueryHandler

	actions *Registry[*Action] // named action registry
	tasks   *Registry[*Task]   // named task registry
}

// broadcast dispatches a message to all registered IPC handlers.
// Each handler is wrapped in panic recovery. All handlers fire regardless of individual results.
func (c *Core) broadcast(msg Message) Result {
	c.ipc.ipcMu.RLock()
	handlers := slices.Clone(c.ipc.ipcHandlers)
	c.ipc.ipcMu.RUnlock()

	for _, h := range handlers {
		func() {
			defer func() {
				if r := recover(); r != nil {
					Error("ACTION handler panicked", "panic", r)
				}
			}()
			h(c, msg)
		}()
	}
	return Result{OK: true}
}

// Query dispatches a request — first handler to return OK wins.
//
//	r := c.Query(MyQuery{})
func (c *Core) Query(q Query) Result {
	c.ipc.queryMu.RLock()
	handlers := slices.Clone(c.ipc.queryHandlers)
	c.ipc.queryMu.RUnlock()

	for _, h := range handlers {
		r := h(c, q)
		if r.OK {
			return r
		}
	}
	return Result{}
}

// QueryAll dispatches a request — collects all OK responses.
//
//	r := c.QueryAll(countQuery{})
//	results := r.Value.([]any)
func (c *Core) QueryAll(q Query) Result {
	c.ipc.queryMu.RLock()
	handlers := slices.Clone(c.ipc.queryHandlers)
	c.ipc.queryMu.RUnlock()

	var results []any
	for _, h := range handlers {
		r := h(c, q)
		if r.OK && r.Value != nil {
			results = append(results, r.Value)
		}
	}
	return Result{results, true}
}

// RegisterQuery registers a handler for QUERY dispatch.
//
//	c.RegisterQuery(func(_ *core.Core, q core.Query) core.Result { ... })
func (c *Core) RegisterQuery(handler QueryHandler) {
	c.ipc.queryMu.Lock()
	c.ipc.queryHandlers = append(c.ipc.queryHandlers, handler)
	c.ipc.queryMu.Unlock()
}

// --- IPC Registration (handlers) ---

// RegisterAction registers a broadcast handler for ACTION messages.
//
//	c.RegisterAction(func(c *core.Core, msg core.Message) core.Result {
//	    if ev, ok := msg.(AgentCompleted); ok { ... }
//	    return core.Result{OK: true}
//	})
func (c *Core) RegisterAction(handler func(*Core, Message) Result) {
	c.ipc.ipcMu.Lock()
	c.ipc.ipcHandlers = append(c.ipc.ipcHandlers, handler)
	c.ipc.ipcMu.Unlock()
}

// RegisterActions registers multiple broadcast handlers.
func (c *Core) RegisterActions(handlers ...func(*Core, Message) Result) {
	c.ipc.ipcMu.Lock()
	c.ipc.ipcHandlers = append(c.ipc.ipcHandlers, handlers...)
	c.ipc.ipcMu.Unlock()
}

