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

### 2. MustServiceFor Uses Panic (Resolved)

```go
func MustServiceFor[T any](c *Core, name string) T {
    panic(...)
}
```

**Resolution:** Keep but document clearly. `Must` prefix is a Go convention that signals "this panics." In the guardrail context (Issue 10/11), `MustServiceFor` is valid for startup-time code where a missing service means the app can't function. The alternative — `ServiceFor` + `if !ok` + manual error handling — adds LOC that contradicts the Result philosophy.

**Rule:** Use `MustServiceFor` only in `OnStartup` / init paths where failure is fatal. Never in request handlers or runtime code. Document this in RFC-025.

### 3. Embed() Legacy Accessor (Resolved)

```go
func (c *Core) Embed() Result { return c.data.Get("app") }
```

**Resolution:** Remove. It's a shortcut to `c.Data().Get("app")` with a misleading name. `Embed` sounds like it embeds something — it actually reads. An agent seeing `c.Embed()` can't know it's reading from Data. Dead code, remove in next refactor.

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

### 5. RegisterAction Lives in task.go (Resolved by Issue 16)

Resolved — task.go splits into ipc.go (registration) + action.go (execution). See Issue 16.

### 6. serviceRegistry Is Unexported (Resolved by Section 20)

Resolved — `serviceRegistry` becomes `ServiceRegistry` embedding `Registry[*Service]`. See Section 20.

### 7. No c.Process() Accessor

Spec'd in Section 17. Blocked on go-process v0.7.0 update.

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

## Changelog

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
