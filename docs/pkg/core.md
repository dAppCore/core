# pkg/core -- Dependency Injection & Service Framework

`pkg/core` is the foundation of the Core application framework. It provides a dependency injection container, service lifecycle management, and a message bus for inter-service communication. Every other package in the ecosystem builds on top of it.

The package is designed for use with Wails v3 (desktop GUI) but is equally useful in CLI and headless applications.

---

## Core Struct

`Core` is the central application object. It owns the service registry, the message bus, embedded assets, and feature flags.

```go
type Core struct {
    App      any         // GUI runtime (e.g. Wails App), set by WithApp
    Features *Features   // Feature flags
    // unexported: svc *serviceManager, bus *messageBus, assets embed.FS
}
```

### Creating a Core Instance

`New()` is the sole constructor. It accepts a variadic list of `Option` functions that configure the instance before it is returned. After all options are applied, the service lock is finalised.

```go
c, err := core.New(
    core.WithService(mypackage.NewMyService),
    core.WithAssets(embeddedFS),
    core.WithServiceLock(),
)
```

If any option returns an error, `New()` returns `nil` and that error immediately.

### Options

| Option | Purpose |
|--------|---------|
| `WithService(factory)` | Register a service via factory function. Auto-discovers the service name from the factory's return type package path and auto-registers an IPC handler if the service has a `HandleIPCEvents` method. |
| `WithName(name, factory)` | Register a service with an explicit name. Does **not** auto-discover IPC handlers. |
| `WithApp(app)` | Inject a GUI runtime (e.g. Wails `*application.App`) into `Core.App`. |
| `WithAssets(fs)` | Attach an `embed.FS` containing frontend assets. |
| `WithServiceLock()` | Prevent any further service registration after `New()` completes. Calls to `RegisterService` after the lock is applied return an error. |

The `Option` type is defined as:

```go
type Option func(*Core) error
```

### Service Retrieval

Services are retrieved by name. Two generic helpers provide type-safe access:

```go
// Returns (T, error) -- safe version
svc, err := core.ServiceFor[*MyService](c, "myservice")

// Panics if not found or wrong type -- use in init paths
svc := core.MustServiceFor[*MyService](c, "myservice")
```

The untyped `Service(name)` method returns `any` (or `nil` if not found).

### Convenience Accessors

`Core` provides shorthand methods for well-known services:

```go
c.Config()     // returns Config interface
c.Display()    // returns Display interface
c.Workspace()  // returns Workspace interface
c.Crypt()      // returns Crypt interface
```

Each calls `MustServiceFor` internally and will panic if the named service is not registered.

### Global Instance

For GUI runtimes that require global access, a singleton pattern is available:

```go
core.SetInstance(c)       // store globally (thread-safe)
app := core.App()         // retrieve Core.App (panics if not set)
inst := core.GetInstance() // retrieve *Core (returns nil if not set)
core.ClearInstance()      // reset to nil (primarily for tests)
```

### Feature Flags

The `Features` struct holds a simple string slice of enabled flags:

```go
c.Features.Flags = []string{"dark-mode", "beta-api"}
c.Features.IsEnabled("dark-mode") // true
```

---

## Service Pattern

### Factory Functions

Services are created via factory functions that receive the `*Core` and return `(any, error)`:

```go
func NewMyService(c *core.Core) (any, error) {
    return &MyService{
        ServiceRuntime: core.NewServiceRuntime(c, MyOptions{BufferSize: 64}),
    }, nil
}
```

The factory is called during `New()` when the corresponding `WithService` or `WithName` option is processed.

### ServiceRuntime[T]

`ServiceRuntime[T]` is a generic helper struct that services embed to gain access to the `Core` instance and typed options:

```go
type ServiceRuntime[T any] struct {
    core *core.Core
    opts T
}
```

Constructor:

```go
rt := core.NewServiceRuntime[MyOptions](c, MyOptions{BufferSize: 64})
```

Methods:

| Method | Returns |
|--------|---------|
| `Core()` | `*Core` -- the parent container |
| `Opts()` | `T` -- the service's typed options |
| `Config()` | `Config` -- shorthand for `Core().Config()` |

Example service:

```go
type MyService struct {
    *core.ServiceRuntime[MyOptions]
    items map[string]string
}

type MyOptions struct {
    BufferSize int
}

func NewMyService(c *core.Core) (any, error) {
    return &MyService{
        ServiceRuntime: core.NewServiceRuntime(c, MyOptions{BufferSize: 128}),
        items:          make(map[string]string),
    }, nil
}
```

### Startable and Stoppable Interfaces

