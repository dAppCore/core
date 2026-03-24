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
//	    core.WithOptions(core.NewOptions(core.Option{Key: "name", Value: "myapp"})),
//	    core.WithService(auth.Register),
//	    core.WithServiceLock(),
//	)
//	if !r.OK { log.Fatal(r.Value) }
//	c := r.Value.(*Core)
func New(opts ...CoreOption) *Core {
	c := &Core{
		app:      &App{},
		data:     &Data{},
		drive:    &Drive{},
		fs:       (&Fs{}).New("/"),
		config:   (&Config{}).New(),
		error:    &ErrorPanic{},
		log:      &ErrorLog{},
		lock:     &Lock{},
		ipc:      &Ipc{},
		info:     systemInfo,
		i18n:     &I18n{},
		services: &serviceRegistry{services: make(map[string]*Service)},
		commands: &commandRegistry{commands: make(map[string]*Command)},
	}
	c.context, c.cancel = context.WithCancel(context.Background())

	// Core services
	CliRegister(c)

	for _, opt := range opts {
		if r := opt(c); !r.OK {
			Error("core.New failed", "err", r.Value)
			break
		}
	}

	// Apply service lock after all opts — v0.3.3 parity
	c.LockApply()

	return c
}

// WithOptions applies key-value configuration to Core.
//
//	core.WithOptions(core.NewOptions(core.Option{Key: "name", Value: "myapp"}))
func WithOptions(opts Options) CoreOption {
	return func(c *Core) Result {
		c.options = &opts
		if name := opts.String("name"); name != "" {
			c.app.Name = name
		}
		return Result{OK: true}
	}
}

// WithService registers a service via its factory function.
// If the factory returns a non-nil Value, WithService auto-discovers the
// service name from the factory's package path (last path segment, lowercase,
// with any "_test" suffix stripped) and calls RegisterService on the instance.
// IPC handler auto-registration is handled by RegisterService.
//
// If the factory returns nil Value (it registered itself), WithService
// returns success without a second registration.
//
//	core.WithService(agentic.Register)
//	core.WithService(display.Register(nil))
func WithService(factory func(*Core) Result) CoreOption {
	return func(c *Core) Result {
		r := factory(c)
		if !r.OK {
			return r
		}
		if r.Value == nil {
			// Factory self-registered — nothing more to do.
			return Result{OK: true}
		}
		// Auto-discover the service name from the instance's package path.
		instance := r.Value
		typeOf := reflect.TypeOf(instance)
		if typeOf.Kind() == reflect.Ptr {
			typeOf = typeOf.Elem()
		}
		pkgPath := typeOf.PkgPath()
		parts := Split(pkgPath, "/")
		name := Lower(parts[len(parts)-1])
		if name == "" {
			return Result{E("core.WithService", Sprintf("service name could not be discovered for type %T", instance), nil), false}
		}

		// RegisterService handles Startable/Stoppable/HandleIPCEvents discovery
		return c.RegisterService(name, instance)
	}
}

// WithName registers a service with an explicit name (no reflect discovery).
//
//	core.WithName("ws", func(c *Core) Result {
//	    return Result{Value: hub, OK: true}
//	})
func WithName(name string, factory func(*Core) Result) CoreOption {
	return func(c *Core) Result {
		r := factory(c)
		if !r.OK {
			return r
		}
		if r.Value == nil {
			return Result{E("core.WithName", Sprintf("failed to create service %q", name), nil), false}
		}
		return c.RegisterService(name, r.Value)
	}
}

// WithOption is a convenience for setting a single key-value option.
//
//	core.New(
//	    core.WithOption("name", "myapp"),
//	    core.WithOption("port", 8080),
//	)
func WithOption(key string, value any) CoreOption {
	return func(c *Core) Result {
		if c.options == nil {
			opts := NewOptions()
			c.options = &opts
		}
		c.options.Set(key, value)
		if key == "name" {
			if s, ok := value.(string); ok {
				c.app.Name = s
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
		return Result{OK: true}
	}
}
