---
title: Getting Started
description: Build a first CoreGO application with the current API.
---

# Getting Started

This page shows the shortest path to a useful CoreGO application using the API that exists in this repository today.

## Install

```bash
go get dappco.re/go
```

## Create a Core

`New` takes zero or more `core.Options` slices, but the current implementation only reads the first one. In practice, treat the constructor as `core.New(core.Options{...})`.

```go
package main

import "dappco.re/go"

func main() {
	c := core.New(core.Options{
		{Key: "name", Value: "agent-workbench"},
	})

	_ = c
}
```

The `name` option is copied into `c.App().Name`.

## Register a Service

Services are registered explicitly with a name and a `core.Service` DTO.

```go
c.Service("audit", core.Service{
	OnStart: func() core.Result {
		core.Info("audit service started", "app", c.App().Name)
		return core.Result{OK: true}
	},
	OnStop: func() core.Result {
		core.Info("audit service stopped", "app", c.App().Name)
		return core.Result{OK: true}
	},
})
```

This registry stores `core.Service` values. It is a lifecycle registry, not a typed object container.

## Register a Query, Task, and Command

```go
type workspaceCountQuery struct{}

type createWorkspaceTask struct {
	Name string
}

c.RegisterQuery(func(_ *core.Core, q core.Query) core.Result {
	switch q.(type) {
	case workspaceCountQuery:
		return core.Result{Value: 1, OK: true}
	}
	return core.Result{}
})

c.Action("workspace.create", func(_ context.Context, opts core.Options) core.Result {
	name := opts.String("name")
	path := "/tmp/agent-workbench/" + name
	return core.Result{Value: path, OK: true}
})

c.Command("workspace/create", core.Command{
	Action: func(opts core.Options) core.Result {
		return c.Action("workspace.create").Run(context.Background(), opts)
	},
})
```

## Start the Runtime

```go
if !c.ServiceStartup(context.Background(), nil).OK {
	panic("startup failed")
}
```

`ServiceStartup` returns `core.Result`, not `error`.

## Run Through the CLI Surface

```go
r := c.Cli().Run("workspace", "create", "--name=alpha")
if r.OK {
	fmt.Println("created:", r.Value)
}
```

For flags with values, the CLI stores the value as a string. `--name=alpha` becomes `opts.String("name") == "alpha"`.

## Query the System

```go
count := c.QUERY(workspaceCountQuery{})
if count.OK {
	fmt.Println("workspace count:", count.Value)
}
```

## Shut Down Cleanly

```go
_ = c.ServiceShutdown(context.Background())
```

Shutdown cancels `c.Context()`, broadcasts `ActionServiceShutdown{}`, waits for background tasks to finish, and then runs service stop hooks.

## Full Example

```go
package main

import (
	"context"
	"fmt"

	"dappco.re/go"
)

type workspaceCountQuery struct{}

type createWorkspaceTask struct {
	Name string
}

func main() {
	c := core.New(core.Options{
		{Key: "name", Value: "agent-workbench"},
	})

	c.Config().Set("workspace.root", "/tmp/agent-workbench")
	c.Config().Enable("workspace.templates")

	c.Service("audit", core.Service{
		OnStart: func() core.Result {
			core.Info("service started", "service", "audit")
			return core.Result{OK: true}
		},
		OnStop: func() core.Result {
			core.Info("service stopped", "service", "audit")
			return core.Result{OK: true}
		},
	})

	c.RegisterQuery(func(_ *core.Core, q core.Query) core.Result {
		switch q.(type) {
		case workspaceCountQuery:
			return core.Result{Value: 1, OK: true}
		}
		return core.Result{}
	})

	c.Action("workspace.create", func(_ context.Context, opts core.Options) core.Result {
		name := opts.String("name")
		path := c.Config().String("workspace.root") + "/" + name
		return core.Result{Value: path, OK: true}
	})

	c.Command("workspace/create", core.Command{
		Action: func(opts core.Options) core.Result {
			return c.Action("workspace.create").Run(context.Background(), opts)
		},
	})

	if !c.ServiceStartup(context.Background(), nil).OK {
		panic("startup failed")
	}

	created := c.Cli().Run("workspace", "create", "--name=alpha")
	fmt.Println("created:", created.Value)

	count := c.QUERY(workspaceCountQuery{})
	fmt.Println("workspace count:", count.Value)

	_ = c.ServiceShutdown(context.Background())
}
```

## Next Steps

- Read [primitives.md](primitives.md) next so the repeated shapes are clear.
- Read [commands.md](commands.md) if you are building a CLI-first system.
- Read [messaging.md](messaging.md) if services need to collaborate without direct imports.
