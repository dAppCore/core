# CLAUDE.md

Guidance for Claude Code and Codex when working with this repository.

## Module

`dappco.re/go/core` — dependency injection, service lifecycle, command routing, and message-passing for Go.

Source files live at the module root (not `pkg/core/`). Tests live in `tests/`.

## Build & Test

```bash
go test ./tests/...          # run all tests
go build .                   # verify compilation
GOWORK=off go test ./tests/  # test without workspace
```

Or via the Core CLI:

```bash
core go test
core go qa                   # fmt + vet + lint + test
```

## API Shape

CoreGO uses the DTO/Options/Result pattern, not functional options:

```go
c := core.New(core.Options{
    {Key: "name", Value: "myapp"},
})

c.Service("cache", core.Service{
    OnStart: func() core.Result { return core.Result{OK: true} },
    OnStop:  func() core.Result { return core.Result{OK: true} },
})

c.Command("deploy/to/homelab", core.Command{
    Action: func(opts core.Options) core.Result {
        return core.Result{Value: "deployed", OK: true}
    },
})

r := c.Cli().Run("deploy", "to", "homelab")
```

**Do not use:** `WithService`, `WithName`, `WithApp`, `WithServiceLock`, `Must*`, `ServiceFor[T]` — these no longer exist.

## Subsystems

| Accessor | Returns | Purpose |
|----------|---------|---------|
| `c.Options()` | `*Options` | Input configuration |
| `c.App()` | `*App` | Application identity |
| `c.Data()` | `*Data` | Embedded filesystem mounts |
| `c.Drive()` | `*Drive` | Named transport handles |
| `c.Fs()` | `*Fs` | Local filesystem I/O |
| `c.Config()` | `*Config` | Runtime settings |
| `c.Cli()` | `*Cli` | CLI surface |
| `c.Command("path")` | `Result` | Command tree |
| `c.Service("name")` | `Result` | Service registry |
| `c.Lock("name")` | `*Lock` | Named mutexes |
| `c.IPC()` | `*Ipc` | Message bus |
| `c.I18n()` | `*I18n` | Locale + translation |

## Messaging

| Method | Pattern |
|--------|---------|
| `c.ACTION(msg)` | Broadcast to all handlers |
| `c.QUERY(q)` | First responder wins |
| `c.QUERYALL(q)` | Collect all responses |
| `c.PERFORM(task)` | First executor wins |
| `c.PerformAsync(task)` | Background goroutine |

## Error Handling

Use `core.E()` for structured errors:

```go
return core.E("service.Method", "what failed", underlyingErr)
```

## Test Naming

`_Good` (happy path), `_Bad` (expected errors), `_Ugly` (panics/edge cases).

## Docs

Full documentation in `docs/`. Start with `docs/getting-started.md`.

## Go Workspace

Part of `~/Code/go.work`. Use `GOWORK=off` to test in isolation.
