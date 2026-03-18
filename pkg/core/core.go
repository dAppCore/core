package core

import (
	"context"
	"embed"
	"errors"
	"io/fs"
	"fmt"
	"reflect"
	"slices"
	"strings"
	"sync"
)

var (
	instance   *Core
	instanceMu sync.RWMutex
)

// New initialises a Core instance using the provided options and performs the necessary setup.
// It is the primary entry point for creating a new Core application.
//
// Example:
//
//	core, err := core.New(
//		core.WithService(&MyService{}),
//		core.WithAssets(assets),
//	)
func New(opts ...Option) (*Core, error) {
	// Default IO rooted at "/" (full access, like os package)
	defaultIO, _ := NewIO("/")
	c := &Core{
		io:    defaultIO,
		etc:   NewEtc(),
		crash: NewCrashHandler(),
		svc:   newServiceManager(),
	}
	c.bus = newMessageBus(c)

	for _, o := range opts {
		if err := o(c); err != nil {
			return nil, err
		}
	}

	c.svc.applyLock()
	return c, nil
}

// WithService creates an Option that registers a service. It automatically discovers
// the service name from its package path and registers its IPC handler if it
// implements a method named `HandleIPCEvents`.
//
// Example:
//
//	// In myapp/services/calculator.go
//	package services
//
//	type Calculator struct{}
//
//	func (s *Calculator) Add(a, b int) int { return a + b }
//
//	// In main.go
//	import "myapp/services"
//
//	core.New(core.WithService(services.NewCalculator))
func WithService(factory func(*Core) (any, error)) Option {
	return func(c *Core) error {
		serviceInstance, err := factory(c)

		if err != nil {
			return E("core.WithService", "failed to create service", err)
		}
		if serviceInstance == nil {
			return E("core.WithService", "service factory returned nil instance", nil)
		}

		// --- Service Name Discovery ---
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

		// --- IPC Handler Discovery ---
		instanceValue := reflect.ValueOf(serviceInstance)
		handlerMethod := instanceValue.MethodByName("HandleIPCEvents")
		if handlerMethod.IsValid() {
			if handler, ok := handlerMethod.Interface().(func(*Core, Message) error); ok {
				c.RegisterAction(handler)
			} else {
				return E("core.WithService", fmt.Sprintf("service %q has HandleIPCEvents but wrong signature; expected func(*Core, Message) error", name), nil)
			}
		}

		return c.RegisterService(name, serviceInstance)
	}
}

// WithName creates an option that registers a service with a specific name.
// This is useful when the service name cannot be inferred from the package path,
// such as when using anonymous functions as factories.
// Note: Unlike WithService, this does not automatically discover or register
// IPC handlers. If your service needs IPC handling, implement HandleIPCEvents
// and register it manually.
func WithName(name string, factory func(*Core) (any, error)) Option {
	return func(c *Core) error {
		serviceInstance, err := factory(c)
		if err != nil {
			return E("core.WithName", fmt.Sprintf("failed to create service %q", name), err)
		}
		return c.RegisterService(name, serviceInstance)
	}
}

// WithApp creates an Option that injects the GUI runtime (e.g., Wails App) into the Core.
// This is essential for services that need to interact with the GUI runtime.
func WithApp(app any) Option {
	return func(c *Core) error {
		c.App = app
		return nil
	}
}

// WithAssets creates an Option that mounts embedded assets.
// The assets are accessible via c.Mnt().
func WithAssets(efs embed.FS) Option {
	return func(c *Core) error {
		sub, err := Mount(efs, ".")
		if err != nil {
			return E("core.WithAssets", "failed to mount assets", err)
		}
		c.mnt = sub
		return nil
	}
}

// WithIO creates an Option that sandboxes filesystem I/O to a root path.
// Default is "/" (full access). Use this to restrict c.Io() operations.
func WithIO(root string) Option {
	return func(c *Core) error {
		io, err := NewIO(root)
		if err != nil {
			return E("core.WithIO", "failed to create IO at "+root, err)
		}
		c.io = io
		return nil
	}
}

