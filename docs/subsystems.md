---
title: Subsystems
description: Built-in accessors for app metadata, embedded data, filesystem, transport handles, i18n, and CLI.
---

# Subsystems

`Core` gives you a set of built-in subsystems so small applications do not need extra plumbing before they can do useful work.

## Accessor Map

| Accessor | Purpose |
|----------|---------|
| `App()` | Application identity and external runtime |
| `Data()` | Named embedded filesystem mounts |
| `Drive()` | Named transport handles |
| `Fs()` | Local filesystem access |
| `I18n()` | Locale collection and translation delegation |
| `Cli()` | Command-line surface over the command tree |

## `App`

`App` stores process identity and optional GUI runtime state.

```go
app := c.App()
app.Name = "agent-workbench"
app.Version = "0.25.0"
app.Description = "workspace runner"
app.Runtime = myRuntime
```

`Find` resolves an executable on `PATH` and returns an `*App`.

```go
r := core.Find("go", "Go toolchain")
```

## `Data`

`Data` mounts named embedded filesystems and makes them addressable through paths like `mount-name/path/to/file`.

```go
c.Data().New(core.Options{
	{Key: "name", Value: "app"},
	{Key: "source", Value: appFS},
	{Key: "path", Value: "templates"},
})
```

Read content:

```go
text := c.Data().ReadString("app/agent.md")
bytes := c.Data().ReadFile("app/agent.md")
list := c.Data().List("app")
names := c.Data().ListNames("app")
```

Extract a mounted directory:

```go
r := c.Data().Extract("app/workspace", "/tmp/workspace", nil)
```

### Path Rule

The first path segment is always the mount name.

## `Drive`

`Drive` is a registry for named transport handles.

```go
c.Drive().New(core.Options{
	{Key: "name", Value: "api"},
	{Key: "transport", Value: "https://api.lthn.ai"},
})

c.Drive().New(core.Options{
	{Key: "name", Value: "mcp"},
	{Key: "transport", Value: "mcp://mcp.lthn.sh"},
})
```

Read them back:

```go
handle := c.Drive().Get("api")
hasMCP := c.Drive().Has("mcp")
names := c.Drive().Names()
```

## `Fs`

`Fs` wraps local filesystem operations with a consistent `Result` shape.

```go
c.Fs().Write("/tmp/core-go/example.txt", "hello")
r := c.Fs().Read("/tmp/core-go/example.txt")
```

Other helpers:

```go
c.Fs().EnsureDir("/tmp/core-go/cache")
c.Fs().List("/tmp/core-go")
c.Fs().Stat("/tmp/core-go/example.txt")
c.Fs().Rename("/tmp/core-go/example.txt", "/tmp/core-go/example-2.txt")
c.Fs().Delete("/tmp/core-go/example-2.txt")
```

### Important Details

- the default `Core` starts with `Fs{root:"/"}`
- relative paths resolve from the current working directory
- `Delete` and `DeleteAll` refuse to remove `/` and `$HOME`

## `I18n`

`I18n` collects locale mounts and forwards translation work to a translator implementation when one is registered.

```go
c.I18n().SetLanguage("en-GB")
```

Without a translator, `Translate` returns the message key itself:

```go
r := c.I18n().Translate("cmd.deploy.description")
```

With a translator:

```go
c.I18n().SetTranslator(myTranslator)
```

Then:

```go
langs := c.I18n().AvailableLanguages()
current := c.I18n().Language()
```

## `Cli`

`Cli` exposes the command registry through a terminal-facing API.

```go
c.Cli().SetBanner(func(_ *core.Cli) string {
	return "Agent Workbench"
})

r := c.Cli().Run("workspace", "create", "--name=alpha")
```

Use [commands.md](commands.md) for the full command and flag model.
