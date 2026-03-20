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

// TaskWithIdentifier is an optional interface for tasks that need to know their assigned identifier.
type TaskWithIdentifier interface {
	Task
	SetTaskIdentifier(id string)
	GetTaskIdentifier() string
}

// QueryHandler handles Query requests. Returns Result{Value, OK}.
type QueryHandler func(*Core, Query) Result

// TaskHandler handles Task requests. Returns Result{Value, OK}.
type TaskHandler func(*Core, Task) Result

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
	TaskIdentifier string
	Task           Task
}

type ActionTaskProgress struct {
	TaskIdentifier string
	Task           Task
	Progress       float64
	Message        string
}

type ActionTaskCompleted struct {
	TaskIdentifier string
	Task           Task
	Result         any
	Error          error
}

// --- Constructor ---

// New creates a Core instance.
//
//	c := core.New(core.Options{
//	    {Key: "name", Value: "myapp"},
//	})
func New(opts ...Options) *Core {
	c := &Core{
		app:      &App{},
		data:     &Data{},
		drive:    &Drive{},
		fs:       &Fs{root: "/"},
		config:   &Config{ConfigOptions: &ConfigOptions{}},
		error:    &ErrorPanic{},
		log:      &ErrorLog{log: Default()},
		lock:     &Lock{},
		ipc:      &Ipc{},
		i18n:     &I18n{},
		services: &serviceRegistry{services: make(map[string]*Service)},
		commands: &commandRegistry{commands: make(map[string]*Command)},
	}
	c.ctx, c.cancel = context.WithCancel(context.Background())

	if len(opts) > 0 {
		cp := make(Options, len(opts[0]))
		copy(cp, opts[0])
		c.options = &cp
		name := cp.String("name")
		if name != "" {
			c.app.Name = name
		}
	}

	// Init Cli surface with Core reference
	c.cli = &Cli{core: c}

	return c
}
