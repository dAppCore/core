---
title: Getting Started
description: How to create a Core application and register services.
---

# Getting Started

This guide walks you through creating a Core application, registering services, and running the lifecycle.

## Installation

```bash
go get forge.lthn.ai/core/go
```

## Creating a Core Instance

Everything starts with `core.New()`. It accepts a variadic list of `Option` functions that configure the container before it is returned.

```go
package main

import "forge.lthn.ai/core/go/pkg/core"

func main() {
    c, err := core.New()
    if err != nil {
        panic(err)
    }
    _ = c // empty container, ready for use
}
```

In practice you will pass options to register services, embed assets, or lock the registry:

```go
c, err := core.New(
    core.WithService(mypackage.NewService),
    core.WithAssets(embeddedFS),
    core.WithServiceLock(),
)
```

See [Configuration](configuration.md) for the full list of options.

## Registering a Service

Services are registered via **factory functions**. A factory receives the `*Core` and returns `(any, error)`:

```go
package greeter

import "forge.lthn.ai/core/go/pkg/core"

type Service struct {
    greeting string
}

func (s *Service) Hello(name string) string {
    return s.greeting + ", " + name + "!"
}

func NewService(c *core.Core) (any, error) {
    return &Service{greeting: "Hello"}, nil
}
```

Register it with `WithService`:

```go
c, err := core.New(
    core.WithService(greeter.NewService),
)
```

`WithService` automatically discovers the service name from the package path. In this case, the service is registered under the name `"greeter"`.

If you need to control the name explicitly, use `WithName`:

```go
c, err := core.New(
    core.WithName("greet", greeter.NewService),
)
```

See [Services](services.md) for the full registration API and the `ServiceRuntime` helper.

## Retrieving a Service

Once registered, services can be retrieved by name:

```go
// Untyped retrieval (returns any)
svc := c.Service("greeter")

// Type-safe retrieval (returns error if not found or wrong type)
greet, err := core.ServiceFor[*greeter.Service](c, "greeter")

// Panicking retrieval (for init-time wiring where failure is fatal)
greet := core.MustServiceFor[*greeter.Service](c, "greeter")
```

## Running the Lifecycle

Services that implement `Startable` and/or `Stoppable` are automatically called during startup and shutdown:

```go
import "context"

// Start all Startable services (in registration order)
err := c.ServiceStartup(context.Background(), nil)

// ... application runs ...

// Stop all Stoppable services (in reverse registration order)
err = c.ServiceShutdown(context.Background())
```

See [Lifecycle](lifecycle.md) for details on the `Startable` and `Stoppable` interfaces.

## Sending Messages

Services communicate through the message bus without needing direct imports of each other:

```go
// Broadcast to all handlers (fire-and-forget)
err := c.ACTION(MyEvent{Data: "something happened"})

// Request data from the first handler that responds
result, handled, err := c.QUERY(MyQuery{Key: "setting"})

// Ask a handler to perform work
result, handled, err := c.PERFORM(MyTask{Input: "data"})
```

See [Messaging](messaging.md) for the full message bus API.

## Putting It All Together

Here is a minimal but complete application:

```go
package main

import (
    "context"
    "fmt"

    "forge.lthn.ai/core/go/pkg/core"
    "forge.lthn.ai/core/go/pkg/log"
)

func main() {
    c, err := core.New(
        core.WithName("log", log.NewService(log.Options{Level: log.LevelInfo})),
        core.WithServiceLock(),
    )
    if err != nil {
        panic(err)
    }

    // Start lifecycle
    if err := c.ServiceStartup(context.Background(), nil); err != nil {
        panic(err)
    }

    // Use services
    logger := core.MustServiceFor[*log.Service](c, "log")
    fmt.Println("Logger started at level:", logger.Level())

    // Query the log level through the message bus
    level, handled, _ := c.QUERY(log.QueryLevel{})
    if handled {
        fmt.Println("Log level via QUERY:", level)
    }

    // Clean shutdown
    if err := c.ServiceShutdown(context.Background()); err != nil {
        fmt.Println("shutdown error:", err)
    }
}
```

## Next Steps

- [Services](services.md) -- service registration patterns in depth
- [Lifecycle](lifecycle.md) -- startup/shutdown ordering and error handling
- [Messaging](messaging.md) -- ACTION, QUERY, and PERFORM
- [Configuration](configuration.md) -- all `With*` options
- [Errors](errors.md) -- the `E()` error helper
- [Testing](testing.md) -- test conventions and helpers
