---
title: Services
description: Service registration, retrieval, ServiceRuntime, and factory patterns.
---

# Services

Services are the building blocks of a Core application. They are plain Go structs registered into a named registry and retrieved by name with optional type assertions.

## Registration

### Factory Functions

The primary way to register a service is via a **factory function** -- a function with the signature `func(*Core) (any, error)`. The factory receives the `Core` instance so it can access other services or register message handlers during construction.

```go
func NewMyService(c *core.Core) (any, error) {
    return &MyService{}, nil
}
```

### WithService (auto-named)

`WithService` registers a service and automatically discovers its name from the Go package path. The last segment of the package path becomes the service name, lowercased.

```go
// If MyService lives in package "myapp/services/calculator",
// it is registered as "calculator".
c, err := core.New(
    core.WithService(calculator.NewService),
)
```

`WithService` also performs **IPC handler discovery**: if the returned service has a method named `HandleIPCEvents` with the signature `func(*Core, Message) error`, it is automatically registered as an action handler.

```go
type Service struct{}

func (s *Service) HandleIPCEvents(c *core.Core, msg core.Message) error {
    // Handle messages
    return nil
}
```

### WithName (explicitly named)

When you need to control the service name (or the factory is an anonymous function), use `WithName`:

```go
c, err := core.New(
    core.WithName("my-service", func(c *core.Core) (any, error) {
        return &MyService{}, nil
    }),
)
```

Unlike `WithService`, `WithName` does **not** auto-discover IPC handlers. Register them manually if needed.

### Direct Registration

You can also register a service directly on an existing `Core` instance:

```go
err := c.RegisterService("my-service", &MyService{})
```

This is useful for tests or when constructing services outside the `New()` options flow.

### Registration Rules

- Service names **must not be empty**.
- **Duplicate names** are rejected with an error.
- If `WithServiceLock()` was passed to `New()`, registration after initialisation is rejected.

## Retrieval

### By Name (untyped)

```go
svc := c.Service("calculator")
if svc == nil {
    // not found
}
```

Returns `nil` if no service is registered under that name.

### Type-Safe Retrieval

`ServiceFor[T]` retrieves and type-asserts in one step:

```go
calc, err := core.ServiceFor[*calculator.Service](c, "calculator")
if err != nil {
    // "service 'calculator' not found"
    // or "service 'calculator' is of type *Foo, but expected *calculator.Service"
}
```

### Panicking Retrieval

For init-time wiring where a missing service is a fatal programming error:

```go
calc := core.MustServiceFor[*calculator.Service](c, "calculator")
// panics if not found or wrong type
```

## ServiceRuntime

`ServiceRuntime[T]` is a generic helper you embed in your service struct. It provides typed access to the `Core` instance and your service's options struct.

```go
type Options struct {
    Precision int
}

type Service struct {
    *core.ServiceRuntime[Options]
}

func NewService(opts Options) func(*core.Core) (any, error) {
    return func(c *core.Core) (any, error) {
        return &Service{
            ServiceRuntime: core.NewServiceRuntime(c, opts),
        }, nil
    }
}
```

`ServiceRuntime` provides these methods:

| Method | Returns | Description |
|--------|---------|-------------|
| `Core()` | `*Core` | The central Core instance |
| `Opts()` | `T` | The service's typed options |
| `Config()` | `Config` | Convenience shortcut for `Core().Config()` |

### Real-World Example: The Log Service

The `pkg/log` package in this repository is the reference implementation of a Core service:

```go
type Service struct {
    *core.ServiceRuntime[Options]
    *Logger
}

func NewService(opts Options) func(*core.Core) (any, error) {
    return func(c *core.Core) (any, error) {
        logger := New(opts)
        return &Service{
            ServiceRuntime: core.NewServiceRuntime(c, opts),
            Logger:         logger,
        }, nil
    }
}

func (s *Service) OnStartup(ctx context.Context) error {
    s.Core().RegisterQuery(s.handleQuery)
    s.Core().RegisterTask(s.handleTask)
    return nil
}
```

Key patterns to note:

1. The factory is a **closure** -- `NewService` takes options and returns a factory function.
2. `ServiceRuntime` is embedded, giving access to `Core()` and `Opts()`.
3. The service implements `Startable` to register its query/task handlers at startup.

## Runtime and NewWithFactories

For applications that wire services from a map of named factories, `NewWithFactories` offers a bulk registration path:

```go
type ServiceFactory func() (any, error)

rt, err := core.NewWithFactories(app, map[string]core.ServiceFactory{
    "config":   configFactory,
    "database": dbFactory,
    "cache":    cacheFactory,
})
```

Factories are called in sorted key order. The resulting `Runtime` wraps a `Core` and exposes `ServiceStartup`/`ServiceShutdown` for GUI runtime integration.

For the simplest case with no custom services:

```go
rt, err := core.NewRuntime(app)
```

## Well-Known Services

Core provides convenience methods for commonly needed services. These use `MustServiceFor` internally and will panic if the service is not registered:

| Method | Expected Name | Expected Interface |
|--------|--------------|-------------------|
| `c.Config()` | `"config"` | `Config` |
| `c.Display()` | `"display"` | `Display` |
| `c.Workspace()` | `"workspace"` | `Workspace` |
| `c.Crypt()` | `"crypt"` | `Crypt` |

These are optional -- only call them if you have registered the corresponding service.

## Thread Safety

The service registry is protected by `sync.RWMutex`. Registration, retrieval, and lifecycle operations are safe to call from multiple goroutines.

## Related Pages

- [Lifecycle](lifecycle.md) -- `Startable` and `Stoppable` interfaces
- [Messaging](messaging.md) -- how services communicate
- [Configuration](configuration.md) -- all `With*` options
