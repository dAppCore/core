# Daemon Registry & Project Manifest — Design

## Goal

Unified `core start/stop/list` for managing multiple background daemons, driven by a `.core/manifest.yaml` project identity file. Runtime state tracked in `~/.core/daemons/`. Release snapshots frozen as `core.json` for marketplace indexing.

## Problem

Daemon management is per-binary with no central awareness. Each consumer (go-ai, core-ide, go-html gallery, LEM chat, blockchain node) picks its own PID path and health port. No way to:
- List all running daemons across projects
- Start a project's services from one command
- Let the marketplace know what a project can run

## Design

### Manifest Schema (go-scm/manifest)

`.core/manifest.yaml` is the project identity file. Extends the existing manifest (UI layout, permissions, modules) with a `daemons` section:

```yaml
code: photo-browser
name: Photo Browser
version: 0.1.0
description: Browse and serve local photo collections

daemons:
  serve:
    binary: core-php
    args: [php, serve]
    health: "127.0.0.1:0"
    default: true
  worker:
    binary: core-mlx
    args: [worker, start]
    health: "127.0.0.1:0"

layout: HLCRF
slots:
  C: photo-grid
  L: folder-tree
permissions:
  read: [./photos/]
modules: [core/media, core/fs]
```

Fields per daemon entry:
- `binary` — executable name (auto-detected if omitted)
- `args` — arguments passed to the binary
- `health` — health check address (port 0 = dynamic)
- `default` — marks the daemon that `core start` runs with no args

### Runtime Registry (go-process)

When a daemon starts, a registration file is written to `~/.core/daemons/`. Removed on stop.

File naming: `{code}-{daemon}.json`

```go
type DaemonEntry struct {
    Code    string    `json:"code"`
    Daemon  string    `json:"daemon"`
    PID     int       `json:"pid"`
    Health  string    `json:"health"`
    Project string    `json:"project"`
    Binary  string    `json:"binary"`
    Started time.Time `json:"started"`
}

type Registry struct { ... }

func NewRegistry() *Registry
func (r *Registry) Register(entry DaemonEntry) error
func (r *Registry) Unregister(code, daemon string) error
func (r *Registry) List() ([]DaemonEntry, error)
func (r *Registry) Get(code, daemon string) (*DaemonEntry, bool)
```

`List()` checks each entry's PID — if dead, removes the stale file and skips it.

`Daemon.Start()` gains an optional registry hook — auto-registers on Start, auto-unregisters on Stop.

### CLI Commands (cli)

Top-level commands, not under `core daemon`:

**`core start [daemon-name]`**
1. Find `.core/manifest.yaml` in cwd or parent dirs
2. Parse manifest, find daemon entry (default if no name given)
3. Check registry — already running? say so
4. Exec the declared binary with args, detached
5. Wait for health, register in `~/.core/daemons/`

**`core stop [daemon-name]`**
1. With name: look up in registry by code+name, send SIGTERM
2. Without name: stop all daemons for current project's code
3. Unregister on exit

**`core list`**
1. Scan `~/.core/daemons/`, prune stale, print table

```
CODE             DAEMON    PID    HEALTH              PROJECT
photo-browser    serve     48291  127.0.0.1:54321     /Users/snider/Code/photo-app
core             mcp       51002  127.0.0.1:9101      -
core-ide         headless  51100  127.0.0.1:9878      -
```

**`core restart [daemon-name]`** — stop then start.

### Release Snapshot (go-devops)

`core build release` generates `core.json` at repo root from `.core/manifest.yaml`:

```json
{
  "schema": 1,
  "code": "photo-browser",
  "name": "Photo Browser",
  "version": "0.1.0",
  "commit": "a1b2c3d4e5f6a1b2c3d4e5f6a1b2c3d4e5f6a1b2",
  "tag": "v0.1.0",
  "built": "2026-03-09T15:00:00Z",
  "daemons": { ... },
  "layout": "HLCRF",
  "slots": { ... },
  "permissions": { ... },
  "modules": [ ... ]
}
```

Differences from YAML manifest:
- `schema` version for forward compatibility
- `commit` — full SHA, immutable
- `tag` — release tag
- `built` — timestamp

Marketplace indexes `core.json` from tagged releases on forge. Self-describing package listings, no manual catalogue updates.

## Where Code Lives

| Component | Repo | Why |
|-----------|------|-----|
| `DaemonEntry`, `Registry` | go-process | Runtime state, alongside PIDFile/Daemon |
| Manifest schema (+ `daemons`) | go-scm/manifest | Already owns the manifest type |
| `core start/stop/list` | cli | Top-level CLI commands |
| `core.json` generation | go-devops | Part of release pipeline |
| Marketplace indexing | go-scm/marketplace | Already owns the catalogue |

## Dependencies

```
go-process (gains Registry, no new deps)

cli (gains start/stop/list commands)
    └── go-process (existing — Registry, ReadPID)
    └── go-scm/manifest (new — parse .core/manifest.yaml)

go-devops (gains core.json generation)
    └── go-scm/manifest (new — read manifest for snapshot)
```

## What Doesn't Change

- Existing `core daemon start/stop/status` in go-ai stays as MCP-specific consumer
- Existing `Daemon`, `PIDFile`, `HealthServer` in go-process — untouched
- Existing `DaemonCommandConfig` / `AddDaemonCommand` in cli — untouched
- go-scm plugin system (plugin.json, registry.json) — separate concern

## Testing

- go-process: Unit tests for Registry (register, unregister, list, stale pruning)
- go-scm: Unit tests for extended manifest parsing (daemons section)
- cli: Integration tests for start/stop/list with mock manifest
- go-devops: Unit test for core.json generation from manifest
