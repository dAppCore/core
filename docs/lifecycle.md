---
title: Lifecycle
description: Startable and Stoppable interfaces, startup and shutdown ordering.
---

# Lifecycle

Core manages the startup and shutdown of services through two opt-in interfaces. Services implement one or both to participate in the application lifecycle.

## Interfaces

### Startable

```go
type Startable interface {
    OnStartup(ctx context.Context) error
}
```

Services implementing `Startable` have their `OnStartup` method called during `ServiceStartup`. This is the place to:

- Open database connections
- Register message bus handlers (queries, tasks)
- Start background workers
- Validate configuration

### Stoppable

```go
type Stoppable interface {
    OnShutdown(ctx context.Context) error
}
```

Services implementing `Stoppable` have their `OnShutdown` method called during `ServiceShutdown`. This is the place to:

- Close database connections
- Flush buffers
- Save state
- Cancel background workers

A service can implement both interfaces:

```go
type Service struct{}

func (s *Service) OnStartup(ctx context.Context) error {
    // Initialise resources
    return nil
}

func (s *Service) OnShutdown(ctx context.Context) error {
    // Release resources
    return nil
}
```

## Ordering

### Startup: Registration Order

Services are started in the order they were registered. If you register services A, B, C (in that order), their `OnStartup` methods are called as A, B, C.

### Shutdown: Reverse Registration Order

Services are stopped in **reverse** registration order. If A, B, C were registered, their `OnShutdown` methods are called as C, B, A.

This ensures that services which depend on earlier services are torn down first.

```go
c, err := core.New()

_ = c.RegisterService("database", dbService)   // started 1st, stopped 3rd
_ = c.RegisterService("cache", cacheService)     // started 2nd, stopped 2nd
_ = c.RegisterService("api", apiService)         // started 3rd, stopped 1st

_ = c.ServiceStartup(ctx, nil)  // database -> cache -> api
_ = c.ServiceShutdown(ctx)      // api -> cache -> database
```

## ServiceStartup

```go
func (c *Core) ServiceStartup(ctx context.Context, options any) error
```

`ServiceStartup` does two things, in order:

1. Calls `OnStartup(ctx)` on every `Startable` service, in registration order.
2. Broadcasts an `ActionServiceStartup{}` message via the message bus.

If any service returns an error, it is collected but does **not** prevent other services from starting. All errors are aggregated with `errors.Join` and returned together.

If the context is cancelled before all services have started, the remaining services are skipped and the context error is included in the aggregate.

## ServiceShutdown

```go
func (c *Core) ServiceShutdown(ctx context.Context) error
```

`ServiceShutdown` does three things, in order:

1. Broadcasts an `ActionServiceShutdown{}` message via the message bus.
2. Calls `OnShutdown(ctx)` on every `Stoppable` service, in reverse registration order.
3. Waits for any in-flight `PerformAsync` background tasks to complete (respecting the context deadline).

As with startup, errors are aggregated rather than short-circuiting. If the context is cancelled during shutdown, the remaining services are skipped but the method still waits for background tasks.

## Built-in Lifecycle Messages

Core broadcasts two action messages as part of the lifecycle. You can listen for these in any registered action handler:

| Message | When |
|---------|------|
| `ActionServiceStartup{}` | After all `Startable` services have been called |
| `ActionServiceShutdown{}` | Before `Stoppable` services are called |

```go
c.RegisterAction(func(c *core.Core, msg core.Message) error {
    switch msg.(type) {
    case core.ActionServiceStartup:
        // All services are up
    case core.ActionServiceShutdown:
        // Shutdown is beginning
    }
    return nil
})
```

## Error Handling

Lifecycle methods never panic. All errors from individual services are collected via `errors.Join` and returned as a single error. You can inspect individual errors with `errors.Is` and `errors.As`:

```go
err := c.ServiceStartup(ctx, nil)
if err != nil {
    // err may contain multiple wrapped errors
    if errors.Is(err, context.Canceled) {
        // context was cancelled
    }
}
```

## Context Cancellation

Both `ServiceStartup` and `ServiceShutdown` respect context cancellation. If the context is cancelled or its deadline is exceeded, the remaining services are skipped:

```go
ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
defer cancel()

err := c.ServiceStartup(ctx, nil)
// If startup takes longer than 5 seconds, remaining services are skipped
```

## Detection

Lifecycle interface detection happens at registration time. When you call `RegisterService`, Core checks whether the service implements `Startable` and/or `Stoppable` and adds it to the appropriate internal list. There is no need to declare anything beyond implementing the interface.

## Related Pages

- [Services](services.md) -- how services are registered
- [Messaging](messaging.md) -- the `ACTION` broadcast used during lifecycle
- [Configuration](configuration.md) -- `WithServiceLock` and other options
