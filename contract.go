// SPDX-License-Identifier: EUPL-1.2

// Contracts, options, and type definitions for the Core framework.

package core

import (
	"context"
	"reflect"
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

// CoreOption is a functional option applied during Core construction.
// Returns Result — if !OK, New() stops and returns the error.
//
//	core.New(
//	    core.WithService(agentic.Register),
//	    core.WithService(monitor.Register),
//	    core.WithServiceLock(),
//	)
type CoreOption func(*Core) Result

// New initialises a Core instance by applying options in order.
// Services registered here form the application conclave — they share
// IPC access and participate in the lifecycle (ServiceStartup/ServiceShutdown).
//
//	r := core.New(
//	    core.WithOptions(core.Options{{Key: "name", Value: "myapp"}}),
//	    core.WithService(auth.Register),
//	    core.WithServiceLock(),
//	)
//	if !r.OK { log.Fatal(r.Value) }
//	c := r.Value.(*Core)
func New(opts ...CoreOption) Result {
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
		info:     systemInfo,
		i18n:     &I18n{},
		services: &serviceRegistry{services: make(map[string]*Service)},
		commands: &commandRegistry{commands: make(map[string]*Command)},
	}
	c.context, c.cancel = context.WithCancel(context.Background())
	c.cli = &Cli{core: c}

	for _, opt := range opts {
		if r := opt(c); !r.OK {
			return r
		}
	}

	return Result{c, true}
}

// WithOptions applies key-value configuration to Core.
//
//	core.WithOptions(core.Options{{Key: "name", Value: "myapp"}})
func WithOptions(opts Options) CoreOption {
	return func(c *Core) Result {
		cp := make(Options, len(opts))
		copy(cp, opts)
		c.options = &cp
		if name := cp.String("name"); name != "" {
			c.app.Name = name
		}
		return Result{OK: true}
	}
}

// WithService registers a service via its factory function.
// The factory receives *Core so the service can wire IPC handlers
// and access other subsystems during construction.
// Service name is auto-discovered from the package path.
// If the service implements HandleIPCEvents, it is auto-registered.
//
//	core.WithService(agentic.Register)
//	core.WithService(display.Register(nil))
func WithService(factory func(*Core) Result) CoreOption {
	return func(c *Core) Result {
		r := factory(c)
		if !r.OK {
			return r
		}

		// If the factory returned a service instance, auto-discover and register
		if r.Value != nil {
			instance := r.Value
			// Service name discovery from package path
			typeOf := reflect.TypeOf(instance)
			if typeOf.Kind() == reflect.Ptr {
				typeOf = typeOf.Elem()
			}
			pkgPath := typeOf.PkgPath()
			parts := Split(pkgPath, "/")
			name := Lower(parts[len(parts)-1])

			if name != "" {
				// IPC handler discovery
				instanceValue := reflect.ValueOf(instance)
				handlerMethod := instanceValue.MethodByName("HandleIPCEvents")
				if handlerMethod.IsValid() {
					if handler, ok := handlerMethod.Interface().(func(*Core, Message) Result); ok {
						c.RegisterAction(handler)
					}
				}

				// Register the service if not already registered by the factory
				if sr := c.Service(name); !sr.OK {
					c.Service(name, Service{})
				}
			}
		}

		return Result{OK: true}
	}
}

// WithServiceLock prevents further service registration after construction.
//
//	core.New(
//	    core.WithService(auth.Register),
//	    core.WithServiceLock(),
//	)
func WithServiceLock() CoreOption {
	return func(c *Core) Result {
		c.LockEnable()
		c.LockApply()
		return Result{OK: true}
	}
}
