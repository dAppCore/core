// SPDX-License-Identifier: EUPL-1.2

// Background action dispatch for the Core framework.
// PerformAsync runs a named Action in a background goroutine with
// panic recovery and progress broadcasting.

package core

// PerformAsync dispatches a named action in a background goroutine.
// Broadcasts ActionTaskStarted, ActionTaskProgress, and ActionTaskCompleted
// as IPC messages so other services can track progress.
//
//	r := c.PerformAsync("agentic.dispatch", opts)
//	taskID := r.Value.(string)
func (c *Core) PerformAsync(action string, opts Options) Result {
	if c.shutdown.Load() {
		return Result{}
	}
	taskID := ID()

	c.ACTION(ActionTaskStarted{TaskIdentifier: taskID, Action: action, Options: opts})

	c.waitGroup.Go(func() {
		defer func() {
			if rec := recover(); rec != nil {
				c.ACTION(ActionTaskCompleted{
					TaskIdentifier: taskID,
					Action:         action,
					Result:         Result{E("core.PerformAsync", Sprint("panic: ", rec), nil), false},
				})
			}
		}()

		r := c.Action(action).Run(Background(), opts)

		c.ACTION(ActionTaskCompleted{
			TaskIdentifier: taskID,
			Action:         action,
			Result:         r,
		})
	})

	return Result{taskID, true}
}

// Progress broadcasts a progress update for a background task.
//
//	c.Progress(taskID, 0.5, "halfway done", "agentic.dispatch")
func (c *Core) Progress(taskID string, progress float64, message string, action string) {
	c.ACTION(ActionTaskProgress{
		TaskIdentifier: taskID,
		Action:         action,
		Progress:       progress,
		Message:        message,
	})
}

// Registration methods (RegisterAction, RegisterActions)
// are in ipc.go — registration is IPC's responsibility.