// WithMount creates an Option that mounts an fs.FS at a specific subdirectory.
// The mounted assets are accessible via c.Mnt().
func WithMount(fsys fs.FS, basedir string) Option {
	return func(c *Core) error {
		sub, err := Mount(fsys, basedir)
		if err != nil {
			return E("core.WithMount", "failed to mount "+basedir, err)
		}
		c.mnt = sub
		return nil
	}
}

// WithServiceLock creates an Option that prevents any further services from being
// registered after the Core has been initialized. This is a security measure to
// prevent late-binding of services that could have unintended consequences.
func WithServiceLock() Option {
	return func(c *Core) error {
		c.svc.enableLock()
		return nil
	}
}

// --- Core Methods ---

// ServiceStartup is the entry point for the Core service's startup lifecycle.
// It is called by the GUI runtime when the application starts.
func (c *Core) ServiceStartup(ctx context.Context, options any) error {
	startables := c.svc.getStartables()

	var agg error
	for _, s := range startables {
		if err := ctx.Err(); err != nil {
			return errors.Join(agg, err)
		}
		if err := s.OnStartup(ctx); err != nil {
			agg = errors.Join(agg, err)
		}
	}

	if err := c.ACTION(ActionServiceStartup{}); err != nil {
		agg = errors.Join(agg, err)
	}

	return agg
}

// ServiceShutdown is the entry point for the Core service's shutdown lifecycle.
// It is called by the GUI runtime when the application shuts down.
func (c *Core) ServiceShutdown(ctx context.Context) error {
	c.shutdown.Store(true)

	var agg error
	if err := c.ACTION(ActionServiceShutdown{}); err != nil {
		agg = errors.Join(agg, err)
	}

	stoppables := c.svc.getStoppables()
	for _, s := range slices.Backward(stoppables) {
		if err := ctx.Err(); err != nil {
			agg = errors.Join(agg, err)
			break // don't return — must still wait for background tasks below
		}
		if err := s.OnShutdown(ctx); err != nil {
			agg = errors.Join(agg, err)
		}
	}

	// Wait for background tasks (PerformAsync), respecting context deadline.
	done := make(chan struct{})
	go func() {
		c.wg.Wait()
		close(done)
	}()
	select {
	case <-done:
	case <-ctx.Done():
		agg = errors.Join(agg, ctx.Err())
	}

	return agg
}

// ACTION dispatches a message to all registered IPC handlers.
// This is the primary mechanism for services to communicate with each other.
func (c *Core) ACTION(msg Message) error {
	return c.bus.action(msg)
}

// RegisterAction adds a new IPC handler to the Core.
func (c *Core) RegisterAction(handler func(*Core, Message) error) {
	c.bus.registerAction(handler)
}

// RegisterActions adds multiple IPC handlers to the Core.
func (c *Core) RegisterActions(handlers ...func(*Core, Message) error) {
	c.bus.registerActions(handlers...)
}

// QUERY dispatches a query to handlers until one responds.
// Returns (result, handled, error). If no handler responds, handled is false.
func (c *Core) QUERY(q Query) (any, bool, error) {
	return c.bus.query(q)
}

// QUERYALL dispatches a query to all handlers and collects all responses.
// Returns all results from handlers that responded.
func (c *Core) QUERYALL(q Query) ([]any, error) {
	return c.bus.queryAll(q)
}

// PERFORM dispatches a task to handlers until one executes it.
// Returns (result, handled, error). If no handler responds, handled is false.
func (c *Core) PERFORM(t Task) (any, bool, error) {
	return c.bus.perform(t)
}

// PerformAsync dispatches a task to be executed in a background goroutine.
// It returns a unique task ID that can be used to track the task's progress.
// The result of the task will be broadcasted via an ActionTaskCompleted message.
func (c *Core) PerformAsync(t Task) string {
	if c.shutdown.Load() {
		return ""
	}

	taskID := fmt.Sprintf("task-%d", c.taskIDCounter.Add(1))

	// If the task supports it, inject the ID
	if tid, ok := t.(TaskWithID); ok {
		tid.SetTaskID(taskID)
	}

	// Broadcast task started
	_ = c.ACTION(ActionTaskStarted{
		TaskID: taskID,
		Task:   t,
	})

	c.wg.Go(func() {
		result, handled, err := c.PERFORM(t)
		if !handled && err == nil {
			err = E("core.PerformAsync", fmt.Sprintf("no handler found for task type %T", t), nil)
		}

		// Broadcast task completed
		_ = c.ACTION(ActionTaskCompleted{
			TaskID: taskID,
			Task:   t,
			Result: result,
			Error:  err,
		})
	})

	return taskID
}

