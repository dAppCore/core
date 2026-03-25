# CoreGO API Contract — RFC Specification

> `dappco.re/go/core` — Dependency injection, service lifecycle, and message-passing framework.
> This document is the authoritative API contract. An agent should be able to write a service
> that registers with Core from this document alone.

**Status:** Living document
**Module:** `dappco.re/go/core`
**Version:** v0.7.0+

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
Run() → ServiceStartup() → Cli.Run() → ServiceShutdown()
```

`Run()` is blocking. `ServiceStartup` calls `OnStartup(ctx)` on all services implementing `Startable`. `ServiceShutdown` calls `OnShutdown(ctx)` on all `Stoppable` services. Shutdown uses `context.Background()` — not the Core context (which is already cancelled).

### 1.3 Subsystem Accessors

Every subsystem is accessed via a method on Core:

```go
c.Options()   // *Options  — input configuration
c.App()       // *App      — application metadata (name, version)
c.Config()    // *Config   — runtime settings, feature flags
c.Data()      // *Data     — embedded assets mounted by packages
c.Drive()     // *Drive    — transport handles (API, MCP, SSH)
c.Fs()        // *Fs       — filesystem I/O (sandboxable)
c.Cli()       // *Cli      — CLI command framework
c.IPC()       // *Ipc      — message bus internals
c.I18n()      // *I18n     — internationalisation
c.Error()     // *ErrorPanic — panic recovery
c.Log()       // *ErrorLog  — structured logging
c.Context()   // context.Context — Core's lifecycle context
c.Env(key)    // string    — environment variable (cached at init)
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

### 2.4 Message, Query, Task

IPC type aliases — all are `any` at the type level, distinguished by usage:

```go
type Message any  // broadcast via ACTION — fire and forget
type Query any    // request/response via QUERY — returns first handler's result
type Task any     // work unit via PERFORM — tracked with progress
```

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
- **Startable interface** → `OnStartup(ctx) error` called during `ServiceStartup`
- **Stoppable interface** → `OnShutdown(ctx) error` called during `ServiceShutdown`
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
    OnStartup(ctx context.Context) error
}

