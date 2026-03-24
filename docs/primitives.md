---
title: Core Primitives
description: The repeated shapes that make CoreGO easy to navigate.
---

# Core Primitives

CoreGO is easiest to use when you read it as a small vocabulary repeated everywhere. Most of the framework is built from the same handful of types.

## Primitive Map

| Type | Used For |
|------|----------|
| `Options` | Input values and lightweight metadata |
| `Result` | Output values and success state |
| `Service` | Lifecycle-managed components |
| `Message` | Broadcast events |
| `Query` | Request-response lookups |
| `Task` | Side-effecting work items |

## `Option` and `Options`

`Option` is one key-value pair. `Options` is an ordered slice of them.

```go
opts := core.Options{
	{Key: "name", Value: "brain"},
	{Key: "path", Value: "prompts"},
	{Key: "debug", Value: true},
}
```

Use the helpers to read values:

```go
name := opts.String("name")
path := opts.String("path")
debug := opts.Bool("debug")
hasPath := opts.Has("path")
raw := opts.Get("name")
```

### Important Details

- `Get` returns the first matching key.
- `String`, `Int`, and `Bool` do not convert between types.
- Missing keys return zero values.
- CLI flags with values are stored as strings, so `--port=8080` should be read with `opts.String("port")`, not `opts.Int("port")`.

## `Result`

`Result` is the universal return shape.

```go
r := core.Result{Value: "ready", OK: true}

if r.OK {
	fmt.Println(r.Value)
}
```

It has two jobs:

- carry a value when work succeeds
- carry either an error or an empty state when work does not succeed

### `Result.Result(...)`

The `Result()` method adapts plain Go values and `(value, error)` pairs into a `core.Result`.

```go
r1 := core.Result{}.Result("hello")
r2 := core.Result{}.Result(file, err)
```

This is how several built-in helpers bridge standard-library calls.

## `Service`

`Service` is the managed lifecycle DTO stored in the registry.

```go
svc := core.Service{
	Name: "cache",
	Options: core.Options{
		{Key: "backend", Value: "memory"},
	},
	OnStart: func() core.Result {
		return core.Result{OK: true}
	},
	OnStop: func() core.Result {
		return core.Result{OK: true}
	},
	OnReload: func() core.Result {
		return core.Result{OK: true}
	},
}
```

### Important Details

- `OnStart` and `OnStop` are used by the framework lifecycle.
- `OnReload` is stored on the service DTO, but CoreGO does not currently call it automatically.
- The registry stores `*core.Service`, not arbitrary typed service instances.

## `Message`, `Query`, and `Task`

These are simple aliases to `any`.

```go
type Message any
type Query any
type Task any
```

That means your own structs become the protocol:

```go
type deployStarted struct {
	Environment string
}

type workspaceCountQuery struct{}

type syncRepositoryTask struct {
	Name string
}
```

## `TaskWithIdentifier`

Long-running tasks can opt into task identifiers.

```go
type indexedTask struct {
	ID string
}

func (t *indexedTask) SetTaskIdentifier(id string) { t.ID = id }
func (t *indexedTask) GetTaskIdentifier() string   { return t.ID }
```

If a task implements `TaskWithIdentifier`, `PerformAsync` injects the generated `task-N` identifier before dispatch.

## `ServiceRuntime[T]`

`ServiceRuntime[T]` is the small helper for packages that want to keep a Core reference and a typed options struct together.

```go
type agentServiceOptions struct {
	WorkspacePath string
}

type agentService struct {
	*core.ServiceRuntime[agentServiceOptions]
}

runtime := core.NewServiceRuntime(c, agentServiceOptions{
	WorkspacePath: "/srv/agent-workspaces",
})
```

It exposes:

- `Core()`
- `Options()`
- `Config()`

This helper does not register anything by itself. It is a composition aid for package authors.
