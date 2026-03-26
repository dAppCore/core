---
title: CoreGO
description: AX-first documentation for the CoreGO framework.
---

# CoreGO

CoreGO is the foundation layer for the Core ecosystem. Module: `dappco.re/go/core`.

## What CoreGO Provides

| Primitive | Purpose |
|-----------|---------|
| `Core` | Central container — everything registers here |
| `Service` | Lifecycle-managed component (Startable/Stoppable return Result) |
| `Action` | Named callable with panic recovery + entitlement |
| `Task` | Composed sequence of Actions |
| `Registry[T]` | Thread-safe named collection (universal brick) |
| `Command` | Path-based CLI command tree |
| `Process` | Managed execution (Action sugar over go-process) |
| `API` | Remote streams (protocol handlers + Drive) |
| `Entitlement` | Permission gate (default permissive, consumer replaces) |
| `ACTION`, `QUERY` | Anonymous broadcast + request/response |
| `Data`, `Drive`, `Fs`, `Config`, `I18n` | Built-in subsystems |

## Quick Example

```go
package main

import "dappco.re/go/core"

func main() {
    c := core.New(
        core.WithOption("name", "agent-workbench"),
        core.WithService(cache.Register),
        core.WithServiceLock(),
    )
    c.Run()
}
```

## API Specification

The full contract is `docs/RFC.md` (21 sections, 1476 lines). An agent should be able to write a service from RFC.md alone.

## Documentation

| Path | Covers |
|------|--------|
| [RFC.md](RFC.md) | Authoritative API contract (21 sections) |
| [primitives.md](primitives.md) | Option, Result, Action, Task, Registry, Entitlement |
| [services.md](services.md) | Service registry, ServiceRuntime, service locks |
| [commands.md](commands.md) | Path-based commands, Managed field |
| [messaging.md](messaging.md) | ACTION, QUERY, named Actions, PerformAsync |
| [lifecycle.md](lifecycle.md) | RunE, ServiceStartup, ServiceShutdown |
| [subsystems.md](subsystems.md) | App, Data, Drive, Fs, Config, I18n |
| [errors.md](errors.md) | core.E(), structured errors, panic recovery |
| [testing.md](testing.md) | AX-7 TestFile_Function_{Good,Bad,Ugly} |
| [configuration.md](configuration.md) | WithOption, WithService, WithServiceLock |
