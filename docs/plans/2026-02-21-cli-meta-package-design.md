# CLI Meta-Package Restructure — Design

**Goal:** Transform `core/cli` from a 35K LOC monolith into a thin assembly repo that ships variant binaries. Domain repos own their commands. `go/pkg/cli` is the only import any domain package needs for CLI concerns.

**Architecture:** Self-registration via `init()` + `cli.RegisterCommands()` (existing pattern, already works). Command code moves from `cli/cmd/*` into domain repos. The cli repo becomes a collection of `main.go` files — each variant blank-imports the domain `cmd` packages it needs.

**Tech Stack:** go/pkg/cli (wraps cobra + charmbracelet), Go workspaces, Taskfile

---

## 1. CLI SDK — The Single Import

`forge.lthn.ai/core/go/pkg/cli` is the **only** import domain packages use for CLI concerns. It wraps cobra, charmbracelet, and stdlib behind a stable API. If the underlying libraries change, only `go/pkg/cli` is touched — every domain repo is insulated.

### Already done (keep as-is)

- **Cobra:** `Command` type alias, `NewCommand()`, `NewGroup()`, `NewRun()`, flag helpers (`StringFlag`, `BoolFlag`, `IntFlag`, `StringSliceFlag`), arg validators
- **Output:** `Success()`, `Error()`, `Warn()`, `Info()`, `Table`, `Section()`, `Label()`, `Task()`, `Hint()`
- **Prompts:** `Confirm()`, `Question()`, `Choose()`, `ChooseMulti()` with grammar-based action variants
- **Styles:** 17 pre-built styles, `AnsiStyle` builder, Tailwind colour constants (47 hex values)
- **Glyphs:** `:check:`, `:cross:`, `:warn:` etc. with Unicode/Emoji/ASCII themes
- **Layout:** HLCRF composite renderer (Header/Left/Content/Right/Footer)
- **Errors:** `Wrap()`, `WrapVerb()`, `ExitError`, `Is()`, `As()`
- **Logging:** `LogDebug()`, `LogInfo()`, `LogWarn()`, `LogError()`, `LogSecurity()`

### New — TUI primitives (charmbracelet under the hood)

Domain packages call these; the charm dependency stays inside `go/pkg/cli`.

- `Spinner(message string) *SpinnerHandle` — async spinner with `.Update(msg)`, `.Done()`, `.Fail()`
- `ProgressBar(total int) *ProgressHandle` — progress bar with `.Increment()`, `.SetMessage(msg)`, `.Done()`
- `List(items []string, opts ...ListOption) (string, error)` — interactive scrollable list selection
- `TextInput(prompt string, opts ...InputOption) (string, error)` — styled single-line text input
- `Viewport(content string, opts ...ViewportOption) error` — scrollable content pane (for long output)
- `RunTUI(model Model) error` — escape hatch for complex interactive UIs (wraps `tea.Model`)

### Stubbed for later (interface exists, returns simple fallback)

- `Form(fields []FormField) (map[string]string, error)` — multi-field form (backed by huh later)
- `FilePicker(opts ...FilePickerOption) (string, error)` — file browser
- `Tabs(items []TabItem) error` — tabbed content panes

### Rule

Domain packages import `forge.lthn.ai/core/go/pkg/cli` and **nothing else** for CLI concerns. No `cobra`, no `lipgloss`, no `bubbletea`. The 34 files in cli/ that currently import cobra directly get rewritten to use `cli.*` helpers during migration.

---

## 2. Domain-Owned Commands

Each domain repo exports its commands via the existing self-registration pattern. The command code moves out of `cli/cmd/*` into the domain repo that owns the business logic.

### Package layout in domain repos

```
go-ml/
├── cmd/                    # CLI commands (self-registering)
│   ├── cmd.go              # init() + AddMLCommands(root)
│   ├── cmd_score.go
│   ├── cmd_chat.go
│   └── ...
├── service.go              # existing business logic
└── go.mod
```

### The contract

```go
// go-ml/cmd/cmd.go
package cmd

import "forge.lthn.ai/core/go/pkg/cli"

func init() {
    cli.RegisterCommands(AddMLCommands)
}

func AddMLCommands(root *cli.Command) {
    mlCmd := cli.NewGroup("ml", "ML inference and training", "")
    root.AddCommand(mlCmd)
    addScoreCommand(mlCmd)
    addChatCommand(mlCmd)
    // ...
}
```

### Migration mapping

