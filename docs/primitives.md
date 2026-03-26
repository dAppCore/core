---
title: Core Primitives
description: The repeated shapes that make CoreGO easy to navigate.
---

# Core Primitives

CoreGO is built from a small vocabulary repeated everywhere.

## Primitive Map

| Type | Used For |
|------|----------|
| `Option` / `Options` | Input values and metadata |
| `Result` | Output values and success state |
| `Service` | Lifecycle-managed components |
| `Action` | Named callable with panic recovery + entitlement |
| `Task` | Composed sequence of Actions |
| `Registry[T]` | Thread-safe named collection |
| `Entitlement` | Permission check result |
| `Message` | Broadcast events (ACTION) |
| `Query` | Request-response lookups (QUERY) |

## `Option` and `Options`

`Option` is one key-value pair. `Options` is an ordered slice of them.

```go
opts := core.NewOptions(
    core.Option{Key: "name", Value: "brain"},
    core.Option{Key: "path", Value: "prompts"},
    core.Option{Key: "debug", Value: true},
)

name := opts.String("name")
debug := opts.Bool("debug")
raw := opts.Get("name")     // Result{Value, OK}
opts.Has("path")             // true
opts.Len()                   // 3
```

## `Result`

Universal return shape. Every Core operation returns Result.

```go
type Result struct {
    Value any
    OK    bool
}

r := c.Config().Get("host")
if r.OK {
    host := r.Value.(string)
}
```

The `Result()` method adapts Go `(value, error)` pairs:

```go
r := core.Result{}.Result(file, err)
```

## `Service`

Managed lifecycle component stored in the `ServiceRegistry`.

```go
core.Service{
    OnStart: func() core.Result { return core.Result{OK: true} },
    OnStop:  func() core.Result { return core.Result{OK: true} },
}
```

Or via `Startable`/`Stoppable` interfaces (preferred for named services):

```go
type Startable interface { OnStartup(ctx context.Context) Result }
type Stoppable interface { OnShutdown(ctx context.Context) Result }
```

## `Action`

Named callable — the atomic unit of work. Registered by name, invoked by name.

```go
type ActionHandler func(context.Context, Options) Result

type Action struct {
    Name        string
    Handler     ActionHandler
    Description string
    Schema      Options
}
```

`Action.Run()` includes panic recovery and entitlement checking.

## `Task`

Composed sequence of Actions:

```go
type Task struct {
    Name        string
    Description string
    Steps       []Step
}

type Step struct {
    Action string
    With   Options
    Async  bool
    Input  string   // "previous" = output of last step
}
```

## `Registry[T]`

Thread-safe named collection with insertion order and 3 lock modes:

```go
r := core.NewRegistry[*MyService]()
r.Set("brain", svc)
r.Get("brain")       // Result
r.Has("brain")       // bool
r.Names()            // []string (insertion order)
r.Each(func(name string, svc *MyService) { ... })
r.Lock()             // fully frozen
r.Seal()             // no new keys, updates OK
```

## `Entitlement`

Permission check result:

```go
type Entitlement struct {
    Allowed   bool
    Unlimited bool
    Limit     int
    Used      int
    Remaining int
    Reason    string
}

e := c.Entitled("social.accounts", 3)
e.NearLimit(0.8)     // true if > 80% used
e.UsagePercent()     // 75.0
```

## `Message` and `Query`

IPC type aliases for the anonymous broadcast system:

```go
type Message any  // broadcast via ACTION
type Query any    // request/response via QUERY
```

For typed, named dispatch use `c.Action("name").Run(ctx, opts)`.

## `ServiceRuntime[T]`

Composition helper for services that need Core access and typed options:

```go
type MyService struct {
    *core.ServiceRuntime[MyOptions]
}

runtime := core.NewServiceRuntime(c, MyOptions{BufferSize: 1024})
runtime.Core()      // *Core
runtime.Options()   // MyOptions
runtime.Config()    // shortcut to Core().Config()
```
