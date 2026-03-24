---
title: Messaging
description: ACTION, QUERY, QUERYALL, PERFORM, and async task flow.
---

# Messaging

CoreGO uses one message bus for broadcasts, lookups, and work dispatch.

## Message Types

```go
type Message any
type Query any
type Task any
```

Your own structs define the protocol.

```go
type repositoryIndexed struct {
	Name string
}

type repositoryCountQuery struct{}

type syncRepositoryTask struct {
	Name string
}
```

## `ACTION`

`ACTION` is a broadcast.

```go
c.RegisterAction(func(_ *core.Core, msg core.Message) core.Result {
	switch m := msg.(type) {
	case repositoryIndexed:
		core.Info("repository indexed", "name", m.Name)
		return core.Result{OK: true}
	}
	return core.Result{OK: true}
})

r := c.ACTION(repositoryIndexed{Name: "core-go"})
```

### Behavior

- all registered action handlers are called in their current registration order
- if a handler returns `OK:false`, dispatch stops and that `Result` is returned
- if no handler fails, `ACTION` returns `Result{OK:true}`

## `QUERY`

`QUERY` is first-match request-response.

```go
c.RegisterQuery(func(_ *core.Core, q core.Query) core.Result {
	switch q.(type) {
	case repositoryCountQuery:
		return core.Result{Value: 42, OK: true}
	}
	return core.Result{}
})

r := c.QUERY(repositoryCountQuery{})
```

### Behavior

- handlers run until one returns `OK:true`
- the first successful result wins
- if nothing handles the query, CoreGO returns an empty `Result`

## `QUERYALL`

`QUERYALL` collects every successful non-nil response.

```go
r := c.QUERYALL(repositoryCountQuery{})
results := r.Value.([]any)
```

### Behavior

- every query handler is called
- only `OK:true` results with non-nil `Value` are collected
- the call itself returns `OK:true` even when the result list is empty

## `PERFORM`

`PERFORM` dispatches a task to the first handler that accepts it.

```go
c.RegisterTask(func(_ *core.Core, t core.Task) core.Result {
	switch task := t.(type) {
	case syncRepositoryTask:
		return core.Result{Value: "synced " + task.Name, OK: true}
	}
	return core.Result{}
})

r := c.PERFORM(syncRepositoryTask{Name: "core-go"})
```

### Behavior

- handlers run until one returns `OK:true`
- the first successful result wins
- if nothing handles the task, CoreGO returns an empty `Result`

## `PerformAsync`

`PerformAsync` runs a task in a background goroutine and returns a generated task identifier.

```go
r := c.PerformAsync(syncRepositoryTask{Name: "core-go"})
taskID := r.Value.(string)
```

### Generated Events

Async execution emits three action messages:

| Message | When |
|---------|------|
| `ActionTaskStarted` | just before background execution begins |
| `ActionTaskProgress` | whenever `Progress` is called |
| `ActionTaskCompleted` | after the task finishes or panics |

Example listener:

```go
c.RegisterAction(func(_ *core.Core, msg core.Message) core.Result {
	switch m := msg.(type) {
	case core.ActionTaskCompleted:
		core.Info("task completed", "task", m.TaskIdentifier, "err", m.Error)
	}
	return core.Result{OK: true}
})
```

## Progress Updates

```go
c.Progress(taskID, 0.5, "indexing commits", syncRepositoryTask{Name: "core-go"})
```

That broadcasts `ActionTaskProgress`.

## `TaskWithIdentifier`

Tasks that implement `TaskWithIdentifier` receive the generated ID before dispatch.

```go
type trackedTask struct {
	ID   string
	Name string
}

func (t *trackedTask) SetTaskIdentifier(id string) { t.ID = id }
func (t *trackedTask) GetTaskIdentifier() string   { return t.ID }
```

## Shutdown Interaction

When shutdown has started, `PerformAsync` returns an empty `Result` instead of scheduling more work.

This is why `ServiceShutdown` can safely drain the outstanding background tasks before stopping services.
