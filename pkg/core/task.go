// SPDX-License-Identifier: EUPL-1.2

// Background task dispatch for the Core framework.

package core

import (
	"fmt"
	"slices"
)

// TaskState holds background task state.
type TaskState struct {
	ID     string
	Task   Task
	Result any
	Error  error
}

// PerformAsync dispatches a task in a background goroutine.
func (c *Core) PerformAsync(t Task) string {
	if c.shutdown.Load() {
		return ""
	}
	taskID := fmt.Sprintf("task-%d", c.taskIDCounter.Add(1))
	if tid, ok := t.(TaskWithID); ok {
		tid.SetTaskID(taskID)
	}
	_ = c.ACTION(ActionTaskStarted{TaskID: taskID, Task: t})
	c.wg.Go(func() {
		result, handled, err := c.PERFORM(t)
		if !handled && err == nil {
			err = E("core.PerformAsync", fmt.Sprintf("no handler found for task type %T", t), nil)
		}
		_ = c.ACTION(ActionTaskCompleted{TaskID: taskID, Task: t, Result: result, Error: err})
	})
	return taskID
}

// Progress broadcasts a progress update for a background task.
func (c *Core) Progress(taskID string, progress float64, message string, t Task) {
	_ = c.ACTION(ActionTaskProgress{TaskID: taskID, Task: t, Progress: progress, Message: message})
}

func (c *Core) Perform(t Task) (any, bool, error) {
	c.ipc.taskMu.RLock()
	handlers := slices.Clone(c.ipc.taskHandlers)
	c.ipc.taskMu.RUnlock()

	for _, h := range handlers {
		result, handled, err := h(c, t)
		if handled {
			return result, true, err
		}
	}
	return nil, false, nil
}

func (c *Core) RegisterTask(handler TaskHandler) {
	c.ipc.taskMu.Lock()
	c.ipc.taskHandlers = append(c.ipc.taskHandlers, handler)
	c.ipc.taskMu.Unlock()
}
