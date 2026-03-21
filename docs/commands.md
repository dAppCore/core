---
title: Commands
description: Path-based command registration and CLI execution.
---

# Commands

Commands are one of the most AX-native parts of CoreGO. The path is the identity.

## Register a Command

```go
c.Command("deploy/to/homelab", core.Command{
	Action: func(opts core.Options) core.Result {
		target := opts.String("target")
		return core.Result{Value: "deploying to " + target, OK: true}
	},
})
```

## Command Paths

Paths must be clean:

- no empty path
- no leading slash
- no trailing slash
- no double slash

These paths are valid:

```text
deploy
deploy/to/homelab
workspace/create
```

These are rejected:

```text
/deploy
deploy/
deploy//to
```

## Parent Commands Are Auto-Created

When you register `deploy/to/homelab`, CoreGO also creates placeholder parents if they do not already exist:

- `deploy`
- `deploy/to`

This makes the path tree navigable without extra setup.

## Read a Command Back

```go
r := c.Command("deploy/to/homelab")
if r.OK {
	cmd := r.Value.(*core.Command)
	_ = cmd
}
```

## Run a Command Directly

```go
cmd := c.Command("deploy/to/homelab").Value.(*core.Command)

r := cmd.Run(core.Options{
	{Key: "target", Value: "uk-prod"},
})
```

If `Action` is nil, `Run` returns `Result{OK:false}` with a structured error.

## Run Through the CLI Surface

```go
r := c.Cli().Run("deploy", "to", "homelab", "--target=uk-prod", "--debug")
```

`Cli.Run` resolves the longest matching command path from the arguments, then converts the remaining args into `core.Options`.

## Flag Parsing Rules

### Double Dash

```text
--target=uk-prod  -> key "target", value "uk-prod"
--debug           -> key "debug", value true
```

### Single Dash

```text
-v      -> key "v", value true
-n=4    -> key "n", value "4"
```

### Positional Arguments

Non-flag arguments after the command path are stored as repeated `_arg` options.

```go
r := c.Cli().Run("workspace", "open", "alpha")
```

That produces an option like:

```go
core.Option{Key: "_arg", Value: "alpha"}
```

### Important Details

- flag values stay as strings
- `opts.Int("port")` only works if some code stored an actual `int`
- invalid flags such as `-verbose` and `--v` are ignored

## Help Output

`Cli.PrintHelp()` prints executable commands:

```go
c.Cli().PrintHelp()
```

It skips:

- hidden commands
- placeholder parents with no `Action` and no `Lifecycle`

Descriptions are resolved through `cmd.I18nKey()`.

## I18n Description Keys

If `Description` is empty, CoreGO derives a key from the path.

```text
deploy                  -> cmd.deploy.description
deploy/to/homelab       -> cmd.deploy.to.homelab.description
workspace/create        -> cmd.workspace.create.description
```

If `Description` is already set, CoreGO uses it as-is.

## Lifecycle Commands

Commands can also delegate to a lifecycle implementation.

```go
type daemonCommand struct{}

func (d *daemonCommand) Start(opts core.Options) core.Result { return core.Result{OK: true} }
func (d *daemonCommand) Stop() core.Result                   { return core.Result{OK: true} }
func (d *daemonCommand) Restart() core.Result                { return core.Result{OK: true} }
func (d *daemonCommand) Reload() core.Result                 { return core.Result{OK: true} }
func (d *daemonCommand) Signal(sig string) core.Result       { return core.Result{Value: sig, OK: true} }

c.Command("agent/serve", core.Command{
	Lifecycle: &daemonCommand{},
})
```

Important behavior:

- `Start` falls back to `Run` when `Lifecycle` is nil
- `Stop`, `Restart`, `Reload`, and `Signal` return an empty `Result` when `Lifecycle` is nil

## List Command Paths

```go
paths := c.Commands()
```

Like the service registry, the command registry is map-backed, so iteration order is not guaranteed.
