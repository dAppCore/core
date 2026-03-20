// SPDX-License-Identifier: EUPL-1.2

// Message bus for the Core framework.
// Dispatches actions (fire-and-forget), queries (first responder),
// and tasks (first executor) between registered handlers.

package core

import (
	"slices"
	"sync"
)

// Ipc holds IPC dispatch data.
type Ipc struct {
	ipcMu       sync.RWMutex
	ipcHandlers []func(*Core, Message) Result

	queryMu       sync.RWMutex
	queryHandlers []QueryHandler

	taskMu       sync.RWMutex
	taskHandlers []TaskHandler
}

func (c *Core) Action(msg Message) Result {
	c.ipc.ipcMu.RLock()
	handlers := slices.Clone(c.ipc.ipcHandlers)
	c.ipc.ipcMu.RUnlock()

	for _, h := range handlers {
		if r := h(c, msg); !r.OK {
			return r
		}
	}
	return Result{OK: true}
}

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
	return Result{Value: results, OK: true}
}

func (c *Core) RegisterQuery(handler QueryHandler) {
	c.ipc.queryMu.Lock()
	c.ipc.queryHandlers = append(c.ipc.queryHandlers, handler)
	c.ipc.queryMu.Unlock()
}
