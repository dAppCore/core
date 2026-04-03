# Specification Mismatches

## Scope
Findings are mismatches between current repository source behavior and existing docs/spec pages under `docs/`.

### 1) `docs/getting-started.md` uses deprecated constructor pattern
- Example and prose show `core.New(core.Options{...})` and say constructor reads only the first `core.Options`.
- Current code uses variadic `core.New(...CoreOption)` only; passing `core.Options` requires `core.WithOptions(core.NewOptions(...))`.
- References: `docs/getting-started.md:18`, `docs/getting-started.md:26`, `docs/getting-started.md:142`.

### 2) `docs/testing.md` and `docs/configuration.md` repeat outdated constructor usage
- Both files document `core.New(core.Options{...})` examples.
- Current constructor is variadic `CoreOption` values.
- References: `docs/testing.md:29`, `docs/configuration.md:16`.

### 3) `docs/lifecycle.md` claims registry order is map-backed and unstable
- File states `Startables()/Stoppables()` are built from a map-backed registry and therefore non-deterministic.
- Current `Registry` stores an explicit insertion-order slice and iterates in insertion order.
- References: `docs/lifecycle.md:64-67`.

### 4) `docs/services.md` stale ordering and lock-name behavior
- Claims registry is map-backed; actual behavior is insertion-order iteration.
- States default service lock name is `"srv"`, but `LockEnable`/`LockApply` do not expose/use a default namespace in implementation.
- References: `docs/services.md:53`, `docs/services.md:86-88`.

### 5) `docs/commands.md` documents removed managed lifecycle field
- Section “Lifecycle Commands” shows `Lifecycle` field with `Start/Stop/Restart/Reload/Signal` callbacks.
- Current `Command` struct has `Managed string` and no `Lifecycle` field.
- References: `docs/commands.md:155-159`.

### 6) `docs/subsystems.md` documents legacy options creation call for subsystem registration
- Uses `c.Data().New(core.Options{...})` and `c.Drive().New(core.Options{...})`.
- `Data.New` and `Drive.New` expect `core.Options` via varargs usage helpers (`core.NewOptions` in current docs/usage pattern).
- References: `docs/subsystems.md:44`, `docs/subsystems.md:75`, `docs/subsystems.md:80`.

### 7) `docs/index.md` RFC summary is stale
- Claims `docs/RFC.md` is 21 sections, 1476 lines, but current RFC content has expanded sections/size.
- Reference: `docs/index.md` table header note.
