// SPDX-License-Identifier: EUPL-1.2

// Contracts, options, and type definitions for the Core framework.

package core

import (
	"context"
	"embed"
	"fmt"
	"reflect"
	"strings"
)

// Contract specifies operational guarantees for Core and its services.
type Contract struct {
	DontPanic      bool
	DisableLogging bool
}

// Option is a function that configures the Core.
type Option func(*Core) error

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

// ConfigService provides access to application configuration.
type ConfigService interface {
	Get(key string, out any) error
	Set(key string, v any) error
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

// New creates a Core instance with the provided options.
func New(opts ...Option) (*Core, error) {
	c := &Core{
		app:  &App{},
		fs:   &Fs{root: "/"},
		cfg:  &Config{ConfigOpts: &ConfigOpts{}},
		err:  &ErrPan{},
		log:  &ErrLog{&ErrOpts{Log: defaultLog}},
		cli:  &Cli{opts: &CliOpts{}},
		srv:  &Service{},
		lock: &Lock{},
		ipc:  &Ipc{},
		i18n: &I18n{},
	}

	for _, o := range opts {
		if err := o(c); err != nil {
			return nil, err
		}
	}

	return c, nil
}

// --- With* Options ---

// WithService registers a service with auto-discovered name and IPC handler.
func WithService(factory func(*Core) (any, error)) Option {
	return func(c *Core) error {
		serviceInstance, err := factory(c)
		if err != nil {
			return E("core.WithService", "failed to create service", err)
		}
		if serviceInstance == nil {
			return E("core.WithService", "service factory returned nil instance", nil)
		}

		typeOfService := reflect.TypeOf(serviceInstance)
		if typeOfService.Kind() == reflect.Ptr {
			typeOfService = typeOfService.Elem()
		}
		pkgPath := typeOfService.PkgPath()
		parts := strings.Split(pkgPath, "/")
		name := strings.ToLower(parts[len(parts)-1])
		if name == "" {
			return E("core.WithService", fmt.Sprintf("service name could not be discovered for type %T (PkgPath is empty)", serviceInstance), nil)
		}

		instanceValue := reflect.ValueOf(serviceInstance)
		handlerMethod := instanceValue.MethodByName("HandleIPCEvents")
		if handlerMethod.IsValid() {
			if handler, ok := handlerMethod.Interface().(func(*Core, Message) error); ok {
				c.RegisterAction(handler)
			} else {
				return E("core.WithService", fmt.Sprintf("service %q has HandleIPCEvents but wrong signature; expected func(*Core, Message) error", name), nil)
			}
		}

		result := c.Service(name, serviceInstance)
		if err, ok := result.(error); ok {
			return err
		}
		return nil
	}
}

// WithName registers a service with an explicit name.
func WithName(name string, factory func(*Core) (any, error)) Option {
	return func(c *Core) error {
		serviceInstance, err := factory(c)
		if err != nil {
			return E("core.WithName", fmt.Sprintf("failed to create service %q", name), err)
		}
		result := c.Service(name, serviceInstance)
		if err, ok := result.(error); ok {
			return err
		}
		return nil
	}
}

// WithApp injects the GUI runtime (e.g., Wails App).
func WithApp(runtime any) Option {
	return func(c *Core) error {
		c.app.Runtime = runtime
		return nil
	}
}

// WithAssets mounts embedded assets.
func WithAssets(efs embed.FS) Option {
	return func(c *Core) error {
		sub, err := Mount(efs, ".")
		if err != nil {
			return E("core.WithAssets", "failed to mount assets", err)
		}
		c.emb = sub
		return nil
	}
}


// WithServiceLock prevents service registration after initialisation.
func WithServiceLock() Option {
	return func(c *Core) error {
		c.LockEnable()
		c.LockApply()
		return nil
	}
}
