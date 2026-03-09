# Go-Blockchain Modernisation Implementation Plan

> **For Claude:** REQUIRED SUB-SKILL: Use superpowers:executing-plans to implement this plan task-by-task.

**Goal:** Modernise go-blockchain from a standalone flag-based binary into a proper `core-chain` binary using `cli.Main()`, DI services, and go-process daemon lifecycle.

**Architecture:** The refactor wraps existing blockchain logic (chain/, consensus/, crypto/, wire/, etc.) in Core framework patterns without modifying domain code. The P2P sync loop becomes a `core.Service` with optional daemon mode via go-process. The CLI entry point migrates from stdlib `flag` to `cli.Main()` + `cli.WithCommands()`. A new `AddChainCommands()` registration function provides `explorer`, `sync`, and `mine` subcommands.

**Tech Stack:** Go 1.26, `forge.lthn.ai/core/cli` (cobra + bubbletea), `forge.lthn.ai/core/go` (DI container), `forge.lthn.ai/core/go-process` (daemon lifecycle), `forge.lthn.ai/core/go-store` (SQLite), `forge.lthn.ai/core/go-p2p` (Levin protocol)

---

### Task 1: Fix go.mod and add to go.work

**Context:** The go.mod has stale replace directives pointing to `/home/claude/Code/core/*` (a different machine). These need to point to `/Users/snider/Code/core/*` for local workspace resolution. The module also needs to be added to go.work.

**Files:**
- Modify: `/Users/snider/Code/core/go-blockchain/go.mod:59-67`
- Modify: `/Users/snider/Code/go.work`

**Step 1: Fix replace directives in go.mod**

Open `/Users/snider/Code/core/go-blockchain/go.mod` and replace all `/home/claude/Code/core/` paths with `/Users/snider/Code/core/`:

```
replace forge.lthn.ai/core/cli => /Users/snider/Code/core/cli

replace forge.lthn.ai/core/go => /Users/snider/Code/core/go

replace forge.lthn.ai/core/go-crypt => /Users/snider/Code/core/go-crypt

replace forge.lthn.ai/core/go-p2p => /Users/snider/Code/core/go-p2p

replace forge.lthn.ai/core/go-store => /Users/snider/Code/core/go-store
```

**Step 2: Add go-blockchain to go.work**

```bash
cd /Users/snider/Code && go work use ./core/go-blockchain
```

**Step 3: Verify the module resolves**

```bash
cd /Users/snider/Code/core/go-blockchain && GOWORK=/Users/snider/Code/go.work go build ./...
```

Expected: Build succeeds (existing code compiles).

**Step 4: Commit**

```bash
cd /Users/snider/Code/core/go-blockchain
git add go.mod
git commit -m "fix: update go.mod replace directives for local workspace"
```

Also commit go.work change:
```bash
cd /Users/snider/Code
git add go.work
git commit -m "chore: add go-blockchain to workspace"
```

---

### Task 2: Create AddChainCommands registration function

**Context:** This is the core of the migration. Instead of `main()` directly creating everything, we create a `AddChainCommands(root *cobra.Command)` function that registers a `chain` parent command with persistent flags and subcommands. The `chain` parent command holds shared state (data dir, seed, testnet flag, chain config).

**Files:**
- Create: `/Users/snider/Code/core/go-blockchain/commands.go`
- Test: `/Users/snider/Code/core/go-blockchain/commands_test.go`

**Step 1: Write the test**

Create `/Users/snider/Code/core/go-blockchain/commands_test.go`:

```go
package blockchain

import (
	"testing"

	"github.com/spf13/cobra"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestAddChainCommands_Good_RegistersParent(t *testing.T) {
	root := &cobra.Command{Use: "test"}
	AddChainCommands(root)

	// Should have a "chain" subcommand
	chainCmd, _, err := root.Find([]string{"chain"})
	require.NoError(t, err)
	assert.Equal(t, "chain", chainCmd.Name())
}

func TestAddChainCommands_Good_HasSubcommands(t *testing.T) {
	root := &cobra.Command{Use: "test"}
	AddChainCommands(root)

	chainCmd, _, _ := root.Find([]string{"chain"})

	// Should have explorer, sync, mine subcommands
	var names []string
	for _, sub := range chainCmd.Commands() {
		names = append(names, sub.Name())
	}
	assert.Contains(t, names, "explorer")
	assert.Contains(t, names, "sync")
}

func TestAddChainCommands_Good_PersistentFlags(t *testing.T) {
	root := &cobra.Command{Use: "test"}
	AddChainCommands(root)

	chainCmd, _, _ := root.Find([]string{"chain"})

	// Should have persistent flags
	assert.NotNil(t, chainCmd.PersistentFlags().Lookup("data-dir"))
	assert.NotNil(t, chainCmd.PersistentFlags().Lookup("seed"))
	assert.NotNil(t, chainCmd.PersistentFlags().Lookup("testnet"))
}
```

