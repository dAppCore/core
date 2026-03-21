---
title: Lifecycle
description: Startup, shutdown, context ownership, and background task draining.
---

# Lifecycle

CoreGO manages lifecycle through `core.Service` callbacks, not through reflection or implicit interfaces.

## Service Hooks

```go
c.Service("cache", core.Service{
	OnStart: func() core.Result {
		return core.Result{OK: true}
	},
	OnStop: func() core.Result {
		return core.Result{OK: true}
	},
})
```

Only services with `OnStart` appear in `Startables()`. Only services with `OnStop` appear in `Stoppables()`.

## `ServiceStartup`

```go
r := c.ServiceStartup(context.Background(), nil)
```

### What It Does

1. clears the shutdown flag
2. stores a new cancellable context on `c.Context()`
3. runs each `OnStart`
4. broadcasts `ActionServiceStartup{}`

### Failure Behavior

- if the input context is already cancelled, startup returns that error
- if any `OnStart` returns `OK:false`, startup stops immediately and returns that result

## `ServiceShutdown`

```go
r := c.ServiceShutdown(context.Background())
```

### What It Does

1. sets the shutdown flag
2. cancels `c.Context()`
3. broadcasts `ActionServiceShutdown{}`
4. waits for background tasks created by `PerformAsync`
5. runs each `OnStop`

### Failure Behavior

- if draining background tasks hits the shutdown context deadline, shutdown returns that context error
- when service stop hooks fail, CoreGO returns the first error it sees

## Ordering

The current implementation builds `Startables()` and `Stoppables()` by iterating over a map-backed registry.

That means lifecycle order is not guaranteed today.

If your application needs strict startup or shutdown ordering, orchestrate it explicitly inside a smaller number of service callbacks instead of relying on registry order.

## `c.Context()`

`ServiceStartup` creates the context returned by `c.Context()`.

Use it for background work that should stop when the application shuts down:

```go
c.Service("watcher", core.Service{
	OnStart: func() core.Result {
		go func(ctx context.Context) {
			<-ctx.Done()
		}(c.Context())
		return core.Result{OK: true}
	},
})
```

## Built-In Lifecycle Actions

You can listen for lifecycle state changes through the action bus.

```go
c.RegisterAction(func(_ *core.Core, msg core.Message) core.Result {
	switch msg.(type) {
	case core.ActionServiceStartup:
		core.Info("core startup completed")
	case core.ActionServiceShutdown:
		core.Info("core shutdown started")
	}
	return core.Result{OK: true}
})
```

## Background Task Draining

`ServiceShutdown` waits for the internal task waitgroup to finish before calling stop hooks.

This is what makes `PerformAsync` safe for long-running work that should complete before teardown.

## `OnReload`

`Service` includes an `OnReload` callback field, but CoreGO does not currently expose a top-level lifecycle runner for reload operations.
