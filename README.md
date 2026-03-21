# Core

[![codecov](https://codecov.io/gh/dAppCore/core/branch/main/graph/badge.svg)](https://codecov.io/gh/dAppCore/core)
[![Go Version](https://img.shields.io/github/go-mod/go-version/dAppCore/core)](https://go.dev/)
[![License](https://img.shields.io/badge/License-EUPL--1.2-blue.svg)](https://opensource.org/licenses/EUPL-1.2)
[![Go Reference](https://pkg.go.dev/badge/dappco.re/go/core.svg)](https://pkg.go.dev/dappco.re/go/core)

Dependency injection and service lifecycle framework for Go. Zero external dependencies beyond `testify` for tests.

```go
import "dappco.re/go/core"
```

## Quick Start

```go
c := core.New(core.Options{
    {Key: "name", Value: "myapp"},
})

// Register a service
c.Service("auth", core.Service{
    OnStart:  func() core.Result { return core.Result{OK: true} },
    OnStop:   func() core.Result { return core.Result{OK: true} },
})

// Retrieve it
r := c.Service("auth")
if r.OK { /* use r.Value */ }

// Register and run commands
c.Command("deploy", handler)
c.Cli().Run()
```

## Primitives

### Options

Key-value pairs that flow through all subsystems:

```go
opts := core.Options{
    {Key: "name", Value: "brain"},
    {Key: "port", Value: 8080},
}

name := opts.String("name")
port := opts.Int("port")
ok   := opts.Has("debug")
```

### Result

Universal return type replacing `(value, error)`:

```go
r := c.Data().New(core.Options{{Key: "name", Value: "store"}})
if r.OK { use(r.Value) }

// Map from Go conventions
r.Result(file, err)   // OK = err == nil, Value = file
```

### Service

Managed component with optional lifecycle hooks:

```go
core.Service{
    Name:     "cache",
    Options:  opts,
    OnStart:  func() core.Result { /* ... */ },
    OnStop:   func() core.Result { /* ... */ },
    OnReload: func() core.Result { /* ... */ },
}
```

## Subsystems

| Accessor | Purpose |
|----------|---------|
| `c.Options()` | Input configuration |
| `c.App()` | Application identity |
| `c.Data()` | Embedded/stored content |
| `c.Drive()` | Resource handle registry |
| `c.Fs()` | Local filesystem I/O |
| `c.Config()` | Configuration + feature flags |
| `c.Cli()` | CLI surface layer |
| `c.Command("path")` | Command tree |
| `c.Service("name")` | Service registry |
| `c.Lock("name")` | Named mutexes |
| `c.IPC()` | Message bus |

## IPC / Message Bus

Fire-and-forget actions, request/response queries, and task dispatch:

```go
// Register a handler
c.IPC().On(func(c *core.Core, msg core.Message) core.Result {
    // handle message
    return core.Result{OK: true}
})

// Dispatch
c.IPC().Action(core.Message{Action: "cache.flush"})
```

## Install

```bash
go get dappco.re/go/core@latest
```

## License

[EUPL-1.2](https://opensource.org/licenses/EUPL-1.2)
