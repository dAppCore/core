---
title: Messaging
description: ACTION, QUERY, QUERYALL, named Actions, and async dispatch.
---

# Messaging

CoreGO has two messaging layers: anonymous broadcast (ACTION/QUERY) and named Actions.

## Anonymous Broadcast

### `ACTION`

Fire-and-forget broadcast to all registered handlers. Each handler is wrapped in panic recovery. Handler return values are ignored — all handlers fire regardless.

```go
c.RegisterAction(func(_ *core.Core, msg core.Message) core.Result {
    if ev, ok := msg.(repositoryIndexed); ok {
        core.Info("indexed", "name", ev.Name)
    }
    return core.Result{OK: true}
})

c.ACTION(repositoryIndexed{Name: "core-go"})
```

### `QUERY`

First handler to return `OK:true` wins.

```go
c.RegisterQuery(func(_ *core.Core, q core.Query) core.Result {
    if _, ok := q.(repositoryCountQuery); ok {
        return core.Result{Value: 42, OK: true}
    }
    return core.Result{}
})

r := c.QUERY(repositoryCountQuery{})
```

### `QUERYALL`

Collects every successful non-nil response.

```go
r := c.QUERYALL(repositoryCountQuery{})
results := r.Value.([]any)
```

## Named Actions

Named Actions are the typed, inspectable replacement for anonymous dispatch. See Section 18 of `RFC.md`.

### Register and Invoke

```go
// Register during OnStartup
c.Action("repo.sync", func(ctx context.Context, opts core.Options) core.Result {
    name := opts.String("name")
    return core.Result{Value: "synced " + name, OK: true}
})

// Invoke by name
r := c.Action("repo.sync").Run(ctx, core.NewOptions(
    core.Option{Key: "name", Value: "core-go"},
))
```

### Capability Check

```go
if c.Action("process.run").Exists() {
    // go-process is registered
}

c.Actions()  // []string of all registered action names
```

### Permission Gate

Every `Action.Run()` checks `c.Entitled(action.Name)` before executing. See Section 21 of `RFC.md`.

## Task Composition

A Task is a named sequence of Actions:

```go
c.Task("deploy", core.Task{
    Steps: []core.Step{
        {Action: "go.build"},
        {Action: "go.test"},
        {Action: "docker.push"},
        {Action: "notify.slack", Async: true},
    },
})

r := c.Task("deploy").Run(ctx, c, opts)
```

Sequential steps stop on first failure. `Async: true` steps fire without blocking. `Input: "previous"` pipes output.

## Background Execution

```go
r := c.PerformAsync("repo.sync", opts)
taskID := r.Value.(string)

c.Progress(taskID, 0.5, "indexing commits", "repo.sync")
```

Broadcasts `ActionTaskStarted`, `ActionTaskProgress`, `ActionTaskCompleted` as ACTION messages.

### Completion Listener

```go
c.RegisterAction(func(_ *core.Core, msg core.Message) core.Result {
    if ev, ok := msg.(core.ActionTaskCompleted); ok {
        core.Info("done", "task", ev.TaskIdentifier, "ok", ev.Result.OK)
    }
    return core.Result{OK: true}
})
```

## Shutdown

When shutdown has started, `PerformAsync` returns an empty `Result`. `ServiceShutdown` drains outstanding background work before stopping services.