**Step 2: Run test to verify it fails**

```bash
cd /Users/snider/Code/core/go-blockchain && GOWORK=/Users/snider/Code/go.work go test -run TestAddChainCommands -v .
```

Expected: FAIL — `AddChainCommands` not defined.

**Step 3: Write the implementation**

Create `/Users/snider/Code/core/go-blockchain/commands.go`:

```go
// Copyright (c) 2017-2026 Lethean (https://lt.hn)
//
// Licensed under the European Union Public Licence (EUPL) version 1.2.
// SPDX-License-Identifier: EUPL-1.2

package blockchain

import (
	"fmt"
	"os"
	"path/filepath"

	"forge.lthn.ai/core/go-blockchain/config"
	"github.com/spf13/cobra"
)

// AddChainCommands registers the "chain" command group with explorer,
// sync, and mine subcommands.
func AddChainCommands(root *cobra.Command) {
	var (
		dataDir string
		seed    string
		testnet bool
	)

	chainCmd := &cobra.Command{
		Use:   "chain",
		Short: "Lethean blockchain node",
		Long:  "Manage the Lethean blockchain — sync, explore, and mine.",
	}

	chainCmd.PersistentFlags().StringVar(&dataDir, "data-dir", defaultDataDir(), "blockchain data directory")
	chainCmd.PersistentFlags().StringVar(&seed, "seed", "seeds.lthn.io:36942", "seed peer address (host:port)")
	chainCmd.PersistentFlags().BoolVar(&testnet, "testnet", false, "use testnet")

	chainCmd.AddCommand(
		newExplorerCmd(&dataDir, &seed, &testnet),
		newSyncCmd(&dataDir, &seed, &testnet),
	)

	root.AddCommand(chainCmd)
}

// resolveConfig returns the chain config and forks for the current network.
func resolveConfig(testnet bool, seed *string) (config.ChainConfig, []config.HardFork) {
	if testnet {
		if *seed == "seeds.lthn.io:36942" {
			*seed = "localhost:46942"
		}
		return config.Testnet, config.TestnetForks
	}
	return config.Mainnet, config.MainnetForks
}

func defaultDataDir() string {
	home, err := os.UserHomeDir()
	if err != nil {
		return ".lethean"
	}
	return filepath.Join(home, ".lethean", "chain")
}

// ensureDataDir creates the data directory if it doesn't exist.
func ensureDataDir(dataDir string) error {
	if err := os.MkdirAll(dataDir, 0o755); err != nil {
		return fmt.Errorf("create data dir: %w", err)
	}
	return nil
}
```

**Step 4: Run tests to verify they pass**

```bash
cd /Users/snider/Code/core/go-blockchain && GOWORK=/Users/snider/Code/go.work go test -run TestAddChainCommands -v .
```

Expected: PASS (3 tests).

**Step 5: Commit**

```bash
cd /Users/snider/Code/core/go-blockchain
git add commands.go commands_test.go
git commit -m "feat: add AddChainCommands registration function"
```

---

### Task 3: Create explorer subcommand

**Context:** The explorer subcommand is the TUI block explorer — the current default mode of the binary. It creates a store, chain, node, TUI models, and runs `cli.NewFrame("HCF")`. This replaces the bulk of the current `main()`.

**Files:**
- Create: `/Users/snider/Code/core/go-blockchain/cmd_explorer.go`

**Step 1: Write the implementation**

Create `/Users/snider/Code/core/go-blockchain/cmd_explorer.go`:

