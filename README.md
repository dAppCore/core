# CoreGO

Dependency injection, service lifecycle, command routing, and message-passing for Go.

Import path:

```go
import "dappco.re/go/core"
```

CoreGO is the foundation layer for the Core ecosystem. It gives you:

- one container: `Core`
- one input shape: `Options`
- one output shape: `Result`
- one command tree: `Command`
- one message bus: `ACTION`, `QUERY`, `PERFORM`

## Why It Exists

Most non-trivial Go systems end up needing the same small set of infrastructure:

- a place to keep runtime state and shared subsystems
- a predictable way to start and stop managed components
- a clean command surface for CLI-style workflows
- decoupled communication between components without tight imports

CoreGO keeps those pieces small and explicit.

## Quick Example

```go
package main

import (
	"context"
	"fmt"

	"dappco.re/go/core"
)

type flushCacheTask struct {
	Name string
}

func main() {
	c := core.New(core.Options{
		{Key: "name", Value: "agent-workbench"},
	})

	c.Service("cache", core.Service{
		OnStart: func() core.Result {
			core.Info("cache started", "app", c.App().Name)
			return core.Result{OK: true}
		},
		OnStop: func() core.Result {
			core.Info("cache stopped", "app", c.App().Name)
			return core.Result{OK: true}
		},
	})

	c.RegisterTask(func(_ *core.Core, task core.Task) core.Result {
		switch t := task.(type) {
		case flushCacheTask:
			return core.Result{Value: "cache flushed for " + t.Name, OK: true}
		}
		return core.Result{}
	})

	c.Command("cache/flush", core.Command{
		Action: func(opts core.Options) core.Result {
			return c.PERFORM(flushCacheTask{
				Name: opts.String("name"),
			})
		},
	})

	if !c.ServiceStartup(context.Background(), nil).OK {
		panic("startup failed")
	}

	r := c.Cli().Run("cache", "flush", "--name=session-store")
	fmt.Println(r.Value)

	_ = c.ServiceShutdown(context.Background())
}
```

## Core Surfaces

| Surface | Purpose |
|---------|---------|
| `Core` | Central container and access point |
| `Service` | Managed lifecycle component |
| `Command` | Path-based executable operation |
| `Cli` | CLI surface over the command tree |
| `Data` | Embedded filesystem mounts |
| `Drive` | Named transport handles |
| `Fs` | Local filesystem operations |
| `Config` | Runtime settings and feature flags |
| `I18n` | Locale collection and translation delegation |
| `E`, `Wrap`, `ErrorLog`, `ErrorPanic` | Structured failures and panic recovery |

## AX-Friendly Model

CoreGO follows the same design direction as the AX spec:

- predictable names over compressed names
- paths as documentation, such as `deploy/to/homelab`
- one repeated vocabulary across the framework
- examples that show how to call real APIs

## Install

```bash
go get dappco.re/go/core
```

Requires Go 1.26 or later.

## Test

```bash
core go test
```

Or with the standard toolchain:

```bash
go test ./...
```

## Docs

The full documentation set lives in `docs/`.

| Path | Covers |
|------|--------|
| `docs/getting-started.md` | First runnable CoreGO app |
| `docs/primitives.md` | `Options`, `Result`, `Service`, `Message`, `Query`, `Task` |
| `docs/services.md` | Service registry, runtime helpers, service locks |
| `docs/commands.md` | Path-based commands and CLI execution |
| `docs/messaging.md` | `ACTION`, `QUERY`, `QUERYALL`, `PERFORM`, `PerformAsync` |
| `docs/lifecycle.md` | Startup, shutdown, context, and task draining |
| `docs/subsystems.md` | `App`, `Data`, `Drive`, `Fs`, `I18n`, `Cli` |
| `docs/errors.md` | Structured errors, logging helpers, panic recovery |
| `docs/testing.md` | Test naming and framework testing patterns |

## License

EUPL-1.2
