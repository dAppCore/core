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

**Implemented 2026-03-25.** `c.Action("name")` is the named action primitive. `c.ACTION(msg)` calls internal `broadcast()`. The rename from `Action(msg)` → `broadcast(msg)` is done.

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

### 2. MustServiceFor Uses Panic (Resolved)

```go
func MustServiceFor[T any](c *Core, name string) T {
    panic(...)
}
```

**Resolution:** Keep but document clearly. `Must` prefix is a Go convention that signals "this panics." In the guardrail context (Issue 10/11), `MustServiceFor` is valid for startup-time code where a missing service means the app can't function. The alternative — `ServiceFor` + `if !ok` + manual error handling — adds LOC that contradicts the Result philosophy.

**Rule:** Use `MustServiceFor` only in `OnStartup` / init paths where failure is fatal. Never in request handlers or runtime code. Document this in RFC-025.

### 3. Embed() Legacy Accessor (Resolved — Removed 2026-03-25)

**Removed.** `c.Embed()` deleted from core.go. Use `c.Data().Get("app")` instead.

### 4. Package-Level vs Core-Level Logging (Resolved)

```go
core.Info("msg")       // global default logger — no Core needed
c.Log().Info("msg")    // Core's logger instance
```

**Resolution:** Both stay. Document the boundary:

| Context | Use | Why |
|---------|-----|-----|
| Has `*Core` (services, handlers) | `c.Log().Info()` | Logger may be configured per-Core |
| No `*Core` (init, package-level, helpers) | `core.Info()` | Global logger, always available |
| Test code | `core.Info()` | Tests may not have Core |

This is the same dual pattern as `process.Run()` (global) vs `c.Process().Run()` (Core). Package-level functions are the bootstrap path. Core methods are the runtime path.

### 5. RegisterAction Lives in task.go (Resolved — Implemented 2026-03-25)

**Done.** RegisterAction/RegisterActions/RegisterTask moved to ipc.go. action.go has ActionDef, TaskDef, execution. task.go has PerformAsync/Progress.

### 6. serviceRegistry Is Unexported (Resolved — Implemented 2026-03-25)

**Done.** All 5 registries migrated to `Registry[T]`:
- `serviceRegistry` → `ServiceRegistry` embedding `Registry[*Service]`
- `commandRegistry` → `CommandRegistry` embedding `Registry[*Command]`
- `Drive` embedding `Registry[*DriveHandle]`
- `Data` embedding `Registry[*Embed]`
- `Lock.locks` using `Registry[*sync.RWMutex]`

### 7. No c.Process() Accessor (Resolved — Implemented 2026-03-25)

**Done.** `c.Process()` returns `*Process` — sugar over `c.Action("process.run")`. No deps added to core/go. go-process v0.7.0 registers the actual handlers.

### 8. NewRuntime / NewWithFactories — GUI Bridge, Not Legacy (Resolved)

NOT dead code. `Runtime` is the **GUI binding container** — it bridges Core to frontend frameworks like Wails. `App.Runtime` holds the Wails app reference (`any`). This is the core-webview-bridge: CoreGO exported methods → Wails WebView2 → CoreTS fronts them.

```go
type Runtime struct {
    app  any    // Wails app or equivalent
    Core *Core  // the Core instance
}
```

**Issue:** `NewWithFactories` uses the old factory pattern (`func() Result` instead of `func(*Core) Result`). Factories don't receive Core, so they can't use DI during construction.

**Resolution:** Update `NewWithFactories` to accept `func(*Core) Result` factories (same as `WithService`). The `Runtime` struct stays — it's the GUI bridge, not a replacement for Core. Consider adding `core.WithRuntime(app)` as a CoreOption:

```go
// Current:
r := core.NewWithFactories(wailsApp, factories)

// Target:
c := core.New(
    core.WithRuntime(wailsApp),
    core.WithService(display.Register),
    core.WithService(gui.Register),
)
```

This unifies CLI and GUI bootstrap — same `core.New()`, just with `WithRuntime` added.

### 9. CommandLifecycle — The Three-Layer CLI Architecture (Resolved — Implemented 2026-03-25)

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

### 13. Lock() Allocates on Every Call (Resolved)

```go
func (c *Core) Lock(name string) *Lock {
    return &Lock{Name: name, Mutex: m}  // new struct every call
}
```

The mutex is cached. The `Lock` wrapper is not.

**Resolution:** With `Registry[T]` (Section 20), Lock becomes a Registry:

```go
// Lock registry becomes:
type LockRegistry struct {
    *Registry[*sync.RWMutex]
}

// Usage stays the same but no allocation:
c.Lock("drain")  // returns cached *Lock, not new allocation
```

The `Lock` struct can cache itself in the registry alongside the mutex. Or simpler: `c.Lock("drain")` returns `*sync.RWMutex` directly — the caller already knows the name.

### 14. Startables() / Stoppables() Return Result (Resolved)

```go
func (c *Core) Startables() Result  // Result{[]*Service, true}
func (c *Core) Stoppables() Result  // Result{[]*Service, true}
```

**Resolution:** With `Registry[T]` and `ServiceRegistry`, these become registry queries:

```go
// Current — type assertion required
startables := c.Startables()
for _, s := range startables.Value.([]*Service) { ... }

// Target — registry filter
for _, s := range c.Registry("services").Each(func(name string, svc *Service) bool {
    return svc.OnStart != nil
}) { ... }
```

Or simpler: change return type to `[]*Service` directly. These are internal — only `ServiceStartup`/`ServiceShutdown` call them. No need for Result wrapping.

### 15. contract.go Comment Says New() Returns Result (Resolved)

```go
//	r := core.New(...)         // WRONG — stale comment
//	if !r.OK { ... }           // WRONG
//	c := r.Value.(*Core)       // WRONG
func New(opts ...CoreOption) *Core {
```

**Resolution:** Fix the comment. Simple mechanical fix:

```go
//	c := core.New(
//	    core.WithOption("name", "myapp"),
//	    core.WithService(mypackage.Register),
//	)
//	c.Run()
func New(opts ...CoreOption) *Core {
```

`New()` returns `*Core` directly — it's the one constructor that can't wrap its own creation error in Result.

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

## Pass Two — Architectural Audit

> Second review of the spec against the actual codebase. Looking for
> anti-patterns, security concerns, and unexported fields that indicate
> unfinished design.

### P2-1. Core Struct Is Fully Unexported — Contradicts Lego Bricks

Every field on `Core` is unexported. The Design Philosophy says "exported fields are intentional" but the most important type in the system hides everything behind accessors.

```go
type Core struct {
    options  *Options          // unexported
    services *serviceRegistry  // unexported
    commands *commandRegistry  // unexported
    ipc      *Ipc             // unexported
    // ... all 15 fields unexported
}
```

**Split:** Some fields SHOULD be exported (registries, subsystems — they're bricks). Some MUST stay unexported (lifecycle internals — they're safety).

| Should Export | Why |
|--------------|-----|
| `Services *ServiceRegistry` | Downstream extends service management |
| `Commands *CommandRegistry` | Downstream extends command routing |
| `IPC *Ipc` | Inspection, capability queries |
| `Data *Data` | Mount inspection |
| `Drive *Drive` | Transport inspection |

| Must Stay Unexported | Why |
|---------------------|-----|
| `context` / `cancel` | Lifecycle is Core's responsibility — exposing cancel is dangerous |
| `waitGroup` | Concurrency is Core's responsibility |
| `shutdown` | Shutdown state must be atomic and Core-controlled |
| `taskIDCounter` | Implementation detail |

**Rule:** Export the bricks. Hide the safety mechanisms.

### P2-2. Fs.root Is Unexported — Correctly

```go
type Fs struct {
    root string  // the sandbox boundary
}
```

This is the ONE field that's correctly unexported. `root` controls path validation — if exported, any consumer bypasses sandboxing by setting `root = "/"`. Security boundaries are the exception to Lego Bricks.

**Rule:** Security boundaries stay unexported. Everything else exports.

### P2-3. Config.Settings Is `map[string]any` — Untyped Bag

```go
type ConfigOptions struct {
    Settings map[string]any   // anything goes
    Features map[string]bool  // at least typed
}
```

`Settings` is a raw untyped map. Any code can stuff anything in. No validation, no schema, no way to know what keys are valid. `ConfigVar[T]` (Issue 11) was designed to fix this but is unused.

**Resolution:** `Settings` should use `Registry[ConfigVar[any]]` — each setting tracked with set/unset state. Or at minimum, `c.Config().Set()` should validate against a declared schema.

The Features map is better — `map[string]bool` is at least typed. But it's still a raw map with no declared feature list.

### P2-4. Global `assetGroups` — State Outside the Conclave

```go
var (
    assetGroups   = make(map[string]*AssetGroup)
    assetGroupsMu sync.RWMutex
)
```

Package-level mutable state with its own mutex. `AddAsset()` and `GetAsset()` work without a Core reference. This bypasses the conclave — there's no permission check, no lifecycle, no IPC. Any imported package's `init()` can modify this.

**Intent:** Generated code from `GeneratePack` calls `AddAsset()` in `init()` — before Core exists.

**Tension:** This is the bootstrap problem. Assets must be available before Core is created because `WithService` factories may need them during `New()`. But global state outside Core is an anti-pattern.

**Resolution:** Accept this as a pre-Core bootstrap layer. Document that `AddAsset`/`GetAsset` are init-time only — after `core.New()`, all asset access goes through `c.Data()`. Consider `c.Data().Import()` to move global assets into Core's registry during construction.

### P2-5. SysInfo Frozen at init() — Untestable

```go
var systemInfo = &SysInfo{values: make(map[string]string)}

func init() {
    systemInfo.values["DIR_HOME"] = homeDir
    // ... populated once, never updated
}
```

`core.Env("DIR_HOME")` returns the init-time value. `t.Setenv("DIR_HOME", temp)` has no effect because the map was populated before the test ran. We hit this exact bug repeatedly in core/agent testing.

**Resolution:** `core.Env()` should check the live `os.Getenv()` as fallback when the cached value doesn't match. Or provide `core.EnvRefresh()` for test contexts. The cached values are a performance optimisation — they shouldn't override actual environment changes.

```go
func Env(key string) string {
    if v := systemInfo.values[key]; v != "" {
        return v
    }
    return os.Getenv(key)  // already does this — but cached value wins
}
```

The issue is that cached values like `DIR_HOME` are set to the REAL home dir at init. `os.Getenv("DIR_HOME")` returns the test override, but `systemInfo.values["DIR_HOME"]` returns the cached original. The cache should only hold values that DON'T exist in os env — computed values like `PID`, `NUM_CPU`, `OS`, `ARCH`.

### P2-6. ErrorPanic.onCrash Is Unexported

```go
type ErrorPanic struct {
    filePath string
    meta     map[string]string
    onCrash  func(CrashReport)  // hidden callback
}
```

The crash callback can't be set by downstream packages. If core/agent wants to send crash reports to OpenBrain, or if a monitoring service wants to capture panics, they can't wire into this.

**Resolution:** Export `OnCrash` or provide `c.Error().SetCrashHandler(fn)`. Crash handling is sensitive but it's also a cross-cutting concern that monitoring services need access to.

