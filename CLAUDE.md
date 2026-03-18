# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Session Context

Running on **Claude Max20 plan** with **1M context window** (Opus 4.6). This enables marathon sessions — use the full context for complex multi-repo work, dispatch coordination, and ecosystem-wide operations. Compact when needed, but don't be afraid of long sessions.

## Project Overview

Core (`forge.lthn.ai/core/go`) is a **dependency injection and service lifecycle framework** for Go. It provides a typed service registry, lifecycle hooks, and a message-passing bus for decoupled communication between services.

This is the foundation layer — it has no CLI, no GUI, and minimal dependencies (`go-io`, `go-log`, `testify`).

## Build & Development Commands

This project uses `core go` commands (no Taskfile). Build configuration lives in `.core/build.yaml`.

```bash
# Run all tests
core go test

# Generate test coverage
core go cov
core go cov --open      # Opens coverage HTML report

# Format, lint, vet
core go fmt
core go lint
core go vet

# Quality assurance
core go qa              # fmt + vet + lint + test
core go qa full         # + race, vuln, security

# Build
core build              # Auto-detects project type
core build --ci         # All targets, JSON output
```

Run a single test: `core go test --run TestName`

## Architecture

### Core Framework (`pkg/core/`)

The `Core` struct is the central application container managing:
- **Services**: Named service registry with type-safe retrieval via `ServiceFor[T]()`
- **Actions/IPC**: Message-passing system where services communicate via `ACTION(msg Message)` and register handlers via `RegisterAction()`
- **Lifecycle**: Services implementing `Startable` (OnStartup) and/or `Stoppable` (OnShutdown) interfaces are automatically called during app lifecycle

Creating a Core instance:
```go
core, err := core.New(
    core.WithService(myServiceFactory),
    core.WithAssets(assets),
    core.WithServiceLock(),  // Prevents late service registration
)
```

### Service Registration Pattern

Services are registered via factory functions that receive the Core instance:
```go
func NewMyService(c *core.Core) (any, error) {
    return &MyService{runtime: core.NewServiceRuntime(c, opts)}, nil
}

core.New(core.WithService(NewMyService))
```

- `WithService`: Auto-discovers service name from package path, registers IPC handler if service has `HandleIPCEvents` method
- `WithName`: Explicitly names a service

### ServiceRuntime Generic Helper (`runtime_pkg.go`)

Embed `ServiceRuntime[T]` in services to get access to Core and typed options:
```go
type MyService struct {
    *core.ServiceRuntime[MyServiceOptions]
}
```

### Error Handling (go-log)

All errors MUST use `E()` from `go-log` (re-exported in `e.go`), never `fmt.Errorf`:
```go
return core.E("service.Method", "what failed", underlyingErr)
return core.E("service.Method", fmt.Sprintf("service %q not found", name), nil)
```

### Test Naming Convention

Tests use `_Good`, `_Bad`, `_Ugly` suffix pattern:
- `_Good`: Happy path tests
- `_Bad`: Expected error conditions
- `_Ugly`: Panic/edge cases

## Packages

| Package | Description |
|---------|-------------|
| `pkg/core` | DI container, service registry, lifecycle, query/task bus |
| `pkg/log` | Structured logger service with Core integration |

## Go Workspace

Uses Go 1.26 workspaces. This module is part of the workspace at `~/Code/go.work`.

After adding modules: `go work sync`
