---
title: Configuration
description: Constructor options, runtime settings, and feature flags.
---

# Configuration

CoreGO uses two different configuration layers:

- constructor-time `core.Options`
- runtime `c.Config()`

## Constructor-Time Options

```go
c := core.New(core.Options{
	{Key: "name", Value: "agent-workbench"},
})
```

### Current Behavior

- `New` accepts `opts ...Options`
- the current implementation copies only the first `Options` slice
- the `name` key is applied to `c.App().Name`

If you need more constructor data, put it in the first `core.Options` slice.

## Runtime Settings with `Config`

Use `c.Config()` for mutable process settings.

```go
c.Config().Set("workspace.root", "/srv/workspaces")
c.Config().Set("max_agents", 8)
c.Config().Set("debug", true)
```

Read them back with:

```go
root := c.Config().String("workspace.root")
maxAgents := c.Config().Int("max_agents")
debug := c.Config().Bool("debug")
raw := c.Config().Get("workspace.root")
```

### Important Details

- missing keys return zero values
- typed accessors do not coerce strings into ints or bools
- `Get` returns `core.Result`

## Feature Flags

`Config` also tracks named feature flags.

```go
c.Config().Enable("workspace.templates")
c.Config().Enable("agent.review")
c.Config().Disable("agent.review")
```

Read them with:

```go
enabled := c.Config().Enabled("workspace.templates")
features := c.Config().EnabledFeatures()
```

Feature names are case-sensitive.

## `ConfigVar[T]`

Use `ConfigVar[T]` when you need a typed value that can also represent “set versus unset”.

```go
theme := core.NewConfigVar("amber")

if theme.IsSet() {
	fmt.Println(theme.Get())
}

theme.Unset()
```

This is useful for package-local state where zero values are not enough to describe configuration presence.

## Recommended Pattern

Use the two layers for different jobs:

- put startup identity such as `name` into `core.Options`
- put mutable runtime values and feature switches into `c.Config()`

That keeps constructor intent separate from live process state.