```go
// Copyright (c) 2017-2026 Lethean (https://lt.hn)
//
// Licensed under the European Union Public Licence (EUPL) version 1.2.
// SPDX-License-Identifier: EUPL-1.2

package blockchain

import (
	"context"
	"log"
	"os"
	"os/signal"
	"path/filepath"

	cli "forge.lthn.ai/core/cli/pkg/cli"
	store "forge.lthn.ai/core/go-store"

	"forge.lthn.ai/core/go-blockchain/chain"
	"forge.lthn.ai/core/go-blockchain/tui"
	"github.com/spf13/cobra"
)

func newExplorerCmd(dataDir, seed *string, testnet *bool) *cobra.Command {
	return &cobra.Command{
		Use:   "explorer",
		Short: "TUI block explorer",
		Long:  "Interactive terminal block explorer with live sync status.",
		RunE: func(cmd *cobra.Command, args []string) error {
			return runExplorer(*dataDir, *seed, *testnet)
		},
	}
}

func runExplorer(dataDir, seed string, testnet bool) error {
	if err := ensureDataDir(dataDir); err != nil {
		return err
	}

	dbPath := filepath.Join(dataDir, "chain.db")
	s, err := store.New(dbPath)
	if err != nil {
		log.Fatalf("open store: %v", err)
	}
	defer s.Close()

	c := chain.New(s)

	cfg, forks := resolveConfig(testnet, &seed)

	ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt)
	defer cancel()

	// Start P2P sync in background.
	go syncLoop(ctx, c, &cfg, forks, seed)

	node := tui.NewNode(c)
	status := tui.NewStatusModel(node)
	explorer := tui.NewExplorerModel(c)
	hints := tui.NewKeyHintsModel()

	frame := cli.NewFrame("HCF")
	frame.Header(status)
	frame.Content(explorer)
	frame.Footer(hints)
	frame.Run()

	return nil
}
```

**Step 2: Verify it compiles**

```bash
cd /Users/snider/Code/core/go-blockchain && GOWORK=/Users/snider/Code/go.work go build ./...
```

Expected: Build succeeds.

**Step 3: Commit**

```bash
cd /Users/snider/Code/core/go-blockchain
git add cmd_explorer.go
git commit -m "feat: add explorer subcommand (TUI block explorer)"
```

---

### Task 4: Create sync subcommand with daemon support

**Context:** The sync subcommand runs the P2P sync loop headless (no TUI). When `--daemon` is passed, it runs as a go-process Daemon with PID file and registry entry. `--stop` sends a signal to stop a running daemon.

**Files:**
- Create: `/Users/snider/Code/core/go-blockchain/cmd_sync.go`
- Create: `/Users/snider/Code/core/go-blockchain/sync_service.go`

**Step 1: Add go-process dependency**

```bash
cd /Users/snider/Code/core/go-blockchain && GOWORK=/Users/snider/Code/go.work go get forge.lthn.ai/core/go-process
```

Add replace directive to go.mod:
```
replace forge.lthn.ai/core/go-process => /Users/snider/Code/core/go-process
```

**Step 2: Write sync_service.go — the sync loop extracted from main.go**

Create `/Users/snider/Code/core/go-blockchain/sync_service.go`:

```go
// Copyright (c) 2017-2026 Lethean (https://lt.hn)
//
// Licensed under the European Union Public Licence (EUPL) version 1.2.
// SPDX-License-Identifier: EUPL-1.2

package blockchain

import (
	"context"
	"crypto/rand"
	"encoding/binary"
	"fmt"
	"log"
	"net"
	"time"

	"forge.lthn.ai/core/go-blockchain/chain"
	"forge.lthn.ai/core/go-blockchain/config"
	"forge.lthn.ai/core/go-blockchain/p2p"
	levin "forge.lthn.ai/core/go-p2p/node/levin"
)

// syncLoop continuously syncs the chain from the seed peer.
// It retries on error and polls every 30s when synced.
func syncLoop(ctx context.Context, c *chain.Chain, cfg *config.ChainConfig, forks []config.HardFork, seed string) {
	opts := chain.SyncOptions{
		VerifySignatures: false,
		Forks:            forks,
	}

	for {
		select {
		case <-ctx.Done():
			return
		default:
		}

		if err := syncOnce(ctx, c, cfg, opts, seed); err != nil {
			log.Printf("sync: %v (retrying in 10s)", err)
			select {
			case <-ctx.Done():
				return
			case <-time.After(10 * time.Second):
			}
			continue
		}

		// Synced — wait before polling again.
		select {
		case <-ctx.Done():
			return
		case <-time.After(30 * time.Second):
		}
	}
}

func syncOnce(ctx context.Context, c *chain.Chain, cfg *config.ChainConfig, opts chain.SyncOptions, seed string) error {
	conn, err := net.DialTimeout("tcp", seed, 10*time.Second)
	if err != nil {
		return fmt.Errorf("dial %s: %w", seed, err)
	}
	defer conn.Close()

	lc := levin.NewConnection(conn)

	var peerIDBuf [8]byte
	rand.Read(peerIDBuf[:])
	peerID := binary.LittleEndian.Uint64(peerIDBuf[:])

	localHeight, _ := c.Height()

	req := p2p.HandshakeRequest{
		NodeData: p2p.NodeData{
			NetworkID: cfg.NetworkID,
			PeerID:    peerID,
			LocalTime: time.Now().Unix(),
			MyPort:    0,
		},
		PayloadData: p2p.CoreSyncData{
			CurrentHeight:  localHeight,
			ClientVersion:  config.ClientVersion,
			NonPruningMode: true,
		},
	}
	payload, err := p2p.EncodeHandshakeRequest(&req)
	if err != nil {
		return fmt.Errorf("encode handshake: %w", err)
	}
	if err := lc.WritePacket(p2p.CommandHandshake, payload, true); err != nil {
		return fmt.Errorf("write handshake: %w", err)
	}

	hdr, data, err := lc.ReadPacket()
	if err != nil {
		return fmt.Errorf("read handshake: %w", err)
	}
	if hdr.Command != uint32(p2p.CommandHandshake) {
		return fmt.Errorf("unexpected command %d", hdr.Command)
	}

	var resp p2p.HandshakeResponse
	if err := resp.Decode(data); err != nil {
		return fmt.Errorf("decode handshake: %w", err)
	}

	localSync := p2p.CoreSyncData{
		CurrentHeight:  localHeight,
		ClientVersion:  config.ClientVersion,
		NonPruningMode: true,
	}
	p2pConn := chain.NewLevinP2PConn(lc, resp.PayloadData.CurrentHeight, localSync)

	return c.P2PSync(ctx, p2pConn, opts)
}
```

**Step 3: Write cmd_sync.go — the sync subcommand**

Create `/Users/snider/Code/core/go-blockchain/cmd_sync.go`:

```go
// Copyright (c) 2017-2026 Lethean (https://lt.hn)
//
// Licensed under the European Union Public Licence (EUPL) version 1.2.
// SPDX-License-Identifier: EUPL-1.2

package blockchain

import (
	"context"
	"fmt"
	"log"
	"os"
	"os/signal"
	"path/filepath"
	"syscall"

	"forge.lthn.ai/core/go-blockchain/chain"
	"forge.lthn.ai/core/go-process"
	store "forge.lthn.ai/core/go-store"
	"github.com/spf13/cobra"
)

func newSyncCmd(dataDir, seed *string, testnet *bool) *cobra.Command {
	var (
		daemon bool
		stop   bool
	)

	cmd := &cobra.Command{
		Use:   "sync",
		Short: "Headless P2P chain sync",
		Long:  "Sync the blockchain from P2P peers without the TUI explorer.",
		RunE: func(cmd *cobra.Command, args []string) error {
			if stop {
				return stopSyncDaemon(*dataDir)
			}
			if daemon {
				return runSyncDaemon(*dataDir, *seed, *testnet)
			}
			return runSyncForeground(*dataDir, *seed, *testnet)
		},
	}

	cmd.Flags().BoolVar(&daemon, "daemon", false, "run as background daemon")
	cmd.Flags().BoolVar(&stop, "stop", false, "stop a running sync daemon")

	return cmd
}

func runSyncForeground(dataDir, seed string, testnet bool) error {
	if err := ensureDataDir(dataDir); err != nil {
		return err
	}

	dbPath := filepath.Join(dataDir, "chain.db")
	s, err := store.New(dbPath)
	if err != nil {
		return fmt.Errorf("open store: %w", err)
	}
	defer s.Close()

	c := chain.New(s)
	cfg, forks := resolveConfig(testnet, &seed)

	ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer cancel()

	log.Println("Starting headless P2P sync...")
	syncLoop(ctx, c, &cfg, forks, seed)
	log.Println("Sync stopped.")
	return nil
}

func runSyncDaemon(dataDir, seed string, testnet bool) error {
	if err := ensureDataDir(dataDir); err != nil {
		return err
	}

	pidFile := filepath.Join(dataDir, "sync.pid")

	d := process.NewDaemon(process.DaemonOptions{
		PIDFile:  pidFile,
		Registry: process.DefaultRegistry(),
		RegistryEntry: process.DaemonEntry{
			Code:   "forge.lthn.ai/core/go-blockchain",
			Daemon: "sync",
		},
	})

	if err := d.Start(); err != nil {
		return fmt.Errorf("daemon start: %w", err)
	}

	dbPath := filepath.Join(dataDir, "chain.db")
	s, err := store.New(dbPath)
	if err != nil {
		_ = d.Stop()
		return fmt.Errorf("open store: %w", err)
	}
	defer s.Close()

	c := chain.New(s)
	cfg, forks := resolveConfig(testnet, &seed)

	ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer cancel()

	d.SetReady(true)
	log.Println("Sync daemon started.")

	// Run sync loop in a goroutine; daemon.Run blocks until signal.
	go syncLoop(ctx, c, &cfg, forks, seed)

	return d.Run(ctx)
}

func stopSyncDaemon(dataDir string) error {
	pidFile := filepath.Join(dataDir, "sync.pid")
	pid, err := process.ReadPID(pidFile)
	if err != nil {
		return fmt.Errorf("no running sync daemon found: %w", err)
	}

	proc, err := os.FindProcess(pid)
	if err != nil {
		return fmt.Errorf("find process %d: %w", pid, err)
	}

	if err := proc.Signal(syscall.SIGTERM); err != nil {
		return fmt.Errorf("signal process %d: %w", pid, err)
	}

	log.Printf("Sent SIGTERM to sync daemon (PID %d)", pid)
	return nil
}
```