### P2-7. Data.mounts Is Unexported — Can't Iterate

```go
type Data struct {
    mounts map[string]*Embed
    mu     sync.RWMutex
}
```

`c.Data().Mounts()` returns names (`[]string`) but you can't iterate the actual `*Embed` objects. With `Registry[T]` (Section 20), Data should embed `Registry[*Embed]` and expose `c.Data().Each()`.

**Resolution:** Data becomes:
```go
type Data struct {
    *Registry[*Embed]
}
```

### P2-8. Logging Timing Gap — Bootstrap vs Runtime

```go
// During init/startup — before Core exists:
core.Info("starting up")       // global default logger — unconfigured

// After core.New() — Core configures its logger:
c.Log().Info("running")        // Core's logger — configured
```

Between program start and `core.New()`, log messages go to the unconfigured global logger. After `core.New()`, they go to Core's configured logger. The transition is invisible — no warning, no redirect.

**Resolution:** Document the boundary clearly. Consider `core.New()` auto-redirecting the global logger to Core's logger:

```go
func New(opts ...CoreOption) *Core {
    c := &Core{...}
    // ... apply options ...

    // Redirect global logger to Core's logger
    SetDefault(c.log.logger())

    return c
}
```

After `New()`, `core.Info()` and `c.Log().Info()` go to the same destination. The timing gap still exists during construction, but it closes as soon as Core is built.

---

## Pass Three — Spec Contradictions

> Third review. The structural issues are documented, the architectural concerns
> are documented. Now looking for contradictions within the spec itself — places
> where design decisions fight each other.

### P3-1. Startable/Stoppable Return `error`, Everything Else Returns `Result`

The lifecycle interfaces are the only Core contracts that return Go's `error`:

```go
type Startable interface { OnStartup(ctx context.Context) error }
type Stoppable interface { OnShutdown(ctx context.Context) error }
```

But every other Core contract uses `Result`: `WithService` factories, `RegisterAction` handlers, `QueryHandler`, `TaskHandler`. This forces wrapping at the boundary:

```go
// Inside RegisterService — wrapping error into Result
srv.OnStart = func() Result {
    if err := s.OnStartup(c.context); err != nil {
        return Result{err, false}
    }
    return Result{OK: true}
}
```

**Resolution:** Change lifecycle interfaces to return `Result`:

```go
type Startable interface { OnStartup(ctx context.Context) Result }
type Stoppable interface { OnShutdown(ctx context.Context) Result }
```

This is a breaking change but it unifies the contract. Every service author already returns `Result` from their factory — the lifecycle should match. The `error` return was inherited from Go convention, not from Core's design.

### P3-2. Section 17 Process Returns `(string, error)`, Section 18 Action Returns `Result`

Section 17 spec'd:
```go
func (p *Process) Run(ctx, cmd, args) (string, error)
```

Section 18 spec'd:
```go
func (a *ActionDef) Run(ctx, opts) Result
```

Section 18.9 says "Process IS an Action under the hood." But they have different return types. If Process is sugar over Actions, it should return Result:

```go
// Consistent:
r := c.Process().Run(ctx, "git", "log")
if r.OK { output := r.Value.(string) }

// Or convenience accessor:
output := c.Process().RunString(ctx, "git", "log")  // returns string, logs error
```

**Resolution:** Process methods return `Result` for consistency. Add typed convenience methods (`RunString`, `RunOK`) that unwrap Result for common patterns — same as `Options.String()` unwraps `Options.Get()`.

### P3-3. Three Patterns for "Get a Value"

```go
Options.Get("key")         → Result        // check r.OK
Options.String("key")      → string        // zero value on miss
ServiceFor[T](c, "name")  → (T, bool)     // Go tuple
Drive.Get("name")          → Result
```

Three different patterns. An agent must learn all three.

**Resolution:** `Registry[T]` unifies the base: `Registry.Get()` returns `Result`. Typed convenience accessors (`String()`, `Int()`, `Bool()`) return zero values — same as Options already does. `ServiceFor[T]` returns `(T, bool)` because generics can't return through `Result.Value.(T)` ergonomically.

The three patterns are actually two:
- `Get()` → Result (when you need to check existence)
- `String()`/`Int()`/typed accessor → zero value (when missing is OK)

`ServiceFor` is the generic edge case. Accept it.

### P3-4. Dual-Purpose Methods Are Anti-AX

```go
c.Service("name")       // GET — returns Result with service
c.Service("name", svc)  // SET — registers service
```

Same method, different behaviour based on arg count. An agent reading `c.Service("auth")` can't tell if it's a read or write without checking the arg count.