| Current location | Destination | Files |
|-----------------|-------------|-------|
| `cmd/ml` | `go-ml/cmd/` | 40 |
| `cmd/ai` | `go-agent/cmd/` | 10 |
| `cmd/dev` | `go-devops/cmd/` | 20 |
| `cmd/forge` | `go-scm/cmd/` | 12 |
| `cmd/gitea` | `go-scm/cmd/` | 7 |
| `cmd/collect` | `go-scm/cmd/` | 8 |
| `cmd/security` | `go-devops/cmd/` | 7 |
| `cmd/deploy` | `go-devops/cmd/` | 3 |
| `cmd/prod` | `go-devops/cmd/` | 7 |
| `cmd/setup` | `go-devops/cmd/` | 14 |
| `cmd/go` | `go-devops/cmd/` | 8 |
| `cmd/qa` | `go-devops/cmd/` | 6 |
| `cmd/test` | `go-devops/cmd/` | 5 |
| `cmd/vm` | `go-devops/cmd/` | 4 |
| `cmd/monitor` | `go-devops/cmd/` | — |
| `cmd/crypt` | `go-crypt/cmd/` | 5 |
| `cmd/rag` | `go-rag/cmd/` | 5 |
| `cmd/unifi` | `go-netops/cmd/` | 7 |
| `cmd/api` | `go-api/cmd/` | 4 |
| `cmd/session` | `go-session/cmd/` | 1 |
| `cmd/gitcmd` | `go-git/cmd/` | 1 |
| `cmd/mcpcmd` | `go-ai/cmd/` | 1 |

### Stays in cli/ (meta/framework commands)

These are CLI-specific concerns, not domain logic:

`config`, `workspace`, `doctor`, `help`, `updater`, `daemon`, `lab`, `module`, `pkgcmd`, `plugin`, `docs`, `vanity-import`

---

## 3. Variant Binaries

The cli/ repo becomes a build assembly point. Each variant is a `main.go` that blank-imports the command packages it needs.

### Directory layout

```
cli/
├── cmd/
│   ├── core/main.go           # Full CLI — everything
│   ├── core-ci/main.go        # CI agent dispatch + SCM
│   ├── core-mlx/main.go       # ML inference subprocess
│   ├── core-ops/main.go       # DevOps + infra management
│   └── core-gui/main.go       # Wails desktop app
├── cmd/                        # Meta commands that stay in cli/
│   ├── config/
│   ├── doctor/
│   ├── help/
│   ├── updater/
│   └── ...
├── go.mod
├── go.work
└── Taskfile.yaml
```

### Variant definitions

**core** (full kitchen sink):
```go
package main

import (
    "forge.lthn.ai/core/go/pkg/cli"

    // Meta commands (local to cli/)
    _ "forge.lthn.ai/core/cli/cmd/config"
    _ "forge.lthn.ai/core/cli/cmd/doctor"
    _ "forge.lthn.ai/core/cli/cmd/help"
    _ "forge.lthn.ai/core/cli/cmd/updater"
    _ "forge.lthn.ai/core/cli/cmd/workspace"

    // Domain commands (self-register from domain repos)
    _ "forge.lthn.ai/core/go-ml/cmd"
    _ "forge.lthn.ai/core/go-agent/cmd"
    _ "forge.lthn.ai/core/go-ai/cmd"
    _ "forge.lthn.ai/core/go-devops/cmd"
    _ "forge.lthn.ai/core/go-scm/cmd"
    _ "forge.lthn.ai/core/go-crypt/cmd"
    _ "forge.lthn.ai/core/go-rag/cmd"
    _ "forge.lthn.ai/core/go-netops/cmd"
    _ "forge.lthn.ai/core/go-api/cmd"
    _ "forge.lthn.ai/core/go-git/cmd"
    _ "forge.lthn.ai/core/go-session/cmd"
)

func main() { cli.Main() }
```

**core-ci** (lightweight CI agent):
```go
package main

import (
    "forge.lthn.ai/core/go/pkg/cli"
    _ "forge.lthn.ai/core/cli/cmd/config"
    _ "forge.lthn.ai/core/go-agent/cmd"
    _ "forge.lthn.ai/core/go-scm/cmd"
    _ "forge.lthn.ai/core/go-devops/cmd"
)

func main() { cli.Main() }
```

**core-mlx** (ML inference as external process):
```go
package main

import (
    "forge.lthn.ai/core/go/pkg/cli"
    _ "forge.lthn.ai/core/cli/cmd/config"
    _ "forge.lthn.ai/core/go-ml/cmd"
)

func main() { cli.Main() }
```

