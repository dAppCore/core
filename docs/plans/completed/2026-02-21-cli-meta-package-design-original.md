# CLI Meta-Package Restructure — Design

**Goal:** Transform `core/cli` from a 35K LOC monolith into a thin assembly repo that ships variant binaries. Domain repos own their commands. `go/pkg/cli` is the only import any domain package needs for CLI concerns.

**Architecture:** Commands register as framework services via `cli.WithCommands()`, passed to `cli.Main()`. Command code lives in the domain repos that own the business logic. The cli repo is a thin `main.go` that wires them together.

**Tech Stack:** go/pkg/cli (wraps cobra + charmbracelet), Core framework lifecycle, Taskfile

---

## 1. CLI SDK — The Single Import

`forge.lthn.ai/core/go/pkg/cli` is the **only** import domain packages use for CLI concerns. It wraps cobra, charmbracelet, and stdlib behind a stable API. If the underlying libraries change, only `go/pkg/cli` is touched — every domain repo is insulated.

### Already done

- **Cobra:** `Command` type alias, `NewCommand()`, `NewGroup()`, `NewRun()`, flag helpers (`StringFlag`, `BoolFlag`, `IntFlag`, `StringSliceFlag`), arg validators
- **Output:** `Success()`, `Error()`, `Warn()`, `Info()`, `Table`, `Section()`, `Label()`, `Task()`, `Hint()`
- **Prompts:** `Confirm()`, `Question()`, `Choose()`, `ChooseMulti()` with grammar-based action variants
- **Styles:** 17 pre-built styles, `AnsiStyle` builder, Tailwind colour constants (47 hex values)
- **Glyphs:** `:check:`, `:cross:`, `:warn:` etc. with Unicode/Emoji/ASCII themes
- **Layout:** HLCRF composite renderer (Header/Left/Content/Right/Footer)
- **Errors:** `Wrap()`, `WrapVerb()`, `ExitError`, `Is()`, `As()`
- **Logging:** `LogDebug()`, `LogInfo()`, `LogWarn()`, `LogError()`, `LogSecurity()`
- **TUI primitives:** `Spinner`, `ProgressBar`, `InteractiveList`, `TextInput`, `Viewport`, `RunTUI`
- **Command registration:** `WithCommands(name, fn)` — registers commands as framework services

### Stubbed for later (interface exists, returns simple fallback)

- `Form(fields []FormField) (map[string]string, error)` — multi-field form (backed by huh later)
- `FilePicker(opts ...FilePickerOption) (string, error)` — file browser
- `Tabs(items []TabItem) error` — tabbed content panes

### Rule

Domain packages import `forge.lthn.ai/core/go/pkg/cli` and **nothing else** for CLI concerns. No `cobra`, no `lipgloss`, no `bubbletea`.

---

## 2. Command Registration — Framework Lifecycle

Commands register through the Core framework's service lifecycle, not through global state or `init()` functions.

### The contract

Each domain repo exports an `Add*Commands(root *cli.Command)` function. The CLI binary wires it in via `cli.WithCommands()`:

```go
// go-ai/cmd/daemon/cmd.go
package daemon

import "forge.lthn.ai/core/go/pkg/cli"

// AddDaemonCommand adds the 'daemon' command group to the root.
func AddDaemonCommand(root *cli.Command) {
    daemonCmd := cli.NewGroup("daemon", "Manage the core daemon", "")
    root.AddCommand(daemonCmd)
    // subcommands...
}
```

No `init()`. No blank imports. No `cli.RegisterCommands()`.

### How it works

`cli.WithCommands(name, fn)` wraps the registration function as a framework service implementing `Startable`. During `Core.ServiceStartup()`, the service's `OnStartup()` casts `Core.App` to `*cobra.Command` and calls the registration function. Core services (i18n, log, workspace) start first since they're registered before command services.

```go
// cli/main.go
func main() {
    cli.Main(
        cli.WithCommands("config", config.AddConfigCommands),
        cli.WithCommands("doctor", doctor.AddDoctorCommands),
        // ...
    )
}
```

### Migration status (completed)

| Source | Destination | Status |
|--------|-------------|--------|
| `cmd/dev, setup, qa, docs, gitcmd, monitor` | `go-devops/cmd/` | Done |
| `cmd/lab` | `go-ai/cmd/` | Done |
| `cmd/workspace` | `go-agentic/cmd/` | Done |
| `cmd/go` | `core/go/cmd/gocmd` | Done |
| `cmd/vanity-import, community` | `go-devops/cmd/` | Done |
| `cmd/updater` | `go-update` | Done (own repo) |
| `cmd/daemon, mcpcmd, security` | `go-ai/cmd/` | Done |
| `cmd/crypt` | `go-crypt/cmd/` | Done |
| `cmd/rag` | `go-rag/cmd/` | Done |
| `cmd/unifi` | `go-netops/cmd/` | Done |
| `cmd/api` | `go-api/cmd/` | Done |
| `cmd/collect, forge, gitea` | `go-scm/cmd/` | Done |
| `cmd/deploy, prod, vm` | `go-devops/cmd/` | Done |

### Stays in cli/ (meta/framework commands)

`config`, `doctor`, `help`, `module`, `pkgcmd`, `plugin`, `session`

---

## 3. Variant Binaries (future)

The cli/ repo can produce variant binaries by creating multiple `main.go` files that wire different sets of commands.

```
cli/
├── main.go                    # Current — meta commands only
├── cmd/core-full/main.go      # Full CLI — all ecosystem commands
├── cmd/core-ci/main.go        # CI agent dispatch + SCM
├── cmd/core-mlx/main.go       # ML inference subprocess
└── cmd/core-ops/main.go       # DevOps + infra management
```

Each variant calls `cli.Main()` with its specific `cli.WithCommands()` set. No blank imports needed.

### Why variants matter

- `core-mlx` ships to the homelab as a ~10MB binary, not 50MB with devops/forge/netops
- `core-ci` deploys to agent machines without ML or CGO dependencies
- Adding a new variant = one new `main.go` with the right `WithCommands` calls

---

## 4. Current State

cli/ has 7 meta packages, one `main.go`, and zero business logic. Everything else lives in the domain repos that own it. Total cli/ LOC is ~2K.
