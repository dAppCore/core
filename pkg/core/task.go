// SPDX-License-Identifier: EUPL-1.2

// Background task dispatch for the Core framework.

package core

import (
	"reflect"
	"slices"
	"strconv"
)

// TaskState holds background task state.
type TaskState struct {
	Identifier string
	Task   Task
	Result any
	Error  error
}

// PerformAsync dispatches a task in a background goroutine.
func (c *Core) PerformAsync(t Task) Result {
	if c.shutdown.Load() {
		return Result{}
	}
	taskID := Concat("task-", strconv.FormatUint(c.taskIDCounter.Add(1), 10))
	if tid, ok := t.(TaskWithIdentifier); ok {
		tid.SetTaskIdentifier(taskID)
	}
	c.ACTION(ActionTaskStarted{TaskIdentifier: taskID, Task: t})
	c.waitGroup.Go(func() {
		defer func() {
			if rec := recover(); rec != nil {
				err := E("core.PerformAsync", Sprint("panic: ", rec), nil)
				c.ACTION(ActionTaskCompleted{TaskIdentifier: taskID, Task: t, Result: nil, Error: err})
			}
		}()
		r := c.PERFORM(t)
		var err error
		if !r.OK {
			if e, ok := r.Value.(error); ok {
				err = e
			} else {
				taskType := reflect.TypeOf(t)
				typeName := "<nil>"
				if taskType != nil {
					typeName = taskType.String()
				}
				err = E("core.PerformAsync", Join(" ", "no handler found for task type", typeName), nil)
			}
		}
		c.ACTION(ActionTaskCompleted{TaskIdentifier: taskID, Task: t, Result: r.Value, Error: err})
	})
	return Result{taskID, true}
}

// Progress broadcasts a progress update for a background task.
func (c *Core) Progress(taskID string, progress float64, message string, t Task) {
	c.ACTION(ActionTaskProgress{TaskIdentifier: taskID, Task: t, Progress: progress, Message: message})
}

func (c *Core) Perform(t Task) Result {
	c.ipc.taskMu.RLock()
	handlers := slices.Clone(c.ipc.taskHandlers)
	c.ipc.taskMu.RUnlock()

	for _, h := range handlers {
		r := h(c, t)
		if r.OK {
			return r
		}
	}
	return Result{}
}

func (c *Core) RegisterAction(handler func(*Core, Message) Result) {
	c.ipc.ipcMu.Lock()
	c.ipc.ipcHandlers = append(c.ipc.ipcHandlers, handler)
	c.ipc.ipcMu.Unlock()
}

func (c *Core) RegisterActions(handlers ...func(*Core, Message) Result) {
	c.ipc.ipcMu.Lock()
	c.ipc.ipcHandlers = append(c.ipc.ipcHandlers, handlers...)
	c.ipc.ipcMu.Unlock()
}

func (c *Core) RegisterTask(handler TaskHandler) {
	c.ipc.taskMu.Lock()
	c.ipc.taskHandlers = append(c.ipc.taskHandlers, handler)
	c.ipc.taskMu.Unlock()
}