type Stoppable interface {
    OnShutdown(ctx context.Context) error
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

All handlers receive all messages. Type-switch to filter. Return `Result{OK: true}` always (errors are logged, not propagated).

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

### 4.3 PERFORM (tracked task)

```go
// Execute with progress tracking
c.PERFORM(MyTask{Data: payload})

// Register task handler
c.RegisterTask(func(c *core.Core, t core.Task) core.Result {
    // do work, report progress
    c.Progress(taskID, 0.5, "halfway done", t)
    return core.Result{Value: output, OK: true}
})
```

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
r := c.Data().Mounts() // []string of mount names
```

---

## 7. Drive — Transport Handles

Registry of named transport handles (API endpoints, MCP servers, etc):

```go
c.Drive().New(core.NewOptions(
    core.Option{Key: "name", Value: "forge"},
    core.Option{Key: "transport", Value: "https://forge.lthn.ai"},
))

r := c.Drive().Get("forge")     // Result with DriveHandle
c.Drive().Has("forge")          // true
c.Drive().Names()               // []string
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

// Delete
r := fs.Delete(path)      // single file
r := fs.DeleteAll(path)   // recursive
r := fs.Rename(old, new)
r := fs.Stat(path)        // Result{Value: os.FileInfo}
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

## 17. Process — Core Primitive (Planned)

> Status: Design spec. Not yet implemented. go-process v0.7.0 will implement this.

### 17.1 The Primitive

`c.Process()` is a Core subsystem accessor — same pattern as `c.Fs()`, `c.Config()`, `c.Log()`. It provides the **interface** for process management. go-process provides the **implementation** via service registration.

```go
c.Process()          // *Process — primitive (defined in core/go)
c.Process().Run()    // executes via IPC → go-process handles it (if registered)
```

If go-process is not registered, process IPC messages go unanswered. No capability = no execution. This is permission-by-registration, not permission-by-config.

### 17.2 Primitive Interface (core/go provides)

Core defines the Process primitive as a thin struct with methods that emit IPC messages:

```go
// Process is the Core primitive for process management.
// Methods emit IPC messages — actual execution is handled by
// whichever service registers to handle ProcessRun/ProcessStart messages.
type Process struct {
    core *Core
}

// Accessor on Core
func (c *Core) Process() *Process { return c.process }
```

### 17.3 Synchronous Execution

```go
// Run executes a command and waits for completion.
// Returns (output, error). Emits ProcessRun via IPC.
//
//   out, err := c.Process().Run(ctx, "git", "log", "--oneline")
func (p *Process) Run(ctx context.Context, command string, args ...string) (string, error)

// RunIn executes in a specific directory.
//
//   out, err := c.Process().RunIn(ctx, "/path/to/repo", "go", "test", "./...")
func (p *Process) RunIn(ctx context.Context, dir string, command string, args ...string) (string, error)

// RunWithEnv executes with additional environment variables.
//
//   out, err := c.Process().RunWithEnv(ctx, dir, []string{"GOWORK=off"}, "go", "test")
func (p *Process) RunWithEnv(ctx context.Context, dir string, env []string, command string, args ...string) (string, error)
```

### 17.4 Async / Detached Execution

```go
// Start spawns a detached process. Returns a handle for monitoring.
// The process survives Core shutdown if Detach is true.
//
//   handle, err := c.Process().Start(ctx, ProcessOptions{
//       Command: "docker", Args: []string{"run", "..."},
//       Dir: repoDir, Detach: true,
//   })
func (p *Process) Start(ctx context.Context, opts ProcessOptions) (*ProcessHandle, error)

// ProcessOptions configures process execution.
type ProcessOptions struct {
    Command string
    Args    []string
    Dir     string
    Env     []string
    Detach  bool          // survives parent, own process group
    Timeout time.Duration // 0 = no timeout
}
```

### 17.5 Process Handle

```go
// ProcessHandle is returned by Start for monitoring and control.
type ProcessHandle struct {
    ID     string         // go-process managed ID
    PID    int            // OS process ID
}

// Methods on the handle — all emit IPC messages
func (h *ProcessHandle) IsRunning() bool
func (h *ProcessHandle) Kill() error
func (h *ProcessHandle) Done() <-chan struct{}
func (h *ProcessHandle) Output() string
func (h *ProcessHandle) Wait() error
func (h *ProcessHandle) Info() ProcessInfo

type ProcessInfo struct {
    ID        string
    PID       int
    Status    string        // pending, running, exited, failed, killed
    ExitCode  int
    Duration  time.Duration
    StartedAt time.Time
}
```

### 17.6 IPC Messages (core/go defines)

Core defines the message types. go-process registers handlers for them. If no handler is registered, calls return `Result{OK: false}` — no process capability.

```go
// Request messages — emitted by c.Process() methods
type ProcessRun struct {
    Command string
    Args    []string
    Dir     string
    Env     []string
}

type ProcessStart struct {
    Command string
    Args    []string
    Dir     string
    Env     []string
    Detach  bool
    Timeout time.Duration
}

type ProcessKill struct {
    ID  string  // by go-process ID
    PID int     // fallback by OS PID
}

// Event messages — emitted by go-process implementation
type ProcessStarted struct {
    ID      string
    PID     int
    Command string
}

type ProcessOutput struct {
    ID   string
    Line string
}

type ProcessExited struct {
    ID       string
    PID      int
    ExitCode int
    Status   string
    Duration time.Duration
    Output   string
}

type ProcessKilled struct {
    ID  string
    PID int
}
```

### 17.7 Permission by Registration

This is the key security model. The IPC bus is the permission boundary:

```go
// If go-process IS registered:
c.Process().Run(ctx, "git", "log")
// → emits ProcessRun via IPC
// → go-process handler receives it
// → executes, returns output
// → Result{Value: output, OK: true}

// If go-process is NOT registered:
c.Process().Run(ctx, "git", "log")
// → emits ProcessRun via IPC
// → no handler registered
// → Result{OK: false}
// → caller gets empty result, no execution happened
```

No config flags, no permission files, no capability tokens. The service either exists in the conclave or it doesn't. Registration IS permission.

This means:
- **A sandboxed Core** (no go-process registered) cannot execute any external commands
- **A full Core** (go-process registered) can execute anything the OS allows
- **A restricted Core** could register a filtered go-process that only allows specific commands
- **Tests** can register a mock process service that records calls without executing

### 17.8 Convenience Helpers (per-package)

Packages that frequently run commands can create local helpers that delegate to `c.Process()`:

```go
// In pkg/agentic/proc.go:
func (s *PrepSubsystem) gitCmd(ctx context.Context, dir string, args ...string) (string, error) {
    return s.core.Process().RunIn(ctx, dir, "git", args...)
}

func (s *PrepSubsystem) gitCmdOK(ctx context.Context, dir string, args ...string) bool {
    _, err := s.gitCmd(ctx, dir, args...)
    return err == nil
}
```

These replace the current standalone `proc.go` helpers that bootstrap their own process service. The helpers become methods on the service that owns `*Core`.

### 17.9 go-process Implementation (core/go-process provides)

go-process registers itself as the ProcessRun/ProcessStart handler:

```go
// go-process service registration
func Register(c *core.Core) core.Result {
    svc := &Service{
        ServiceRuntime: core.NewServiceRuntime(c, Options{}),
        processes:      make(map[string]*ManagedProcess),
    }

    // Register as IPC handler for process messages
    c.RegisterAction(func(c *core.Core, msg core.Message) core.Result {
        switch m := msg.(type) {
        case core.ProcessRun:
            return svc.handleRun(m)
        case core.ProcessStart:
            return svc.handleStart(m)
        case core.ProcessKill:
            return svc.handleKill(m)
        }
        return core.Result{OK: true}
    })

    return core.Result{Value: svc, OK: true}
}
```

### 17.10 Migration Path

Current state → target state:

| Current | Target |
|---------|--------|
| `proc.go` standalone helpers with `ensureProcess()` | Methods on PrepSubsystem using `s.core.Process()` |
| `process.RunWithOptions(ctx, opts)` global function | `c.Process().Run(ctx, cmd, args...)` via IPC |
| `process.StartWithOptions(ctx, opts)` global function | `c.Process().Start(ctx, opts)` via IPC |
| `syscall.Kill(pid, 0)` direct OS calls | `handle.IsRunning()` via go-process |
| `syscall.Kill(pid, SIGTERM)` direct OS calls | `handle.Kill()` via go-process |
| `process.SetDefault(svc)` global singleton | Service registered in Core conclave |
| `agentic.ProcessRegister` bridge wrapper | `process.Register` direct factory |

---

## 18. Action and Task — The Execution Primitives (Planned)

> Status: Design spec. Replaces the current `ACTION`/`PERFORM` broadcast model
> with named, composable execution units.

### 18.1 The Concept

The current IPC has three verbs:
- `ACTION(msg)` — broadcast fire-and-forget
- `QUERY(q)` — first responder wins
- `PERFORM(t)` — first executor wins

This works but treats everything as anonymous messages. There's no way to:
- Name a callable and invoke it by name
- Chain callables into flows
- Schedule a callable for later
- Inspect what callables are registered

**Action** is the fix. An Action is a named, registered callable. The atomic unit of work in Core.

### 18.2 core.Action() — The Atomic Unit

```go
// Register a named action
c.Action("git.log", func(ctx context.Context, opts core.Options) core.Result {
    dir := opts.String("dir")
    return c.Process().RunIn(ctx, dir, "git", "log", "--oneline", "-20")
})

// Invoke by name
r := c.Action("git.log").Run(ctx, core.NewOptions(
    core.Option{Key: "dir", Value: "/path/to/repo"},
))
if r.OK {
    log := r.Value.(string)
}

// Check if an action exists (permission check)
if c.Action("process.run").Exists() {
    // process capability is available
}
```

`c.Action(name)` is dual-purpose like `c.Service(name)`:
- With a handler arg → registers the action
- Without → returns the action for invocation

### 18.3 Action Signature

```go
// ActionHandler is the function signature for all actions.
type ActionHandler func(context.Context, Options) Result

// ActionDef is a registered action.
type ActionDef struct {
    Name        string
    Handler     ActionHandler
    Description string        // AX: human + agent readable
    Schema      Options       // declares expected input keys (optional)
}
```

### 18.4 Where Actions Come From

Services register their actions during `OnStartup`. This is the same pattern as command registration — services own their capabilities:

```go
func (s *MyService) OnStartup(ctx context.Context) error {
    c := s.Core()

    c.Action("process.run", s.handleRun)
    c.Action("process.start", s.handleStart)
    c.Action("process.kill", s.handleKill)

    c.Action("git.clone", s.handleGitClone)
    c.Action("git.push", s.handleGitPush)

    return nil
}
```

go-process registers `process.*` actions. core/agent registers `agentic.*` actions. The action namespace IS the capability map.

### 18.5 The Permission Model

If `process.run` is not registered, calling it returns `Result{OK: false}`. This is the same "registration IS permission" model from Section 17.7, but generalised to ALL capabilities:

```go
// Full Core — everything available
c := core.New(
    core.WithService(process.Register),   // registers process.* actions
    core.WithService(agentic.Register),   // registers agentic.* actions
    core.WithService(brain.Register),     // registers brain.* actions
)

// Sandboxed Core — no process, no brain
c := core.New(
    core.WithService(agentic.Register),   // only agentic.* actions
)
// c.Action("process.run").Run(...)  → Result{OK: false}
// c.Action("brain.recall").Run(...) → Result{OK: false}
```

### 18.6 core.Task() — Composing Actions

A Task is a named sequence, chain, or graph of Actions. Think n8n nodes but in code.

```go
// Sequential chain — stops on first failure
c.Task("deploy", core.TaskDef{
    Description: "Build, test, and deploy to production",
    Steps: []core.Step{
        {Action: "go.build",   With: core.Options{...}},
        {Action: "go.test",    With: core.Options{...}},
        {Action: "docker.push", With: core.Options{...}},
        {Action: "ansible.deploy", With: core.Options{...}},
    },
})

// Run the task
r := c.Task("deploy").Run(ctx, core.NewOptions(
    core.Option{Key: "target", Value: "production"},
))
```

### 18.7 Task Composition Patterns

```go
// Chain — sequential, output of each feeds next
c.Task("review-pipeline", core.TaskDef{
    Steps: []core.Step{
        {Action: "agentic.dispatch", With: opts},
        {Action: "agentic.verify",   Input: "previous"},  // gets output of dispatch
        {Action: "agentic.merge",    Input: "previous"},
    },
})

// Parallel — all run concurrently, wait for all
c.Task("multi-repo-sweep", core.TaskDef{
    Parallel: []core.Step{
        {Action: "agentic.dispatch", With: optsGoIO},
        {Action: "agentic.dispatch", With: optsGoLog},
        {Action: "agentic.dispatch", With: optsGoMCP},
    },
})

// Conditional — branch on result
c.Task("qa-gate", core.TaskDef{
    Steps: []core.Step{
        {Action: "go.test"},
        {
            If:   "previous.OK",
            Then: core.Step{Action: "agentic.merge"},
            Else: core.Step{Action: "agentic.flag-review"},
        },
    },
})

// Scheduled — run at a specific time or interval
c.Task("nightly-sweep", core.TaskDef{
    Schedule: "0 2 * * *",  // cron: 2am daily
    Steps: []core.Step{
        {Action: "agentic.scan"},
        {Action: "agentic.dispatch-fixes", Input: "previous"},
    },
})
```

### 18.8 How This Relates to Existing IPC

The current IPC verbs become invocation modes for Actions:

| Current | Becomes | Purpose |
|---------|---------|---------|
| `c.ACTION(msg)` | `c.Action("name").Broadcast(opts)` | Fire-and-forget to ALL handlers |
| `c.QUERY(q)` | `c.Action("name").Query(opts)` | First responder wins |
| `c.PERFORM(t)` | `c.Action("name").Run(opts)` | Execute and return result |
| `c.PerformAsync(t)` | `c.Action("name").RunAsync(opts)` | Background with progress |

The anonymous message types (`Message`, `Query`, `Task`) still work for backwards compatibility. Named Actions are the AX-native way forward.

### 18.9 How Process Fits

Section 17's `c.Process()` is syntactic sugar over Actions:

```go
// c.Process().Run(ctx, "git", "log") is equivalent to:
c.Action("process.run").Run(ctx, core.NewOptions(
    core.Option{Key: "command", Value: "git"},
    core.Option{Key: "args", Value: []string{"log"}},
))

// c.Process().Start(ctx, opts) is equivalent to:
c.Action("process.start").Run(ctx, core.NewOptions(
    core.Option{Key: "command", Value: opts.Command},
    core.Option{Key: "args", Value: opts.Args},
    core.Option{Key: "detach", Value: true},
))
```

The `Process` primitive is a typed convenience layer. Under the hood, it's Actions all the way down.

### 18.10 Inspecting the Action Registry

```go
// List all registered actions
actions := c.Actions()  // []string{"process.run", "process.start", "agentic.dispatch", ...}

// Check capabilities
c.Action("process.run").Exists()    // true if go-process registered
c.Action("brain.recall").Exists()   // true if brain registered

// Get action metadata
def := c.Action("agentic.dispatch").Def()
// def.Description = "Dispatch a subagent to work on a task"
// def.Schema = Options with expected keys
```

This makes the capability map queryable. An agent can inspect what Actions are available before attempting to use them.

---

## 19. API — Remote Streams (Planned)

> Status: Design spec. The transport primitive for remote communication.

### 19.1 The Concept

HTTP is a stream. WebSocket is a stream. SSE is a stream. MCP over HTTP is a stream. The transport protocol is irrelevant to the consumer — you write bytes, you read bytes.

```
c.IPC()      → local conclave (in-process, same binary)
c.API()      → remote streams (cross-process, cross-machine)
c.Process()  → managed execution (via IPC Actions)
```

IPC is local. API is remote. The consumer doesn't care which one resolves their Action — if `process.run` is local it goes through IPC, if it's on Charon it goes through API. Same Action name, same Result type.

### 19.2 The Primitive

```go
// API is the Core primitive for remote communication.
// All remote transports are streams — the protocol is a detail.
type API struct {
    core    *Core
    streams map[string]*Stream
}

// Accessor on Core
func (c *Core) API() *API { return c.api }
```

### 19.3 Streams

A Stream is a named, bidirectional connection to a remote endpoint. How it connects (HTTP, WebSocket, TCP, unix socket) is configured in `c.Drive()`.

```go
// Open a stream to a named endpoint
s, err := c.API().Stream("charon")
// → looks up "charon" in c.Drive()
// → Drive has: transport="http://10.69.69.165:9101/mcp"
// → API opens HTTP connection, returns Stream

// Stream interface
type Stream interface {
    Send(data []byte) error
    Receive() ([]byte, error)
    Close() error
}
```

### 19.4 Relationship to Drive

`c.Drive()` holds the connection config. `c.API()` opens the actual streams.

```go
// Drive holds WHERE to connect
c.Drive().New(core.NewOptions(
    core.Option{Key: "name", Value: "charon"},
    core.Option{Key: "transport", Value: "http://10.69.69.165:9101/mcp"},
    core.Option{Key: "token", Value: agentToken},
))

// API handles HOW to connect
s, _ := c.API().Stream("charon")  // reads config from Drive("charon")
s.Send(payload)                    // HTTP POST under the hood
resp, _ := s.Receive()             // SSE/response parsing under the hood
```

Drive is the phone book. API is the phone.

### 19.5 Protocol Handlers

Different transports register as protocol handlers — same pattern as Actions:

```go
// In a hypothetical core/http package:
func Register(c *core.Core) core.Result {
    c.API().RegisterProtocol("http", httpStreamFactory)
    c.API().RegisterProtocol("https", httpStreamFactory)
    return core.Result{OK: true}
}

// In core/mcp:
func Register(c *core.Core) core.Result {
    c.API().RegisterProtocol("mcp", mcpStreamFactory)
    return core.Result{OK: true}
}
```

When `c.API().Stream("charon")` is called:
1. Look up "charon" in Drive → get transport URL
2. Parse protocol from URL → "http"
3. Find registered protocol handler → httpStreamFactory
4. Factory creates the Stream

No protocol handler = no capability. Same permission model as Process and Actions.

### 19.6 Remote Action Dispatch

The killer feature: Actions that transparently cross machine boundaries.

```go
// Local action — goes through IPC
c.Action("agentic.status").Run(ctx, opts)

// Remote action — goes through API
c.Action("charon:agentic.status").Run(ctx, opts)
// → splits on ":" → host="charon", action="agentic.status"
// → c.API().Stream("charon") → sends JSON-RPC call
// → remote core-agent handles it → result comes back
```

The current `dispatchRemote` function in core/agent does exactly this manually — builds MCP JSON-RPC, opens HTTP, parses SSE. With `c.API()`, it becomes one line.

### 19.7 Where This Already Exists (Partially)

The pieces are scattered across the ecosystem:

| Current | Becomes |
|---------|---------|
| `dispatchRemote` in core/agent — manual HTTP + SSE + MCP | `c.Action("charon:agentic.dispatch").Run(opts)` |
| `statusRemote` in core/agent — same manual HTTP | `c.Action("charon:agentic.status").Run(opts)` |
| `mcpInitialize` / `mcpCall` in core/agent — MCP handshake | `c.API().Stream("charon")` (MCP protocol handler) |
| `brainRecall` in core/agent — HTTP POST to brain API | `c.Action("brain.recall").Run(opts)` or `c.API().Stream("brain")` |
| Forge API calls — custom HTTP client | `c.API().Stream("forge")` |
| `DriveHandle.Transport` — stores URLs | `c.Drive()` already does this — API reads from it |

### 19.8 The Full Subsystem Map

```
c.Registry()  — universal named collection (the brick all registries use)
c.Options()   — input configuration (what was passed to New)
c.App()       — identity (name, version)
c.Config()    — runtime settings
c.Data()      — embedded assets
c.Drive()     — connection config (WHERE to reach things)
c.API()       — remote streams (HOW to reach things)
c.Fs()        — filesystem
c.Process()   — managed execution
c.Action()    — named callables (register, invoke, inspect)
c.IPC()       — local message bus (consumes Action registry)
c.Cli()       — command tree
c.Log()       — logging
c.Error()     — panic recovery
c.I18n()      — internationalisation
```

14 subsystems. `c.Registry()` is the foundation — most other subsystems build on it.

---

## 20. Registry — The Universal Collection Primitive (Planned)

> Status: Design spec. Extracts the pattern shared by 5+ existing registries.

### 20.1 The Problem

Core has multiple independent registry implementations that all do the same thing:

```
serviceRegistry  — map[string]*Service + mutex + locked
commandRegistry  — map[string]*Command + mutex
Ipc handlers     — []func + mutex
Drive            — map[string]*DriveHandle + mutex
Data             — map[string]*Embed
```

Five registries, five implementations of: named map + thread safety + optional locking.

### 20.2 The Primitive

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
c.Registry("services")              // the service registry
c.Registry("commands")              // the command tree
c.Registry("actions")               // IPC action handlers
c.Registry("drives")                // transport handles
c.Registry("data")                  // mounted filesystems
```

Cross-cutting queries become natural:

```go
c.Registry("actions").List("process.*")  // all process capabilities
c.Registry("drives").Names()              // all configured transports
c.Registry("services").Has("brain")       // is brain service loaded?
c.Registry("actions").Len()               // how many actions registered?
```

### 20.5 Typed Accessors Are Sugar

The existing subsystem accessors become typed convenience over Registry:

```go
// These are equivalent:
c.Service("brain")                         // typed sugar
c.Registry("services").Get("brain")        // universal access

c.Drive().Get("forge")                     // typed sugar
c.Registry("drives").Get("forge")          // universal access

c.Action("process.run")                    // typed sugar
c.Registry("actions").Get("process.run")   // universal access
```

The typed accessors stay — they're ergonomic and type-safe. `c.Registry()` adds the universal query layer on top.

### 20.6 Why This Matters for IPC

This resolves Issue 6 (serviceRegistry unexported) and Issue 12 (Ipc data-only struct) cleanly:

**IPC is safe to expose Actions and Handlers** because it doesn't control the write path. The Registry does:

```go
// IPC reads from the registry (safe, read-only)
c.IPC().Actions()    // reads c.Registry("actions").Names()
c.IPC().Handlers()   // reads c.Registry("handlers").Len()

// Registration goes through the primitive (controlled)
c.Action("process.run", handler)  // writes to c.Registry("actions")

// Locking goes through the primitive
c.Registry("actions").Lock()      // no more registration after startup
```

IPC is a consumer of the registry, not the owner of the data. The Registry primitive owns the data. This is the separation that makes it safe to export everything.

### 20.7 ServiceRegistry and CommandRegistry Become Exported

With `Registry[T]` as the brick:

```go
type ServiceRegistry struct {
    *Registry[*Service]
}

type CommandRegistry struct {
    *Registry[*Command]
}
```

These are now exported types — consumers can extend service management and command routing. But they can't bypass the lock because `Registry.Set()` checks `locked`. The primitive enforces the contract.

### 20.8 What This Replaces

| Current | Becomes |
|---------|---------|
| `serviceRegistry` (unexported, custom map+mutex) | `ServiceRegistry` embedding `Registry[*Service]` |
| `commandRegistry` (unexported, custom map+mutex) | `CommandRegistry` embedding `Registry[*Command]` |
| `Drive.handles` (internal map+mutex) | `Drive` embedding `Registry[*DriveHandle]` |
| `Data.mounts` (internal map) | `Data` embedding `Registry[*Embed]` |
| `Ipc.ipcHandlers` (internal slice+mutex) | `Registry[ActionHandler]` in IPC |
| 5 separate lock implementations | One `Registry.Lock()` / `Registry.Locked()` |
| `WithServiceLock()` | `c.Registry("services").Lock()` |

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
| Registry types (`serviceRegistry`) | Lets consumers extend service management |
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
core/go-cli      — CLI framework (if separated)
core/mcp         — MCP server (adds go-sdk)
core/agent       — orchestration (adds forge, yaml, mcp)
```

Each layer imports the one below. core/go imports nothing from the ecosystem — everything imports core/go.

## Known Issues

### 1. Naming Convention — UPPERCASE vs CamelCase (Resolved)

The naming convention encodes the architecture:

| Style | Meaning | Example |
|-------|---------|---------|
| `CamelCase()` | **Primitive** — the Lego brick, the building block | `c.Action("name")`, `c.Service("name")`, `c.Config()` |
| `UPPERCASE()` | **Consumer convenience** — sugar over primitives, works out of the box | `c.ACTION(msg)`, `c.QUERY(q)`, `c.PERFORM(t)` |

**Current code has this backwards.** `ACTION()` is the uppercase method but it's mapped to the raw dispatch. `Action()` is CamelCase but it's just an alias.

**Resolution:**

```go
// CamelCase = primitive (the registry, the brick)
c.Action("process.run")           // get/register a named Action
c.Action("process.run").Run(opts) // invoke by name
c.Action("process.run").Exists()  // capability check

// UPPERCASE = consumer convenience (sugar, shortcuts)
c.ACTION(msg)    // broadcast to all — sugar over c.Action("broadcast").Run()
c.QUERY(q)       // first responder — sugar over c.Action("query").Run()
c.PERFORM(t)     // execute task — sugar over c.Action("perform").Run()

// CamelCase subsystem = owns the registry
c.IPC()          // the conclave's bus — owns the Action registry
c.IPC().Actions() // all registered action names
```

The UPPERCASE methods stay for backwards compatibility and convenience — a service that just wants to broadcast uses `c.ACTION(msg)`. A service that needs to inspect capabilities or invoke by name uses `c.Action("name")`.

### 2. MustServiceFor Uses Panic

```go
func MustServiceFor[T any](c *Core, name string) T {
    panic(...)
}
```

RFC-025 says "no hidden panics." `Must` prefix signals it, but the pattern contradicts the Result philosophy. Consider deprecating in favour of `ServiceFor` + `if !ok` pattern.

### 3. Embed() Legacy Accessor

```go
func (c *Core) Embed() Result { return c.data.Get("app") }
```

Dead accessor with "use Data()" comment. Should be removed — it's API surface clutter that confuses agents.

### 4. Package-Level vs Core-Level Logging

```go
core.Info("msg")       // global default logger
c.Log().Info("msg")    // Core's logger instance
```

Both work. Global functions exist for code without Core access (early init, proc.go helpers). Services with Core access should use `c.Log()`. Document the boundary.

### 5. RegisterAction Lives in task.go

IPC registration (`RegisterAction`, `RegisterActions`, `RegisterTask`) is in `task.go` but the dispatch functions (`Action`, `Query`, `QueryAll`) are in `ipc.go`. All IPC should be in one file or the split should follow a clear boundary (dispatch vs registration).

### 6. serviceRegistry Is Unexported

`serviceRegistry` is unexported, meaning consumers can't extend service management. Per the Lego Bricks philosophy, this should be exported so downstream packages can build on it.

### 7. No c.Process() Accessor

Process management (go-process) should be a Core subsystem accessor like `c.Fs()`, not a standalone service retrieved via `ServiceFor`. Planned for go-process v0.7.0 update.

### 8. NewRuntime / NewWithFactories — Legacy

These pre-v0.7.0 functions take `app any` instead of `*Core`. `Runtime` is a separate struct from `Core` with its own `ServiceStartup`/`ServiceShutdown`. This was the original bootstrap pattern before `core.New()` + `WithService` replaced it.

**Question:** Is anything still using `NewRuntime`/`NewWithFactories`? If not, remove. If yes, migrate to `core.New()`.

### 9. CommandLifecycle — The Three-Layer CLI Architecture

```go
type CommandLifecycle interface {
    Start(Options) Result
    Stop() Result
    Restart() Result
    Reload() Result
    Signal(string) Result
}
```

Lives on `Command.Lifecycle` as an optional field. Comment says "provided by go-process" but nobody implements it yet.

**Intent:** Every CLI command can potentially be a daemon. The `Command` struct is a **primitive declaration** — it carries enough information for multiple consumers to act on it:

```
Service registers:     c.Command("serve", Command{Action: handler, Managed: "process.daemon"})
core.Cli() provides:   basic arg parsing, runs the Action
core/cli extends:      rich help, --stop/--restart/--status flags, shell completion
go-process extends:    PID file, health check, signal handling, daemon registry
```

Each layer reads the same `Command` struct. No layer modifies it. The struct IS the contract — services declare, packages consume.

**The three layers:**

| Layer | Package | Provides | Reads From |
|-------|---------|----------|------------|
| Primitive | core/go `core.Cli()` | Command tree, basic parsing, minimal runner | `Command.Action`, `Command.Path`, `Command.Flags` |
| Rich CLI | core/cli | Cobra-style help, subcommands, completion, man pages | Same `Command` struct — builds UI from declarations |
| Process | go-process | PID file, health, signals, daemon registry | `Command.Managed` field — wraps the Action in lifecycle |

This is why `CommandLifecycle` is on the struct as a field, not on Core as a method. It's data, not behaviour. The behaviour comes from whichever package reads it.

**Resolution:** Replace the `CommandLifecycle` interface with a `Managed` field:

```go
type Command struct {
    Name        string
    Description string
    Path        string
    Action      CommandAction     // the business logic
    Managed     string            // "" = one-shot, "process.daemon" = managed lifecycle
    Flags       Options
    Hidden      bool
}
```

When `Managed` is set:
- `core.Cli()` sees it's a daemon, adds basic `--stop`/`--status` flag handling
- `core/cli` adds full daemon management UI (start/stop/restart/reload/status)
- `go-process` provides the actual mechanics (PID, health, signals, registry)
- `core-agent serve` → go-process starts the Action as a daemon
- `core-agent serve --stop` → go-process sends SIGTERM via PID file

The `CommandLifecycle` interface disappears. The lifecycle verbs become process Actions (Section 18):

```
process.start    — start managed daemon
process.stop     — graceful SIGTERM → wait → SIGKILL
process.restart  — stop + start
process.reload   — SIGHUP
process.signal   — arbitrary signal
process.status   — is it running? PID? uptime?
```

Any command with `Managed: "process.daemon"` gets these for free when go-process is in the conclave.

### 10. Array[T] — Guardrail Primitive (Resolved)

`array.go` exports a generic ordered collection: `NewArray[T]`, `Add`, `AddUnique`, `Contains`, `Filter`, `Each`, `Remove`, `Deduplicate`, `Len`, `Clear`, `AsSlice`.

Currently unused by any consumer. Originally appeared speculative.

**Actual intent:** Array[T] is a **guardrail primitive** — same category as the string helpers (`core.Contains`, `core.Split`, `core.Trim`). The purpose is not capability (Go's `slices` package can do it all). The purpose is:

1. **One import, one pattern** — an agent sees `core.Array` and knows "this is how we do collections here"
2. **Attack surface reduction** — no inline `for i := range` with off-by-one bugs, no hand-rolled dedup with subtle equality issues
3. **Scannable** — `grep "Array\[" *.go` finds every collection operation
4. **Model-proof** — weaker models (Gemini, Codex) can't mess up `arr.AddUnique(item)` the way they can mess up a custom implementation. They generate inline collection ops every time because they don't recognise "this is the same operation from 3 files ago"

**The primitive taxonomy:**

| Primitive | Go Stdlib | Core Guardrail | Why Both Exist |
|-----------|-----------|---------------|----------------|
| Strings | `strings.*` | `core.Contains/Split/Trim` | Single import, scannable, model-proof |
| Paths | `filepath.*` | `core.JoinPath/PathBase` | Single import, scannable |
| Errors | `fmt.Errorf` | `core.E()` | Structured, no silent swallowing |
| Named maps | `map[string]T` | `Registry[T]` | Thread-safe, lockable, queryable |
| Ordered slices | `[]T` + `slices.*` | `Array[T]` | Dedup, unique-add, filter — one pattern |

**Resolution:** Keep Array[T]. It's the ordered counterpart to Registry[T]:
- `Registry[T]` — named collection (map), lookup by key
- `Array[T]` — ordered collection (slice), access by index + filter/each

Both are guardrail primitives that force a single codepath for common operations. Document in RFC-025 as part of AX Principle 6 (Core Primitives).

### 11. ConfigVar[T] — Typed Config with Set Tracking (Resolved)

```go
type ConfigVar[T any] struct { val T; set bool }
func (v *ConfigVar[T]) Get() T
func (v *ConfigVar[T]) Set(val T)
func (v *ConfigVar[T]) IsSet() bool
func (v *ConfigVar[T]) Unset()
```

Currently only used internally in `config.go`.

**Intent:** Distinguishes "explicitly set to false" from "never set." Essential for layered config (defaults → file → env → flags → runtime) where you need to know WHICH layer set a value, not just what the value is.

**Resolution:** Promote to a documented primitive. ConfigVar[T] solves the same guardrail problem as Array[T] — without it, every config consumer writes their own "was this set?" tracking with a separate `*bool` or sentinel values. That's exactly the kind of inline reimplementation that weaker models get wrong.

```go
// Without ConfigVar — every consumer reinvents this
var debug bool
var debugSet bool  // or *bool, or sentinel value

// With ConfigVar — one pattern
var debug core.ConfigVar[bool]
debug.Set(true)
debug.IsSet()  // true — explicitly set
debug.Unset()
debug.IsSet()  // false — reverted to unset
debug.Get()    // zero value of T
```

ConfigVar[T] is the typed counterpart to Option (which is `any`-typed). Both hold a value, but ConfigVar tracks whether it was explicitly set.

### 12. Ipc — From Data-Only Struct to Registry Owner (Resolved)

`Ipc` currently holds handler slices and mutexes but has zero methods. All IPC methods live on `*Core`. The `c.IPC()` accessor returns the raw struct with nothing useful on it.

**Resolution:** With the naming convention from Issue 1 resolved, the roles are clear:

```
c.Action("name")       — CamelCase primitive: register/invoke/inspect named Actions
c.ACTION(msg)          — UPPERCASE convenience: broadcast (sugar over primitives)
c.IPC()                — CamelCase subsystem: OWNS the Action registry
```

`c.IPC()` becomes the conclave's brain — the registry of all capabilities:

```go
// Registry inspection (on Ipc)
c.IPC().Actions()                     // []string — all registered action names
c.IPC().Action("process.run")         // *ActionDef — metadata, handler, schema
c.IPC().Handlers()                    // int — total registered handlers
c.IPC().Tasks()                       // []string — registered task flows

// Primitive operations (on Core — delegates to IPC)
c.Action("process.run", handler)      // register
c.Action("process.run").Run(opts)     // invoke
c.Action("process.run").Exists()      // check

// Consumer convenience (on Core — sugar)
c.ACTION(msg)                         // broadcast to all handlers
c.QUERY(q)                            // first responder wins
c.PERFORM(t)                          // execute task
```

Three layers, one registry:
- **`c.IPC()`** owns the data (Action registry, handler slices, task flows)
- **`c.Action()`** is the primitive API for interacting with it
- **`c.ACTION()`** is the convenience shortcut for common patterns

This is the same pattern as:
- **`c.Drive()`** owns connection config, **`c.API()`** opens streams using it
- **`c.Data()`** owns mounts, **`c.Embed()`** was the legacy shortcut

### 13. Lock() Allocates on Every Call

```go
func (c *Core) Lock(name string) *Lock {
    // ... looks up or creates mutex ...
    return &Lock{Name: name, Mutex: m}  // new struct every call
}
```

The mutex is cached and reused. The `Lock` wrapper struct is allocated fresh every call. `c.Lock("drain").Mutex.Lock()` allocates a throwaway struct just to access the mutex.

**Resolution:** Cache the `Lock` struct alongside the mutex, or return `*sync.RWMutex` directly since the `Lock` struct's only purpose is carrying the `Name` field (which the caller already knows).

### 14. Startables() / Stoppables() Return Result

```go
func (c *Core) Startables() Result  // Result{[]*Service, true}
func (c *Core) Stoppables() Result  // Result{[]*Service, true}
```

These return `Result` wrapping `[]*Service`, requiring a type assertion. They're only called by `ServiceStartup`/`ServiceShutdown` internally. Since they always succeed and always return `[]*Service`, they should return `[]*Service` directly.

**Resolution:** Change signatures to `func (c *Core) Startables() []*Service`. Or if keeping Result for consistency, document that `r.Value.([]*Service)` is the expected assertion.

### 15. contract.go Comment Says New() Returns Result

```go
//	r := core.New(...)
//	if !r.OK { log.Fatal(r.Value) }
//	c := r.Value.(*Core)
func New(opts ...CoreOption) *Core {
```

The comment shows the old API where `New()` returned `Result`. The actual signature returns `*Core` directly. This was changed during the v0.4.0 restructure but the comment wasn't updated.

**Resolution:** Fix the comment to match the signature. `New()` returns `*Core` because Core is the one type that can't wrap its own creation in `Result` (it doesn't exist yet).

### 16. task.go Mixes Concerns (Resolved)

`task.go` contains six functions that belong in two different files:

**Current task.go → splits into:**

| Function | Target File | Role | Why |
|----------|------------|------|-----|
| `RegisterAction` | `ipc.go` | IPC registry | Registers handlers in `c.IPC()`'s registry |
| `RegisterActions` | `ipc.go` | IPC registry | Batch variant of above |
| `RegisterTask` | `ipc.go` | IPC registry | Same pattern, different handler type |
| `Perform` | `action.go` (new) | Action primitive | `c.Action("name").Run()` — synchronous execution |
| `PerformAsync` | `action.go` | Action primitive | `c.Action("name").RunAsync()` — background with panic recovery + progress |
| `Progress` | `action.go` | Action primitive | Progress is per-Action, broadcasts via `c.ACTION()` |

**The file rename tells the story:** `task.go` → `action.go`. Actions are the atom (Section 18). Tasks are compositions of Actions (Section 18.6) — they get their own file when the flow system is built.

**What stays in contract.go (message types):**

```go
type ActionTaskStarted struct {
    TaskIdentifier string
    Task           Task
}

type ActionTaskProgress struct {
    TaskIdentifier string
    Task           Task
    Progress       float64
    Message        string
}

type ActionTaskCompleted struct {
    TaskIdentifier string
    Task           Task
    Result         any
    Error          error
}
```

These names are already correct — they're `ACTION` messages (broadcast events) about Task lifecycle. The naming convention from Issue 1 validates them: `Action` prefix = it's a broadcast message type. `Task` in the name = it's about task lifecycle. No rename needed.

**The semantic clarity after the split:**

```
ipc.go      — registry: where handlers are stored
action.go   — execution: where Actions run (sync, async, progress)
contract.go — types: message definitions, interfaces, options
```

Registration is IPC's job. Execution is Action's job. Types are shared contracts. Three files, three concerns, zero overlap.

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

## Changelog

- 2026-03-25: Added Section 20 — Registry universal collection primitive. Resolved Issues 1, 6, 9, 12, 16. Updated subsystem map to 14.
- 2026-03-25: Added Section 19 — API/Stream remote transport primitive
- 2026-03-25: Added Known Issues 9-16 (ADHD brain dump recovery — CommandLifecycle, Array[T], ConfigVar[T], Ipc struct, Lock allocation, Startables/Stoppables, stale comment, task.go concerns)
- 2026-03-25: Added Section 18 — Action and Task execution primitives
- 2026-03-25: Added Section 17 — c.Process() primitive spec
- 2026-03-25: Added Design Philosophy + Known Issues 1-8
- 2026-03-25: Initial specification — matches v0.7.0 implementation
