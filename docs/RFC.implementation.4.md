# Implementation Plan 4 — c.Process() Primitive (Section 17)

> Depends on: Plan 2 (Registry), Plan 3 (Actions).
> go-process v0.7.0 update required.

## Phase A: Update go-process to v0.7.0 Core API

**Repo:** core/go-process

1. Change `NewService` factory to return `core.Result` (not `(any, error)`)
2. Add `Register` function matching `WithService` signature
3. Remove `ProcessRegister` bridge from core/agent
4. Register process Actions during OnStartup:

```go
func (s *Service) OnStartup(ctx context.Context) core.Result {
    c := s.Core()
    c.Action("process.run", s.handleRun)
    c.Action("process.start", s.handleStart)
    c.Action("process.kill", s.handleKill)
    return core.Result{OK: true}
}
```

## Phase B: Add Process Primitive to core/go

**New file in core/go:** `process.go`

```go
type Process struct {
    core *Core
}

func (c *Core) Process() *Process { return c.process }

func (p *Process) Run(ctx context.Context, command string, args ...string) Result {
    return p.core.Action("process.run").Run(ctx, NewOptions(
        Option{Key: "command", Value: command},
        Option{Key: "args", Value: args},
    ))
}

func (p *Process) RunIn(ctx context.Context, dir, command string, args ...string) Result {
    return p.core.Action("process.run").Run(ctx, NewOptions(
        Option{Key: "command", Value: command},
        Option{Key: "args", Value: args},
        Option{Key: "dir", Value: dir},
    ))
}
```

No new dependencies in core/go — Process is just Action sugar.

## Phase C: Migrate core/agent proc.go

Replace standalone helpers with Core methods:

```go
// Current: proc.go standalone with ensureProcess() hack
out, err := runCmd(ctx, dir, "git", "log")

// Target: method on PrepSubsystem using Core
out := s.core.Process().RunIn(ctx, dir, "git", "log")
```

Delete `proc.go`. Delete `ensureProcess()`. Delete `ProcessRegister`.

## Phase D: Replace syscall.Kill calls

5 call sites in core/agent use `syscall.Kill(pid, 0)` and `syscall.Kill(pid, SIGTERM)`.

Replace with:
```go
processIsRunning(st.ProcessID, st.PID)  // already in proc.go
processKill(st.ProcessID, st.PID)       // already in proc.go
```

These use `process.Get(id).IsRunning()` when ProcessID is available.

## Phase E: Atomic File Writes (P4-10)

Add to core/go `fs.go`:

```go
func (m *Fs) WriteAtomic(p, content string) Result {
    tmp := p + ".tmp." + strconv.FormatInt(time.Now().UnixNano(), 36)
    if r := m.Write(tmp, content); !r.OK { return r }
    if err := os.Rename(tmp, m.path(p)); err != nil {
        m.Delete(tmp)
        return Result{err, false}
    }
    return Result{OK: true}
}
```

Migrate `writeStatus` in core/agent to use `WriteAtomic`. Fixes P4-9 (51 race sites).

## Resolves

I7 (no c.Process), P9-1 (os/exec in core), P4-9 (status.json race), P4-10 (Fs.Write not atomic), P11-2 (unsafe.Pointer Fs bypass — add Fs.NewUnrestricted instead).