**core-ops** (infra management):
```go
package main

import (
    "forge.lthn.ai/core/go/pkg/cli"
    _ "forge.lthn.ai/core/cli/cmd/config"
    _ "forge.lthn.ai/core/go-devops/cmd"
    _ "forge.lthn.ai/core/go-scm/cmd"
    _ "forge.lthn.ai/core/go-netops/cmd"
)

func main() { cli.Main() }
```

### Taskfile

```yaml
tasks:
  build:all:
    cmds:
      - go build -o bin/core ./cmd/core
      - go build -o bin/core-ci ./cmd/core-ci
      - go build -o bin/core-mlx ./cmd/core-mlx
      - go build -o bin/core-ops ./cmd/core-ops

  build:core:
    cmds: [go build -o bin/core ./cmd/core]

  build:ci:
    cmds: [go build -o bin/core-ci ./cmd/core-ci]

  build:mlx:
    cmds: [go build -o bin/core-mlx ./cmd/core-mlx]

  build:ops:
    cmds: [go build -o bin/core-ops ./cmd/core-ops]
```

### Why variants matter

- `core-mlx` ships to the homelab as a ~10MB binary, not 50MB with devops/forge/netops
- `core-ci` deploys to agent machines without ML or CGO dependencies
- Other packages use `exec.Command("core-mlx", "serve")` to consume heavy subsystems as external processes rather than linking them in
- Adding a new variant = one new `main.go` with the right blank imports

---

## 4. Migration Order

Gradual migration, largest packages first, cli/ works at every step. Each phase is one session's worth of work.

### Phase 0: CLI SDK expansion (prerequisite)

Extend `go/pkg/cli` with charmbracelet TUI wrappers (Spinner, ProgressBar, List, TextInput, Viewport, RunTUI). Stub Form, FilePicker, Tabs. Ensure all `cli.*` helpers cover what the 34 direct-cobra files need. This unblocks all subsequent phases.

### Phase 1: cmd/ml → go-ml/cmd/ (40 files)

The ML pipeline is the largest command package and the primary candidate for the `core-mlx` variant. Moving it out proves the pattern and shrinks cli/ by a third.

### Phase 2: cmd/ai → go-agent/cmd/ (10 files)

AgentCI dispatch, task management. Natural fit — go-agent already has the orchestration logic. Unblocks `core-ci` variant.

### Phase 3: cmd/forge + cmd/gitea + cmd/collect → go-scm/cmd/ (27 files)

All three use go-scm packages directly. Bundle into one move since they share the same domain repo.

### Phase 4: cmd/dev + cmd/deploy + cmd/prod + cmd/setup + cmd/security + cmd/go + cmd/qa + cmd/test + cmd/vm + cmd/monitor → go-devops/cmd/ (74 files)

All ops/infra/dev tooling belongs in go-devops. Can split across multiple sessions if needed. Unblocks `core-ops` variant.

### Phase 5: Small moves (one session, batch them)

- `cmd/crypt` → `go-crypt/cmd/` (5 files)
- `cmd/rag` → `go-rag/cmd/` (5 files)
- `cmd/unifi` → `go-netops/cmd/` (7 files)
- `cmd/api` → `go-api/cmd/` (4 files, mostly done)
- `cmd/session` → `go-session/cmd/` (1 file)
- `cmd/gitcmd` → `go-git/cmd/` (1 file)
- `cmd/mcpcmd` → `go-ai/cmd/` (1 file)

### Phase 6: Variant assembly

Create `cmd/core/main.go`, `cmd/core-ci/main.go`, `cmd/core-mlx/main.go`, `cmd/core-ops/main.go`. Update Taskfile. The current root `main.go` becomes `cmd/core/main.go`. Old `cli/cmd/*` directories that moved out get deleted.

### Per-phase checklist

1. Copy files to domain repo's `cmd/`
2. Rewrite any direct cobra imports → `cli.*` helpers
3. `go test ./...` in domain repo
4. Update cli's `main.go` blank import from `cli/cmd/X` → `go-X/cmd`
5. Delete old `cli/cmd/X`
6. `go test ./...` in cli
7. Commit and push both repos

### End state

cli/ has ~12 meta packages, ~5 variant `main.go` files, and zero business logic. Everything else lives in the domain repos that own it. Total cli/ LOC drops from ~35K to ~2K.
