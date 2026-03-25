# RFC Plan 2 — Registry + Actions Sessions

> After Plan 1 bugs are fixed and AX-7 rename is done.

## Session Goal: Registry[T] + First Migration

1. Build `registry.go` with full AX-7 tests
2. Migrate `serviceRegistry` → `ServiceRegistry` embedding `Registry[*Service]`
3. Verify all existing tests still pass
4. Commit + push

## Session Goal: Action System

1. Rename `task.go` → `action.go`
2. Move `RegisterAction`/`RegisterActions`/`RegisterTask` to `ipc.go`
3. Build `ActionDef` type with `Run()`, `Exists()`, `Def()`
4. Wire `c.Action("name")` dual-purpose accessor
5. Full AX-7 tests
6. Commit + push

## Session Goal: Migrate core/agent Handlers

1. Register named Actions in `agentic.Register()`
2. Replace nested `c.ACTION()` cascade with Task pipeline
3. Test that queue drains properly after agent completion
4. This is the P6-1 fix — the queue starvation bug

## Session Goal: c.Process() + go-process v0.7.0

1. Update go-process factory to return `core.Result`
2. Add `process.Register` direct factory
3. Remove `agentic.ProcessRegister` bridge
4. Add `Process` primitive to core/go (sugar over Actions)
5. Migrate core/agent `proc.go` → `s.core.Process()` calls
6. Delete `proc.go` and `ensureProcess()`

## Between Sessions

Each session should produce:
- Working code (all tests pass)
- A commit with conventional message
- Updated coverage numbers
- Any new findings added to RFC.md passes
