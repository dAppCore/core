---
title: Services
description: Register, inspect, and lock CoreGO services.
---

# Services

In CoreGO, a service is a named lifecycle entry stored in the Core registry.

## Register a Service

```go
c := core.New()

r := c.Service("audit", core.Service{
	OnStart: func() core.Result {
		core.Info("audit started")
		return core.Result{OK: true}
	},
	OnStop: func() core.Result {
		core.Info("audit stopped")
		return core.Result{OK: true}
	},
})
```

Registration succeeds when:

- the name is not empty
- the registry is not locked
- the name is not already in use

## Read a Service Back

```go
r := c.Service("audit")
if r.OK {
	svc := r.Value.(*core.Service)
	_ = svc
}
```

The returned value is `*core.Service`.

## List Registered Services

```go
names := c.Services()
```

### Important Detail

The current registry is map-backed. `Services()`, `Startables()`, and `Stoppables()` do not promise a stable order.

## Lifecycle Snapshots

Use these helpers when you want the current set of startable or stoppable services:

```go
startables := c.Startables()
stoppables := c.Stoppables()
```

They return `[]*core.Service` inside `Result.Value`.

## Lock the Registry

CoreGO has a service-lock mechanism, but it is explicit.

```go
c := core.New()

c.LockEnable()
c.Service("audit", core.Service{})
c.Service("cache", core.Service{})
c.LockApply()
```

After `LockApply`, new registrations fail:

```go
r := c.Service("late", core.Service{})
fmt.Println(r.OK) // false
```

The default lock name is `"srv"`. You can pass a different name if you need a custom lock namespace.

For the service registry itself, use the default `"srv"` lock path. That is the path used by `Core.Service(...)`.

## `NewWithFactories`

For GUI runtimes or factory-driven setup, CoreGO provides `NewWithFactories`.

```go
r := core.NewWithFactories(nil, map[string]core.ServiceFactory{
	"audit": func() core.Result {
		return core.Result{Value: core.Service{
			OnStart: func() core.Result {
				return core.Result{OK: true}
			},
		}, OK: true}
	},
	"cache": func() core.Result {
		return core.Result{Value: core.Service{}, OK: true}
	},
})
```

### Important Details

- each factory must return a `core.Service` in `Result.Value`
- factories are executed in sorted key order
- nil factories are skipped
- the return value is `*core.Runtime`

## `Runtime`

`Runtime` is a small wrapper used for external runtimes such as GUI bindings.

```go
r := core.NewRuntime(nil)
rt := r.Value.(*core.Runtime)

_ = rt.ServiceStartup(context.Background(), nil)
_ = rt.ServiceShutdown(context.Background())
```

`Runtime.ServiceName()` returns `"Core"`.

## `ServiceRuntime[T]` for Package Authors

If you are writing a package on top of CoreGO, use `ServiceRuntime[T]` to keep a typed options struct and the parent `Core` together.

```go
type repositoryServiceOptions struct {
	BaseDirectory string
}

type repositoryService struct {
	*core.ServiceRuntime[repositoryServiceOptions]
}

func newRepositoryService(c *core.Core) *repositoryService {
	return &repositoryService{
		ServiceRuntime: core.NewServiceRuntime(c, repositoryServiceOptions{
			BaseDirectory: "/srv/repos",
		}),
	}
}
```

This is a package-authoring helper. It does not replace the `core.Service` registry entry.
