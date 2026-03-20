// SPDX-License-Identifier: EUPL-1.2

// Contracts, options, and type definitions for the Core framework.

package core

import (
	"context"
)

// Message is the type for IPC broadcasts (fire-and-forget).
type Message any

// Query is the type for read-only IPC requests.
type Query any

// Task is the type for IPC requests that perform side effects.
type Task any

// TaskWithID is an optional interface for tasks that need to know their assigned ID.
type TaskWithID interface {
	Task
	SetTaskID(id string)
	GetTaskID() string
}

// QueryHandler handles Query requests. Returns (result, handled, error).
type QueryHandler func(*Core, Query) (any, bool, error)

// TaskHandler handles Task requests. Returns (result, handled, error).
type TaskHandler func(*Core, Task) (any, bool, error)

// Startable is implemented by services that need startup initialisation.
type Startable interface {
	OnStartup(ctx context.Context) error
}

// Stoppable is implemented by services that need shutdown cleanup.
type Stoppable interface {
	OnShutdown(ctx context.Context) error
}

// --- Action Messages ---

type ActionServiceStartup struct{}
type ActionServiceShutdown struct{}

type ActionTaskStarted struct {
	TaskID string
	Task   Task
}

type ActionTaskProgress struct {
	TaskID   string
	Task     Task
	Progress float64
	Message  string
}

type ActionTaskCompleted struct {
	TaskID string
	Task   Task
	Result any
	Error  error
}

// --- Constructor ---

// New creates a Core instance.
//
//	c := core.New(core.Options{
//	    {K: "name", V: "myapp"},
//	})
func New(opts ...Options) *Core {
	c := &Core{
		app:     &App{},
		data:    &Data{},
		drive:   &Drive{},
		fs:      &Fs{root: "/"},
		config:  &Config{ConfigOpts: &ConfigOpts{}},
		error:   &ErrorPanic{},
		log:     &ErrorLog{log: defaultLog},
		cli:     &Cli{opts: &CliOpts{}},
		service: &Service{},
		lock:    &Lock{},
		ipc:     &Ipc{},
		i18n:    &I18n{},
	}

	if len(opts) > 0 {
		c.options = &opts[0]
		name := opts[0].String("name")
		if name != "" {
			c.app.Name = name
		}
	}

	return c
}
