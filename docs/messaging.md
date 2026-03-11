---
title: Messaging
description: ACTION, QUERY, and PERFORM -- the message bus for decoupled service communication.
---

# Messaging

The message bus enables services to communicate without importing each other. It supports three patterns:

| Pattern | Method | Semantics |
|---------|--------|-----------|
| **ACTION** | `c.ACTION(msg)` | Broadcast to all handlers (fire-and-forget) |
| **QUERY** | `c.QUERY(q)` | First responder wins (read-only) |
| **PERFORM** | `c.PERFORM(t)` | First responder executes (side effects) |

All three are type-safe at the handler level through Go type switches, while the bus itself uses `any` to avoid import cycles.

## Message Types

```go
// Any struct can be a message -- no interface to implement.
type Message any   // Used with ACTION
type Query any     // Used with QUERY / QUERYALL
type Task any      // Used with PERFORM / PerformAsync
```

Define your message types as plain structs:

```go
// In your package
type UserCreated struct {
    UserID string
    Email  string
}

type GetUserCount struct{}

type SendEmail struct {
    To      string
    Subject string
    Body    string
}
```

## ACTION -- Broadcast

`ACTION` dispatches a message to **every** registered action handler. Handlers receive the message and can inspect it via type switch. All handlers are called regardless of whether they handle the specific message type.

### Dispatching

```go
err := c.ACTION(UserCreated{UserID: "123", Email: "user@example.com"})
```

Errors from all handlers are aggregated with `errors.Join`. If no handlers are registered, `ACTION` returns `nil`.

### Handling

```go
c.RegisterAction(func(c *core.Core, msg core.Message) error {
    switch m := msg.(type) {
    case UserCreated:
        fmt.Printf("New user: %s (%s)\n", m.UserID, m.Email)
    }
    return nil
})
```

You can register multiple handlers. Each handler receives every message -- use a type switch to filter.

```go
// Register multiple handlers at once
c.RegisterActions(handler1, handler2, handler3)
```

### Auto-Discovery

If a service registered via `WithService` has a method called `HandleIPCEvents` with the signature `func(*Core, Message) error`, it is automatically registered as an action handler:

```go
type Service struct{}

func (s *Service) HandleIPCEvents(c *core.Core, msg core.Message) error {
    switch msg.(type) {
    case UserCreated:
        // react to event
    }
    return nil
}
```

## QUERY -- Request/Response

`QUERY` dispatches a query to handlers in registration order. The **first** handler that returns `handled == true` wins -- subsequent handlers are not called.

### Dispatching

```go
result, handled, err := c.QUERY(GetUserCount{})
if !handled {
    // no handler recognised this query
}
count := result.(int)
```

### Handling

```go
c.RegisterQuery(func(c *core.Core, q core.Query) (any, bool, error) {
    switch q.(type) {
    case GetUserCount:
        return 42, true, nil
    }
    return nil, false, nil // not handled -- pass to next handler
})
```

Return `false` for `handled` to let the query fall through to the next handler.

### QUERYALL -- Collect All Responses

`QUERYALL` calls **every** handler and collects all responses where `handled == true`:

```go
results, err := c.QUERYALL(GetPluginInfo{})
// results contains one entry per handler that responded
```

Errors from all handlers are aggregated. Results from handlers that returned `handled == false` or `result == nil` are excluded.

## PERFORM -- Execute a Task

`PERFORM` dispatches a task to handlers in registration order. Like `QUERY`, the first handler that returns `handled == true` wins.

### Dispatching

```go
result, handled, err := c.PERFORM(SendEmail{
    To:      "user@example.com",
    Subject: "Welcome",
    Body:    "Hello!",
})
if !handled {
    // no handler could execute this task
}
```

### Handling

```go
c.RegisterTask(func(c *core.Core, t core.Task) (any, bool, error) {
    switch m := t.(type) {
    case SendEmail:
        err := sendMail(m.To, m.Subject, m.Body)
        return nil, true, err
    }
    return nil, false, nil
})
```

## PerformAsync -- Background Tasks

`PerformAsync` dispatches a task to be executed in a background goroutine. It returns a task ID immediately.

```go
taskID := c.PerformAsync(SendEmail{
    To:      "user@example.com",
    Subject: "Report",
    Body:    "...",
})
// taskID is something like "task-1"
```

The lifecycle of an async task produces three action messages:

| Message | When |
|---------|------|
| `ActionTaskStarted{TaskID, Task}` | Immediately, before execution begins |
| `ActionTaskProgress{TaskID, Task, Progress, Message}` | When `c.Progress()` is called |
| `ActionTaskCompleted{TaskID, Task, Result, Error}` | After execution finishes |

### Listening for Completion

```go
c.RegisterAction(func(c *core.Core, msg core.Message) error {
    switch m := msg.(type) {
    case core.ActionTaskCompleted:
        fmt.Printf("Task %s finished: result=%v err=%v\n",
            m.TaskID, m.Result, m.Error)
    }
    return nil
})
```

### Reporting Progress

From within a task handler (or anywhere that has the task ID):

```go
c.Progress(taskID, 0.5, "halfway done", myTask)
```

This broadcasts an `ActionTaskProgress` message.

### TaskWithID

If your task struct implements `TaskWithID`, `PerformAsync` will inject the assigned task ID before dispatching:

```go
type TaskWithID interface {
    Task
    SetTaskID(id string)
    GetTaskID() string
}
```

```go
type MyLongTask struct {
    id string
}

func (t *MyLongTask) SetTaskID(id string) { t.id = id }
func (t *MyLongTask) GetTaskID() string   { return t.id }
```

### Shutdown Behaviour

- `PerformAsync` returns an empty string if the Core is already shut down.
- `ServiceShutdown` waits for all in-flight async tasks to complete (respecting the context deadline).

## Real-World Example: Log Service

The `pkg/log` service demonstrates both query and task handling:

```go
// Query type: "what is the current log level?"
type QueryLevel struct{}

// Task type: "change the log level"
type TaskSetLevel struct {
    Level Level
}

func (s *Service) OnStartup(ctx context.Context) error {
    s.Core().RegisterQuery(s.handleQuery)
    s.Core().RegisterTask(s.handleTask)
    return nil
}

func (s *Service) handleQuery(c *core.Core, q core.Query) (any, bool, error) {
    switch q.(type) {
    case QueryLevel:
        return s.Level(), true, nil
    }
    return nil, false, nil
}

func (s *Service) handleTask(c *core.Core, t core.Task) (any, bool, error) {
    switch m := t.(type) {
    case TaskSetLevel:
        s.SetLevel(m.Level)
        return nil, true, nil
    }
    return nil, false, nil
}
```

Other services can query or change the log level without importing the log package:

```go
// Query the level
level, handled, _ := c.QUERY(log.QueryLevel{})

// Change the level
_, _, _ = c.PERFORM(log.TaskSetLevel{Level: log.LevelDebug})
```

## Thread Safety

The message bus uses `sync.RWMutex` for each handler list (actions, queries, tasks). Registration and dispatch are safe to call concurrently from multiple goroutines. Handler lists are snapshot-cloned before dispatch, so handlers registered during dispatch will not be called until the next dispatch.

## Related Pages

- [Services](services.md) -- how services are registered
- [Lifecycle](lifecycle.md) -- `ActionServiceStartup` and `ActionServiceShutdown` messages
- [Testing](testing.md) -- testing message handlers