Services that need lifecycle hooks implement one or both of:

```go
type Startable interface {
    OnStartup(ctx context.Context) error
}

type Stoppable interface {
    OnShutdown(ctx context.Context) error
}
```

The service manager detects these interfaces at registration time and stores references internally.

- **Startup**: `ServiceStartup()` calls `OnStartup` on every `Startable` service in registration order, then broadcasts `ActionServiceStartup{}` via the message bus.
- **Shutdown**: `ServiceShutdown()` first broadcasts `ActionServiceShutdown{}`, then calls `OnShutdown` on every `Stoppable` service in **reverse** registration order. This ensures that services which were started last are stopped first, respecting dependency order.

Errors from individual services are aggregated via `errors.Join` and returned together, so one failing service does not prevent others from completing their lifecycle.

### Service Lock

When `WithServiceLock()` is passed to `New()`, the `serviceManager` sets `lockEnabled = true` during option processing. After all options have been applied, `applyLock()` sets `locked = true`. Any subsequent call to `RegisterService` returns an error:

```
core: service "late-service" is not permitted by the serviceLock setting
```

This prevents accidental late-binding of services after the application has been fully wired.

### Service Name Discovery

`WithService` derives the service name from the Go package path of the returned struct. For a type `myapp/services.Calculator`, the name becomes `services`. For `myapp/calculator.Service`, it becomes `calculator`.

To control the name explicitly, use `WithName("calc", factory)`.

### IPC Handler Discovery

`WithService` also checks whether the service has a method named `HandleIPCEvents` with signature `func(*Core, Message) error`. If found, it is automatically registered as an ACTION handler via `RegisterAction`.

`WithName` does **not** perform this discovery. Register handlers manually if needed.

---

## Message Bus

The message bus provides three distinct communication patterns, all thread-safe:

### 1. ACTION -- Fire-and-Forget Broadcast

`ACTION` dispatches a message to **all** registered handlers. Every handler is called; errors are aggregated.

```go
// Define a message type
type OrderPlaced struct {
    OrderID string
    Total   float64
}

// Dispatch
err := c.ACTION(OrderPlaced{OrderID: "abc", Total: 42.50})

// Register a handler
c.RegisterAction(func(c *core.Core, msg core.Message) error {
    switch m := msg.(type) {
    case OrderPlaced:
        log.Printf("Order %s placed for %.2f", m.OrderID, m.Total)
    }
    return nil
})
```

Multiple handlers can be registered at once with `RegisterActions(h1, h2, h3)`.

The `Message` type is defined as `any`, so any struct can serve as a message. Handlers use a type switch to filter messages they care about.

**Built-in action messages:**

| Message | Broadcast when |
|---------|---------------|
| `ActionServiceStartup{}` | After all `Startable.OnStartup` calls complete |
| `ActionServiceShutdown{}` | Before `Stoppable.OnShutdown` calls begin |
| `ActionTaskStarted{TaskID, Task}` | A `PerformAsync` task begins |
| `ActionTaskProgress{TaskID, Task, Progress, Message}` | A background task reports progress |
| `ActionTaskCompleted{TaskID, Task, Result, Error}` | A `PerformAsync` task finishes |

### 2. QUERY -- Read-Only Request/Response

`QUERY` dispatches a query to handlers until the **first** one responds (returns `handled = true`). Remaining handlers are skipped.

```go
type GetUserByID struct {
    ID string
}

// Register
c.RegisterQuery(func(c *core.Core, q core.Query) (any, bool, error) {
    switch req := q.(type) {
    case GetUserByID:
        user, err := db.Find(req.ID)
        return user, true, err
    }
    return nil, false, nil // not handled -- pass to next handler
})

// Dispatch
result, handled, err := c.QUERY(GetUserByID{ID: "u-123"})
if !handled {
    // no handler recognised this query
}
user := result.(*User)
```

`QUERYALL` dispatches the query to **all** handlers and collects every non-nil result:

```go
results, err := c.QUERYALL(ListPlugins{})
// results is []any containing responses from every handler that responded
```

The `Query` type is `any`. The `QueryHandler` signature is:

```go
type QueryHandler func(*Core, Query) (any, bool, error)
```

### 3. TASK -- Side-Effect Request/Response

`PERFORM` dispatches a task to handlers until the **first** one executes it (returns `handled = true`). Semantically identical to `QUERY` but intended for operations with side effects.

