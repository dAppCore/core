# Implementation Plan 6 — Ecosystem Sweep (Phase 3)

> Depends on: Plans 1-5 complete. 44 repos need updating.
> Use Codex dispatch with RFC as the spec.

## Breaking Changes

### StartableV2 (if P3-1 adopted)

```go
type StartableV2 interface {
    OnStartup(ctx context.Context) Result
}
```

Core checks V2 first, falls back to V1. No breakage. Deprecation notice on V1.

**Codex template:** "If this repo implements OnStartup returning error, add a V2 variant returning core.Result alongside it."

### RunE() exists alongside Run()

No breakage — Run() still works. New code uses RunE().

### Command.Managed replaces CommandLifecycle

```go
// Old:
Command{Lifecycle: myLifecycle}

// New:
Command{Managed: "process.daemon"}
```

**Codex template:** "If this repo uses Command.Lifecycle, replace with Managed field."

## Per-Repo Sweep Checklist

For each of 44 repos:

1. [ ] Update `go.mod` to core v0.8.0
2. [ ] Replace `OnStartup() error` with `OnStartup() Result` (if V2 adopted)
3. [ ] Replace `os/exec` with `c.Process()` calls
4. [ ] Replace manual `unsafe.Pointer` Fs hacks with `Fs.NewUnrestricted()`
5. [ ] Run AX-7 gap analysis, fill missing Good/Bad/Ugly
6. [ ] Rename tests to `TestFile_Function_{Good,Bad,Ugly}`
7. [ ] Add `llm.txt` and `AGENTS.md`
8. [ ] Verify `go test ./...` passes
9. [ ] Create PR to dev

## Dispatch Strategy

```
Batch 1 (core packages — we control):
  core/go, core/agent, core/mcp, go-process, go-io, go-log

Batch 2 (utility packages):
  go-config, go-git, go-scm, go-cache, go-store, go-session

Batch 3 (domain packages):
  go-forge, go-api, go-build, go-devops, go-container

Batch 4 (application packages):
  gui, cli, ts, php, ide, lint

Batch 5 (research/experimental):
  go-ai, go-ml, go-mlx, go-inference, go-rag, go-rocm
```

Each batch is one Codex sweep. Verify batch N before starting N+1.

## Success Criteria for v0.8.0

- [ ] All 3 critical bugs fixed (Plan 1)
- [ ] Registry[T] implemented and migrated (Plan 2)
- [ ] Action system working (Plan 3 Phase A-C)
- [ ] c.Process() primitive working (Plan 4 Phase A-C)
- [ ] Missing primitives added (Plan 5)
- [ ] core/go AX-7 at 100%
- [ ] core/agent AX-7 at 100%
- [ ] Zero os/exec in core/go
- [ ] Zero unsafe.Pointer on Core types in ecosystem
- [ ] All Phase 1-2 changes shipped
- [ ] Phase 3 sweep at least started (batch 1-2)
