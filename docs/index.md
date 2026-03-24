---
title: CoreGO
description: AX-first documentation for the CoreGO framework.
---

# CoreGO

CoreGO is the foundation layer for the Core ecosystem. It gives you one container, one command tree, one message bus, and a small set of shared primitives that repeat across the whole framework.

The current module path is `dappco.re/go/core`.

## AX View

CoreGO already follows the main AX ideas from RFC-025:

- predictable names such as `Core`, `Service`, `Command`, `Options`, `Result`, `Message`
- path-shaped command registration such as `deploy/to/homelab`
- one repeated input shape (`Options`) and one repeated return shape (`Result`)
- comments and examples that show real usage instead of restating the type signature

## What CoreGO Owns

| Surface | Purpose |
|---------|---------|
| `Core` | Central container and access point |
| `Service` | Managed lifecycle component |
| `Command` | Path-based command tree node |
| `ACTION`, `QUERY`, `PERFORM` | Decoupled communication between components |
| `Data`, `Drive`, `Fs`, `Config`, `I18n`, `Cli` | Built-in subsystems for common runtime work |
| `E`, `Wrap`, `ErrorLog`, `ErrorPanic` | Structured failures and panic recovery |

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
			core.Info("cache ready", "app", c.App().Name)
			return core.Result{OK: true}
		},
		OnStop: func() core.Result {
			core.Info("cache stopped", "app", c.App().Name)
			return core.Result{OK: true}
		},
	})

	c.RegisterTask(func(_ *core.Core, task core.Task) core.Result {
		switch task.(type) {
		case flushCacheTask:
			return core.Result{Value: "cache flushed", OK: true}
		}
		return core.Result{}
	})

	c.Command("cache/flush", core.Command{
		Action: func(opts core.Options) core.Result {
			return c.PERFORM(flushCacheTask{Name: opts.String("name")})
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

## Documentation Paths

| Path | Covers |
|------|--------|
| [getting-started.md](getting-started.md) | First runnable CoreGO app |
| [primitives.md](primitives.md) | `Options`, `Result`, `Service`, `Message`, `Query`, `Task` |
| [services.md](services.md) | Service registry, service locks, runtime helpers |
| [commands.md](commands.md) | Path-based commands and CLI execution |
| [messaging.md](messaging.md) | `ACTION`, `QUERY`, `QUERYALL`, `PERFORM`, `PerformAsync` |
| [lifecycle.md](lifecycle.md) | Startup, shutdown, context, background task draining |
| [configuration.md](configuration.md) | Constructor options, config state, feature flags |
| [subsystems.md](subsystems.md) | `App`, `Data`, `Drive`, `Fs`, `I18n`, `Cli` |
| [errors.md](errors.md) | Structured errors, logging helpers, panic recovery |
| [testing.md](testing.md) | Test naming and framework-level testing patterns |
| [pkg/core.md](pkg/core.md) | Package-level reference summary |
| [pkg/log.md](pkg/log.md) | Logging reference for the root package |
| [pkg/PACKAGE_STANDARDS.md](pkg/PACKAGE_STANDARDS.md) | AX package-authoring guidance |

## Good Reading Order

1. Start with [getting-started.md](getting-started.md).
2. Learn the repeated shapes in [primitives.md](primitives.md).
3. Pick the integration path you need next: [services.md](services.md), [commands.md](commands.md), or [messaging.md](messaging.md).
4. Use [subsystems.md](subsystems.md), [errors.md](errors.md), and [testing.md](testing.md) as reference pages while building.
