# CoreGO API Contract — RFC Specification

> `dappco.re/go/core` — Dependency injection, service lifecycle, permission, and message-passing framework.
> This document is the authoritative API contract. An agent should be able to write a service
> that registers with Core from this document alone.

**Status:** Living document
**Module:** `dappco.re/go/core`
**Version:** v0.8.0

---

## 1. Core — The Container

Core is the central application container. Everything registers with Core, communicates through Core, and has its lifecycle managed by Core.

### 1.1 Creation

```go
c := core.New(
    core.WithOption("name", "my-app"),
    core.WithService(mypackage.Register),
    core.WithService(anotherpackage.Register),
    core.WithServiceLock(),
)
c.Run()
```

`core.New()` returns `*Core` (not Result — Core is the one type that can't wrap its own creation error). Functional options are applied in order. `WithServiceLock()` prevents late service registration.

### 1.2 Lifecycle

```
New() → WithService factories called → LockApply()
RunE() → defer ServiceShutdown() → ServiceStartup() → Cli.Run() → returns error
Run()  → RunE() → os.Exit(1) on error
```

`RunE()` is the primary lifecycle — returns `error`, always calls `ServiceShutdown` via defer (even on startup failure or panic). `Run()` is sugar that calls `RunE()` and exits on error. `ServiceStartup` calls `OnStartup(ctx)` on all `Startable` services in registration order. `ServiceShutdown` calls `OnShutdown(ctx)` on all `Stoppable` services.

### 1.3 Subsystem Accessors

Every subsystem is accessed via a method on Core:

```go
c.Options()      // *Options     — input configuration
c.App()          // *App         — application metadata (name, version)
c.Config()       // *Config      — runtime settings, feature flags
c.Data()         // *Data        — embedded assets (Registry[*Embed])
c.Drive()        // *Drive       — transport handles (Registry[*DriveHandle])
c.Fs()           // *Fs          — filesystem I/O (sandboxable)
c.Cli()          // *Cli         — CLI command framework
c.IPC()          // *Ipc         — message bus internals
c.I18n()         // *I18n        — internationalisation
c.Error()        // *ErrorPanic  — panic recovery
c.Log()          // *ErrorLog    — structured logging
c.Process()      // *Process     — managed execution (Action sugar)
c.API()          // *API         — remote streams (protocol handlers)
c.Action(name)   // *Action      — named callable (register/invoke)
c.Task(name)     // *Task        — composed Action sequence
c.Entitled(name) // Entitlement  — permission check
c.RegistryOf(n)  // *Registry    — cross-cutting queries
c.Context()      // context.Context
c.Env(key)       // string       — environment variable (cached at init)
```

---

## 2. Primitive Types

### 2.1 Option

The atom. A single key-value pair.

```go
core.Option{Key: "name", Value: "brain"}
core.Option{Key: "port", Value: 8080}
core.Option{Key: "debug", Value: true}
```

### 2.2 Options

A collection of Option with typed accessors.

```go
opts := core.NewOptions(
    core.Option{Key: "name", Value: "myapp"},
    core.Option{Key: "port", Value: 8080},
    core.Option{Key: "debug", Value: true},
)

opts.String("name")  // "myapp"
opts.Int("port")     // 8080
opts.Bool("debug")   // true
opts.Has("name")     // true
opts.Len()           // 3

opts.Set("name", "new-name")
opts.Get("name")     // Result{Value: "new-name", OK: true}
```

### 2.3 Result

Universal return type. Every Core operation returns Result.

```go
type Result struct {
    Value any
    OK    bool
}
```

Usage patterns:

```go
// Check success
r := c.Config().Get("database.host")
if r.OK {
    host := r.Value.(string)
}

// Service factory returns Result
func Register(c *core.Core) core.Result {
    svc := &MyService{}
    return core.Result{Value: svc, OK: true}
}

// Error as Result
return core.Result{Value: err, OK: false}
```

No generics on Result. Type-assert the Value when needed. This is deliberate — `Result` is universal across all subsystems without carrying type parameters.

### 2.4 Message, Query

IPC type aliases for the broadcast/request system:

```go
type Message any  // broadcast via ACTION — fire and forget
type Query any    // request/response via QUERY — returns first handler's result
```

For tracked work, use named Actions: `c.PerformAsync("action.name", opts)`.

---

## 3. Service System

### 3.1 Registration

Services register via factory functions passed to `WithService`:

```go
core.New(
    core.WithService(mypackage.Register),
)
```

The factory signature is `func(*Core) Result`. The returned `Result.Value` is the service instance.

### 3.2 Factory Pattern

```go
func Register(c *core.Core) core.Result {
    svc := &MyService{
        runtime: core.NewServiceRuntime(c, MyOptions{}),
    }
    return core.Result{Value: svc, OK: true}
}
```

`NewServiceRuntime[T]` gives the service access to Core and typed options:

```go
type MyService struct {
    *core.ServiceRuntime[MyOptions]
}

// Access Core from within the service:
func (s *MyService) doSomething() {
    c := s.Core()
    cfg := s.Config().String("my.setting")
}
```

### 3.3 Auto-Discovery

`WithService` reflects on the returned instance to discover:
- **Package name** → service name (from reflect type path)
- **Startable interface** → `OnStartup(ctx) Result` called during `ServiceStartup`
- **Stoppable interface** → `OnShutdown(ctx) Result` called during `ServiceShutdown`
- **HandleIPCEvents method** → auto-registered as IPC handler

### 3.4 Retrieval

```go
// Type-safe retrieval
svc, ok := core.ServiceFor[*MyService](c, "mypackage")
if !ok {
    // service not registered
}

// Must variant (panics if not found)
svc := core.MustServiceFor[*MyService](c, "mypackage")

// List all registered services
names := c.Services() // []string
```

### 3.5 Lifecycle Interfaces

```go
type Startable interface {
    OnStartup(ctx context.Context) Result
}

type Stoppable interface {
    OnShutdown(ctx context.Context) Result
}
```

Services implementing these are automatically called during `c.Run()`.

---

## 4. IPC — Message Passing

### 4.1 ACTION (broadcast)

Fire-and-forget broadcast to all registered handlers:

```go
// Send
c.ACTION(messages.AgentCompleted{
    Agent: "codex", Repo: "go-io", Status: "completed",
})

// Register handler
c.RegisterAction(func(c *core.Core, msg core.Message) core.Result {
    if ev, ok := msg.(messages.AgentCompleted); ok {
        // handle completion
    }
    return core.Result{OK: true}
})
```

All handlers receive all messages. Type-switch to filter. Handler return values are ignored — broadcast calls ALL handlers regardless. Each handler is wrapped in panic recovery.

### 4.2 QUERY (request/response)

First handler to return a non-empty result wins:

```go
// Send
result := c.QUERY(MyQuery{Name: "brain"})
if result.OK {
    svc := result.Value
}

// Register handler
c.RegisterQuery(func(c *core.Core, q core.Query) core.Result {
    if mq, ok := q.(MyQuery); ok {
        return core.Result{Value: found, OK: true}
    }
    return core.Result{OK: false} // not my query
})
```

### 4.3 PerformAsync (background action)

```go
// Execute a named action in background with progress tracking
r := c.PerformAsync("agentic.dispatch", opts)
taskID := r.Value.(string)

// Report progress
c.Progress(taskID, 0.5, "halfway done", "agentic.dispatch")
```

Broadcasts `ActionTaskStarted`, `ActionTaskProgress`, `ActionTaskCompleted` as ACTION messages.

---

## 5. Config

Runtime configuration with typed accessors and feature flags.

```go
c.Config().Set("database.host", "localhost")
c.Config().Set("database.port", 5432)

host := c.Config().String("database.host")  // "localhost"
port := c.Config().Int("database.port")      // 5432

// Feature flags
c.Config().Enable("dark-mode")
c.Config().Enabled("dark-mode")     // true
c.Config().Disable("dark-mode")
c.Config().EnabledFeatures()         // []string

// Type-safe generic getter
val := core.ConfigGet[string](c.Config(), "database.host")
```

---

## 6. Data — Embedded Assets

Mount embedded filesystems and read from them:

```go
//go:embed prompts/*
var promptFS embed.FS

// Mount during service registration
c.Data().New(core.NewOptions(
    core.Option{Key: "name", Value: "prompts"},
    core.Option{Key: "source", Value: promptFS},
    core.Option{Key: "path", Value: "prompts"},
))

// Read
r := c.Data().ReadString("prompts/coding.md")
if r.OK {
    content := r.Value.(string)
}

// List
r := c.Data().List("prompts/")
r := c.Data().ListNames("prompts/")
r := c.Data().Mounts() // []string (insertion order)

// Data embeds Registry[*Embed] — all Registry methods available:
c.Data().Has("prompts")
c.Data().Each(func(name string, emb *Embed) { ... })
```

---

## 7. Drive — Transport Handles

Registry of named transport handles (API endpoints, MCP servers, etc):

```go
c.Drive().New(core.NewOptions(
    core.Option{Key: "name", Value: "forge"},
    core.Option{Key: "transport", Value: "https://forge.lthn.ai"},
))

r := c.Drive().Get("forge")     // Result with *DriveHandle
c.Drive().Has("forge")          // true
c.Drive().Names()               // []string (insertion order)

// Drive embeds Registry[*DriveHandle] — all Registry methods available.
```

---

## 8. Fs — Filesystem

Sandboxable filesystem I/O. All paths are validated against the root.

```go
fs := c.Fs()

// Read/Write
r := fs.Read("/path/to/file")           // Result{Value: string}
r := fs.Write("/path/to/file", content) // Result{OK: bool}
r := fs.WriteMode(path, content, 0600)  // With permissions

// Directory ops
r := fs.EnsureDir("/path/to/dir")
r := fs.List("/path/to/dir")            // Result{Value: []os.DirEntry}
fs.IsDir(path)                           // bool
fs.IsFile(path)                          // bool
fs.Exists(path)                          // bool

// Streams
r := fs.Open(path)        // Result{Value: *os.File}
r := fs.Create(path)      // Result{Value: *os.File}
r := fs.Append(path)      // Result{Value: io.WriteCloser}
r := fs.ReadStream(path)  // Result{Value: io.ReadCloser}
r := fs.WriteStream(path) // Result{Value: io.WriteCloser}

// Atomic write (write-to-temp-then-rename, safe for concurrent readers)
r := fs.WriteAtomic(path, content)

// Delete
r := fs.Delete(path)      // single file
r := fs.DeleteAll(path)   // recursive
r := fs.Rename(old, new)
r := fs.Stat(path)        // Result{Value: os.FileInfo}

// Sandbox control
fs.Root()                  // sandbox root path
fs.NewUnrestricted()       // Fs with root "/" — full access
```

---

## 9. CLI

Command tree with path-based routing:

```go
c.Command("issue/get", core.Command{
    Description: "Get a Forge issue",
    Action: s.cmdIssueGet,
})

c.Command("issue/list", core.Command{
    Description: "List Forge issues",
    Action: s.cmdIssueList,
})

// Action signature
func (s *MyService) cmdIssueGet(opts core.Options) core.Result {
    repo := opts.String("_arg")  // positional arg
    num := opts.String("number") // --number=N flag
    // ...
    return core.Result{OK: true}
}
```

Path = command hierarchy. `issue/get` becomes `myapp issue get` in CLI.

Managed commands have lifecycle provided by go-process:

```go
c.Command("serve", core.Command{
    Action:  handler,
    Managed: "process.daemon",  // go-process provides start/stop/restart
})
```

---

## 10. Error Handling

All errors use `core.E()`:

```go
// Standard error
return core.E("service.Method", "what failed", underlyingErr)

// With format
return core.E("service.Method", core.Sprintf("not found: %s", name), nil)

// Error inspection
core.Operation(err)      // "service.Method"
core.ErrorMessage(err)   // "what failed"
core.ErrorCode(err)      // code if set via WrapCode
core.Root(err)           // unwrap to root cause
core.Is(err, target)     // errors.Is
core.As(err, &target)    // errors.As
```

**NEVER use `fmt.Errorf`, `errors.New`, or `log.*`.** Core handles all error reporting.

---

## 11. Logging

```go
core.Info("server started", "port", 8080)
core.Debug("processing", "item", name)
core.Warn("deprecated", "feature", "old-api")
core.Error("failed", "err", err)
core.Security("access denied", "user", username)
```

Key-value pairs after the message. Structured, not formatted strings.

---

## 12. String Helpers

Core re-exports string operations to avoid `strings` import:

```go
core.Contains(s, substr)
core.HasPrefix(s, prefix)
core.HasSuffix(s, suffix)
core.TrimPrefix(s, prefix)
core.TrimSuffix(s, suffix)
core.Split(s, sep)
core.SplitN(s, sep, n)
core.Join(sep, parts...)
core.Replace(s, old, new)
core.Lower(s) / core.Upper(s)
core.Trim(s)
core.Sprintf(format, args...)
core.Concat(parts...)
core.NewBuilder() / core.NewReader(s)
```

---

## 13. Path Helpers

```go
core.Path(segments...)      // ~/segments joined
core.JoinPath(segments...)  // filepath.Join
core.PathBase(p)            // filepath.Base
core.PathDir(p)             // filepath.Dir
core.PathExt(p)             // filepath.Ext
core.PathIsAbs(p)           // filepath.IsAbs
core.PathGlob(pattern)      // filepath.Glob
core.CleanPath(p, sep)      // normalise separators
```

---

## 14. Utility Functions

```go
core.Print(writer, format, args...)  // formatted output
core.Env(key)                         // cached env var (set at init)
core.EnvKeys()                        // all available env keys

// Arg extraction (positional)
core.Arg(0, args...)       // Result
core.ArgString(0, args...) // string
core.ArgInt(0, args...)    // int
core.ArgBool(0, args...)   // bool

// Flag parsing
core.IsFlag("--name")              // true
core.ParseFlag("--name=value")    // "name", "value", true
core.FilterArgs(args)              // strip flags, keep positional

// Identifiers and validation
core.ID()                          // "id-42-a3f2b1" — unique per process
core.ValidateName("brain")        // Result{OK: true} — rejects "", ".", "..", path seps
core.SanitisePath("../../x")      // "x" — extracts safe base, "invalid" for dangerous
```

---

## 15. Lock System

Per-Core mutex registry for coordinating concurrent access:

```go
c.Lock("drain").Mutex.Lock()
defer c.Lock("drain").Mutex.Unlock()

// Enable named locks
c.LockEnable("service-registry")

// Apply lock (prevents further registration)
c.LockApply()
```

---

## 16. ServiceRuntime Generic Helper

Embed in services to get Core access and typed options:

```go
type MyService struct {
    *core.ServiceRuntime[MyOptions]
}

type MyOptions struct {
    BufferSize int
    Timeout    time.Duration
}

func NewMyService(c *core.Core) core.Result {
    svc := &MyService{
        ServiceRuntime: core.NewServiceRuntime(c, MyOptions{
            BufferSize: 1024,
            Timeout:    30 * time.Second,
        }),
    }
    return core.Result{Value: svc, OK: true}
}

// Within the service:
func (s *MyService) DoWork() {
    c := s.Core()           // access Core
    opts := s.Options()     // MyOptions{BufferSize: 1024, ...}
    cfg := s.Config()       // shortcut to s.Core().Config()
}
```

---

## 17. Process — Managed Execution

`c.Process()` is sugar over named Actions. core/go defines the primitive. go-process provides the implementation via `c.Action("process.run", handler)`.

```go
// Synchronous — returns Result
r := c.Process().Run(ctx, "git", "log", "--oneline")
r := c.Process().RunIn(ctx, "/repo", "go", "test", "./...")
r := c.Process().RunWithEnv(ctx, dir, []string{"GOWORK=off"}, "go", "test")

// Async — returns process ID
r := c.Process().Start(ctx, opts)

// Control
c.Process().Kill(ctx, core.NewOptions(core.Option{Key: "id", Value: processID}))

// Capability check
if c.Process().Exists() { /* go-process is registered */ }
```

**Permission by registration:** No go-process registered → `c.Process().Run()` returns `Result{OK: false}`. No config, no tokens. The service either exists or it doesn't.

```go
// Sandboxed Core — no process capability
c := core.New()
c.Process().Run(ctx, "rm", "-rf", "/")  // Result{OK: false} — nothing happens

// Full Core — process registered
c := core.New(core.WithService(process.Register))
c.Process().Run(ctx, "git", "log")      // executes, returns output
```

> Consumer implementation: see `go-process/docs/RFC.md`

---

## 18. Action and Task — The Execution Primitives

An Action is a named, registered callable. A Task is a composed sequence of Actions.

### 18.1 Action — The Atomic Unit

```go
// Register
c.Action("git.log", func(ctx context.Context, opts core.Options) core.Result {
    dir := opts.String("dir")
    return c.Process().RunIn(ctx, dir, "git", "log", "--oneline")
})

// Invoke
r := c.Action("git.log").Run(ctx, core.NewOptions(
    core.Option{Key: "dir", Value: "/repo"},
))

// Check capability
c.Action("process.run").Exists()  // true if go-process registered

// List all
c.Actions()  // []string{"process.run", "agentic.dispatch", ...}
```

`c.Action(name)` is dual-purpose: with handler arg → register; without → return for invocation.

### 18.2 Action Type

```go
type ActionHandler func(context.Context, Options) Result

type Action struct {
    Name        string
    Handler     ActionHandler
    Description string
    Schema      Options       // expected input keys
}
```

`Action.Run()` has panic recovery and entitlement checking (Section 21) built in.

### 18.3 Where Actions Come From

Services register during `OnStartup`:

```go
func (s *MyService) OnStartup(ctx context.Context) core.Result {
    c := s.Core()
    c.Action("process.run", s.handleRun)
    c.Action("git.clone", s.handleGitClone)
    return core.Result{OK: true}
}
```

The action namespace IS the capability map. go-process registers `process.*`, core/agent registers `agentic.*`.

### 18.4 Permission Model

Three states for any action:

| State | `Exists()` | `Entitled()` | `Run()` |
|-------|-----------|-------------|---------|
| Not registered | false | — | `Result{OK: false}` not registered |
| Registered, not entitled | true | false | `Result{OK: false}` not entitled |
| Registered and entitled | true | true | executes handler |

### 18.5 Task — Composing Actions

```go
c.Task("deploy", core.Task{
    Description: "Build, test, deploy",
    Steps: []core.Step{
        {Action: "go.build"},
        {Action: "go.test"},
        {Action: "docker.push"},
        {Action: "ansible.deploy", Async: true},  // doesn't block
    },
})

r := c.Task("deploy").Run(ctx, c, opts)
```

Sequential steps stop on first failure. `Async: true` steps fire without blocking.
`Input: "previous"` pipes last step's output to next step.

### 18.6 Background Execution

```go
r := c.PerformAsync("agentic.dispatch", opts)
taskID := r.Value.(string)

// Broadcasts ActionTaskStarted, ActionTaskProgress, ActionTaskCompleted
c.Progress(taskID, 0.5, "halfway", "agentic.dispatch")
```

### 18.7 How Process Fits

`c.Process()` is sugar over Actions:

```go
c.Process().Run(ctx, "git", "log")
// equivalent to:
c.Action("process.run").Run(ctx, core.NewOptions(
    core.Option{Key: "command", Value: "git"},
    core.Option{Key: "args", Value: []string{"log"}},
))
```

---

## 19. API — Remote Streams

Drive is the phone book (WHERE). API is the phone (HOW). Consumer packages register protocol handlers.

```go
// Configure endpoint in Drive
c.Drive().New(core.NewOptions(
    core.Option{Key: "name", Value: "charon"},
    core.Option{Key: "transport", Value: "http://10.69.69.165:9101/mcp"},
))

// Open stream — looks up Drive, finds protocol handler
r := c.API().Stream("charon")
if r.OK {
    stream := r.Value.(core.Stream)
    stream.Send(payload)
    resp, _ := stream.Receive()
    stream.Close()
}
```

### 19.1 Stream Interface

```go
type Stream interface {
    Send(data []byte) error
    Receive() ([]byte, error)
    Close() error
}
```

### 19.2 Protocol Handlers

Consumer packages register factories per URL scheme:

```go
// In a transport package's OnStartup:
c.API().RegisterProtocol("http", httpStreamFactory)
c.API().RegisterProtocol("mcp", mcpStreamFactory)
```

Resolution: `c.API().Stream("charon")` → Drive lookup → extract scheme → find factory → create Stream.

No protocol handler = no capability.

### 19.3 Remote Action Dispatch

Actions transparently cross machine boundaries via `host:action` syntax:

```go
// Local
r := c.RemoteAction("agentic.status", ctx, opts)

// Remote — same API, different host
r := c.RemoteAction("charon:agentic.status", ctx, opts)
// → splits on ":" → endpoint="charon", action="agentic.status"
// → c.API().Call("charon", "agentic.status", opts)

// Web3 — Lethean dVPN routed
r := c.RemoteAction("snider.lthn:brain.recall", ctx, opts)
```

### 19.4 Direct Call

```go
r := c.API().Call("charon", "agentic.dispatch", opts)
// Opens stream, sends JSON-RPC, receives response, closes stream
```

---

## 20. Registry — The Universal Collection Primitive

Thread-safe named collection. The brick all registries build on.

### 20.1 The Type

```go
// Registry is a thread-safe named collection. The universal brick
// for all named registries in Core.
type Registry[T any] struct {
    items  map[string]T
    mu     sync.RWMutex
    locked bool
}
```

### 20.3 Operations

```go
r := core.NewRegistry[*Service]()

r.Set("brain", brainSvc)          // register
r.Get("brain")                     // Result{brainSvc, true}
r.Has("brain")                     // true
r.Names()                          // []string{"brain", "monitor", ...}
r.List("brain.*")                  // glob/prefix match
r.Each(func(name string, item T))  // iterate
r.Len()                            // count
r.Lock()                           // prevent further Set calls
r.Locked()                         // bool
r.Delete("brain")                  // remove (if not locked)
```

### 20.4 Core Accessor

`c.Registry(name)` accesses named registries. Each subsystem's registry is accessible through it:

```go
c.RegistryOf("services")              // the service registry
c.Registry("commands")              // the command tree
c.RegistryOf("actions")               // IPC action handlers
c.RegistryOf("drives")                // transport handles
c.Registry("data")                  // mounted filesystems
```

Cross-cutting queries become natural:

```go
c.RegistryOf("actions").List("process.*")  // all process capabilities
c.RegistryOf("drives").Names()              // all configured transports
c.RegistryOf("services").Has("brain")       // is brain service loaded?
c.RegistryOf("actions").Len()               // how many actions registered?
```

### 20.5 Typed Accessors Are Sugar

The existing subsystem accessors become typed convenience over Registry:

```go
// These are equivalent:
c.Service("brain")                         // typed sugar
c.RegistryOf("services").Get("brain")        // universal access

c.Drive().Get("forge")                     // typed sugar
c.RegistryOf("drives").Get("forge")          // universal access

c.Action("process.run")                    // typed sugar
c.RegistryOf("actions").Get("process.run")   // universal access
```

The typed accessors stay — they're ergonomic and type-safe. `c.Registry()` adds the universal query layer on top.

### 20.6 What Embeds Registry

All named collections in Core embed `Registry[T]`:

- `ServiceRegistry` → `Registry[*Service]`
- `CommandRegistry` → `Registry[*Command]`
- `Drive` → `Registry[*DriveHandle]`
- `Data` → `Registry[*Embed]`
- `Lock.locks` → `Registry[*sync.RWMutex]`
- `IPC.actions` → `Registry[*Action]`
- `IPC.tasks` → `Registry[*Task]`

---

## Design Philosophy

### Core Is Lego Bricks

Core is infrastructure, not an encapsulated library. Downstream packages (core/agent, core/mcp, go-process) compose with Core's primitives. **Exported fields are intentional, not accidental.** Every unexported field that forces a consumer to write a wrapper method adds LOC downstream — the opposite of Core's purpose.

```go
// Core reduces downstream code:
if r.OK { use(r.Value) }

// vs Go convention that adds downstream LOC:
val, err := thing.Get()
if err != nil {
    return fmt.Errorf("get: %w", err)
}
```

This is why `core.Result` exists — it replaces multiple lines of error handling with `if r.OK {}`. That's the design: expose the primitive, reduce consumer code.

### Export Rules

| Should Export | Why |
|--------------|-----|
| Struct fields used by consumers | Removes accessor boilerplate downstream |
| Registry types (`ServiceRegistry`) | Lets consumers extend service management |
| IPC internals (`Ipc` handlers) | Lets consumers build custom dispatch |
| Lifecycle hooks (`OnStart`, `OnStop`) | Composable without interface overhead |

| Should NOT Export | Why |
|------------------|-----|
| Mutexes and sync primitives | Concurrency must be managed by Core |
| Context/cancel pairs | Lifecycle is Core's responsibility |
| Internal counters | Implementation detail, not a brick |

### Why core/go Is Minimal

core/go deliberately avoids importing anything beyond stdlib + go-io + go-log. This keeps it as a near-pure stdlib implementation. Packages that add external dependencies (CLI frameworks, HTTP routers, MCP SDK) live in separate repos:

```
core/go          — pure primitives (stdlib only)
core/go-process  — process management (adds os/exec)
core/mcp         — MCP server (adds go-sdk)
core/agent       — orchestration (adds forge, yaml, mcp)
```

Each layer imports the one below. core/go imports nothing from the ecosystem — everything imports core/go.

## Known Issues — All Resolved

| # | Issue | Resolution |
|---|-------|-----------|
| 1 | UPPERCASE vs CamelCase naming | `c.Action("name")` = primitive, `c.ACTION(msg)` = sugar. `broadcast()` internal. |
| 2 | MustServiceFor uses panic | Kept — `Must` prefix is Go convention. Use only in startup paths. |
| 3 | Embed() legacy accessor | Removed. Use `c.Data().Get("app")`. |
| 4 | Package vs Core logging | Both stay. `core.Info()` = bootstrap. `c.Log().Info()` = runtime. |
| 5 | RegisterAction in task.go | Moved to ipc.go. action.go has execution. task.go has PerformAsync. |
| 6 | serviceRegistry unexported | `ServiceRegistry` embedding `Registry[*Service]`. All 5 registries migrated. |
| 7 | No c.Process() | Implemented — Action sugar over `process.run/start/kill`. |
| 8 | NewRuntime/NewWithFactories | GUI bridge, not legacy. Needs `WithRuntime(app)` CoreOption. |
| 9 | CommandLifecycle interface | Removed. `Command.Managed` string field. Lifecycle verbs are Actions. |
| 10 | Array[T] guardrail | Kept — ordered counterpart to Registry[T]. Model-proof collection ops. |
| 11 | ConfigVar[T] | Kept — distinguishes "set to false" from "never set". Layered config. |
| 12 | Ipc data-only struct | IPC owns Action registry. `c.Action()` is API. `c.IPC()` owns data. |
| 13 | Lock() allocates per call | Lock uses `Registry[*sync.RWMutex]`. Mutex cached. |
| 14 | Startables/Stoppables return type | Return `Result`. Registry.Each iterates in insertion order. |
| 15 | New() comment stale | Fixed — shows `*Core` return with correct example. |
| 16 | task.go mixes concerns | Split: ipc.go (registration), action.go (execution), task.go (async). `type Task any` removed. |

> Full discussion of each issue preserved in git history (commit `0704a7a` and earlier).
> The resolution column IS the current implementation.

## AX Principles Applied

This API follows RFC-025 Agent Experience (AX):

1. **Predictable names** — `Config` not `Cfg`, `Service` not `Srv`
2. **Usage-example comments** — every public function shows HOW with real values
3. **Path is documentation** — `c.Data().ReadString("prompts/coding.md")`
4. **Universal types** — Option, Options, Result everywhere
5. **Event-driven** — ACTION/QUERY/PERFORM, not direct function calls between services
6. **Tests as spec** — `TestFile_Function_{Good,Bad,Ugly}` for every function
7. **Export primitives** — Core is Lego bricks, not an encapsulated library
8. **Naming encodes architecture** — CamelCase = primitive brick, UPPERCASE = consumer convenience
9. **File = concern** — one file, one job (ipc.go = registry, action.go = execution, contract.go = types)



## Findings Summary — 108 Findings Across 13 Passes (All Resolved)

The full discovery process (Passes 2-13) produced 108 findings that reduce to 5 root causes.
All findings are resolved in v0.8.0. Full pass detail preserved in git history.

| Pass | Focus | Findings | Key Resolutions |
|------|-------|----------|----------------|
| 2 | Architecture | 8 | Core fields exported (Lego Bricks), Fs.root exception, Config untyped bag |
| 3 | Spec contradictions | 8 | Startable returns Result, Process returns Result, Registry lock modes |
| 4 | Concurrency | 12 | Registry[T] per-subsystem mutex, WriteAtomic, PerformAsync shutdown race |
| 5 | Consumer experience | 8 | ServiceRuntime + manual .core both valid, HandleIPCEvents documented |
| 6 | Cross-repo cascade | 8 | TaskDef replaces nested ACTION cascade (P6-1), MCP aggregator pattern |
| 7 | Failure modes | 8 | RunE() + defer shutdown, panic recovery in ACTION/Action, SafeGo |
| 8 | Type safety | 8 | Result trade-off accepted, typed convenience methods, AX-7 Ugly tests |
| 9 | Missing primitives | 8 | core.ID(), ValidateName, WriteAtomic, Fs.NewUnrestricted, c.Process() |
| 10 | Spec self-audit | 8 | Subsystem count corrected, cross-reference table, findings index |
| 11 | Security model | 8 | Entitlement primitive (Section 21), Fs.NewUnrestricted, audit logging |
| 12 | Migration risk | 8 | Phase 1 additive (zero breakage), Phase 3 breaking (Startable, CommandLifecycle) |
| 13 | Hidden assumptions | 8 | Single Core per process, Go only, Linux/macOS, static conclave, ephemeral IPC |

## Synthesis — Five Root Causes

> With all 108 findings in context, the patterns emerge. 60 findings
> cluster into 5 root causes. Fix the root, fix the cluster.

### Root Cause 1: Type Erasure via Result{any} — 16 findings

`Result{Value: any, OK: bool}` erases all type information. Every consumer writes bare type assertions that panic on wrong types. The LOC reduction is real but the compile-time safety loss creates 50+ hidden panic sites.

**The tension is fundamental.** Result exists to reduce downstream LOC. But `any` means the compiler can't help. This isn't fixable without abandoning Result — which defeats Core's purpose.

**Mitigation, not fix:** Typed convenience methods (`ReadString`, `ListEntries`, `ConfigGet[T]`). AX-7 Ugly tests for every type assertion. `Registry[T]` where generics work. Accept `Result` as the integration seam where types meet.

### Root Cause 2: No Internal Boundaries — 14 findings

`*Core` grants God Mode. Every service sees everything. The unexported fields were an attempt at boundaries but `unsafe.Pointer` proves they don't work. The conclave has no isolation.

**This is by design for v0.8.0.** All services are first-party trusted code. The Lego Bricks philosophy says "export everything." The tension is: Lego Bricks vs Least Privilege.

**Resolution:** Section 21 (Entitlement primitive) — implemented. `c.Entitled()` gates Actions. Default permissive, consumer replaces checker. Port of RFC-004 concept:

```
Registration = capability  ("process.run action exists")
Entitlement  = permission  ("this Core is ALLOWED to run processes")
```

```go
c.Entitlement("process.run")  // true if both registered AND permitted
c.Action("process.run").Run() // checks entitlement before executing
```

The Entitlement system (RFC-004) answers "can this workspace do this action?" with package-based feature gating, usage limits, and boost mechanics. Config Channels (RFC-003) add context — "what settings apply on this surface (CLI vs MCP vs HTTP)?" Together they provide the boundary model without removing Lego Bricks — all bricks exist, entitlements control which ones are usable.

See: RFC-003 (Config Channels), RFC-004 (Entitlements), RFC-005 (Commerce Matrix).

### Root Cause 3: Synchronous Everything — 12 findings

IPC dispatch is synchronous. Startup is synchronous. File I/O assumes no concurrency. The one async path (`PerformAsync`) is unbounded. When anything runs concurrently — which it does in production — races emerge.

**The cascade (P6-1) is the symptom.** The root cause is that Core was designed for sequential execution and concurrency was added incrementally without revisiting the foundations.

**Resolution:** The Action/Task system (Section 18) is implemented in core/go. `TaskDef` with `Steps` supports sequential chains, async dispatch, and previous-input piping. The cascade fix requires core/agent to wire its handlers as named Actions and replace the nested `c.ACTION()` calls with `c.Task("agent.completion").Run()`. See `core/agent/docs/plans/2026-03-25-core-go-v0.8.0-migration.md` Priority 5.

### Root Cause 4: No Recovery Path — 10 findings

Every failure mode is "log and crash." `os.Exit(1)` bypasses defers. Startup failure leaks running services. Panicking handlers crash the process. `SafeGo` exists but isn't used.

**One fix resolves most of this cluster:**

```go
func (c *Core) Run() error {
    defer c.ServiceShutdown(context.Background())
    // ... no os.Exit, return errors
}
```

`defer` ensures cleanup always runs. Returning `error` lets `main()` handle the exit. Panic recovery in ACTION handlers prevents cascade crashes. Wire `SafeGo` as the standard goroutine launcher.

### Root Cause 5: Missing Primitives — 8 findings

The guardrail coverage is incomplete. Strings have primitives. Paths have primitives. Errors have primitives. But JSON, time, IDs, validation, and health don't. Each gap means consumers reinvent the wheel — and weaker models get it wrong.

**Resolution:** Prioritise by usage frequency:
1. `core.ID()` — used everywhere, 3 different patterns today
2. `core.Validate(name/path)` — copy-pasted 3 times today
3. `core.Health()` — needed for production monitoring
4. `core.Time()` / timestamp convention — document RFC3339
5. JSON — judgment call, may be unnecessary wrapping

### What This Means for v0.8.0

The five root causes map to a priority order:

| Priority | Root Cause | v0.8.0 Action |
|----------|-----------|---------------|
| 1 | No recovery (10) | Fix Run(), add defer, panic recovery — **Phase 1** |
| 2 | Synchronous (12) | Fix ACTION chain bug, design Task system — **Phase 1-2** |
| 3 | Missing primitives (8) | Add ID, Validate, Health — **Phase 1** |
| 4 | Type erasure (16) | Add typed convenience methods, AX-7 tests — **ongoing** |
| 5 | No boundaries (14) | Section 21 Entitlement primitive — implemented. `c.Entitled()` + `Action.Run()` enforcement |

Root causes 1-4 are resolved. Root cause 5 (boundaries) is designed (Section 21) and implementation is v0.8.0 scope.

### Cross-References — Existing RFCs That Solve Open Problems

Core/go provides the INTERFACE (stdlib only). Consumer packages bring the IMPLEMENTATION. These existing RFCs map directly to open findings:

| Finding | Existing RFC | Core Provides (interface) | Consumer Provides (impl) |
|---------|-------------|--------------------------|-------------------------|
| P13-5: Sync startup | RFC-002 (Event-Driven Modules) | `Startable` + event declarations | Lazy instantiation based on `$listens` pattern |
| P11-1: God Mode | RFC-004 (Entitlements) | `c.Entitlement(action) bool` | Package/feature gating, usage limits |
| P11-3: Secret exposure | RFC-012 (SMSG) | `c.Secret(name) string` | SMSG decrypt, Vault, env fallback |
| P9-6: No validation | RFC-009 (Sigil Transforms) | Composable transform chain interface | Validators, sanitisers, reversible transforms |
| P11-2: Fs sandbox bypass | RFC-014 (TIM) | `c.Fs()` sandbox root | TIM container = OS-level isolation boundary |
| P13-2: Go only | RFC-013 (DataNode) | `c.Data()` mounts `fs.FS` | DataNode = in-memory fs, test mounts, cross-language data |
| P2-8: Logging gap | RFC-003 (Config Channels) | `c.Config()` with channel context | Settings vary by surface (CLI vs MCP vs HTTP) |

**The pattern:** Core defines a primitive with a Go interface. The RFC describes the concept. A consumer package implements it. Core stays stdlib-only. The ecosystem gets rich features via composition.

```
core/go:          c.Secret(name) → looks up in Registry["secrets"]
go-smsg:          registers SMSG decryptor as secret provider
go-vault:         registers HashiCorp Vault as secret provider
env fallback:     built into core/go (os.Getenv) — no extra dependency

core/go:          c.Entitlement(action) → looks up in Registry["entitlements"]
go-entitlements:  ports RFC-004 from CorePHP, registers package/feature checker
default:          built into core/go — returns true (no restrictions, trusted conclave)
```

No dependency injected into core/go. The interface is the primitive. The implementation is the consumer.

---

## Consumer RFCs

core/go provides the primitives. These RFCs describe how consumers use them:

| Package | RFC | Scope |
|---------|-----|-------|
| go-process | `core/go-process/docs/RFC.md` | Action handlers for process.run/start/kill, ManagedProcess, daemon registry |
| core/agent | `core/agent/docs/RFC.md` | Named Actions, completion pipeline (P6-1 fix), WriteAtomic migration, Process migration, Entitlement gating |

Each consumer RFC is self-contained — an agent can implement it from the document alone.

---

## Versioning

### Release Model

```
v0.7.x  — previous stable
v0.8.0  — production release: all primitives, all boundaries, all consumers aligned
          Sections 1-21 implemented. 483 tests, 84.7% coverage, 100% AX-7 naming.
v0.8.*  — patches tell us where the agentic process missed things
```

### The Cadence

1. **RFC spec** — design the target version in prose (this document)
2. **v0.7.x patches** — mechanical fixes that don't change the API contract
3. **Implementation** — build Sections 17-20, resolve design issues
4. **AX-7 at 100%** — every function has Good/Bad/Ugly tests
5. **Tag v0.8.0** — only when 100% confident it's production ready
6. **Measure v0.8.x** — each patch tells you what the spec missed

The fallout versions are the feedback loop. v0.8.1 means the spec missed one thing. v0.8.15 means the spec missed fifteen things. The patch count per release IS the quality metric — it tells you how wrong you were.

### What v0.8.0 Requires

| Requirement | Status |
|-------------|--------|
| All 16 Known Issues resolved in code | **Done** (2026-03-25) |
| Section 17: c.Process() primitive | **Done** — Action sugar |
| Section 18: Action/Task system | **Done** — ActionDef→Action, TaskDef→Task, type Task any removed |
| Section 19: c.API() streams | **Done** — Stream interface, protocol handlers, RemoteAction |
| Section 20: Registry[T] primitive | **Done** — all 5 registries migrated |
| Section 21: Entitlement primitive | **Done** — Entitled(), SetEntitlementChecker(), RecordUsage(), Action.Run() enforcement |
| AX-7 test coverage at 100% | **Done** — core/go 456/456 (100%) |
| Zero os/exec in core/go | **Done** — App.Find() uses os.Stat |
| type Task any removed | **Done** — PerformAsync takes named action + Options |
| Startable/Stoppable return Result | **Done** — breaking, clean |
| CommandLifecycle removed | **Done** → Command.Managed field |
| Consumer RFCs written | **Done** — go-process/docs/RFC.md, core/agent/docs/RFC.md |

### What Blocks v0.8.0 Tag

- go-process v0.7.0 alignment (consumer RFC written, ready to implement)
- core/agent v0.8.0 migration (consumer RFC written, Phase 1 ready)

### What Does NOT Block v0.8.0

- Ecosystem sweep (Plan 6 — after consumers align)
- core/cli update (extension, not primitive)

## 21. Entitlement — The Permission Primitive

> Status: Design spec. v0.8.0 scope — the permission boundary for the ecosystem.
> Core provides the primitive. go-entitlements and commerce-matrix provide implementations.

### 21.1 The Problem

`*Core` grants God Mode (P11-1). Every service sees everything. The 14 findings in Root Cause 2 all stem from this. The conclave is trusted — but the SaaS platform (RFC-004), the commerce hierarchy (RFC-005), and the agent sandbox all need boundaries.

Three systems ask the same question with different vocabulary:

```
Can [subject] do [action] with [quantity] in [context]?
```

| System | Subject | Action | Quantity | Context |
|--------|---------|--------|----------|---------|
| RFC-004 Entitlements | workspace | feature.code | N | active packages |
| RFC-005 Commerce Matrix | entity (M1/M2/M3) | permission.key | 1 | hierarchy path |
| Core Actions | this Core instance | action.name | 1 | registered services |

### 21.2 The Primitive

```go
// Entitlement is the result of a permission check.
// Carries context for both boolean gates (Allowed) and usage limits (Limit/Used/Remaining).
// Maps directly to RFC-004 EntitlementResult and RFC-005 PermissionResult.
type Entitlement struct {
    Allowed   bool   // permission granted
    Unlimited bool   // no cap (agency tier, admin, trusted conclave)
    Limit     int    // total allowed (0 = boolean gate, no quantity dimension)
    Used      int    // current consumption
    Remaining int    // Limit - Used
    Reason    string // denial reason — for UI feedback and audit logging
}

// Entitled checks if an action is permitted in the current context.
// Default: always returns Allowed=true, Unlimited=true (trusted conclave).
// With go-entitlements: checks workspace packages, features, usage, boosts.
// With commerce-matrix: checks entity hierarchy, lock cascade.
//
//   e := c.Entitled("process.run")           // boolean — can this Core run processes?
//   e := c.Entitled("social.accounts", 3)    // quantity — can workspace create 3 more accounts?
//   if e.Allowed { proceed() }
//   if e.NearLimit(0.8) { showWarning() }
func (c *Core) Entitled(action string, quantity ...int) Entitlement
```

### 21.3 The Checker — Consumer-Provided

Core defines the interface. Consumer packages provide the implementation.

```go
// EntitlementChecker answers "can [subject] do [action] with [quantity]?"
// Subject comes from context (workspace, entity, user — consumer's concern).
type EntitlementChecker func(action string, quantity int, ctx context.Context) Entitlement
```

Registration via Core:

```go
// SetEntitlementChecker replaces the default (permissive) checker.
// Called by go-entitlements or commerce-matrix during OnStartup.
//
//   func (s *EntitlementService) OnStartup(ctx context.Context) core.Result {
//       s.Core().SetEntitlementChecker(s.check)
//       return core.Result{OK: true}
//   }
func (c *Core) SetEntitlementChecker(checker EntitlementChecker)
```

Default checker (no entitlements package loaded):

```go
// defaultChecker — trusted conclave, everything permitted
func defaultChecker(action string, quantity int, ctx context.Context) Entitlement {
    return Entitlement{Allowed: true, Unlimited: true}
}
```

### 21.4 Enforcement Point — Action.Run()

The entitlement check lives in `Action.Run()`, before execution. One enforcement point for all capabilities.

```go
func (a *Action) Run(ctx context.Context, opts Options) (result Result) {
    if !a.Exists() { return not-registered }
    if !a.enabled { return disabled }

    // Entitlement check — permission boundary
    if e := a.core.Entitled(a.Name); !e.Allowed {
        return Result{E("action.Run",
            Concat("not entitled: ", a.Name, " — ", e.Reason), nil), false}
    }

    defer func() { /* panic recovery */ }()
    return a.Handler(ctx, opts)
}
```

Three states for any action:

| State | Exists() | Entitled() | Run() |
|-------|----------|------------|-------|
| Not registered | false | — | Result{OK: false} "not registered" |
| Registered, not entitled | true | false | Result{OK: false} "not entitled" |
| Registered and entitled | true | true | executes handler |

### 21.5 How RFC-004 (SaaS Entitlements) Plugs In

go-entitlements registers as a service and replaces the checker:

```go
// In go-entitlements:
func (s *Service) OnStartup(ctx context.Context) core.Result {
    s.Core().SetEntitlementChecker(func(action string, qty int, ctx context.Context) core.Entitlement {
        workspace := s.workspaceFromContext(ctx)
        if workspace == nil {
            return core.Entitlement{Allowed: true, Unlimited: true} // no workspace = system context
        }

        result := s.Can(workspace, action, qty)

        return core.Entitlement{
            Allowed:   result.IsAllowed(),
            Unlimited: result.IsUnlimited(),
            Limit:     result.Limit,
            Used:      result.Used,
            Remaining: result.Remaining,
            Reason:    result.Message(),
        }
    })
    return core.Result{OK: true}
}
```

Maps 1:1 to RFC-004's `EntitlementResult`:
- `$result->isAllowed()` → `e.Allowed`
- `$result->isUnlimited()` → `e.Unlimited`
- `$result->limit` → `e.Limit`
- `$result->used` → `e.Used`
- `$result->remaining` → `e.Remaining`
- `$result->getMessage()` → `e.Reason`
- `$result->isNearLimit()` → `e.NearLimit(0.8)`
- `$result->getUsagePercentage()` → `e.UsagePercent()`

### 21.6 How RFC-005 (Commerce Matrix) Plugs In

commerce-matrix registers and replaces the checker with hierarchy-aware logic:

```go
// In commerce-matrix:
func (s *MatrixService) OnStartup(ctx context.Context) core.Result {
    s.Core().SetEntitlementChecker(func(action string, qty int, ctx context.Context) core.Entitlement {
        entity := s.entityFromContext(ctx)
        if entity == nil {
            return core.Entitlement{Allowed: true, Unlimited: true}
        }

        result := s.Can(entity, action, "")

        return core.Entitlement{
            Allowed: result.IsAllowed(),
            Reason:  result.Reason,
        }
    })
    return core.Result{OK: true}
}
```

Maps to RFC-005's cascade model:
- `M1 says NO → everything below is NO` → checker walks hierarchy, returns `{Allowed: false, Reason: "Locked by M1"}`
- Training mode → checker returns `{Allowed: false, Reason: "undefined — training required"}`
- Production strict mode → undefined = denied

### 21.7 Composing Both Systems

When a SaaS platform ALSO has commerce hierarchy (Host UK), the checker composes internally:

```go
func (s *CompositeService) check(action string, qty int, ctx context.Context) core.Entitlement {
    // Check commerce matrix first (hard permissions)
    matrixResult := s.matrix.Can(entityFromCtx(ctx), action, "")
    if matrixResult.IsDenied() {
        return core.Entitlement{Allowed: false, Reason: matrixResult.Reason}
    }

    // Then check entitlements (usage limits)
    entResult := s.entitlements.Can(workspaceFromCtx(ctx), action, qty)
    return core.Entitlement{
        Allowed:   entResult.IsAllowed(),
        Unlimited: entResult.IsUnlimited(),
        Limit:     entResult.Limit,
        Used:      entResult.Used,
        Remaining: entResult.Remaining,
        Reason:    entResult.Message(),
    }
}
```

Matrix (hierarchy) gates first. Entitlements (usage) gate second. One checker, composed.

### 21.8 Convenience Methods on Entitlement

```go
// NearLimit returns true if usage exceeds the threshold percentage.
// RFC-004: $result->isNearLimit() uses 80% threshold.
//
//   if e.NearLimit(0.8) { showUpgradePrompt() }
func (e Entitlement) NearLimit(threshold float64) bool

// UsagePercent returns current usage as a percentage of the limit.
// RFC-004: $result->getUsagePercentage()
//
//   pct := e.UsagePercent()  // 75.0
func (e Entitlement) UsagePercent() float64

// RecordUsage is called after a gated action succeeds.
// Delegates to the entitlement service for usage tracking.
// This is the equivalent of RFC-004's $workspace->recordUsage().
//
//   e := c.Entitled("ai.credits", 10)
//   if e.Allowed {
//       doWork()
//       c.RecordUsage("ai.credits", 10)
//   }
func (c *Core) RecordUsage(action string, quantity ...int)
```

### 21.9 Audit Trail — RFC-004 Section: Audit Logging

Every entitlement check can be logged via `core.Security()`:

```go
func (c *Core) Entitled(action string, quantity ...int) Entitlement {
    qty := 1
    if len(quantity) > 0 {
        qty = quantity[0]
    }

    e := c.entitlementChecker(action, qty, c.Context())

    // Audit logging for denials (P11-6)
    if !e.Allowed {
        Security("entitlement.denied", "action", action, "quantity", qty, "reason", e.Reason)
    }

    return e
}
```

### 21.10 Core Struct Changes

```go
type Core struct {
    // ... existing fields ...
    entitlementChecker EntitlementChecker  // default: everything permitted
}
```

Constructor:

```go
func New(opts ...CoreOption) *Core {
    c := &Core{
        // ... existing ...
        entitlementChecker: defaultChecker,
    }
    // ...
}
```

### 21.11 What This Does NOT Do

- **Does not add database dependencies** — Core is stdlib only. Usage tracking, package management, billing — all in consumer packages.
- **Does not define features** — The feature catalogue (social.accounts, ai.credits, etc.) is defined by the SaaS platform, not Core.
- **Does not manage subscriptions** — Commerce (RFC-005) and billing (Blesta/Stripe) are consumer concerns.
- **Does not replace Action registration** — Registration IS capability. Entitlement IS permission. Both must be true.
- **Does not enforce at Config/Data/Fs level** — v0.8.0 gates Actions. Config/Data/Fs gating requires per-subsystem entitlement checks (same pattern, more integration points).

### 21.12 The Subsystem Map (Updated)

```
c.Registry()     — universal named collection
c.Options()      — input configuration
c.App()          — identity
c.Config()       — runtime settings
c.Data()         — embedded assets
c.Drive()        — connection config (WHERE)
c.API()          — remote streams (HOW) [planned]
c.Fs()           — filesystem
c.Process()      — managed execution (Action sugar)
c.Action()       — named callables (register, invoke, inspect)
c.Task()         — composed Action sequences
c.IPC()          — local message bus
c.Cli()          — command tree
c.Log()          — logging
c.Error()        — panic recovery
c.I18n()         — internationalisation
c.Entitled()     — permission check (NEW)
c.RecordUsage()  — usage tracking (NEW)
```

---

## Changelog

- 2026-03-25: v0.8.0 — All 21 sections implemented. 483 tests, 84.7% coverage, 100% AX-7 naming.
- 2026-03-25: Initial specification created from 500k token discovery session. 108 findings, 5 root causes, 13 review passes. Discovery detail preserved in git history.