**Step 4: Verify it compiles**

```bash
cd /Users/snider/Code/core/go-blockchain && GOWORK=/Users/snider/Code/go.work go build ./...
```

Expected: Build succeeds.

**Step 5: Commit**

```bash
cd /Users/snider/Code/core/go-blockchain
git add sync_service.go cmd_sync.go go.mod go.sum
git commit -m "feat: add sync subcommand with daemon support"
```

---

### Task 5: Create cmd/core-chain/main.go entry point

**Context:** The new standalone binary entry point. Uses `cli.WithAppName("core-chain")` and `cli.Main()` with `WithCommands()`.

**Files:**
- Create: `/Users/snider/Code/core/go-blockchain/cmd/core-chain/main.go`

**Step 1: Create the directory**

```bash
mkdir -p /Users/snider/Code/core/go-blockchain/cmd/core-chain
```

**Step 2: Write the entry point**

Create `/Users/snider/Code/core/go-blockchain/cmd/core-chain/main.go`:

```go
// Copyright (c) 2017-2026 Lethean (https://lt.hn)
//
// Licensed under the European Union Public Licence (EUPL) version 1.2.
// SPDX-License-Identifier: EUPL-1.2

package main

import (
	cli "forge.lthn.ai/core/cli/pkg/cli"
	blockchain "forge.lthn.ai/core/go-blockchain"
)

func main() {
	cli.WithAppName("core-chain")
	cli.Main(
		cli.WithCommands("chain", blockchain.AddChainCommands),
	)
}
```

**Step 3: Build the binary**

```bash
cd /Users/snider/Code/core/go-blockchain && GOWORK=/Users/snider/Code/go.work go build -o ./bin/core-chain ./cmd/core-chain
```

Expected: Produces `bin/core-chain` binary.

**Step 4: Verify help output**

```bash
./bin/core-chain chain --help
```

Expected output should show:
```
Manage the Lethean blockchain — sync, explore, and mine.

Usage:
  core-chain chain [command]

Available Commands:
  explorer    TUI block explorer
  sync        Headless P2P chain sync

Flags:
      --data-dir string   blockchain data directory (default "~/.lethean/chain")
      --seed string       seed peer address (host:port) (default "seeds.lthn.io:36942")
      --testnet           use testnet
```

**Step 5: Commit**

```bash
cd /Users/snider/Code/core/go-blockchain
git add cmd/core-chain/main.go
git commit -m "feat: add core-chain binary entry point with cli.Main()"
```

---

### Task 6: Remove old cmd/chain/main.go

**Context:** The old entry point is now replaced by `cmd/core-chain/main.go` + the package-level `commands.go`, `cmd_explorer.go`, `cmd_sync.go`, and `sync_service.go`. The sync logic was moved to `sync_service.go` (package-level), and `defaultDataDir` was moved to `commands.go`.

**Files:**
- Delete: `/Users/snider/Code/core/go-blockchain/cmd/chain/main.go`
- Delete: `/Users/snider/Code/core/go-blockchain/cmd/chain/` (directory)

