# CLAUDE.md

Guidance for Claude Code and Codex when working with this repository.

## Module

`dappco.re/go/core` — dependency injection, service lifecycle, permission, and message-passing for Go.

Source files and tests live at the module root. No `pkg/` nesting.

## Build & Test

```bash
go test ./... -count=1       # run all tests (483 tests, 84.7% coverage)
go build ./...               # verify compilation
```

Or via the Core CLI:

```bash
core go test
core go qa                   # fmt + vet + lint + test
```

## API Shape

```go
c := core.New(
    core.WithOption("name", "myapp"),
    core.WithService(mypackage.Register),
    core.WithServiceLock(),
)
c.Run()    // or: if err := c.RunE(); err != nil { ... }
```

Service factory:

```go
func Register(c *core.Core) core.Result {
    svc := &MyService{ServiceRuntime: core.NewServiceRuntime(c, MyOpts{})}
    return core.Result{Value: svc, OK: true}
}
```

## Subsystems

| Accessor | Returns | Purpose |
|----------|---------|---------|
| `c.Options()` | `*Options` | Input configuration |
| `c.App()` | `*App` | Application identity |
| `c.Config()` | `*Config` | Runtime settings, feature flags |
| `c.Data()` | `*Data` | Embedded assets (Registry[*Embed]) |
| `c.Drive()` | `*Drive` | Transport handles (Registry[*DriveHandle]) |
| `c.Fs()` | `*Fs` | Filesystem I/O (sandboxable) |
| `c.Cli()` | `*Cli` | CLI command framework |
| `c.IPC()` | `*Ipc` | Message bus internals |
| `c.Process()` | `*Process` | Managed execution (Action sugar) |
| `c.API()` | `*API` | Remote streams (protocol handlers) |
| `c.Action(name)` | `*Action` | Named callable (register/invoke) |
| `c.Task(name)` | `*Task` | Composed Action sequence |
| `c.Entitled(name)` | `Entitlement` | Permission check |
| `c.RegistryOf(n)` | `*Registry` | Cross-cutting queries |
| `c.I18n()` | `*I18n` | Internationalisation |

## Messaging

| Method | Pattern |
|--------|---------|
| `c.ACTION(msg)` | Broadcast to all handlers (panic recovery per handler) |
| `c.QUERY(q)` | First responder wins |
| `c.QUERYALL(q)` | Collect all responses |
| `c.PerformAsync(action, opts)` | Background goroutine with progress |

## Lifecycle

```go
type Startable interface { OnStartup(ctx context.Context) Result }
type Stoppable interface { OnShutdown(ctx context.Context) Result }
```

`RunE()` always calls `defer ServiceShutdown` — even on startup failure or panic.

## Error Handling

Use `core.E()` for structured errors:

```go
return core.E("service.Method", "what failed", underlyingErr)
```

**Never** use `fmt.Errorf`, `errors.New`, `os/exec`, or `unsafe.Pointer` on Core types.

## Test Naming (AX-7)

`TestFile_Function_{Good,Bad,Ugly}` — 100% compliance.

## Docs

Full API contract: `docs/RFC.md` (1476 lines, 21 sections).

## Go Workspace

Part of `~/Code/go.work`. Use `GOWORK=off` to test in isolation.
