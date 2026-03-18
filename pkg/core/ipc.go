// SPDX-License-Identifier: EUPL-1.2

// Message bus for the Core framework.
// Dispatches actions (fire-and-forget), queries (first responder),
// and tasks (first executor) between registered handlers.

package core

import (
	"errors"
	"slices"
	"sync"
)

// Ipc owns action, query, and task dispatch between services.
type Ipc struct {
	core *Core

	ipcMu       sync.RWMutex
	ipcHandlers []func(*Core, Message) error

	queryMu       sync.RWMutex
	queryHandlers []QueryHandler

	taskMu       sync.RWMutex
	taskHandlers []TaskHandler
}

// NewBus creates an empty message bus bound to the given Core.
func NewBus(c *Core) *Ipc {
	return &Ipc{core: c}
}

// Action dispatches a message to all registered IPC handlers.
func (b *Ipc) Action(msg Message) error {
	b.ipcMu.RLock()
	handlers := slices.Clone(b.ipcHandlers)
	b.ipcMu.RUnlock()

	var agg error
	for _, h := range handlers {
		if err := h(b.core, msg); err != nil {
			agg = errors.Join(agg, err)
		}
	}
	return agg
}

// RegisterAction adds a single IPC handler.
func (b *Ipc) RegisterAction(handler func(*Core, Message) error) {
	b.ipcMu.Lock()
	b.ipcHandlers = append(b.ipcHandlers, handler)
	b.ipcMu.Unlock()
}

// RegisterActions adds multiple IPC handlers.
func (b *Ipc) RegisterActions(handlers ...func(*Core, Message) error) {
	b.ipcMu.Lock()
	b.ipcHandlers = append(b.ipcHandlers, handlers...)
	b.ipcMu.Unlock()
}

// Query dispatches a query to handlers until one responds.
func (b *Ipc) Query(q Query) (any, bool, error) {
	b.queryMu.RLock()
	handlers := slices.Clone(b.queryHandlers)
	b.queryMu.RUnlock()

	for _, h := range handlers {
		result, handled, err := h(b.core, q)
		if handled {
			return result, true, err
		}
	}
	return nil, false, nil
}

// QueryAll dispatches a query to all handlers and collects all responses.
func (b *Ipc) QueryAll(q Query) ([]any, error) {
	b.queryMu.RLock()
	handlers := slices.Clone(b.queryHandlers)
	b.queryMu.RUnlock()

	var results []any
	var agg error
	for _, h := range handlers {
		result, handled, err := h(b.core, q)
		if err != nil {
			agg = errors.Join(agg, err)
		}
		if handled && result != nil {
			results = append(results, result)
		}
	}
	return results, agg
}

// RegisterQuery adds a query handler.
func (b *Ipc) RegisterQuery(handler QueryHandler) {
	b.queryMu.Lock()
	b.queryHandlers = append(b.queryHandlers, handler)
	b.queryMu.Unlock()
}

// Perform dispatches a task to handlers until one executes it.
func (b *Ipc) Perform(t Task) (any, bool, error) {
	b.taskMu.RLock()
	handlers := slices.Clone(b.taskHandlers)
	b.taskMu.RUnlock()

	for _, h := range handlers {
		result, handled, err := h(b.core, t)
		if handled {
			return result, true, err
		}
	}
	return nil, false, nil
}

// RegisterTask adds a task handler.
func (b *Ipc) RegisterTask(handler TaskHandler) {
	b.taskMu.Lock()
	b.taskHandlers = append(b.taskHandlers, handler)
	b.taskMu.Unlock()
}