```go
type SendEmail struct {
    To      string
    Subject string
    Body    string
}

c.RegisterTask(func(c *core.Core, t core.Task) (any, bool, error) {
    switch task := t.(type) {
    case SendEmail:
        err := mailer.Send(task.To, task.Subject, task.Body)
        return nil, true, err
    }
    return nil, false, nil
})

result, handled, err := c.PERFORM(SendEmail{
    To:      "user@example.com",
    Subject: "Welcome",
    Body:    "Hello!",
})
```

The `Task` type is `any`. The `TaskHandler` signature is:

```go
type TaskHandler func(*Core, Task) (any, bool, error)
```

### Background Tasks

`PerformAsync` runs a `PERFORM` dispatch in a background goroutine and returns a task ID immediately:

```go
taskID := c.PerformAsync(BuildProject{Path: "/src"})
// taskID is "task-1", "task-2", etc.
```

The framework automatically broadcasts:

1. `ActionTaskStarted` -- when the goroutine begins
2. `ActionTaskCompleted` -- when the task finishes (includes `Result` and `Error`)

If the task implements `TaskWithID`, the framework injects the assigned ID before execution:

```go
type TaskWithID interface {
    Task
    SetTaskID(id string)
    GetTaskID() string
}
```

Services can report progress during long-running tasks:

```go
c.Progress(taskID, 0.5, "Compiling 50%...", task)
// Broadcasts ActionTaskProgress{TaskID: taskID, Progress: 0.5, Message: "..."}
```

### Thread Safety

The message bus uses `sync.RWMutex` for each handler slice (IPC, query, task). Handler registration acquires a write lock; dispatch acquires a read lock and copies the handler slice before iterating, so dispatches never block registrations.

---

## Error Handling

The `Error` struct provides contextual error wrapping:

```go
type Error struct {
    Op  string // operation, e.g. "config.Load"
    Msg string // human-readable description
    Err error  // underlying error (optional)
}
```

### E() Helper

`E()` is the primary constructor:

```go
return core.E("config.Load", "failed to read config file", err)
// Output: "config.Load: failed to read config file: <underlying error>"

return core.E("auth.Login", "invalid credentials", nil)
// Output: "auth.Login: invalid credentials"
```

When `err` is `nil`, the resulting `Error` has no wrapped cause.

### Error Chain Compatibility

`Error` implements `Unwrap()`, so it works with `errors.Is()` and `errors.As()`:

```go
var coreErr *core.Error
if errors.As(err, &coreErr) {
    log.Printf("Operation: %s, Message: %s", coreErr.Op, coreErr.Msg)
}
```

### Convention

The `Op` field should follow `package.Function` or `service.Method` format. The `Msg` field should be a human-readable sentence suitable for display to end users.

---

## Runtime (Wails Integration)

The `Runtime` struct wraps `Core` for use as a Wails service. It implements the Wails service interface (`ServiceName`, `ServiceStartup`, `ServiceShutdown`).

```go
type Runtime struct {
    app  any    // GUI runtime
    Core *Core
}
```

### NewRuntime

Creates a minimal runtime with no custom services:

```go
rt, err := core.NewRuntime(wailsApp)
```

### NewWithFactories

Creates a runtime with named service factories. Factories are called in sorted (alphabetical) order to ensure deterministic initialisation:

```go
rt, err := core.NewWithFactories(wailsApp, map[string]core.ServiceFactory{
    "calculator": func() (any, error) { return &Calculator{}, nil },
    "logger":     func() (any, error) { return &Logger{}, nil },
})
```

`ServiceFactory` is defined as `func() (any, error)` -- note it does **not** receive `*Core`, unlike the `WithService` factory. The `Runtime` wraps each factory result with `WithName` internally.

### Lifecycle Delegation

`Runtime.ServiceStartup` and `Runtime.ServiceShutdown` delegate directly to `Core.ServiceStartup` and `Core.ServiceShutdown`. The Wails runtime calls these automatically.

```go
func (r *Runtime) ServiceStartup(ctx context.Context, options any) {
    _ = r.Core.ServiceStartup(ctx, options)
}

func (r *Runtime) ServiceShutdown(ctx context.Context) {
    if r.Core != nil {
        _ = r.Core.ServiceShutdown(ctx)
    }
}
```

---

## Interfaces

`pkg/core` defines several interfaces that services may implement or consume. These decouple services from concrete implementations.

### Lifecycle Interfaces

| Interface | Method | Purpose |
|-----------|--------|---------|
| `Startable` | `OnStartup(ctx) error` | Initialisation on app start |
| `Stoppable` | `OnShutdown(ctx) error` | Cleanup on app shutdown |

### Well-Known Service Interfaces

