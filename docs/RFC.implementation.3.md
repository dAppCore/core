# Implementation Plan 3 — Action/Task System (Section 18)

> Depends on: Plan 2 (Registry). The execution primitive.

## Phase A: Action Registry

**New file:** `action.go` (renamed from `task.go`)

```go
type ActionDef struct {
    Name        string
    Handler     ActionHandler
    Description string
    Schema      Options        // expected input keys
    enabled     bool           // for Disable()
}

type ActionHandler func(context.Context, Options) Result
```

**Core accessor:**

```go
// Dual-purpose (like Service):
c.Action("process.run", handler)           // register
c.Action("process.run").Run(ctx, opts)     // invoke
c.Action("process.run").Exists()           // check
c.Action("process.run").Def()              // metadata
```

Internally uses `Registry[*ActionDef]`.

## Phase B: Move Registration to IPC

Move from `task.go`:
- `RegisterAction` → `ipc.go` (registers in `Registry[ActionHandler]`)
- `RegisterActions` → `ipc.go`
- `RegisterTask` → `ipc.go`

Keep in `action.go`:
- `Perform` → becomes `c.Action("name").Run()`
- `PerformAsync` → becomes `c.Action("name").RunAsync()`
- `Progress` → stays

## Phase C: Panic Recovery on All Actions

Every `Action.Run()` wraps in recover:

```go
func (a *ActionDef) Run(ctx context.Context, opts Options) (result Result) {
    defer func() {
        if r := recover(); r != nil {
            result = Result{E("action.Run", Sprint("panic: ", r), nil), false}
        }
    }()
    return a.Handler(ctx, opts)
}
```

## Phase D: Task Composition (v0.8.0 stretch)

```go
type TaskDef struct {
    Name  string
    Steps []Step
}

type Step struct {
    Action string
    With   Options
    Async  bool     // run in background
    Input  string   // "previous" = output of last step
}
```

`c.Task("name", def)` registers. `c.Task("name").Run(ctx)` executes steps in order.

## Resolves

I16 (task.go concerns), P6-1 (cascade → Task pipeline), P6-2 (O(N×M) → direct dispatch), P7-3 (panic recovery), P7-8 (circuit breaker via Disable), P10-2 (PERFORM not ACTION for request/response).

## Migration in core/agent

After Actions exist, refactor `handlers.go`:

```go
// Current: nested c.ACTION() cascade
// Target:
c.Task("agent.completion", TaskDef{
    Steps: []Step{
        {Action: "agentic.qa"},
        {Action: "agentic.auto-pr"},
        {Action: "agentic.verify"},
        {Action: "agentic.ingest", Async: true},
        {Action: "agentic.poke", Async: true},
    },
})
```