**Resolution:** Keep dual-purpose on Core as sugar (it's ergonomic). But the underlying `Registry[T]` has explicit verbs:

```go
// Registry — explicit, no ambiguity
c.Registry("services").Get("auth")        // always read
c.Registry("services").Set("auth", svc)   // always write

// Core sugar — dual-purpose, ergonomic
c.Service("auth")                          // read (no second arg)
c.Service("auth", svc)                     // write (second arg present)
```

Document the dual-purpose pattern. Agents that need clarity use Registry. Agents that want brevity use the sugar.

### P3-5. `error` Leaks Through Despite Result

Core says "use Result, not (val, error)." But `error` appears at multiple boundaries:

| Where | Returns | Why |
|-------|---------|-----|
| `Startable.OnStartup` | `error` | Go interface convention |
| `core.E()` | `error` | It IS the error constructor |
| `embed.go` internals | `error` | stdlib fs.FS requires it |
| `ServiceStartup` wrapping | `error` → `Result` | Bridging two worlds |

**Resolution:** Accept the boundary. `error` exists at the Go stdlib interface level. `Result` exists at the Core contract level. The rule is:

- **Implementing a Go interface** (fs.FS, io.Reader, etc.) → return `error`
- **Implementing a Core contract** (factory, handler, lifecycle) → return `Result`
- **`core.E()`** returns `error` because it creates errors — it's a constructor, not a contract

With P3-1 resolved (Startable/Stoppable return Result), the only `error` returns in Core contracts are the stdlib interface implementations.

### P3-6. Data Has Overlapping APIs

```go
c.Data().Get("prompts")           // returns *Embed (the mount itself)
c.Data().ReadString("prompts/x")  // reads a file through the mount
c.Data().ReadFile("prompts/x")    // same but returns []byte
c.Data().List("prompts/")         // lists files in mount
```

Two levels of abstraction on the same type: mount management (`Get`, `New`, `Mounts`) and file access (`ReadString`, `ReadFile`, `List`).

**Resolution:** With `Registry[*Embed]`, Data splits cleanly:

```go
// Mount management — via Registry
c.Data().Get("prompts")         // Registry[*Embed].Get() → the mount
c.Data().Set("prompts", embed)  // Registry[*Embed].Set() → register mount
c.Data().Names()                // Registry[*Embed].Names() → all mounts

// File access — on the Embed itself
embed := c.Data().Get("prompts").Value.(*Embed)
embed.ReadString("coding.md")   // read from this specific mount
```

The current convenience methods (`c.Data().ReadString("prompts/coding.md")`) can stay as sugar — they resolve the mount from the path prefix then delegate. But the separation is clear: Data manages mounts, Embed reads files.

### P3-7. Action System Has No Error Propagation Model

Section 18 defines `c.Action("name").Run(opts) → Result`. But:

- What if the handler panics? `PerformAsync` has `defer recover()`. Does `Run`?
- What about timeouts? `ctx` is passed but no deadline enforcement.
- What about retries? A failed Action just returns `Result{OK: false}`.

**Resolution:** Actions inherit the safety model from `PerformAsync`:

```go
// Action.Run always wraps in panic recovery:
func (a *ActionDef) Run(ctx context.Context, opts Options) Result {
    defer func() {
        if r := recover(); r != nil {
            // capture, log, return Result{panic, false}
        }
    }()
    return a.Handler(ctx, opts)
}
```

Timeouts come from `ctx` — the caller sets the deadline, the Action respects it. Retries are a Task concern (Section 18.6), not an Action concern — a Task can define `Retry: 3` on a step.

### P3-8. Registry.Lock() Is a One-Way Door

`Registry.Lock()` prevents `Set()` permanently. No `Unlock()`. This was intentional for ServiceLock — once startup completes, no more service registration.

But with Actions and hot-reload:
- A service updates its Action handler after config change
- A plugin system loads/unloads Actions at runtime
- A test wants to replace an Action with a mock

**Resolution:** Registry supports three lock modes:

```go
r.Lock()          // no new keys, no updates — fully frozen (ServiceLock)
r.Seal()          // no new keys, but existing keys CAN be updated (Action hot-reload)
r.Open()          // (default) anything goes
```

`Seal()` is the middle ground — the registry shape is fixed (no new capabilities appear) but implementations can change (handler updated, service replaced). This supports hot-reload without opening the door to uncontrolled registration.

---

## Pass Four — Concurrency and Performance

> Fourth review. Structural, architectural, and spec contradictions are documented.
> Now looking at concurrency edge cases and performance implications.

### P4-1. ServiceStartup Order Is Non-Deterministic

`Startables()` iterates `map[string]*Service` — Go map iteration is random. If service B depends on service A being started first, it works sometimes and fails sometimes.

**Resolution:** `Registry[T]` should maintain insertion order (use a slice alongside the map). Services start in registration order — the order they appear in `core.New()`. This is predictable and matches the programmer's intent.

### P4-2. ACTION Dispatch Is Synchronous and Blocking

`c.ACTION(msg)` clones the handler slice then calls every handler sequentially. A slow handler blocks all subsequent handlers. No timeouts, no parallelism.

**Resolution:** Document that ACTION handlers MUST be fast. Long work should be dispatched via `PerformAsync`. Consider a timeout per handler or parallel dispatch option for fire-and-forget broadcasts.

### P4-3. ACTION Handler Returning !OK Stops the Chain

```go
for _, h := range handlers {
    if r := h(c, msg); !r.OK {
        return r  // remaining handlers never called
    }
}
```

For "fire-and-forget" broadcast, one failing handler silencing the rest is surprising.

**Resolution:** ACTION (broadcast) should call ALL handlers regardless of individual results. Collect errors, log them, but don't stop. QUERY (first responder) correctly stops on first OK. PERFORM (first executor) correctly stops on first OK. Only ACTION has the wrong stop condition.

### P4-4. IPC Clone-and-Iterate Is Safe but Undocumented

Handlers are cloned before iteration, lock released before calling. A handler CAN call `RegisterAction` without deadlocking. But newly registered handlers aren't called for the current message — only the next dispatch.

**Resolution:** Document this behaviour. It's correct but surprising. "Handlers registered during dispatch are visible on the next dispatch, not the current one."

### P4-5. No Backpressure on PerformAsync

`PerformAsync` spawns unlimited goroutines via `waitGroup.Go()`. 1000 concurrent calls = 1000 goroutines. No pool, no queue, no backpressure.

**Resolution:** `PerformAsync` should respect a configurable concurrency limit. With the Action system (Section 18), this becomes a property of the Action: `c.Action("heavy.work").MaxConcurrency(10)`. The registry tracks running count.

### P4-6. ConfigVar.Set() Has No Lock

```go
func (v *ConfigVar[T]) Set(val T) { v.val = val; v.set = true }
```

Two goroutines setting the same ConfigVar is a data race. `Config.Set()` holds a mutex, but `ConfigVar.Set()` does not. If a consumer holds a reference to a ConfigVar and sets it from multiple goroutines, undefined behaviour.

**Resolution:** ConfigVar should be atomic or protected by Config's mutex. Since ConfigVar is meant to be accessed through `Config.Get/Set`, the lock should be on Config. Direct ConfigVar manipulation should be documented as single-goroutine only.

### P4-7. PerformAsync Has a Shutdown Race (TOCTOU)

```go
func (c *Core) PerformAsync(t Task) Result {
    if c.shutdown.Load() { return Result{} }  // check
    // ... time passes ...
    c.waitGroup.Go(func() { ... })             // act
}
```

Between the check and the `Go()`, `ServiceShutdown` can set `shutdown=true` and start `waitGroup.Wait()`. The goroutine spawns after shutdown begins.

**Resolution:** Go 1.26's `WaitGroup.Go()` may handle this (it calls `Add(1)` internally before spawning). Verify. If not safe, use a mutex around the shutdown check + goroutine spawn.

### P4-8. Named Lock "srv" Is Shared Across All Service Operations

All service operations lock on `c.Lock("srv")` — registration, lookup, listing, startables, stoppables. A write (registration) blocks all reads (lookups). The `RWMutex` helps (reads don't block reads) but there's no isolation between subsystems.

**Resolution:** With `Registry[T]`, each registry has its own mutex. `ServiceRegistry` has its own lock, `CommandRegistry` has its own lock. No cross-subsystem blocking. This is already the plan — Registry[T] includes its own `sync.RWMutex`.

---

## Pass Four (Revisited) — Concurrency Deep Dive

> Re-examining concurrency with 9 additional passes of context.
> Looking at cross-subsystem races that P4 couldn't see.

### P4-9. status.json Has 51 Read-Modify-Write Sites — No Locking

`ReadStatus()` and `writeStatus()` are called from 51 sites across core/agent. They read a JSON file, modify the struct, and write it back. No mutex. No file lock. No atomic swap.

```go
// Goroutine A (onAgentComplete):
st, _ := ReadStatus(wsDir)   // reads status.json
st.Status = "completed"
writeStatus(wsDir, st)        // writes status.json

// Goroutine B (status MCP tool, concurrent):
st, _ := ReadStatus(wsDir)   // reads status.json — may see partial write
st.Status = "blocked"         // overwrites A's "completed"
writeStatus(wsDir, st)        // last write wins
```

Concurrent callers:
- **spawnAgent goroutine** — onAgentComplete writes final status
- **status MCP tool** — detects dead PIDs, writes status corrections
- **drainOne** — writes "running" when spawning queued agent
- **shutdownNow** — writes "failed" for killed processes
- **resume** — writes "running" when relaunching

All on different goroutines. All targeting the same file. Classic TOCTOU.

**Impact:** Status corruption. An agent marked "completed" by the goroutine gets overwritten to "blocked" by the status tool. Or vice versa. The queue sees the wrong status and makes wrong decisions.

**Resolution:** Per-workspace mutex in the workspace status manager. Or atomic file writes (write to temp, rename). Or move status into Core's `Registry[*WorkspaceStatus]` where the registry mutex protects access.

### P4-10. Fs.Write Uses os.WriteFile — Not Atomic

```go
func (m *Fs) WriteMode(p, content string, mode os.FileMode) Result {
    // os.WriteFile truncates then writes
    return Result{}.New(os.WriteFile(full, []byte(content), mode))
}
```

`os.WriteFile` truncates the file to zero length, then writes content. A concurrent reader during the truncate-to-write window sees an empty file. `ReadStatus` during that window returns an unmarshal error — the file exists but is empty.

This is P4-9's underlying cause. Even with a mutex on ReadStatus/writeStatus, the Fs primitive itself isn't atomic.

**Resolution:** Fs needs an atomic write option:

```go
// Atomic: write to temp file, then rename (rename is atomic on POSIX)
fs.WriteAtomic(path, content)

// Implementation:
tmp := path + ".tmp"
os.WriteFile(tmp, content, mode)
os.Rename(tmp, path)  // atomic on same filesystem
```

### P4-11. Config.Set and Config.Get Can Interleave Across Goroutines

P4-6 found ConfigVar.Set has no lock. But the deeper issue:

```go
// Goroutine A:
c.Config().Set("agents.concurrency", newMap)

// Goroutine B (same moment):
val := c.Config().Get("agents.concurrency")
// may see partial map — some keys updated, some not
```

Config.Set holds the mutex during Set. Config.Get holds it during Get. But the VALUE is `any` — if it's a map, the map itself isn't copied. Both goroutines hold a reference to the same map. Mutations to the map after Set aren't protected.

**Resolution:** Config.Set should deep-copy map values. Or document: "Config values must be immutable after Set. Never mutate a map retrieved from Config."

### P4-12. The Global Logger Race Window

P2-8 noted the logging timing gap. The concurrency angle:

```go
// main goroutine:
c := core.New(...)  // eventually calls SetDefault(logger)

// init() goroutine or early startup:
core.Info("starting")  // uses defaultLogPtr.Load() — may be nil or stale
```

`defaultLogPtr` is `atomic.Pointer[Log]` — the Load/Store is safe. But `core.Info()` calls `Default().Info()`. If `Default()` returns nil (before any logger is set), it panics.

```go
func Default() *Log {
    if l := defaultLogPtr.Load(); l != nil {
        return l
    }
    // falls through to create default — but this creates a NEW default every time
    // until SetDefault is called
}
```

Actually let me check:

```go
func Default() *Log {
```

The real question is whether Default() handles nil safely.

**Resolution:** Verify Default() is nil-safe. If it creates a fallback logger on first call, the race is benign (just uses the fallback until Core sets the real one). If it returns nil, it's a panic.

---

## Pass Five — Consumer Experience

> Fifth review. Looking at Core from the outside — what does a service author
> actually experience? Where does the API confuse, contradict, or surprise?

### P5-1. ServiceRuntime Is Not Used by core/agent — Two Registration Camps

The RFC documents `ServiceRuntime[T]` as THE pattern. But core/agent's three main services don't use it:

```go
// core/agent (agentic, brain, monitor) — manual pattern:
func Register(c *core.Core) core.Result {
    svc := NewPrep()
    svc.core = c     // manual Core assignment
    return core.Result{Value: svc, OK: true}
}

// core/gui (display, window, menu, etc.) — ServiceRuntime pattern:
func Register(c *core.Core) core.Result {
    svc := &DisplayService{
        ServiceRuntime: core.NewServiceRuntime(c, DisplayOpts{}),
    }
    return core.Result{Value: svc, OK: true}
}
```

Two camps exist: CLI services set `.core` manually. GUI services embed `ServiceRuntime[T]`. Both work. Neither is wrong. But the RFC only documents one pattern.

**Resolution:** Document both patterns. `ServiceRuntime[T]` is for services that need typed options. Manual `.core = c` is for services that manage their own config. Neither is preferred — use whichever fits.

### P5-2. Register Returns Result but OnStartup Returns error — P3-1 Impact on Consumer

From a consumer's perspective, they write two functions with different return types for the same service:

```go
func Register(c *core.Core) core.Result { ... }      // Result
func (s *Svc) OnStartup(ctx context.Context) error { ... }  // error
```

This is P3-1 from the consumer's side. The cognitive load is: "my factory returns Result but my lifecycle returns error." Every service author encounters this inconsistency.

### P5-3. No Way to Declare Service Dependencies

Services register independently. If brain depends on process being started first, there's no way to express that:

```go
core.New(
    core.WithService(process.Register),  // must be first?
    core.WithService(brain.Register),    // depends on process?
    core.WithService(agentic.Register),  // depends on brain + process?
)
```

The order in `core.New()` is the implicit dependency graph. But P4-1 says startup order is non-deterministic (map iteration). So even if you order them in `New()`, they might start in different order.

**Resolution:** With the Action system (Section 18), dependencies become capability checks:

```go
func (s *Brain) OnStartup(ctx context.Context) error {
    if !s.Core().Action("process.run").Exists() {
        return core.E("brain", "requires process service", nil)
    }
    // safe to use process
}
```

Or with Registry (Section 20), declare dependencies:

```go
core.WithService(brain.Register).DependsOn("process")
```

### P5-4. HandleIPCEvents Is Auto-Discovered via Reflect — Magic

```go
if handler, ok := instance.(interface {
    HandleIPCEvents(*Core, Message) Result
}); ok {
    c.ipc.ipcHandlers = append(c.ipc.ipcHandlers, handler.HandleIPCEvents)
}
```

If your service has a method called `HandleIPCEvents` with exactly the right signature, it's automatically registered as an IPC handler. No explicit opt-in. A service author might not realise their method is being called for EVERY IPC message.

**Resolution:** Document this clearly. Or replace with explicit registration during `OnStartup`:

```go
func (s *Svc) OnStartup(ctx context.Context) error {
    s.Core().RegisterAction(s.handleEvents)  // explicit
    return nil
}
```

Auto-discovery is convenient but anti-AX — an agent can't tell from reading the code that `HandleIPCEvents` is special. There's no annotation, no interface declaration, just a magic method name.

### P5-5. Commands Are Registered During OnStartup — Invisible Dependency

```go
func (s *PrepSubsystem) OnStartup(ctx context.Context) error {
    s.registerCommands(ctx)       // registers CLI commands
    s.registerForgeCommands()     // more commands
    s.registerWorkspaceCommands() // more commands
    return nil
}
```

Commands are registered inside OnStartup, not during factory construction. This means:
- Commands don't exist until ServiceStartup runs
- `c.Commands()` returns empty before startup
- If startup fails partway, some commands exist and some don't

**Resolution:** Consider allowing command registration during factory (before startup). Or document that commands are only available after `ServiceStartup` completes. With Actions (Section 18), commands ARE Actions — registered the same way, available the same time.

### P5-6. No Service Discovery — Only Lookup by Name

```go
svc, ok := core.ServiceFor[*Brain](c, "brain")
```

You must know the service name AND type. There's no:
- "give me all services that implement interface X"
- "give me all services in the 'monitoring' category"
- "what services are registered?"

`c.Services()` returns names but not types or capabilities.

**Resolution:** With `Registry[T]` and the Action system, this becomes capability-based:

```go
c.Registry("actions").List("monitor.*")   // all monitoring capabilities
c.Registry("services").Each(func(name string, svc *Service) {
    // iterate all services
})
```

### P5-7. Factory Receives *Core But Can't Tell If Other Services Exist

During `WithService` execution, factories run in order. Factory #3 can see services #1 and #2 (they're registered). But it can't safely USE them because `OnStartup` hasn't run yet.

```go
// In factory — service exists but isn't started:
func Register(c *core.Core) core.Result {
    brain, ok := core.ServiceFor[*Brain](c, "brain")
    // ok=true — brain's factory already ran
    // but brain.OnStartup() hasn't — brain isn't ready
}
```

**Resolution:** Factories should only create and return. All inter-service communication happens after startup, via IPC. Document: "factories create, OnStartup connects."

### P5-8. The MCP Subsystem Pattern (core/mcp) — Not in the RFC

core/mcp's `Register` function discovers ALL other services to build the MCP tool list:

```go
func Register(c *core.Core) core.Result {
    // discovers agentic, brain, monitor subsystems
    // registers their tools with the MCP server
}
```

This is a cross-cutting service that consumes other services' capabilities. It's a real pattern but not documented in the RFC. With Actions, it becomes: "MCP service lists all registered Actions and exposes them as MCP tools."

**Resolution:** Document the "aggregator service" pattern — a service that reads from `c.Registry("actions")` and builds an external interface (MCP, REST, CLI) from it. This is the bridge between Core's internal Action registry and external protocols.

---

## Pass Six — Cross-Repo Patterns and Cascade Analysis

> Sixth review. Tracing how real services compose with Core across repos.
> Following the actual execution path of a single event through the system.

### P6-1. ACTION Cascade Is Synchronous, Nested, and Unbounded

When `AgentCompleted` fires, it triggers a synchronous cascade 4 levels deep:

```
c.ACTION(AgentCompleted)
  → QA handler: runs build+test (30+ seconds), then:
    c.ACTION(QAResult)
      → PR handler: git push + Forge API (10+ seconds), then:
        c.ACTION(PRCreated)
          → Verify handler: test + merge API (20+ seconds), then:
            c.ACTION(PRMerged)
              → all 5 handlers called again (type-check, skip)
          → remaining handlers called
      → remaining handlers called
  → Ingest handler: runs for AgentCompleted
  → Poke handler: runs for AgentCompleted
```

**Problems:**
1. **Blocking** — the entire cascade runs on ONE goroutine. If QA takes 30s and merge takes 20s, the Poke handler doesn't fire for 50+ seconds.
2. **No timeout** — if the Forge merge API hangs, everything blocks indefinitely.
3. **Nested depth** — 4 levels of `c.ACTION()` inside `c.ACTION()`. Stack grows linearly.
4. **Handler fanout** — 5 handlers × 4 nested broadcasts ≈ 30+ handler invocations for one event.
5. **Queue starvation** — Poke (which drains the queue) can't fire until the entire pipeline completes. Other agent completions wait behind this one.

**This explains observed behaviour:** sometimes agents complete but the queue doesn't drain for minutes. The Poke handler is blocked behind a slow Forge API call 3 levels deep in the cascade.

**Resolution:** The pipeline should not be nested ACTION broadcasts. It should be a **Task** (Section 18.6):

```go
c.Task("agent-completion-pipeline", core.TaskDef{
    Steps: []core.Step{
        {Action: "agentic.qa",     Async: false},
        {Action: "agentic.auto-pr", Async: false},
        {Action: "agentic.verify",  Async: false},
        {Action: "agentic.ingest",  Async: true},  // parallel, don't block
        {Action: "agentic.poke",    Async: true},  // parallel, don't block
    },
})
```

The Task executor runs steps in order, with `Async: true` steps dispatched in parallel. Ingest and Poke don't wait for the pipeline — they fire immediately. The pipeline has a timeout. Each step has its own error handling.

### P6-2. Every Handler Receives Every Message — O(handlers × messages)

All 5 handlers are called for every ACTION. Each handler type-checks and skips if it's not their message. With N handlers and M message types, this is O(N×M) per event — every handler processes every message even if it only cares about one type.

With 12 message types and 5 handlers, that's 60 type-checks per agent completion cascade. Scales poorly as more handlers are added.

**Resolution:** The Action system (Section 18) fixes this — named Actions route directly:

```go
c.Action("on.agent.completed").Run(opts)  // only handlers for THIS action
```

No broadcast to all handlers. No type-checking. Direct dispatch by name.

### P6-3. Handlers Can't Tell If They're Inside a Nested Dispatch

A handler receiving `QAResult` can't tell if it was triggered by a top-level `c.ACTION(QAResult)` or by a nested call from inside the AgentCompleted handler. There's no dispatch context, no parent message ID, no trace.

**Resolution:** With the Action system, each invocation gets a context that tracks the chain:

```go
func handler(ctx context.Context, opts core.Options) core.Result {
    // ctx carries: parent action, trace ID, depth
}
```

### P6-4. Monitor Package Is Half-Migrated

```go
notifier ChannelNotifier // TODO(phase3): remove — replaced by c.ACTION()
```

Monitor has a `ChannelNotifier` field with a TODO to replace it with IPC. It's using a direct callback pattern alongside the IPC system. Two notification paths for the same events.

**Resolution:** Complete the migration. Remove `ChannelNotifier`. All events go through `c.ACTION()` (or named Actions once Section 18 is implemented).

### P6-5. Two Patterns for "Service Needs Core" — .core Field vs ServiceRuntime

Pass five found that core/agent sets `.core = c` manually while core/gui embeds `ServiceRuntime[T]`. But there's a third pattern:

```go
// Pattern 3: RegisterHandlers receives *Core as parameter
func RegisterHandlers(c *core.Core, s *PrepSubsystem) {
    c.RegisterAction(func(c *core.Core, msg core.Message) core.Result {
        // c is passed again — which one to use?
    })
}
```

The handler receives `*Core` as a parameter, but the service already has `s.core`. Two references to the same Core. If they ever diverge (multiple Core instances in tests), bugs ensue.

**Resolution:** Handlers should use the Core from the handler parameter (it's the dispatching Core). Services should use their embedded Core. Document: "handler's `c` parameter is the current Core. `s.core` is the Core from registration time. In single-Core apps they're the same."

### P6-6. Message Types Are Untyped — Any Handler Can Emit Any Message

```go
c.ACTION(messages.PRMerged{...})  // anyone can emit this
```

There's no ownership model. The Verify handler emits `PRMerged` but any code with a Core reference could emit `PRMerged` at any time. A rogue service could emit `AgentCompleted` for an agent that never existed.

**Resolution:** With named Actions, emission becomes explicit:

```go
c.Action("event.pr.merged").Emit(opts)  // must be registered
```

If no service registered the `event.pr.merged` action, the emit does nothing. Registration IS permission — same model as process execution.

### P6-7. The Aggregator Pattern (MCP) Has No Formal Support

core/mcp discovers all services and registers their capabilities as MCP tools. This is a cross-cutting "aggregator" pattern that reads the full registry. But there's no API for "give me all capabilities" — MCP hand-rolls it by checking each known service.

**Resolution:** `c.Registry("actions").List("*")` gives the full capability map. MCP's Register function becomes:

```go
func Register(c *core.Core) core.Result {
    for _, name := range c.Registry("actions").Names() {
        def := c.Action(name).Def()
        mcpServer.AddTool(name, def.Description, def.Schema)
    }
}
```

The aggregator reads from Registry. Any service's Actions are auto-exposed via MCP. No hand-wiring.

### P6-8. go-process OnShutdown Kills All Processes — But Core Doesn't Wait

go-process's `OnShutdown` sends SIGTERM to all running processes. But `ServiceShutdown` calls `OnStop` sequentially. If process service stops first, other services that depend on running processes lose their child processes mid-operation.

**Resolution:** Shutdown order should be reverse registration order (Section P4-1 — Registry maintains insertion order). Services registered last stop first. Since go-process is typically registered early, it stops last — giving other services time to finish their process-dependent work before processes are killed.

---

## Pass Seven — Failure Modes and Recovery

> Seventh review. What happens when things go wrong? Panics, partial failures,
> resource leaks, unrecoverable states.

### P7-1. New() Returns Half-Built Core on Option Failure

```go
func New(opts ...CoreOption) *Core {
    c := &Core{...}
    for _, opt := range opts {
        if r := opt(c); !r.OK {
            Error("core.New failed", "err", r.Value)
            break  // stops applying options, but returns c anyway
        }
    }
    c.LockApply()
    return c  // half-built: some services registered, some not
}
```

If `WithService` #3 fails, services #1 and #2 are registered but #3-5 are not. The caller gets a `*Core` with no indication it's incomplete. No error return. Just a log message.

**Resolution:** `New()` should either:
- Return `(*Core, error)` — breaking change but honest
- Store the error on Core: `c.Error().HasFailed()` — queryable
- Panic on construction failure — it's unrecoverable anyway

### P7-2. ServiceStartup Fails — No Rollback, No Cleanup

If service 3 of 5 fails `OnStartup`:
- Services 1+2 are running (started successfully)
- Service 3 failed
- Services 4+5 never started
- `Run()` calls `os.Exit(1)`
- Services 1+2 never get `OnShutdown` called
- Open files, goroutines, connections leak

```go
func (c *Core) Run() {
    r := c.ServiceStartup(c.context, nil)
    if !r.OK {
        Error(err.Error())
        os.Exit(1)  // no cleanup — started services leak
    }
```

**Resolution:** `Run()` should call `ServiceShutdown` even on startup failure — to clean up services that DID start:

```go
r := c.ServiceStartup(c.context, nil)
if !r.OK {
    c.ServiceShutdown(context.Background())  // cleanup started services
    os.Exit(1)
}
```

### P7-3. ACTION Handlers Have No Panic Recovery

```go
func (c *Core) Action(msg Message) Result {
    for _, h := range handlers {
        if r := h(c, msg); !r.OK {  // if h panics → unrecovered
            return r
        }
    }
}
```

`PerformAsync` has `defer recover()`. ACTION handlers do not. A single panicking handler crashes the entire application. Given that handlers run user-supplied code (service event handlers), this is high risk.

**Resolution:** Wrap each handler call in panic recovery:

```go
for _, h := range handlers {
    func() {
        defer func() {
            if r := recover(); r != nil {
                Error("ACTION handler panicked", "err", r, "msg", msg)
            }
        }()
        h(c, msg)
    }()
}
```

This matches PerformAsync's pattern. A panicking handler is logged and skipped, not fatal.

### P7-4. Run() Has No Panic Recovery — Shutdown Skipped

```go
func (c *Core) Run() {
    // no defer recover()
    c.ServiceStartup(...)
    cli.Run()               // if this panics...
    c.ServiceShutdown(...)  // ...this never runs
}
```

If `Cli.Run()` panics (user command handler panics), `ServiceShutdown` is never called. All services leak. PID files aren't cleaned. Health endpoints keep responding.

**Resolution:** `Run()` needs `defer c.ServiceShutdown(context.Background())`:

```go
func (c *Core) Run() {
    defer c.ServiceShutdown(context.Background())
    // ... rest of Run
}
```

### P7-5. os.Exit(1) Called Directly — Bypasses defer

`Run()` calls `os.Exit(1)` twice. `os.Exit` bypasses ALL defer statements. Even if we add `defer ServiceShutdown`, `os.Exit(1)` skips it.

```go
r := c.ServiceStartup(...)
if !r.OK {
    os.Exit(1)  // skips all defers
}
```

**Resolution:** Don't call `os.Exit` inside `Run()`. Return an error and let `main()` handle the exit:

```go
func (c *Core) Run() error {
    defer c.ServiceShutdown(context.Background())
    // ...
    if !r.OK { return r.Value.(error) }
    return nil
}

// In main():
if err := c.Run(); err != nil {
    os.Exit(1)
}
```

Or use `c.Error().Recover()` + panic instead of os.Exit — panics respect defer.

### P7-6. ServiceShutdown Continues After First Error But May Miss Services

```go
var firstErr error
for _, s := range stoppables {
    r := s.OnStop()
    if !r.OK && firstErr == nil {
        firstErr = e  // records first error
    }
}
```

Good — it continues calling OnStop even after failures. But it checks `ctx.Err()` in the loop:

```go
if err := ctx.Err(); err != nil {
    return Result{err, false}  // stops remaining services
}
```

If the shutdown context has a timeout and it expires, remaining services never get OnStop. Combined with P7-4 (no panic recovery), a slow-to-stop service can prevent later services from stopping.

**Resolution:** Each service gets its own shutdown timeout, not a shared context:

```go
for _, s := range stoppables {
    svcCtx, cancel := context.WithTimeout(ctx, 10*time.Second)
    s.OnStop(svcCtx)
    cancel()
}
```

### P7-7. ErrorPanic.SafeGo Exists but Nobody Uses It

```go
func (h *ErrorPanic) SafeGo(fn func()) {
    go func() {
        defer h.Recover()
        fn()
    }()
}
```

A protected goroutine launcher exists. But `PerformAsync` doesn't use it — it has its own inline `defer recover()`. `Run()` doesn't use it. `ACTION` dispatch doesn't use it. The safety primitive exists but isn't wired into the system that needs it most.

**Resolution:** `SafeGo` should be the standard way to spawn goroutines in Core. `PerformAsync`, `Run()`, and ACTION handlers should all use `c.Error().SafeGo()` or its equivalent. One pattern for protected goroutine spawning.

### P7-8. No Circuit Breaker on IPC Handlers

If an ACTION handler consistently fails (returns !OK or panics), it's still called for every message forever. There's no:
- Tracking of handler failure rate
- Automatic disabling of broken handlers
- Alert that a handler is failing

Combined with P6-1 (cascade), a consistently failing handler in the middle of the pipeline blocks everything downstream on every event.

**Resolution:** The Action system (Section 18) should track failure rate per action. After N consecutive failures, the action is suspended with an alert. This is the same rate-limiting pattern that core/agent uses for agent dispatch (`trackFailureRate` in dispatch.go).

---

## Pass Eight — Type Safety

> Eighth review. What does the type system promise versus what it delivers?
> Where do panics hide behind type assertions?

### P8-1. 79% of Type Assertions Are Bare — Panic on Wrong Type

Core has 63 type assertions on `Result.Value` in source code. Only 13 use the comma-ok pattern. **50 are bare assertions that panic if the type is wrong:**

```go
// Safe (13 instances):
s, ok := r.Value.(string)

// Unsafe (50 instances):
content := r.Value.(string)  // panics if Value is int, nil, or anything else
entries := r.Value.([]os.DirEntry)  // panics if wrong
```

core/agent has 128 type assertions — the same ratio applies. Roughly 100 hidden panic sites in the consumer code.

**The implicit contract:** "if `r.OK` is true, `Value` is the expected type." But nothing enforces this. A bug in `Fs.Read()` returning `int` instead of `string` would panic at the consumer's type assertion — far from the source.

**Resolution:** Two approaches:
1. **Typed Result methods** — `r.String()`, `r.Bytes()`, `r.Entries()` that use comma-ok internally
2. **Convention enforcement** — AX-7 Ugly tests should test with wrong types in Result.Value

### P8-2. Result Is One Type for Everything — No Compile-Time Safety

```go
type Result struct {
    Value any
    OK    bool
}
```

`Fs.Read()` returns `Result` with `string`. `Fs.List()` returns `Result` with `[]os.DirEntry`. `Service()` returns `Result` with `*Service`. Same type, completely different contents. The compiler can't help.

```go
r := c.Fs().Read(path)
entries := r.Value.([]os.DirEntry)  // compiles fine, panics at runtime
```

**This is the core trade-off of AX:** Result reduces LOC (`if r.OK` vs `if err != nil`) but loses compile-time type safety. Go's `(T, error)` pattern catches this at compile time. Core's Result catches it at runtime.

**Resolution:** Accept the trade-off but mitigate:
- Typed convenience methods on subsystems: `c.Fs().ReadString(path)` returns `string` directly
- These already exist partially: `Options.String()`, `Config.String()`, `Config.Int()`
- Extend the pattern: `c.Fs().ReadString()`, `c.Fs().ListEntries()`, `c.Data().ReadString()`
- Keep `Result` as the universal type for generic operations. Add typed accessors for known patterns.

### P8-3. Message/Query/Task Are All `any` — No Type Routing

```go
type Message any
type Query any
type Task any
```

Every IPC boundary is untyped. A handler receives `any` and must type-switch:

```go
c.RegisterAction(func(c *core.Core, msg core.Message) core.Result {
    ev, ok := msg.(messages.AgentCompleted)
    if !ok { return core.Result{OK: true} }  // not my message, skip
})
```

Wrong type = silent skip. No compile-time guarantee that a handler will ever receive its expected message. No way to know at registration time what messages a handler cares about.

**Resolution:** Named Actions (Section 18) fix this — each Action has a typed handler:

```go
c.Action("agent.completed", func(ctx context.Context, opts core.Options) core.Result {
    // opts is always Options — typed accessor: opts.String("repo")
    // no type-switch needed
})
```

### P8-4. Option.Value Is `any` — Config Becomes Untyped Bag

```go
type Option struct {
    Key   string
    Value any  // string? int? bool? *MyStruct? []byte?
}
```

`WithOption("port", 8080)` stores an int. `WithOption("name", "app")` stores a string. `WithOption("debug", true)` stores a bool. The Key gives no hint about the expected Value type.

```go
c.Options().String("port")  // returns "" — port is int, not string
c.Options().Int("name")     // returns 0 — name is string, not int
```

Silent wrong-type returns. No error, no panic — just zero values.

**Resolution:** `ConfigVar[T]` (Issue 11) solves this for Config. For Options, document expected types per key, or use typed option constructors:

```go
// Instead of:
core.WithOption("port", 8080)

// Consider:
core.WithInt("port", 8080)
core.WithString("name", "app")
core.WithBool("debug", true)
```

This is more verbose but compile-time safe. Or accept the untyped bag and rely on convention + tests.

### P8-5. ServiceFor Uses Generics but Returns (T, bool) Not Result

```go
func ServiceFor[T any](c *Core, name string) (T, bool) {
```

This is the ONE place Core uses Go's tuple return instead of Result. It breaks the universal pattern. A consumer accessing a service writes different error handling than accessing anything else:

```go
// Service — tuple pattern:
svc, ok := core.ServiceFor[*Brain](c, "brain")
if !ok { ... }

// Everything else — Result pattern:
r := c.Config().Get("key")
if !r.OK { ... }
```

**Resolution:** Accept this. `ServiceFor[T]` uses generics which can't work with `Result.Value.(T)` ergonomically — the generic type parameter can't flow through the `any` interface. This is a Go language limitation, not a Core design choice.

### P8-6. Fs Path Validation Returns Result — Then Callers Assert String

```go
func (m *Fs) validatePath(p string) Result {
    // returns Result{validatedPath, true} or Result{error, false}
}

func (m *Fs) Read(p string) Result {
    vp := m.validatePath(p)
    if !vp.OK { return vp }
    data, err := os.ReadFile(vp.Value.(string))  // bare assertion
}
```

`validatePath` returns `Result` with a string inside. Every Fs method then bare-asserts `vp.Value.(string)`. This is safe (validatePath always returns string on OK) but the pattern is fragile — if validatePath ever changes its return type, every Fs method panics.

**Resolution:** `validatePath` should return `(string, error)` since it's internal. Or `validatePath` should return `string` directly since failure means the whole Fs operation fails. Internal methods don't need Result wrapping — that's for the public API boundary.

### P8-7. IPC Handler Registration Accepts Any Function — No Signature Validation

```go
c.RegisterAction(func(c *core.Core, msg core.Message) core.Result { ... })
```

The function signature is enforced by Go's type system. But `HandleIPCEvents` auto-discovery uses interface assertion:

```go
if handler, ok := instance.(interface {
    HandleIPCEvents(*Core, Message) Result
}); ok {
```

If a service has `HandleIPCEvents` with a SLIGHTLY wrong signature (e.g., returns `error` instead of `Result`), it silently doesn't register. No error, no warning. The service author thinks their handler is registered but it's not.

**Resolution:** Log when a service has a method named `HandleIPCEvents` but with the wrong signature. Or better — remove auto-discovery (P5-4) and require explicit registration.

### P8-8. The Guardrail Paradox — Core Removes Type Safety to Reduce LOC

Core's entire design trades compile-time safety for runtime brevity:

| Go Convention | Core | Trade-off |
|--------------|------|-----------|
| `(string, error)` | `Result` | Loses return type info |
| `specific.Type` | `any` | Loses type checking |
| `strings.Contains` | `core.Contains` | Same safety, different import |
| `typed struct` | `Options{Key, Value any}` | Loses field types |

The LOC reduction is real — `if r.OK` IS shorter than `if err != nil`. But every `r.Value.(string)` is a deferred panic that `(string, error)` would have caught at compile time.

**This is not a bug — it's the fundamental design tension of Core.** The resolution is not to add types back (that defeats the purpose). It's to:
1. Cover every type assertion with AX-7 Ugly tests
2. Add typed convenience methods where patterns are known
3. Accept that Result is a runtime contract, not a compile-time one
4. Use `Registry[T]` (typed generic) instead of `map[string]any` where possible

The guardrail primitives (Array[T], Registry[T], ConfigVar[T]) are Core's answer — they're generic and typed. The untyped surface (Result, Option, Message) is the integration layer where different types meet.

---

## Pass Nine — What's Missing, What Shouldn't Be There

> Ninth review. What should Core provide that it doesn't? What does Core
> contain that belongs elsewhere?

### P9-1. core/go Imports os/exec — Violates Its Own Rule

`app.go` imports `os/exec` for `exec.LookPath`:

```go
func (a App) Find(filename, name string) Result {
    path, err := exec.LookPath(filename)
```

Core says "zero os/exec — go-process is the only allowed user" (Section 17). But Core itself violates this. `App.Find()` locates binaries on PATH — a process concern that belongs in go-process.

**Resolution:** Move `App.Find()` to go-process or make it a `process.which` Action. Core shouldn't import the package it tells consumers not to use.

### P9-2. core/go Uses reflect for Service Name Discovery

```go
typeOf := reflect.TypeOf(instance)
pkgPath := typeOf.PkgPath()
name := Lower(parts[len(parts)-1])
```

`WithService` uses `reflect` to auto-discover the service name from the factory's return type's package path. This is magic (P5-4) and fragile — renaming a package changes the service name.

**Resolution:** `WithName` already exists for explicit naming. Consider making `WithService` require a name, or document that the auto-discovered name IS the package name and changing it is a breaking change.

### P9-3. No JSON Primitive — Every Consumer Imports encoding/json

core/agent's agentic package has 15+ files importing `encoding/json`. Core provides string helpers (`core.Contains`, `core.Split`) but no JSON helpers. Every consumer writes:

```go
data, err := json.Marshal(thing)      // no core wrapper
json.NewDecoder(resp.Body).Decode(&v) // no core wrapper
```

**Resolution:** Consider `core.JSON.Marshal()` / `core.JSON.Unmarshal()` or `core.ToJSON()` / `core.FromJSON()`. Same guardrail pattern as strings — one import, one pattern, model-proof. Or accept that `encoding/json` is stdlib and doesn't need wrapping.

**The test:** Does wrapping JSON add value beyond import consistency? Strings helpers add value (they're on `core.` so agents find them). JSON helpers might just be noise. This one is a judgment call.

### P9-4. No Time Primitive — Timestamps Are Ad-Hoc

Every service formats time differently:

```go
time.Now().Format("2006-01-02T15-04-05")   // agentic events
time.Now().Format(time.RFC3339)             // review queue
time.Now()                                  // status.UpdatedAt
```

No standard timestamp format across the ecosystem.

**Resolution:** `core.Now()` returning a Core-standard timestamp, or `core.FormatTime(t)` with one format. Or document the convention: "all timestamps use RFC3339" and let consumers use `time` directly.

### P9-5. No ID Generation Primitive — Multiple Patterns Exist

```go
// plan.go — crypto/rand hex
b := make([]byte, 3)
rand.Read(b)
return slug + "-" + hex.EncodeToString(b)

// task.go — atomic counter
taskID := "task-" + strconv.FormatUint(c.taskIDCounter.Add(1), 10)

// go-process — atomic counter
id := fmt.Sprintf("proc-%d", s.idCounter.Add(1))
```

Three different ID generation patterns. No standard `core.NewID()`.

**Resolution:** `core.ID()` that returns a unique string. Implementation can be atomic counter, UUID, or whatever — but one function, one pattern. Same guardrail logic as everything else.

### P9-6. No Validation Primitive — Input Validation Is Scattered

```go
// Repo name validation — agentic/prep.go
repoName := core.PathBase(input.Repo)
if repoName == "." || repoName == ".." || repoName == "" {
    return core.E("prep", "invalid repo name", nil)
}

// Command path validation — command.go
if path == "" || HasPrefix(path, "/") || HasSuffix(path, "/") || Contains(path, "//") {
    return Result{E("core.Command", "invalid command path", nil), false}
}

// Plan ID sanitisation — agentic/plan.go
safe := core.PathBase(id)
if safe == "." || safe == ".." || safe == "" {
    safe = "invalid"
}
```

Every validation is hand-rolled. The same path traversal check appears in three places with slight variations.

**Resolution:** `core.ValidateName(s)` / `core.ValidatePath(s)` / `core.SanitisePath(s)` — reusable validation primitives. One check, used everywhere. A model can't forget the `..` check if it's in the primitive.

### P9-7. Error Codes Are Optional and Unused

```go
func WrapCode(err error, code, op, msg string) error { ... }
func NewCode(code, msg string) error { ... }
func ErrorCode(err error) string { ... }
```

Error codes exist in the type system but nobody uses them. No consumer calls `WrapCode` or checks `ErrorCode`. The infrastructure is there but the convention isn't established.

**Resolution:** Either remove error codes (dead feature) or define a code taxonomy and use it. Error codes are useful for API responses (HTTP status codes map to error codes) and for monitoring (count errors by code). But they need a convention to be useful.

### P9-8. Core Has No Observability Primitive

Core has:
- `c.Log()` — structured logging
- `c.Error()` — panic recovery + crash reports
- `ActionTaskProgress` — progress tracking

Core does NOT have:
- Metrics (counters, gauges, histograms)
- Tracing (span IDs, parent/child relationships)
- Health checks (readiness, liveness)

go-process has health checks. But they're per-daemon, not per-service. There's no `c.Health()` primitive that aggregates all services' health status.

**Resolution:** Consider `c.Health()` as a primitive:

```go
c.Health().Ready()           // bool — all services started
c.Health().Live()            // bool — not in shutdown
c.Health().Check("brain")    // per-service health
```

Metrics and tracing are extension concerns — they belong in separate packages like go-process. But health is fundamental enough to be a Core primitive. Every service could implement `Healthy() bool` alongside `Startable`/`Stoppable`.

---

## Pass Ten — The Spec Auditing Itself

> Tenth review. Does the RFC contradict itself? Do sections conflict?
> Is the subsystem map accurate? Is the API surface coherent?

### P10-1. Section 17 and 18 Contradict on Process Return Type

Section 17 spec'd `c.Process().Run() → (string, error)`. Section 18 spec'd `c.Action().Run() → Result`. Section 18.9 says "Process IS sugar over Actions." P3-2 flagged this.

**But Section 17 still shows the old signature.** The spec contradicts itself across sections. An agent reading Section 17 implements `(string, error)`. An agent reading Section 18 implements `Result`.

**Resolution:** Update Section 17 to match Section 18. Process returns Result. Add typed convenience: `c.Process().RunString()` for the common case.

### P10-2. Section 17 Uses ACTION for Request/Response — Should Be PERFORM

Section 17.6 lists `ProcessRun` as an IPC Message (broadcast). Section 17.7 shows it going through ACTION dispatch. But `ProcessRun` NEEDS a response (the output string). That's PERFORM, not ACTION.

P6-1 showed that ACTION cascade is synchronous and blocking. Using ACTION for process execution means every process call blocks the entire IPC pipeline.

**Resolution:** Process messages use PERFORM (targeted execution, returns result) not ACTION (broadcast, fire-and-forget). Event messages (ProcessStarted, ProcessExited) use ACTION (notification, no response needed).

```
PERFORM: ProcessRun, ProcessStart, ProcessKill  — need response
ACTION:  ProcessStarted, ProcessExited, ProcessKilled — notification only
```

### P10-3. Subsystem Count Is Wrong

Section 19.8 says "14 subsystems." Actual count:

| Category | Accessor | Count |
|----------|----------|-------|
| Subsystems (struct with methods) | Config, Data, Fs, IPC, Cli, Log, Error, I18n, Drive, App | 10 |
| Value accessors | Options, Context, Env, Core | 4 |
| Dual-purpose (get/set) | Service, Command, Lock | 3 |
| Planned primitives | Action, Registry, API, Process | 4 |
| Proposed (P9-8) | Health | 1 |

That's **22 methods on Core**, not 14 subsystems. The map needs to distinguish between subsystems, accessors, registries, and convenience methods.

**Resolution:** Recount. Subsystems are things with methods you call. Value accessors return data. Dual-purpose methods are Registry sugar. The map should categorise, not just list.

### P10-4. Four API Patterns on One Struct — No Categorisation

Core's methods look uniform but behave differently:

```go
c.Config()       // subsystem — returns struct, call methods on it
c.Options()      // accessor — returns data, read from it
c.Service("x")   // dual-purpose — get or set depending on args
c.Action("x")    // planned registry — get, set, invoke, inspect
```

An agent sees 22 methods on Core and can't tell which pattern each follows without reading the implementation. The names give no hint.

**Resolution:** Document the categories in the RFC. Or use naming to signal:

```
c.Config()        // noun → subsystem (has methods)
c.Options()       // noun → accessor (read-only data)
c.Service("x")    // noun + arg → registry operation
c.ACTION(msg)     // VERB → convenience (does something)
```

The existing naming DOES follow a pattern — it's just undocumented:
- **No-arg noun** → subsystem accessor
- **Noun + arg** → registry get/set
- **UPPERCASE verb** → convenience operation
- **Noun returning primitive** → value accessor

### P10-5. Registry Is Section 20 but Resolves Issues From Pass One

Registry[T] was added in Section 20 but it resolves Issues 6, 12, 13 from Pass One. The RFC reads chronologically (issues found, then solution added later) but an implementer reads it looking for "what do I build?"

**Resolution:** Add a cross-reference table:

| Issue | Resolved By | Section |
|-------|------------|---------|
| 1 (naming) | CamelCase/UPPERCASE convention | Design Philosophy |
| 5 (RegisterAction location) | task.go → action.go split | Issue 16 |
| 6 (serviceRegistry unexported) | Registry[T] | Section 20 |
| 9 (CommandLifecycle) | Managed field + three-layer CLI | Issue 9 |
| 10 (Array[T]) | Guardrail primitive, keep | Issue 10 |
| 11 (ConfigVar[T]) | Promote to primitive | Issue 11 |
| 12 (Ipc data-only) | IPC consumes Action registry | Section 20 |
| 13 (Lock allocation) | Registry[T] caching | Section 20 |
| 16 (task.go concerns) | ipc.go + action.go split | Issue 16 |

### P10-6. The Design Philosophy Section Is Before the Findings That Shaped It

"Core Is Lego Bricks" appears before the passes that discovered WHY it's Lego bricks. The Design Philosophy was written during Pass Two but the reasoning (Passes Three through Nine) comes after. An implementer reading top-to-bottom gets the rules before the reasoning.

**Resolution:** This is actually correct for an RFC — rules first, reasoning appendixed. The passes ARE the appendix. The Design Philosophy is the executive summary. The structure works for both humans (read top-down) and agents (grep for specific sections).

### P10-7. v0.8.0 Requirements Don't Include Pass Two-Nine Findings

The v0.8.0 checklist lists Section 17-20 implementation and AX-7 tests. But Passes Two through Nine found 56 additional issues (P2-1 through P9-8). These aren't in the checklist.

**Resolution:** Add a "findings severity" classification:

| Severity | Count | v0.8.0? | Examples |
|----------|-------|---------|----------|
| Critical (bug) | 3 | Yes | P4-3 (ACTION stops chain), P6-1 (cascade blocks), P7-2 (no cleanup) |
| High (design) | 12 | Yes | P3-1 (error/Result mismatch), P8-1 (bare assertions), P9-1 (exec import) |
| Medium (improve) | 25 | Stretch | P5-4 (magic methods), P9-3 (no JSON), P9-5 (no ID gen) |
| Low (document) | 16 | No | P4-4 (clone-iterate), P10-6 (section ordering) |

3 critical bugs must be fixed for v0.8.0. 12 high design issues should be. The rest are improvements.

### P10-8. No Section Numbers for Passes — Hard to Cross-Reference

Passes are referenced as P2-1, P3-2, etc. but they're not in the table of contents. The RFC has Sections 1-20 (numbered) and Passes One-Ten (named). An agent searching for "P6-1" has to scan the document.

**Resolution:** Either number the passes as sections (Section 21 = Pass One, etc.) or add a findings index at the end with all P-numbers linking to their location. The findings ARE the backlog — they need to be queryable.

---

## Pass Eleven — Security Model

> Eleventh review. The conclave's threat model. What can a service do?
> What's the actual isolation? Where are the secrets?

### P11-1. Every Service Has God Mode

A registered service receives `*Core`. With it, it can:

| Capability | Method | Risk |
|-----------|--------|------|
| Read all config | `c.Config().Get(anything)` | Secret exposure |
| Write all config | `c.Config().Set(anything, anything)` | Config corruption |
| Read all data | `c.Data().ReadString(any/path)` | Data exfiltration |
| List all services | `c.Services()` | Reconnaissance |
| Access any service | `ServiceFor[T](c, name)` | Lateral movement |
| Broadcast any event | `c.ACTION(any_message)` | Event spoofing |
| Register handlers | `c.RegisterAction(spy)` | Eavesdropping |
| Register commands | `c.Command("admin/nuke", fn)` | Capability injection |
| Read environment | `c.Env("FORGE_TOKEN")` | Secret theft |
| Write filesystem | `c.Fs().Write(any, any)` | Data destruction |
| Delete filesystem | `c.Fs().DeleteAll(any)` | Data destruction |

There's no per-service permission model. Registration grants full access. This is fine for a trusted conclave (all services from your own codebase) but dangerous for plugins or third-party services.

**Resolution:** For v0.8.0, accept God Mode — all services are first-party. For v0.9.0+, consider capability-based Core views where a service receives a restricted `*Core` that only exposes permitted subsystems.

### P11-2. The Fs Sandbox Is Bypassed by unsafe.Pointer

Pass Two (P2-2) said `Fs.root` is "correctly unexported — the security boundary." But core/agent bypasses it:

```go
type fsRoot struct{ root string }
f := &core.Fs{}
(*fsRoot)(unsafe.Pointer(f)).root = root
```

Two files (`paths.go` and `detect.go`) use `unsafe.Pointer` to overwrite the private `root` field, creating unrestricted Fs instances. The security boundary that P2-2 praised is already broken by the first consumer.

**Resolution:** Add `Fs.NewUnrestricted()` or `Fs.New("/")` as a legitimate API. If consumers need unrestricted access, give them a door instead of letting them pick the lock. Then `go vet` rules or linting can flag `unsafe.Pointer` usage on Core types.

### P11-3. core.Env() Exposes All Environment Variables — Including Secrets

```go
c.Env("FORGE_TOKEN")      // returns the token
c.Env("OPENAI_API_KEY")   // returns the key
c.Env("ANTHROPIC_API_KEY") // returns the key
```

Any service can read any environment variable. API keys, tokens, database passwords — all accessible via `c.Env()`. There's no secret/non-secret distinction.

**Resolution:** Consider `c.Secret(name)` that reads from a secure store (encrypted file, vault) rather than plain environment variables. Or `c.Env()` with a redaction list — keys matching `*TOKEN*`, `*KEY*`, `*SECRET*`, `*PASSWORD*` are redacted in logs but still accessible to code.

The logging system already has `SetRedactKeys` — extend this to Env access logging.

### P11-4. ACTION Event Spoofing — Any Code Can Emit Any Message

P6-6 noted this from a design perspective. From a security perspective:

```go
// A rogue service can fake agent completions:
c.ACTION(messages.AgentCompleted{
    Agent: "codex", Repo: "go-io", Status: "completed",
})
// This triggers the ENTIRE pipeline: QA → PR → Verify → Merge
// For an agent that never existed
```

No authentication on messages. No sender identity. Any service can emit any message type and trigger any pipeline.

**Resolution:** Named Actions (Section 18) help — `c.Action("event.agent.completed").Emit()` requires the action to be registered. But the emitter still isn't authenticated. Consider adding sender identity to IPC:

```go
c.ACTION(msg, core.Sender("agentic"))  // identifies who sent it
```

Handlers can then filter by sender. Not cryptographic security — but traceability.

### P11-5. RegisterAction Can Install a Spy Handler

```go
c.RegisterAction(func(c *core.Core, msg core.Message) core.Result {
    // sees EVERY message in the conclave
    // can log, exfiltrate, or modify behaviour
    log.Printf("SPY: %T %+v", msg, msg)
    return core.Result{OK: true}
})
```

A single `RegisterAction` call installs a handler that receives every IPC message. Combined with P11-1 (God Mode), a service can silently observe all inter-service communication.

**Resolution:** For trusted conclaves, this is a feature (monitoring, debugging). For untrusted environments, IPC handlers need scope: "this handler only receives messages of type X." Named Actions (Section 18) solve this — each handler is registered for a specific action, not all messages.

### P11-6. No Audit Trail — No Logging of Security-Relevant Operations

There's no automatic logging when:
- A service registers (who registered what)
- A handler is added (who's listening)
- Config is modified (who changed what)
- Environment variables are read (who accessed secrets)
- Filesystem is written (who wrote where)

All operations are silent. The `core.Security()` log level exists but nothing calls it.

**Resolution:** Core should emit `core.Security()` logs for:
- Service registration: `Security("service.registered", "name", name, "caller", caller)`
- Config writes: `Security("config.set", "key", key, "caller", caller)`
- Secret access: `Security("secret.accessed", "key", key, "caller", caller)`
- Fs writes outside Data mounts: `Security("fs.write", "path", path, "caller", caller)`

### P11-7. ServiceLock Has No Authentication

```go
c.LockApply()  // anyone can call this
c.LockEnable() // anyone can call this
```

ServiceLock prevents late registration. But any code with `*Core` can call `LockApply()` prematurely, locking out legitimate services that haven't registered yet. Or call `LockEnable()` to arm the lock before it should be.

**Resolution:** `LockApply` should only be callable from `New()` (during construction). Post-construction, the lock is managed by Core, not by consumers. With Registry[T], `Lock()` becomes a Registry method — and Registry access can be scoped.

### P11-8. The Permission-by-Registration Model Has No Revocation

Section 17.7 says "no handler = no capability." But once a service registers, there's no way to revoke its capabilities:

- Can't unregister a service
- Can't remove an ACTION handler
- Can't revoke filesystem access
- Can't deregister an Action

The conclave is append-only. Services join but never leave (except on shutdown). For hot-reload (P3-8), a misbehaving service can't be ejected.

**Resolution:** Registry needs `Delete(name)` and `Disable(name)`:

```go
c.Registry("actions").Disable("rogue.action")  // stops dispatch, keeps registered
c.Registry("services").Delete("rogue")          // removes entirely
```

`Disable` is soft — the action exists but doesn't execute. `Delete` is hard — the action is gone. Both require the caller to have sufficient authority — which loops back to P11-1 (God Mode).

---

## Pass Twelve — Migration Risk

> Twelfth review. 44 repos import core/go. What breaks when we ship v0.8.0?
> Categorise every RFC change by blast radius.

### P12-1. The Blast Radius — 44 Consumer Repos

Every breaking change in core/go ripples to 44 repositories across the ecosystem. The highest-risk changes:

| Change | Blast Radius | Files Affected |
|--------|-------------|----------------|
| P3-1: Startable returns Result | 26 files | Every OnStartup/OnShutdown impl |
| P7-5: Run() returns error | 15 files | Every main.go calling c.Run() |
| I9: CommandLifecycle → Managed | ~10 files | Services with daemon commands |
| I3: Remove Embed() | 0 files | Nobody uses it — safe |

### P12-2. Migration Phasing — What Can Ship Independently

**Phase 1: Additive (zero breakage)**
- Section 17: `c.Process()` — new accessor
- Section 18: `c.Action()` — new accessor
- Section 19: `c.API()` — new accessor
- Section 20: `Registry[T]` — new type
- P7-4: Add `defer ServiceShutdown` to Run()
- P4-3: Fix ACTION chain (!OK doesn't stop broadcast)
- P7-3: Add panic recovery to ACTION handlers
- I3: Remove `Embed()` (zero consumers)

These can ship as v0.7.x patches. No consumer breakage.

**Phase 2: Internal refactor (no API change)**
- I16: task.go → action.go (file rename, same package)
- I5: Move RegisterAction to ipc.go (same package)
- P9-1: Remove os/exec from app.go
- P11-2: Add Fs.NewUnrestricted() to replace unsafe.Pointer hacks

**Phase 3: Breaking changes (need ecosystem sweep)**
- P3-1: Startable/Stoppable return Result — 26 files to update
- P7-5: Run() returns error — 15 main.go files
- I9: Command.Managed replaces CommandLifecycle

Phase 3 is the only one that requires a coordinated ecosystem update. Phases 1-2 can ship immediately.

### P12-3. The Startable Change Is The Biggest Risk

26 files across the ecosystem implement `OnStartup(ctx) error`. Changing to `OnStartup(ctx) Result` means:

```go
// Before (26 files):
func (s *Svc) OnStartup(ctx context.Context) error {
    return nil
}

// After:
func (s *Svc) OnStartup(ctx context.Context) core.Result {
    return core.Result{OK: true}
}
```

This is a mechanical change but it touches every service in every repo. One Codex dispatch per repo with the template: "change OnStartup/OnShutdown return type from error to core.Result."

**Alternative:** Keep `error` return. Accept the inconsistency (P3-1) as the cost of not breaking 26 files. Add `Result`-returning variants as new interfaces:

```go
type StartableV2 interface { OnStartup(ctx context.Context) Result }
```

Core checks for V2 first, falls back to V1. No breakage. V1 is deprecated. Consumers migrate at their own pace.

### P12-4. The Run() Change Can Be Backwards Compatible

```go
// Current:
func (c *Core) Run() { ... os.Exit(1) }

// Add alongside (no breakage):
func (c *Core) RunE() error { ... return err }
```

`Run()` stays for backwards compatibility. `RunE()` is the new pattern. `Run()` internally calls `RunE()` and handles the exit. Consumers migrate when ready.

### P12-5. The Critical Bugs Are Phase 1 — Ship Immediately

The 3 critical bugs (P4-3, P6-1, P7-2) are all Phase 1 (additive/internal):

| Bug | Fix | Phase | Risk |
|-----|-----|-------|------|
| P4-3: ACTION !OK stops chain | Change loop to continue on !OK | 1 | Low — behaviour improves |
| P7-2: No cleanup on startup failure | Add shutdown call before exit | 1 | Low — adds safety |
| P7-3: ACTION handlers no panic recovery | Wrap in defer/recover | 1 | Low — adds safety |

These can ship tomorrow as v0.7.1 with zero consumer breakage.

### P12-6. The Cascade Fix (P6-1) Requires core/agent Change

Fixing the synchronous cascade (P6-1) means core/agent's handlers.go must change from nested `c.ACTION()` calls to a Task pipeline. This is a core/agent change, not a core/go change.

```go
// Current (core/agent/handlers.go):
c.RegisterAction(func(c *core.Core, msg core.Message) core.Result {
    // runs QA, then calls c.ACTION(QAResult) — nested, blocking
})

// Target:
c.Task("agent.completion", core.TaskDef{
    Steps: []{Action: "agentic.qa"}, {Action: "agentic.pr"}, ...
})
```

This needs Section 18 (Actions/Tasks) implemented first. So the cascade fix depends on the Action system, which is Phase 1 (additive). But the handler refactor in core/agent is a separate change.

### P12-7. 44 Repos Need AX-7 Tests Before v0.8.0 Tagging

The v0.8.0 checklist says "AX-7 at 100%." Currently:
- core/go: 14% AX-7 (83.6% statement coverage, wrong naming)
- core/agent: 92% AX-7 (79.9% statement coverage)
- Other 42 repos: unknown

Getting all 44 repos to AX-7 100% before v0.8.0 is unrealistic. Scope it:
- core/go + core/agent at 100% = v0.8.0 (reference implementations)
- Other repos adopt incrementally via Codex sweeps
- Each repo gets AX-7 as part of the Phase 3 ecosystem sweep

### P12-8. The Migration Tool — Codex Dispatches From the RFC

The RFC IS the migration spec. Each Phase maps to a Codex dispatch:

```
Phase 1: dispatch to core/go — "implement Section 17-20, fix P4-3/P7-2/P7-3"
Phase 2: dispatch to core/go — "refactor task.go→action.go, remove os/exec from app.go"
Phase 3: sweep all 44 repos — "update OnStartup/OnShutdown to Result, Run→RunE"
```

The RFC sections are detailed enough that a Codex agent can implement them. That's the AX promise — the spec IS the implementation guide.

---

## Pass Thirteen — Hidden Assumptions

> Final review. What does this RFC assume without stating? What breaks
> if the assumption is wrong?

### P13-1. Single Core per Process — Not Enforced

The RFC assumes one Core instance per process. But `core.New()` can be called multiple times. Tests do this routinely (each test creates its own Core). The global state that breaks:

| Global | Effect of Multiple Cores |
|--------|------------------------|
| `assetGroups` | Shared — all Cores see same assets |
| `systemInfo` | Shared — all Cores see same env |
| `defaultLogPtr` | Last `New()` wins — earlier Cores' log redirect lost |
| `process.Default()` | Last `SetDefault()` wins |
| `crashMu` | Shared — crash reports interleave |

**Resolution:** Document that Core is designed for single-instance-per-process. Multiple instances work for tests (isolated registries, shared globals) but not for production (global state conflicts). The global state IS the singleton enforcement — it just does it implicitly instead of explicitly.

### P13-2. All Services Are Go — No Plugin Boundary

Every service in the RFC is a Go function compiled into the binary. There's no concept of:
- Loading a service from a shared library
- Running a service in a subprocess
- Connecting to a remote service as if it were local

Section 19 (API/Streams) addresses remote communication. But there's no "remote service that looks local" — no transparent proxy where `c.Service("brain")` could return a stub that forwards over the network.

**Resolution:** With Actions (Section 18), this becomes natural:

```go
// Local action — in-process handler
c.Action("brain.recall").Run(opts)

// Remote action — transparently proxied
c.Action("charon:brain.recall").Run(opts)
```

The Action system doesn't care where the handler is. Local or remote, same API. This is the path to cross-process services without changing the service author's experience.

### P13-3. Go Is The Only Language — But CorePHP and CoreTS Exist

The RFC specs Go types (`struct`, `interface`, generics). But the ecosystem has CorePHP (Laravel) and CoreTS (TypeScript). The primitives (Option, Result, Service lifecycle) should be language-agnostic concepts that each implementation maps to its language.

**Resolution:** The RFC is the Go implementation spec. A separate document (or RFC-025 extension) should define the language-agnostic concepts:

| Concept | Go | PHP | TypeScript |
|---------|-----|-----|-----------|
| Option | `Option{Key, Value}` | `['key' => 'value']` | `{key: string, value: any}` |
| Result | `Result{Value, OK}` | `Result::ok($val)` | `{value: T, ok: boolean}` |
| Service | `Startable` interface | Laravel ServiceProvider | Class with lifecycle |
| IPC | `c.ACTION(msg)` | `event(new Msg())` | `core.emit('msg', data)` |

The concepts map. The syntax differs. The RFC should note this boundary.

### P13-4. Linux/macOS Only — syscall Assumptions

`syscall.Kill`, `syscall.SIGTERM`, `os/exec`, PID files — all assume Unix. Windows has different process management. The RFC doesn't mention platform constraints.

**Resolution:** Document: "Core targets Linux and macOS. Windows is not supported for process management (go-process, daemon lifecycle). Core primitives (Option, Result, Config, IPC) are platform-independent. Process execution is Unix-only."

### P13-5. Startup Is Synchronous — No Lazy Loading

All services start during `ServiceStartup()`, sequentially. A service that takes 30 seconds to connect to a database blocks all subsequent services from starting.

**Resolution:** P5-3 proposed dependency declaration. The deeper fix is async startup:

```go
type AsyncStartable interface {
    OnStartup(ctx context.Context) <-chan Result  // returns immediately, signals when ready
}
```

Services that can start in parallel do so. Services with dependencies wait for their dependencies' channels. This is a v0.9.0+ consideration — v0.8.0 keeps synchronous startup with insertion-order guarantee (P4-1 fix).

### P13-6. The Conclave Is Static — No Runtime Service Addition

Services are registered during `core.New()`. After `LockApply()`, no more services. There's no:
- Hot-loading a new service at runtime
- Plugin system that discovers services from a directory
- Upgrading a service without restart

Section 20 introduced `Registry.Seal()` (update existing, no new keys) for Action hot-reload. But service hot-reload is a harder problem — a service has state, connections, goroutines.

**Resolution:** For v0.8.0, the conclave is static. Accept this. For v0.9.0+, consider:
- `c.Registry("services").Replace("brain", newBrainSvc)` — swap implementation
- The old service gets `OnShutdown`, the new one gets `OnStartup`
- Handlers registered by the old service are replaced

### P13-7. IPC Is Instant — No Persistence or Replay

Messages are dispatched and forgotten. There's no:
- Message queue (messages sent while no handler exists are lost)
- Replay (can't re-process past events)
- Dead letter (failed messages aren't stored)

If a handler is registered AFTER a message was sent, it never sees that message. P4-4 documented this for the clone-and-iterate pattern. But it's a deeper assumption — IPC is ephemeral.

**Resolution:** For event sourcing or replay, use a persistent store (OpenBrain, database) alongside IPC. Core's IPC is real-time coordination, not a message queue. Document this boundary.

### P13-8. The RFC Assumes Adversarial Review — But Who Reviews The RFC?

This RFC was written by one author (Cladius) reviewing one codebase in one session. The findings are thorough but single-perspective. The 96 findings (now 104) all come from the same analytical lens.

What's missing:
- Performance benchmarking (how fast is IPC dispatch actually?)
- Real-world failure data (which bugs have users hit?)
- Comparative analysis (how does Core compare to Wire, Fx, Dig?)
- User research (what do service authors actually struggle with?)

**Resolution:** The RFC is a living document. Each v0.8.x patch adds a finding. Each consumer that struggles adds a P5-type consumer experience finding. The passes continue — they're not a one-time audit but a continuous process.

The meta-assumption: this RFC is complete. It's not. It's the best single-session analysis possible. The next session starts at Pass Fourteen.

---

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

**Resolution for v0.9.0+:** Entitlements, not CoreView. The boundary system already exists in CorePHP (RFC-004: Entitlements). Port it:

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

**Resolution:** The Action/Task system (Section 18) is the fix. Actions execute with concurrency control. Tasks define parallel/sequential composition. The IPC bus stops being the execution engine — it becomes the notification channel. PERFORM replaces ACTION for request/response. Async is opt-in per Action, not per handler.

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
| 5 | No boundaries (14) | Accept for v0.8.0, design CoreView for v0.9.0 — **deferred** |

Root causes 1-3 are fixable. Root cause 4 is mitigable. Root cause 5 is a v0.9.0 architecture change.

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

## Versioning

### Release Model

```
v0.7.x  — current stable (the API this RFC documents)
v0.7.*  — mechanical fixes to align code with RFC spec (issues 2,3,13,14,15)
v0.8.0  — production stable: all 16 issues resolved, Sections 17-20 implemented
v0.8.*  — patches only: each patch is a process gap to investigate
v0.9.0  — next design cycle (repeat: RFC spec → implement → stabilise → tag)
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
| All 16 Known Issues resolved in code | 15/16 resolved in RFC, 1 blocked (Issue 7) |
| Section 17: c.Process() primitive | Spec'd, needs go-process v0.7.0 |
| Section 18: Action/Task system | Spec'd, needs implementation |
| Section 19: c.API() streams | Spec'd, needs implementation |
| Section 20: Registry[T] primitive | Spec'd, needs implementation |
| AX-7 test coverage at 100% | core/go at 14%, core/agent at 92% |
| RFC-025 compliance | All code examples match implementation |
| Zero os/exec in consumer packages | core/agent done, go-process is the only allowed user |
| AGENTS.md + llm.txt on all repos | core/go, core/agent, core/docs done |

### What Does NOT Block v0.8.0

- core/cli v0.7.0 update (extension, not primitive)
- Borg/DataNode integration (separate ecosystem)
- CorePHP/CoreTS alignment (different release cycles)
- Full ecosystem AX-7 coverage (core/go + core/agent are the reference)

## 21. Entitlement — The Permission Primitive (Design)

> Status: Design spec. Brings v0.9.0 boundary model into v0.8.0.
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
- **Does not enforce at Config/Data/Fs level** — v0.8.0 gates Actions only. Config/Data/Fs gating is v0.9.0+ (requires CoreView or scoped Core).

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

### 21.13 Implementation Plan

```
1. Add Entitlement struct to contract.go (DTO)
2. Add EntitlementChecker type to contract.go
3. Add entitlementChecker field to Core struct
4. Add defaultChecker (always permitted)
5. Add c.Entitled() method
6. Add c.SetEntitlementChecker() method
7. Add c.RecordUsage() method (delegates to checker service)
8. Add NearLimit() / UsagePercent() convenience methods
9. Wire into Action.Run() — enforcement point
10. AX-7 tests: Good (permitted), Bad (denied), Ugly (no checker, quantity, near-limit)
11. Update RFC-025 with entitlement pattern
```

Zero new dependencies. ~100 lines of code. The entire permission model for the ecosystem.

---

## Changelog

- 2026-03-25: Added Section 21 — Entitlement primitive design. Bridges RFC-004 (SaaS feature gating), RFC-005 (Commerce Matrix hierarchy), and Core Actions into one permission primitive.
- 2026-03-25: Implementation session — Plans 1-5 complete. 456 tests, 84.4% coverage, 100% AX-7 naming. See RFC.plan.md "What Was Shipped" section.
- 2026-03-25: Pass Three — 8 spec contradictions (P3-1 through P3-8). Lifecycle returns, Process/Action mismatch, getter inconsistency, dual-purpose methods, error leaking, Data overlap, Action error model, Registry lock modes.

- 2026-03-25: Pass Three — 8 spec contradictions (P3-1 through P3-8). Lifecycle returns, Process/Action mismatch, getter inconsistency, dual-purpose methods, error leaking, Data overlap, Action error model, Registry lock modes.
- 2026-03-25: Pass Two — 8 architectural findings (P2-1 through P2-8)
- 2026-03-25: Added versioning model + v0.8.0 requirements
- 2026-03-25: Resolved all 16 Known Issues. Added Section 20 (Registry).
- 2026-03-25: Added Section 19 — API/Stream remote transport primitive
- 2026-03-25: Added Known Issues 9-16 (ADHD brain dump recovery — CommandLifecycle, Array[T], ConfigVar[T], Ipc struct, Lock allocation, Startables/Stoppables, stale comment, task.go concerns)
- 2026-03-25: Added Section 18 — Action and Task execution primitives
- 2026-03-25: Added Section 17 — c.Process() primitive spec
- 2026-03-25: Added Design Philosophy + Known Issues 1-8
- 2026-03-25: Initial specification — matches v0.7.0 implementation
