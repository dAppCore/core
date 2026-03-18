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

// Ipc holds IPC dispatch data.
type Ipc struct {
	ipcMu       sync.RWMutex
	ipcHandlers []func(*Core, Message) error

	queryMu       sync.RWMutex
	queryHandlers []QueryHandler

	taskMu       sync.RWMutex
	taskHandlers []TaskHandler
}

func (c *Core) Action(msg Message) error {
	c.ipc.ipcMu.RLock()
	handlers := slices.Clone(c.ipc.ipcHandlers)
	c.ipc.ipcMu.RUnlock()

	var agg error
	for _, h := range handlers {
		if err := h(c, msg); err != nil {
			agg = errors.Join(agg, err)
		}
	}
	return agg
}

func (c *Core) Query(q Query) (any, bool, error) {
	c.ipc.queryMu.RLock()
	handlers := slices.Clone(c.ipc.queryHandlers)
	c.ipc.queryMu.RUnlock()

	for _, h := range handlers {
		result, handled, err := h(c, q)
		if handled {
			return result, true, err
		}
	}
	return nil, false, nil
}

func (c *Core) QueryAll(q Query) ([]any, error) {
	c.ipc.queryMu.RLock()
	handlers := slices.Clone(c.ipc.queryHandlers)
	c.ipc.queryMu.RUnlock()

	var results []any
	var agg error
	for _, h := range handlers {
		result, handled, err := h(c, q)
		if err != nil {
			agg = errors.Join(agg, err)
		}
		if handled && result != nil {
			results = append(results, result)
		}
	}
	return results, agg
}

func (c *Core) RegisterQuery(handler QueryHandler) {
	c.ipc.queryMu.Lock()
	c.ipc.queryHandlers = append(c.ipc.queryHandlers, handler)
	c.ipc.queryMu.Unlock()
}
