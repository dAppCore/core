# Go-Blockchain Modernisation Design

## Goal

Modernise `forge.lthn.ai/core/go-blockchain` from a standalone binary with stdlib `flag` and bare goroutines into a proper `core-chain` binary using `cli.Main()`, DI services, and go-process daemon lifecycle.

## Architecture

The refactor migrates go-blockchain to the standard Core CLI patterns without touching internal blockchain logic (chain/, consensus/, crypto/, wire/, types/, etc.). The P2P sync loop and wallet scanner become `core.Service` implementations managed by the DI container, with the sync service optionally running as a go-process daemon in headless mode.

## Current State

- **Entry point**: `cmd/chain/main.go` — stdlib `flag`, direct `store.New()`, bare `go syncLoop()`, inline `frame.Run()`
- **Stale go.mod**: Replace directives point to `/home/claude/Code/core/*` (different machine)
- **No build config**: No `.core/build.yaml`, not in `go.work`
- **Dependencies**: core/cli, core/go-p2p, core/go-store, bubbletea, testify, x/crypto

## Design

### 1. Entry Point: `cli.Main()` Migration

New `cmd/core-chain/main.go`:

```go
package main

import (
    "forge.lthn.ai/core/cli/pkg/cli"
    blockchain "forge.lthn.ai/core/go-blockchain"
)

func main() {
    cli.Main(
        cli.WithCommands("chain", blockchain.AddChainCommands),
    )
}
```

`AddChainCommands()` registers subcommands on the `chain` parent cobra command:

| Subcommand | Description |
|------------|-------------|
| `chain explorer` | TUI block explorer (current default mode) |
| `chain sync` | Headless P2P sync (daemon-capable) |
| `chain mine` | Mining (existing mining/ package) |

Persistent flags on `chain` parent: `--data-dir`, `--seed`, `--testnet`.

Old `cmd/chain/` directory is removed after migration.

### 2. SyncService as `core.Service`

Wraps the current `syncLoop`/`syncOnce` logic:

```go
type SyncService struct {
    *core.ServiceRuntime[SyncServiceOptions]
}

type SyncServiceOptions struct {
    DataDir  string
    Seed     string
    Testnet  bool
    Chain    *chain.Chain
}

func (s *SyncService) OnStartup(ctx context.Context) error {
    go s.syncLoop(ctx)
    return nil
}

func (s *SyncService) OnShutdown() error { return nil }
```

In headless mode (`core-chain chain sync`), runs as a **go-process Daemon**:
- PID file at `~/.core/daemons/core-chain-sync.pid`
- Auto-registered in daemon registry
- `core-chain chain sync --stop` to halt

In TUI mode (`core-chain chain explorer`), SyncService starts as a background service within the same process — no daemon wrapper needed.

### 3. WalletService as `core.Service`

Same pattern as SyncService. Wraps existing `wallet/` scanner logic. Only instantiated when wallet-related subcommands are invoked.

### 4. DI Container Wiring

```go
func AddChainCommands(parent *cobra.Command) {
    // Parse persistent flags, create store, chain
    // Wire services into core.New():
    c, _ := core.New(
        core.WithService(NewSyncService),
        core.WithService(NewWalletService),
    )
}
```

The `chain.Chain` instance and `store.Store` are created once and injected into services via the DI container — no globals.

### 5. Wire Protocol Extraction (Deferred)

The `wire/` package is generic Levin binary serialization, reusable by go-p2p and potentially other modules. Candidate for extraction to `forge.lthn.ai/core/go-wire` as a separate module.

**Deferred to a follow-up task** — initial refactor keeps wire/ in-tree.

### 6. Build & Workspace Integration

- **`.core/build.yaml`**: `binary: core-chain`, targets: `darwin/arm64`, `linux/amd64`
- **go.mod**: Fix replace directives from `/home/claude/Code/core/*` to `/Users/snider/Code/core/*`
- **go.work**: Add `./go-blockchain` entry
- **Build**: `core build` or `go build -o ./bin/core-chain ./cmd/core-chain`

### 7. Unchanged Packages

These packages contain blockchain domain logic and are not touched:

- `chain/` — block storage, sync algorithm, validation
- `consensus/` — proof-of-work consensus rules
- `crypto/` — cryptographic primitives
- `difficulty/` — difficulty adjustment
- `mining/` — block mining
- `types/` — block, transaction, header types
- `wire/` — binary serialization (extraction deferred)
- `p2p/` — Levin protocol encoding (called from SyncService)
- `tui/` — TUI models (wired into `cli.NewFrame("HCF")` from explorer subcommand)
- `config/` — network configs, hard forks, client version
- `wallet/` — wallet scanning (wrapped by WalletService)
- `rpc/` — RPC client types

## Binary

Standalone `core-chain` binary (not integrated into main `core` binary). Rationale: go-blockchain pulls in x/crypto and potentially CGo-dependent dependencies that shouldn't bloat the core CLI.

## Key References

| File | Role |
|------|------|
| `cmd/chain/main.go` | Current entry point (to be replaced) |
| `chain/sync.go` | Sync logic (to be wrapped by SyncService) |
| `wallet/scanner.go` | Wallet scanner (to be wrapped by WalletService) |
| `tui/` | TUI models (rewired to explorer subcommand) |
| `core/cli/pkg/cli/main.go` | `cli.Main()` pattern |
| `core/go-process/daemon.go` | Daemon lifecycle |
