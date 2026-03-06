# Core Go Framework — Documentation

Core (`forge.lthn.ai/core/go`) is a **dependency injection and service lifecycle framework** for Go. It provides a typed service registry, lifecycle hooks, and a message-passing bus for decoupled communication between services.

This is the foundation layer — it has no CLI, no GUI, and minimal dependencies (`go-io`, `go-log`, `testify`).

---

## Packages

| Package | Description |
|---------|-------------|
| `pkg/core` | DI container, service registry, lifecycle, query/task bus |
| `pkg/log` | Structured logger service with Core integration |

---

## Quick Start

```go
import (
    "forge.lthn.ai/core/go/pkg/core"
    "forge.lthn.ai/core/go/pkg/log"
)

c, err := core.New(
    core.WithName("log", log.NewService(log.Options{Level: log.LevelInfo})),
    core.WithName("myservice", mypackage.NewService(opts)),
)
// Services implementing Startable/Stoppable are called automatically.
```

### Type-safe service retrieval

```go
svc, err := core.ServiceFor[*mypackage.Service](c, "myservice")
```

### Query/Task bus

Services communicate via typed messages without direct imports:

```go
// Query: request data from a service
result, err := c.Query(log.QueryLevel{})

// Task: ask a service to do something
c.Task(log.TaskSetLevel{Level: log.LevelDebug})
```

---

## Architecture

See [Package Standards](pkg/PACKAGE_STANDARDS.md) for the canonical patterns:
- Service struct with `core.ServiceRuntime[T]`
- Factory functions for registration
- Lifecycle hooks (`Startable`, `Stoppable`)
- Thread-safe APIs
- Query/Task handlers

See [Log Service](pkg/log.md) for the reference implementation within this repo.

---

## Ecosystem

This framework is consumed by 20+ standalone Go modules under `forge.lthn.ai/core/`. The CLI, AI, ML, and infrastructure packages all build on this DI container.

For CLI documentation, see `forge.lthn.ai/core/cli`.
