# Implementation Plan 1 — Critical Bug Fixes (Phase 1)

> Ship as v0.7.1. Zero consumer breakage. Fix the 3 critical bugs.

## P4-3: ACTION !OK Stops Broadcast Chain

**File:** `ipc.go`
**Change:** ACTION dispatch must call ALL handlers regardless of individual results.

```go
// Current (broken):
for _, h := range handlers {
    if r := h(c, msg); !r.OK {
        return r  // STOPS — remaining handlers never called
    }
}

// Fix:
for _, h := range handlers {
    h(c, msg)  // call all, ignore individual results for broadcast
}
```

**Test:** `TestIpc_Action_Ugly` — register 3 handlers, second returns !OK, verify third still fires.

## P7-2: No Cleanup on Startup Failure

**File:** `core.go` — `Run()`
**Change:** Call ServiceShutdown before exit on startup failure.

```go
// Current:
if !r.OK { os.Exit(1) }

// Fix:
if !r.OK {
    c.ServiceShutdown(context.Background())
    os.Exit(1)
}
```

## P7-3: ACTION Handlers No Panic Recovery

**File:** `ipc.go`
**Change:** Wrap each handler in defer/recover.

```go
for _, h := range handlers {
    func() {
        defer func() {
            if r := recover(); r != nil {
                Error("ACTION handler panicked", "panic", r)
            }
        }()
        h(c, msg)
    }()
}
```

**Test:** `TestIpc_Action_Ugly` — handler that panics, verify other handlers still execute.

## P7-4: Run() Needs defer ServiceShutdown

**File:** `core.go`
**Change:** Add defer as first line of Run.

```go
func (c *Core) Run() {
    defer c.ServiceShutdown(context.Background())
    // ... rest unchanged, but remove os.Exit calls
}
```

## Additional Phase 1 (safe)

- **I3:** Remove `Embed()` accessor (0 consumers)
- **I15:** Fix stale comment on `New()` — update to show `*Core` return
- **P9-1:** Remove `os/exec` import from `app.go` — move `App.Find()` to go-process
