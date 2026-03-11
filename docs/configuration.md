---
title: Configuration Options
description: WithService, WithName, WithApp, WithAssets, and WithServiceLock options.
---

# Configuration Options

The `Core` is configured through **options** -- functions with the signature `func(*Core) error`. These are passed to `core.New()` and applied in order during initialisation.

```go
type Option func(*Core) error
```

## Available Options

### WithService

```go
func WithService(factory func(*Core) (any, error)) Option
```

Registers a service using a factory function. The service name is **auto-discovered** from the Go package path of the returned type (the last path segment, lowercased).

```go
// If the returned type is from package "myapp/services/calculator",
// the service name becomes "calculator".
core.New(
    core.WithService(calculator.NewService),
)
```

`WithService` also performs two automatic behaviours:

1. **Name discovery** -- uses `reflect` to extract the package name from the returned type.
2. **IPC handler discovery** -- if the service has a `HandleIPCEvents(c *Core, msg Message) error` method, it is registered as an action handler automatically.

If the factory returns an error or `nil`, `New()` fails with an error.

If the returned type has no package path (e.g. a primitive or anonymous type), `New()` fails with a descriptive error.

### WithName

```go
func WithName(name string, factory func(*Core) (any, error)) Option
```

Registers a service with an **explicit name**. Use this when the auto-discovered name would be wrong (e.g. anonymous functions, or when you want a different name).

```go
core.New(
    core.WithName("greet", func(c *core.Core) (any, error) {
        return &Greeter{}, nil
    }),
)
```

Unlike `WithService`, `WithName` does **not** auto-discover IPC handlers. If your service needs to handle actions, register the handler manually:

```go
core.WithName("greet", func(c *core.Core) (any, error) {
    svc := &Greeter{}
    c.RegisterAction(svc.HandleIPCEvents)
    return svc, nil
}),
```

### WithApp

```go
func WithApp(app any) Option
```

Injects a GUI runtime (e.g. a Wails App instance) into the Core. The app is stored in the `Core.App` field and can be accessed globally via `core.App()` after `SetInstance` is called.

```go
core.New(
    core.WithApp(wailsApp),
)
```

This is primarily used for desktop applications where services need access to the windowing runtime.

### WithAssets

```go
func WithAssets(fs embed.FS) Option
```

Registers the application's embedded assets filesystem. Retrieve it later with `c.Assets()`.

```go
//go:embed frontend/dist
var assets embed.FS

core.New(
    core.WithAssets(assets),
)
```

### WithServiceLock

```go
func WithServiceLock() Option
```

Prevents any services from being registered after `New()` returns. Any call to `RegisterService` after initialisation will return an error.

```go
c, err := core.New(
    core.WithService(myService),
    core.WithServiceLock(), // no more services can be added
)
// c.RegisterService("late", &svc) -> error
```

This is a safety measure to ensure all services are declared upfront, preventing accidental late-binding that could cause ordering or lifecycle issues.

**How it works:** The lock is recorded during option processing but only **applied** after all options have been processed. This means options that register services (like `WithService`) can appear in any order relative to `WithServiceLock`.

## Option Ordering

Options are applied in the order they are passed to `New()`. This means:

- Services registered earlier are available to later factories (via `c.Service()`).
- `WithServiceLock()` can appear at any position -- it only takes effect after all options have been processed.
- `WithApp` and `WithAssets` can appear at any position.

```go
core.New(
    core.WithServiceLock(),           // recorded, not yet applied
    core.WithService(factory1),       // succeeds (lock not yet active)
    core.WithService(factory2),       // succeeds
    // After New() returns, the lock is applied
)
```

## Global Instance

For applications that need global access to the Core (typically GUI runtimes), there is a global instance mechanism:

```go
// Set the global instance (typically during app startup)
core.SetInstance(c)

// Retrieve it (panics if not set)
app := core.App()

// Non-panicking access
c := core.GetInstance()
if c == nil {
    // not set
}

// Clear it (useful in tests)
core.ClearInstance()
```

These functions are thread-safe.

## Features

The `Core` struct includes a `Features` field for simple feature flagging:

```go
c.Features.Flags = []string{"experimental-ui", "beta-api"}

if c.Features.IsEnabled("experimental-ui") {
    // enable experimental UI
}
```

Feature flags are string-matched (case-sensitive). This is a lightweight mechanism -- for complex feature management, register a dedicated service.

## Related Pages

- [Services](services.md) -- service registration and retrieval
- [Lifecycle](lifecycle.md) -- startup/shutdown after configuration
- [Getting Started](getting-started.md) -- end-to-end example