| Interface | Service name | Key methods |
|-----------|-------------|-------------|
| `Config` | `"config"` | `Get(key, out) error`, `Set(key, v) error` |
| `Display` | `"display"` | `OpenWindow(opts...) error` |
| `Workspace` | `"workspace"` | `CreateWorkspace`, `SwitchWorkspace`, `WorkspaceFileGet`, `WorkspaceFileSet` |
| `Crypt` | `"crypt"` | `CreateKeyPair`, `EncryptPGP`, `DecryptPGP` |

These interfaces live in `interfaces.go` and define the contracts that concrete implementations must satisfy.

### Contract

The `Contract` struct configures resilience behaviour:

```go
type Contract struct {
    DontPanic      bool // recover from panics, return errors instead
    DisableLogging bool // suppress all logging
}
```

---

## Complete Example

Putting it all together -- a service that stores items, broadcasts actions, and responds to queries:

```go
package inventory

import (
    "context"
    "sync"

    "forge.lthn.ai/core/go/pkg/core"
)

// Options configures the inventory service.
type Options struct {
    MaxItems int
}

// Service manages an inventory of items.
type Service struct {
    *core.ServiceRuntime[Options]
    items map[string]string
    mu    sync.RWMutex
}

// NewService creates a factory for Core registration.
func NewService(opts Options) func(*core.Core) (any, error) {
    return func(c *core.Core) (any, error) {
        if opts.MaxItems == 0 {
            opts.MaxItems = 1000
        }
        return &Service{
            ServiceRuntime: core.NewServiceRuntime(c, opts),
            items:          make(map[string]string),
        }, nil
    }
}

// OnStartup registers query and task handlers.
func (s *Service) OnStartup(ctx context.Context) error {
    s.Core().RegisterQuery(s.handleQuery)
    s.Core().RegisterTask(s.handleTask)
    return nil
}

// -- Query: look up an item --

type GetItem struct{ ID string }

func (s *Service) handleQuery(c *core.Core, q core.Query) (any, bool, error) {
    switch req := q.(type) {
    case GetItem:
        s.mu.RLock()
        val, ok := s.items[req.ID]
        s.mu.RUnlock()
        if !ok {
            return nil, true, core.E("inventory.GetItem", "not found", nil)
        }
        return val, true, nil
    }
    return nil, false, nil
}

// -- Task: add an item --

type AddItem struct {
    ID   string
    Name string
}

type ItemAdded struct {
    ID   string
    Name string
}

func (s *Service) handleTask(c *core.Core, t core.Task) (any, bool, error) {
    switch task := t.(type) {
    case AddItem:
        s.mu.Lock()
        s.items[task.ID] = task.Name
        s.mu.Unlock()

        _ = c.ACTION(ItemAdded{ID: task.ID, Name: task.Name})
        return task.ID, true, nil
    }
    return nil, false, nil
}

// -- Wiring it up --

func main() {
    c, err := core.New(
        core.WithName("inventory", NewService(Options{MaxItems: 500})),
        core.WithServiceLock(),
    )
    if err != nil {
        panic(err)
    }

    // Start lifecycle
    _ = c.ServiceStartup(context.Background(), nil)

    // Use the bus
    _, _, _ = c.PERFORM(AddItem{ID: "item-1", Name: "Widget"})
    result, _, _ := c.QUERY(GetItem{ID: "item-1"})
    // result == "Widget"

    // Shutdown
    _ = c.ServiceShutdown(context.Background())
}
```

---

## File Map

| File | Responsibility |
|------|---------------|
| `core.go` | `New()`, options (`WithService`, `WithName`, `WithApp`, `WithAssets`, `WithServiceLock`), `ServiceFor[T]`, `MustServiceFor[T]`, lifecycle dispatch, global instance, bus method delegation |
| `interfaces.go` | `Core` struct definition, `Option`, `Message`, `Query`, `Task`, `QueryHandler`, `TaskHandler`, `Startable`, `Stoppable`, `Contract`, `Features`, well-known service interfaces (`Config`, `Display`, `Workspace`, `Crypt`), built-in action message types |
| `message_bus.go` | `messageBus` struct, `action()`, `query()`, `queryAll()`, `perform()`, handler registration |
| `service_manager.go` | `serviceManager` struct, service registry, `Startable`/`Stoppable` tracking, service lock mechanism |
| `runtime_pkg.go` | `ServiceRuntime[T]` generic helper, `Runtime` struct (Wails integration), `NewRuntime()`, `NewWithFactories()` |
| `e.go` | `Error` struct, `E()` constructor, `Unwrap()` for error chain compatibility |
