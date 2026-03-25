# RFC Plan — How to Work With This Spec

> For future Claude sessions. Read this FIRST before touching code.

## What Exists

- `docs/RFC.md` — 3,845-line API spec with 108 findings across 13 passes
- `docs/RFC.implementation.{1-6}.md` — ordered implementation plans
- `llm.txt` — agent entry point
- `CLAUDE.md` — session-specific instructions

## The 108 Findings Reduce to 5 Root Causes

1. **Type erasure** (16 findings) — `Result{Value: any}` loses compile-time safety. Mitigate with typed methods + AX-7 tests. Not fixable without abandoning Result.

2. **No internal boundaries** (14 findings) — `*Core` grants God Mode. Solved by porting RFC-004 (Entitlements) from CorePHP. v0.9.0 work.

3. **Synchronous everything** (12 findings) — IPC dispatch blocks. ACTION cascade in core/agent blocks queue for minutes. Fixed by Action/Task system (Plan 3).

4. **No recovery path** (10 findings) — `os.Exit` bypasses defer. No cleanup on failure. Fixed by Plan 1 (defer + RunE + panic recovery).

5. **Missing primitives** (8 findings) — No ID, validation, health, atomic writes. Fixed by Plan 5.

## Implementation Order

```
Plan 1 → v0.7.1 (ship immediately, zero breakage)
Plan 2 → Registry[T] (foundation — Plans 3-4 depend on this)
Plan 3 → Action/Task (execution primitive — Plan 4 depends on this)
Plan 4 → c.Process() (needs go-process v0.7.0 update first)
Plan 5 → Missing primitives + AX-7 (independent, do alongside 2-4)
Plan 6 → Ecosystem sweep (after 1-5, dispatched via Codex)
```

## 3 Critical Bugs — Fix First

1. **P4-3:** `ipc.go` — ACTION handler returning `!OK` stops entire broadcast chain. Other handlers never fire. Fix: call all handlers, don't stop on failure.

2. **P6-1:** core/agent `handlers.go` — Nested `c.ACTION()` calls create synchronous cascade 4 levels deep. QA → PR → Verify → Merge blocks Poke handler for minutes. Queue doesn't drain. Fix: replace with Task pipeline (needs Plan 3).

3. **P7-2:** `core.go` — `Run()` calls `os.Exit(1)` on startup failure without calling `ServiceShutdown()`. Running services leak. Fix: add `defer c.ServiceShutdown()` + replace `os.Exit` with error return.

## Key Design Decisions Already Made

- **CamelCase = primitive** (brick), **UPPERCASE = convenience** (sugar)
- **Core is Lego bricks** — export the bricks, hide the safety mechanisms
- **Fs.root is the ONE exception** — security boundaries stay unexported
- **Registration IS permission** — no handler = no capability
- **`error` at Go interface boundary, `Result` at Core contract boundary**
- **Dual-purpose methods** (Service, Command, Action) — keep as sugar, Registry has explicit Get/Set
- **Array[T] and ConfigVar[T] are guardrail primitives** — model-proof, not speculative
- **ServiceRuntime[T] and manual `.core = c` are both valid** — document both
- **Startable V2 returns Result** — add alongside V1 for backwards compat
- **`RunE()` alongside `Run()`** — no breakage

## Existing RFCs That Solve Open Problems

| Problem | RFC | Core Provides | Consumer Implements |
|---------|-----|---------------|-------------------|
| Permissions | RFC-004 Entitlements | `c.Entitlement()` interface | go-entitlements package |
| Config context | RFC-003 Config Channels | `c.Config()` with channel | config channel service |
| Secrets | RFC-012 SMSG | `c.Secret()` interface | go-smsg / env fallback |
| Validation | RFC-009 Sigil | Transform chain interface | validator implementations |
| Containers | RFC-014 TIM | `c.Fs()` sandbox | TIM = OS isolation |
| In-memory fs | RFC-013 DataNode | `c.Data()` mounts fs.FS | DataNode / Borg |
| Lazy startup | RFC-002 Event Modules | Event declaration | Lazy instantiation |

Core stays stdlib-only. Consumers bring implementations via WithService.

## What NOT to Do

- Don't add dependencies to core/go (it's stdlib + go-io + go-log only)
- Don't use `os/exec` — go-process is the only allowed user (P9-1: core/go itself violates this in app.go — fix it)
- Don't use `unsafe.Pointer` on Core types — add legitimate APIs instead
- Don't call `os.Exit` inside Core — return errors, let main() exit
- Don't create global mutable state — use Core's Registry
- Don't auto-discover via reflect — use explicit registration (HandleIPCEvents is the last magic method)

## AX-7 Status

- core/agent: 92% (840 tests, 79.9% coverage)
- core/go: 14% (83.6% coverage but wrong naming — needs rename + gap fill)
- Rename script exists (Python, used on core/agent — same script works)
- 212 functions × 3 categories = 636 target for core/go

## Session Context That Won't Be In Memory

- The ACTION cascade (P6-1) is the root cause of "agents finish but queue doesn't drain"
- status.json has 51 unprotected read-modify-write sites (P4-9) — real race condition
- The Fs sandbox is bypassed by 2 files using unsafe.Pointer (P11-2)
- `core.Env("DIR_HOME")` is cached at init — `t.Setenv` doesn't override it (P2-5)
- go-process `NewService` returns `(any, error)` not `core.Result` — needs v0.7.0 update
- Multiple Core instances share global state (assetGroups, systemInfo, defaultLog)
