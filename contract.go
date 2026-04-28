// SPDX-License-Identifier: EUPL-1.2

// Contracts, options, and type definitions for the Core framework.

package core

import "sync"

// Message is the type for IPC broadcasts (fire-and-forget).
//
//	c := core.New()
//	var msg core.Message = core.ActionTaskStarted{TaskIdentifier: "task-42", Action: "agent.run"}
//	c.ACTION(msg)
type Message any

// Query is the type for read-only IPC requests.
//
//	c := core.New()
//	var query core.Query = core.NewOptions(core.Option{Key: "name", Value: "agent"})
//	_ = c.Query(query)
type Query any

// QueryHandler handles Query requests. Returns Result{Value, OK}.
//
//	handler := func(c *core.Core, q core.Query) core.Result {
//	    opts := q.(core.Options)
//	    return core.Result{Value: opts.String("name"), OK: true}
//	}
//	core.New().RegisterQuery(core.QueryHandler(handler))
type QueryHandler func(*Core, Query) Result

// Startable is implemented by services that need startup initialisation.
//
//	func (s *MyService) OnStartup(ctx Context) core.Result {
//	    return core.Result{OK: true}
//	}
type Startable interface {
	OnStartup(ctx Context) Result
}

// Stoppable is implemented by services that need shutdown cleanup.
//
//	func (s *MyService) OnShutdown(ctx Context) core.Result {
//	    return core.Result{OK: true}
//	}
type Stoppable interface {
	OnShutdown(ctx Context) Result
}

// --- Action Messages ---

// ActionServiceStartup is broadcast when a Core service finishes startup.
//
//	c := core.New()
//	c.Action("service.startup", func(ctx Context, opts core.Options) core.Result {
//	    name := opts.String("name")
//	    core.Println(core.Sprintf("service %s started", name))
//	    return core.Result{OK: true}
//	})
type ActionServiceStartup struct{}

// ActionServiceShutdown is broadcast when Core begins service shutdown.
//
//	c := core.New()
//	c.Action("service.shutdown", func(ctx Context, opts core.Options) core.Result {
//	    name := opts.String("name")
//	    core.Println(core.Sprintf("service %s stopped", name))
//	    return core.Result{OK: true}
//	})
type ActionServiceShutdown struct{}

// ActionTaskStarted is broadcast when an asynchronous task begins.
//
//	event := core.ActionTaskStarted{TaskIdentifier: "task-42", Action: "agent.run"}
//	core.New().ACTION(event)
type ActionTaskStarted struct {
	TaskIdentifier string
	Action         string
	Options        Options
}

// ActionTaskProgress is broadcast when an asynchronous task reports progress.
//
//	event := core.ActionTaskProgress{TaskIdentifier: "task-42", Action: "agent.run", Progress: 0.75, Message: "indexing repo"}
//	core.New().ACTION(event)
type ActionTaskProgress struct {
	TaskIdentifier string
	Action         string
	Progress       float64
	Message        string
}

// ActionTaskCompleted is broadcast when an asynchronous task finishes.
//
//	event := core.ActionTaskCompleted{TaskIdentifier: "task-42", Action: "agent.run", Result: core.Result{Value: "done", OK: true}}
//	core.New().ACTION(event)
type ActionTaskCompleted struct {
	TaskIdentifier string
	Action         string
	Result         Result
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
//	c := core.New(
//	    core.WithOption("name", "myapp"),
//	    core.WithService(auth.Register),
//	    core.WithServiceLock(),
//	)
//	c.Run()
func New(opts ...CoreOption) *Core {
	c := &Core{
		app:                &App{},
		data:               &Data{Registry: NewRegistry[*Embed]()},
		drive:              &Drive{Registry: NewRegistry[*DriveHandle]()},
		fs:                 (&Fs{}).New("/"),
		config:             (&Config{}).New(),
		error:              &ErrorPanic{},
		log:                &ErrorLog{},
		lock:               &Lock{locks: NewRegistry[*sync.RWMutex]()},
		ipc:                &Ipc{actions: NewRegistry[*Action](), tasks: NewRegistry[*Task]()},
		info:               systemInfo,
		i18n:               &I18n{},
		api:                &API{protocols: NewRegistry[StreamFactory]()},
		services:           &ServiceRegistry{Registry: NewRegistry[*Service]()},
		commands:           &CommandRegistry{Registry: NewRegistry[*Command]()},
		entitlementChecker: defaultChecker,
	}
	c.context, c.cancel = WithCancel(Background())
	c.api.core = c

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
		typeOf := TypeOf(instance)
		if typeOf.Kind() == KindPointer {
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