// Progress broadcasts a progress update for a background task.
func (c *Core) Progress(taskID string, progress float64, message string, t Task) {
	_ = c.ACTION(ActionTaskProgress{
		TaskID:   taskID,
		Task:     t,
		Progress: progress,
		Message:  message,
	})
}

// RegisterQuery adds a query handler to the Core.
func (c *Core) RegisterQuery(handler QueryHandler) {
	c.bus.registerQuery(handler)
}

// RegisterTask adds a task handler to the Core.
func (c *Core) RegisterTask(handler TaskHandler) {
	c.bus.registerTask(handler)
}

// RegisterService adds a new service to the Core.
// If the service implements LocaleProvider, its locale FS is collected
// for the i18n service to load during startup.
func (c *Core) RegisterService(name string, api any) error {
	// Collect locale filesystems from services that provide them
	if lp, ok := api.(LocaleProvider); ok {
		c.locales = append(c.locales, lp.Locales())
	}
	return c.svc.registerService(name, api)
}

// Service retrieves a registered service by name.
// It returns nil if the service is not found.
func (c *Core) Service(name string) any {
	return c.svc.service(name)
}

// ServiceFor retrieves a registered service by name and asserts its type to the given interface T.
func ServiceFor[T any](c *Core, name string) (T, error) {
	var zero T
	raw := c.Service(name)
	if raw == nil {
		return zero, E("core.ServiceFor", fmt.Sprintf("service %q not found", name), nil)
	}
	typed, ok := raw.(T)
	if !ok {
		return zero, E("core.ServiceFor", fmt.Sprintf("service %q is type %T, expected %T", name, raw, zero), nil)
	}
	return typed, nil
}

// MustServiceFor retrieves a registered service by name and asserts its type to the given interface T.
// It panics if the service is not found or cannot be cast to T.
func MustServiceFor[T any](c *Core, name string) T {
	svc, err := ServiceFor[T](c, name)
	if err != nil {
		panic(err)
	}
	return svc
}

// App returns the global application instance.
// It panics if the Core has not been initialized via SetInstance.
// This is typically used by GUI runtimes that need global access.
func App() any {
	instanceMu.RLock()
	inst := instance
	instanceMu.RUnlock()
	if inst == nil {
		panic("core.App() called before core.SetInstance()")
	}
	return inst.App
}

// SetInstance sets the global Core instance for App() access.
// This is typically called by GUI runtimes during initialization.
func SetInstance(c *Core) {
	instanceMu.Lock()
	instance = c
	instanceMu.Unlock()
}

// GetInstance returns the global Core instance, or nil if not set.
// Use this for non-panicking access to the global instance.
func GetInstance() *Core {
	instanceMu.RLock()
	inst := instance
	instanceMu.RUnlock()
	return inst
}

// ClearInstance resets the global Core instance to nil.
// This is primarily useful for testing to ensure a clean state between tests.
func ClearInstance() {
	instanceMu.Lock()
	instance = nil
	instanceMu.Unlock()
}

// Config returns the registered Config service.
func (c *Core) Config() Config {
	return MustServiceFor[Config](c, "config")
}

// Display returns the registered Display service.
func (c *Core) Display() Display {
	return MustServiceFor[Display](c, "display")
}

// Workspace returns the registered Workspace service.
func (c *Core) Workspace() Workspace {
	return MustServiceFor[Workspace](c, "workspace")
}

// Crypt returns the registered Crypt service.
func (c *Core) Crypt() Crypt {
	return MustServiceFor[Crypt](c, "crypt")
}

// Core returns self, implementing the CoreProvider interface.
func (c *Core) Core() *Core { return c }