**Step 1: Remove old entry point**

```bash
rm -rf /Users/snider/Code/core/go-blockchain/cmd/chain
```

**Step 2: Verify build still works**

```bash
cd /Users/snider/Code/core/go-blockchain && GOWORK=/Users/snider/Code/go.work go build ./...
```

Expected: Build succeeds (old cmd/chain is gone, cmd/core-chain is the only entry point).

**Step 3: Run all tests**

```bash
cd /Users/snider/Code/core/go-blockchain && GOWORK=/Users/snider/Code/go.work go test ./...
```

Expected: All tests pass.

**Step 4: Commit**

```bash
cd /Users/snider/Code/core/go-blockchain
git add -A cmd/chain
git commit -m "refactor: remove old cmd/chain entry point (replaced by cmd/core-chain)"
```

---

### Task 7: Add .core/build.yaml

**Context:** Every Core ecosystem binary needs a `.core/build.yaml` for the `core build` system. This tells the build system the binary name, targets, and ldflags.

**Files:**
- Create: `/Users/snider/Code/core/go-blockchain/.core/build.yaml`

**Step 1: Create the build config**

```bash
mkdir -p /Users/snider/Code/core/go-blockchain/.core
```

Create `/Users/snider/Code/core/go-blockchain/.core/build.yaml`:

```yaml
project: core-chain
binary: core-chain
main: ./cmd/core-chain

targets:
  - os: darwin
    arch: arm64
  - os: linux
    arch: amd64

ldflags:
  - -s -w
  - -X forge.lthn.ai/core/cli/pkg/cli.AppVersion={{.Version}}
  - -X forge.lthn.ai/core/cli/pkg/cli.BuildCommit={{.Commit}}
  - -X forge.lthn.ai/core/cli/pkg/cli.BuildDate={{.Date}}
```

**Step 2: Verify build via core build**

```bash
cd /Users/snider/Code/core/go-blockchain && core build
```

Expected: Produces `core-chain` binary in `./bin/`.

**Step 3: Commit**

```bash
cd /Users/snider/Code/core/go-blockchain
git add .core/build.yaml
git commit -m "chore: add .core/build.yaml for core-chain binary"
```

---

### Task 8: Final verification and push

**Context:** End-to-end verification that everything works: build, tests, binary help output.

**Files:** None (verification only).

**Step 1: Clean build**

```bash
cd /Users/snider/Code/core/go-blockchain && GOWORK=/Users/snider/Code/go.work go build ./...
```

Expected: Clean build, no errors.

**Step 2: Run all tests**

```bash
cd /Users/snider/Code/core/go-blockchain && GOWORK=/Users/snider/Code/go.work go test ./...
```

Expected: All tests pass.

**Step 3: Build binary**

```bash
cd /Users/snider/Code/core/go-blockchain && GOWORK=/Users/snider/Code/go.work go build -o ./bin/core-chain ./cmd/core-chain
```

Expected: Binary built successfully.

**Step 4: Verify CLI help**

```bash
./bin/core-chain --help
./bin/core-chain chain --help
./bin/core-chain chain explorer --help
./bin/core-chain chain sync --help
```

Expected: Clean help output with correct app name and subcommands.

**Step 5: Push to forge**

```bash
cd /Users/snider/Code/core/go-blockchain && git push origin main
```

Expected: Push succeeds to forge.

---

## File Summary

| Action | File | Purpose |
|--------|------|---------|
| Modify | `go.mod` | Fix replace directives |
| Create | `commands.go` | `AddChainCommands()` + shared helpers |
| Create | `commands_test.go` | Tests for command registration |
| Create | `cmd_explorer.go` | TUI block explorer subcommand |
| Create | `sync_service.go` | Extracted sync loop (from old main.go) |
| Create | `cmd_sync.go` | Headless sync subcommand with daemon support |
| Create | `cmd/core-chain/main.go` | Standalone binary entry point |
| Delete | `cmd/chain/main.go` | Old entry point (replaced) |
| Create | `.core/build.yaml` | Build system config |

## Dependency Changes

| Dependency | Status |
|------------|--------|
| `forge.lthn.ai/core/go-process` | **New** — daemon lifecycle, PID file, registry |
| `forge.lthn.ai/core/cli` | Existing — now used for `cli.Main()` + `WithCommands()` |
| `forge.lthn.ai/core/go-store` | Existing — unchanged |
| `forge.lthn.ai/core/go-p2p` | Existing — unchanged |
