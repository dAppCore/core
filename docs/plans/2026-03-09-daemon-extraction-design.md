# Daemon Process Management Extraction — Design

## Goal

Move daemon runtime primitives from `core/cli` and daemon CLI commands from `go-ai` into their natural homes: runtime types into `go-process`, generic CLI commands into `core/cli` as a reusable command builder.

## Problem

Daemon lifecycle management is scattered across three repos:

| Repo | What it has | Problem |
|------|-------------|---------|
| `cli/pkg/cli/daemon.go` | PIDFile, HealthServer, Daemon, Mode detection | Process management primitives don't belong in a CLI/TUI library |
| `go-ai/cmd/daemon/cmd.go` | start/stop/status/run CLI + MCP foreground | Generic daemon commands hardcoded to one consumer |
| `go-process` | Process spawning, signals, output, DAG runner | Missing self-as-daemon management |

## Design

### go-process gains daemon runtime types

New files in `go-process` root package:

- **`daemon.go`** — `Daemon`, `DaemonOptions`, `Mode`, `DetectMode()`
- **`pidfile.go`** — `PIDFile` (acquire, release, read, stale detection)
- **`health.go`** — `HealthServer`, `HealthCheck` type

These are standalone types alongside `Process` and `Runner`. A `Daemon` manages *this process* as a long-running service. A `Process` manages *child processes*. Same domain, complementary concerns.

**Types extracted from `cli/pkg/cli/daemon.go`:**

```go
// Mode represents how the process was launched.
type Mode int
const (
    ModeInteractive Mode = iota // TTY attached
    ModePipe                     // stdin/stdout piped
    ModeDaemon                   // Detached background
)

// PIDFile manages a PID lock file for daemon processes.
type PIDFile struct { ... }
func NewPIDFile(path string) *PIDFile
func (p *PIDFile) Acquire() error
func (p *PIDFile) Release() error
func (p *PIDFile) Read() (int, bool)  // Read PID + check if running

// HealthServer provides HTTP /health and /ready endpoints.
type HealthServer struct { ... }
func NewHealthServer(addr string) *HealthServer
func (h *HealthServer) AddCheck(check HealthCheck)
func (h *HealthServer) SetReady(ready bool)
func (h *HealthServer) Start() error
func (h *HealthServer) Stop(ctx context.Context) error

// Daemon orchestrates PIDFile + HealthServer + signal handling.
type Daemon struct { ... }
type DaemonOptions struct {
    PIDFile         string
    ShutdownTimeout time.Duration
    HealthAddr      string
    HealthChecks    []HealthCheck
    OnReload        func()
}
func NewDaemon(opts DaemonOptions) *Daemon
func (d *Daemon) Start() error
func (d *Daemon) Run(ctx context.Context) error
func (d *Daemon) Stop() error
func (d *Daemon) SetReady(ready bool)
```

**New helper from `go-ai/cmd/daemon/cmd.go`:**

```go
// WaitForHealth polls a health endpoint until it responds OK or timeout.
func WaitForHealth(addr string, timeout time.Duration) bool
```

### core/cli gains generic daemon CLI commands

New file `cmd/daemon/cmd.go` in `core/cli`:

```go
// DaemonCommandConfig configures the generic daemon CLI commands.
type DaemonCommandConfig struct {
    RunForeground func(ctx context.Context) error  // Business logic callback
    PIDFile       string   // Default PID file path
    HealthAddr    string   // Default health address
}

// AddDaemonCommand registers start/stop/status/run subcommands.
func AddDaemonCommand(root *cli.Command, cfg DaemonCommandConfig)
```

Subcommands:
- **`start`** — Re-exec binary as detached process, wait for health
- **`stop`** — Read PID file, send SIGTERM, wait for exit
- **`status`** — Check PID file + health endpoint, display status
- **`run`** — Run foreground (calls `cfg.RunForeground`), manages Daemon lifecycle

All commands import `go-process` for PIDFile, Daemon, WaitForHealth.

### go-ai shrinks

`go-ai/cmd/daemon/` deleted entirely. Registration becomes:

```go
import daemon "forge.lthn.ai/core/cli/cmd/daemon"

daemon.AddDaemonCommand(root, daemon.DaemonCommandConfig{
    RunForeground: func(ctx context.Context) error {
        svc := mcp.New(mcp.WithSubsystem(...))
        return startMCP(ctx, svc, cfg)
    },
    PIDFile:    cfg.PIDFile,
    HealthAddr: cfg.HealthAddr,
})
```

`startMCP()` stays in go-ai — it's MCP-specific business logic.

### Deletion

- `cli/pkg/cli/daemon.go` — deleted (types move to go-process)
- `go-ai/cmd/daemon/cmd.go` — deleted (commands move to cli, MCP wiring stays in go-ai)

### What doesn't change

- go-process existing API (Process, Runner, RunSpec, Service, exec/) — untouched
- go-process core/go dependency — already present
- Any other consumer of cli/daemon.go gets updated to import from go-process

## Dependencies

```
go-process (gains daemon types, no new deps)
    └── core/go (existing)

core/cli (gains daemon commands)
    └── go-process (new — for PIDFile, Daemon, WaitForHealth)

go-ai (shrinks)
    └── core/cli (existing — now uses cli's daemon commands)
```

## Testing

- go-process: Unit tests for PIDFile, HealthServer, Daemon, WaitForHealth, Mode detection (ported from cli + go-ai tests)
- core/cli: Integration tests for daemon commands (mock RunForeground callback)
- go-ai: Verify MCP wiring still works after refactor
