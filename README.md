# CoreGO

Dependency injection, service lifecycle, permission, and message-passing for Go.

```go
import "dappco.re/go/core"
```

CoreGO is the foundation layer for the Core ecosystem. It gives you:

- one container: `Core`
- one input shape: `Options`
- one output shape: `Result`
- one command tree: `Command`
- one message bus: `ACTION`, `QUERY` + named `Action` callables
- one permission gate: `Entitled`
- one collection primitive: `Registry[T]`

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

## Core Surfaces

| Surface | Purpose |
|---------|---------|
| `Core` | Central container and access point |
| `Service` | Managed lifecycle component |
| `Command` | Path-based executable operation |
| `Action` | Named callable with panic recovery + entitlement |
| `Task` | Composed sequence of Actions |
| `Registry[T]` | Thread-safe named collection |
| `Process` | Managed execution (Action sugar) |
| `API` | Remote streams (protocol handlers) |
| `Entitlement` | Permission check result |
| `Data` | Embedded filesystem mounts |
| `Drive` | Named transport handles |
| `Fs` | Local filesystem (sandboxable) |
| `Config` | Runtime settings and feature flags |

## Install

```bash
go get dappco.re/go/core
```

Requires Go 1.26 or later.

## Test

```bash
go test ./...    # 483 tests, 84.7% coverage
```

## Docs

The authoritative API contract is `docs/RFC.md` (21 sections).

## License

EUPL-1.2
