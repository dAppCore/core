---
title: Core Go Framework
description: Dependency injection and service lifecycle framework for Go.
---

# Core Go Framework

Core (`forge.lthn.ai/core/go`) is a dependency injection and service lifecycle framework for Go. It provides a typed service registry, lifecycle hooks, and a message-passing bus for decoupled communication between services.

This is the foundation layer of the ecosystem. It has no CLI, no GUI, and minimal dependencies.

## Installation

```bash
go get forge.lthn.ai/core/go
```

Requires Go 1.26 or later.

## What It Does

Core solves three problems that every non-trivial Go application eventually faces:

1. **Service wiring** -- how do you register, retrieve, and type-check services without import cycles?
2. **Lifecycle management** -- how do you start and stop services in the right order?
3. **Decoupled communication** -- how do services talk to each other without knowing each other's types?

## Packages

| Package | Purpose |
|---------|---------|
| [`pkg/core`](services.md) | DI container, service registry, lifecycle, message bus |
| `pkg/log` | Structured logger service with Core integration |

## Quick Example

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
        core.WithServiceLock(), // Prevent late registration
    )
    if err != nil {
        panic(err)
    }

    // Start all services
    if err := c.ServiceStartup(context.Background(), nil); err != nil {
        panic(err)
    }

    // Type-safe retrieval
    logger, err := core.ServiceFor[*log.Service](c, "log")
    if err != nil {
        panic(err)
    }
    fmt.Println("Log level:", logger.Level())

    // Shut down (reverse order)
    _ = c.ServiceShutdown(context.Background())
}
```

## Documentation

| Page | Covers |
|------|--------|
| [Getting Started](getting-started.md) | Creating a Core app, registering your first service |
| [Services](services.md) | Service registration, `ServiceRuntime`, factory pattern |
| [Lifecycle](lifecycle.md) | `Startable`/`Stoppable` interfaces, startup/shutdown order |
| [Messaging](messaging.md) | ACTION, QUERY, PERFORM -- the message bus |
| [Configuration](configuration.md) | `WithService`, `WithName`, `WithAssets`, `WithServiceLock` options |
| [Testing](testing.md) | Test naming conventions, test helpers, fuzz testing |
| [Errors](errors.md) | `E()` helper, `Error` struct, unwrapping |

## Dependencies

Core is deliberately minimal:

- `forge.lthn.ai/core/go-io` -- abstract storage (local, S3, SFTP, WebDAV)
- `forge.lthn.ai/core/go-log` -- structured logging
- `github.com/stretchr/testify` -- test assertions (test-only)

## Licence

EUPL-1.2
