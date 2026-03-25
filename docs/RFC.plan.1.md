# RFC Plan 1 — First Session Priorities (COMPLETED 2026-03-25)

> All items shipped. See RFC.plan.md "What Was Shipped" section.

## Priority 1: Fix the 3 Critical Bugs (Plan 1)

These are one-line to five-line changes. Ship as v0.7.1.

### Bug 1: ACTION stops on !OK (ipc.go line ~33)

```go
// CURRENT (broken — handler 3 failing silences handlers 4 and 5):
for _, h := range handlers {
    if r := h(c, msg); !r.OK { return r }
}

// FIX:
for _, h := range handlers {
    func() {
        defer func() { if r := recover(); r != nil { Error("handler panic", "err", r) } }()
        h(c, msg)
    }()
}
```

This also fixes P7-3 (no panic recovery) in the same change.

### Bug 2: Run() leaks on startup failure (core.go Run method)

Add one line:
```go
func (c *Core) Run() {
    defer c.ServiceShutdown(context.Background())  // ADD THIS
    // ... rest unchanged
}
```

### Bug 3: Remove stale Embed() and fix comment

Delete `func (c *Core) Embed() Result` from core.go.
Fix the `New()` comment to show `*Core` return.

### Test all 3 with AX-7 naming:
```
TestIpc_Action_Ugly_HandlerFailsChainContinues
TestIpc_Action_Ugly_HandlerPanicsChainContinues
TestCore_Run_Ugly_StartupFailureCallsShutdown
```

## Priority 2: AX-7 Rename for core/go

Run the same Python rename script used on core/agent:

```python
# Same script from core/agent session — applies to any Go package
# Changes TestFoo_Good to TestFile_Foo_Good
```

This is mechanical. No logic changes. Just naming.

Then run gap analysis:
```bash
python3 -c "... same gap analysis script ..."
```

## Priority 3: Start Registry[T] (Plan 2)

Create `registry.go` with the type. Write tests FIRST (AX-7 complete from day one):

```
TestRegistry_Set_Good
TestRegistry_Set_Bad
TestRegistry_Set_Ugly
TestRegistry_Get_Good
...
```

Then migrate `serviceRegistry` first (most tested, most used).

## What Was Skipped (shipped in same session instead)

All items originally marked "skip" were shipped because Registry and Actions were built in the same session:
- Plan 3 (Actions) — DONE: ActionDef, TaskDef, c.Action(), c.Task()
- Plan 4 (Process) — DONE for core/go: c.Process() sugar over Actions
- Breaking changes — DONE: Startable returns Result, CommandLifecycle removed
