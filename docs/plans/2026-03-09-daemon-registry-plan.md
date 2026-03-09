# Daemon Registry & Project Manifest — Implementation Plan

> **For Claude:** REQUIRED SUB-SKILL: Use superpowers:executing-plans to implement this plan task-by-task.

**Goal:** Add daemon declarations to the project manifest, a runtime registry for tracking running daemons, and `core start/stop/list/restart` CLI commands.

**Architecture:** The go-scm manifest gains a `Daemons` map. go-process gains a `Registry` that writes/reads JSON files under `~/.core/daemons/`. CLI commands read the manifest to know what to start, and use the registry to track what's running. A separate task adds `core.json` snapshot generation to go-devops for marketplace indexing.

**Tech Stack:** Go 1.26, `forge.lthn.ai/core/go-scm`, `forge.lthn.ai/core/go-process`, `forge.lthn.ai/core/cli`, `gopkg.in/yaml.v3`, `encoding/json`

**Workspace:** `GOWORK=/Users/snider/Code/go.work` (required for all builds/tests)

**Key source files (read these first):**
- `/Users/snider/Code/core/go-scm/manifest/manifest.go` — manifest struct being extended (51 LOC)
- `/Users/snider/Code/core/go-scm/manifest/manifest_test.go` — existing tests (66 LOC)
- `/Users/snider/Code/core/go-scm/manifest/loader.go` — Load/MarshalYAML (44 LOC)
- `/Users/snider/Code/core/go-process/daemon.go` — Daemon type the registry wraps (153 LOC)
- `/Users/snider/Code/core/go-process/pidfile.go` — ReadPID used for stale detection (93 LOC)
- `/Users/snider/Code/core/cli/pkg/cli/daemon_cmd.go` — existing daemon commands pattern (263 LOC)

---

### Task 1: Extend go-scm manifest with daemons section

**Files:**
- Modify: `go-scm/manifest/manifest.go`
- Modify: `go-scm/manifest/manifest_test.go`

**Step 1: Write the failing test**

Add to `/Users/snider/Code/core/go-scm/manifest/manifest_test.go`:

```go
func TestParse_Good_WithDaemons(t *testing.T) {
	raw := `
code: photo-browser
name: Photo Browser
version: 0.1.0
description: Browse local photo collections

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

modules:
  - core/media
`
	m, err := Parse([]byte(raw))
	require.NoError(t, err)
	assert.Equal(t, "photo-browser", m.Code)
	assert.Equal(t, "Browse local photo collections", m.Description)
	assert.Len(t, m.Daemons, 2)

	serve := m.Daemons["serve"]
	assert.Equal(t, "core-php", serve.Binary)
	assert.Equal(t, []string{"php", "serve"}, serve.Args)
	assert.Equal(t, "127.0.0.1:0", serve.Health)
	assert.True(t, serve.Default)

	worker := m.Daemons["worker"]
	assert.Equal(t, "core-mlx", worker.Binary)
	assert.False(t, worker.Default)
}

func TestManifest_DefaultDaemon_Good(t *testing.T) {
	m := Manifest{
		Daemons: map[string]DaemonSpec{
			"serve":  {Binary: "core-php", Default: true},
			"worker": {Binary: "core-mlx"},
		},
	}
	name, spec, ok := m.DefaultDaemon()
	assert.True(t, ok)
	assert.Equal(t, "serve", name)
	assert.Equal(t, "core-php", spec.Binary)
}

func TestManifest_DefaultDaemon_Bad_NoDaemons(t *testing.T) {
	m := Manifest{}
	_, _, ok := m.DefaultDaemon()
	assert.False(t, ok)
}

func TestManifest_DefaultDaemon_Good_SingleImplicit(t *testing.T) {
	m := Manifest{
		Daemons: map[string]DaemonSpec{
			"serve": {Binary: "core-php"},
		},
	}
	name, _, ok := m.DefaultDaemon()
	assert.True(t, ok)
	assert.Equal(t, "serve", name)
}
```

**Step 2: Run tests to verify they fail**

```bash
cd /Users/snider/Code/core/go-scm && GOWORK=/Users/snider/Code/go.work go test ./manifest/ -run 'TestParse_Good_WithDaemons|TestManifest_DefaultDaemon' -v
```

Expected: FAIL — `DaemonSpec` and `DefaultDaemon` undefined.

**Step 3: Write minimal implementation**

Edit `/Users/snider/Code/core/go-scm/manifest/manifest.go`:

```go
package manifest

import (
	"fmt"

	"gopkg.in/yaml.v3"
)

// DaemonSpec declares a background service this project can run.
type DaemonSpec struct {
	// Binary is the executable name (auto-detected if omitted).
	Binary string `yaml:"binary,omitempty" json:"binary,omitempty"`
	// Args are the arguments passed to the binary.
	Args []string `yaml:"args,omitempty" json:"args,omitempty"`
	// Health is the health check address (port 0 = dynamic).
	Health string `yaml:"health,omitempty" json:"health,omitempty"`
	// Default marks this as the daemon started by `core start` with no args.
	Default bool `yaml:"default,omitempty" json:"default,omitempty"`
}

// Manifest represents a .core/manifest.yaml project identity file.
type Manifest struct {
	Code        string            `yaml:"code" json:"code"`
	Name        string            `yaml:"name" json:"name"`
	Version     string            `yaml:"version" json:"version"`
	Description string            `yaml:"description,omitempty" json:"description,omitempty"`
	Sign        string            `yaml:"sign,omitempty" json:"sign,omitempty"`
	Layout      string            `yaml:"layout,omitempty" json:"layout,omitempty"`
	Slots       map[string]string `yaml:"slots,omitempty" json:"slots,omitempty"`

	Daemons     map[string]DaemonSpec `yaml:"daemons,omitempty" json:"daemons,omitempty"`
	Permissions Permissions           `yaml:"permissions,omitempty" json:"permissions,omitempty"`
	Modules     []string              `yaml:"modules,omitempty" json:"modules,omitempty"`
}

// Permissions declares the I/O capabilities a module requires.
type Permissions struct {
	Read  []string `yaml:"read" json:"read,omitempty"`
	Write []string `yaml:"write" json:"write,omitempty"`
	Net   []string `yaml:"net" json:"net,omitempty"`
	Run   []string `yaml:"run" json:"run,omitempty"`
}

// Parse decodes YAML bytes into a Manifest.
func Parse(data []byte) (*Manifest, error) {
	var m Manifest
	if err := yaml.Unmarshal(data, &m); err != nil {
		return nil, fmt.Errorf("manifest.Parse: %w", err)
	}
	return &m, nil
}

// SlotNames returns a deduplicated list of component names from slots.
func (m *Manifest) SlotNames() []string {
	seen := make(map[string]bool)
	var names []string
	for _, name := range m.Slots {
		if !seen[name] {
			seen[name] = true
			names = append(names, name)
		}
	}
	return names
}

// DefaultDaemon returns the name and spec of the default daemon.
// If no daemon has default:true and there's exactly one, it's used.
// Returns ("", DaemonSpec{}, false) if no default can be determined.
func (m *Manifest) DefaultDaemon() (string, DaemonSpec, bool) {
	if len(m.Daemons) == 0 {
		return "", DaemonSpec{}, false
	}

	for name, spec := range m.Daemons {
		if spec.Default {
			return name, spec, true
		}
	}

	if len(m.Daemons) == 1 {
		for name, spec := range m.Daemons {
			return name, spec, true
		}
	}

	return "", DaemonSpec{}, false
}
```

**Step 4: Update the manifest loader path**

Edit `/Users/snider/Code/core/go-scm/manifest/loader.go` line 12:

Change `const manifestPath = ".core/view.yml"` to `const manifestPath = ".core/manifest.yaml"`

**Step 5: Run tests to verify they pass**

```bash
cd /Users/snider/Code/core/go-scm && GOWORK=/Users/snider/Code/go.work go test ./manifest/ -v
```

Expected: ALL PASS (existing + new tests).

**Step 6: Commit**

```bash
cd /Users/snider/Code/core/go-scm
git add manifest/manifest.go manifest/manifest_test.go manifest/loader.go
git commit -m "feat: add DaemonSpec to manifest, rename path to .core/manifest.yaml"
```

---

### Task 2: Create daemon registry in go-process

**Files:**
- Create: `go-process/registry.go`
- Create: `go-process/registry_test.go`

**Step 1: Write the test file**

Create `/Users/snider/Code/core/go-process/registry_test.go`:

```go
package process

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestRegistry_RegisterAndGet(t *testing.T) {
	dir := t.TempDir()
	reg := NewRegistry(dir)

	entry := DaemonEntry{
		Code:    "test-app",
		Daemon:  "serve",
		PID:     os.Getpid(), // Use own PID so it's "alive"
		Health:  "127.0.0.1:9000",
		Project: "/tmp/test-app",
		Binary:  "/usr/bin/test",
	}

	err := reg.Register(entry)
	require.NoError(t, err)

	got, ok := reg.Get("test-app", "serve")
	require.True(t, ok)
	assert.Equal(t, os.Getpid(), got.PID)
	assert.Equal(t, "127.0.0.1:9000", got.Health)
	assert.Equal(t, "/tmp/test-app", got.Project)
	assert.False(t, got.Started.IsZero())
}

func TestRegistry_Unregister(t *testing.T) {
	dir := t.TempDir()
	reg := NewRegistry(dir)

	entry := DaemonEntry{
		Code:   "test-app",
		Daemon: "serve",
		PID:    os.Getpid(),
	}

	require.NoError(t, reg.Register(entry))

	err := reg.Unregister("test-app", "serve")
	require.NoError(t, err)

	_, ok := reg.Get("test-app", "serve")
	assert.False(t, ok)

	// File should be gone
	path := filepath.Join(dir, "test-app-serve.json")
	_, err = os.Stat(path)
	assert.True(t, os.IsNotExist(err))
}

func TestRegistry_List(t *testing.T) {
	dir := t.TempDir()
	reg := NewRegistry(dir)

	require.NoError(t, reg.Register(DaemonEntry{
		Code: "app-a", Daemon: "serve", PID: os.Getpid(),
	}))
	require.NoError(t, reg.Register(DaemonEntry{
		Code: "app-b", Daemon: "worker", PID: os.Getpid(),
	}))

	entries, err := reg.List()
	require.NoError(t, err)
	assert.Len(t, entries, 2)
}

func TestRegistry_List_PrunesStale(t *testing.T) {
	dir := t.TempDir()
	reg := NewRegistry(dir)

	// Register with a PID that doesn't exist
	require.NoError(t, reg.Register(DaemonEntry{
		Code: "stale-app", Daemon: "serve", PID: 999999999,
	}))

	entries, err := reg.List()
	require.NoError(t, err)
	assert.Len(t, entries, 0)

	// Stale file should have been removed
	path := filepath.Join(dir, "stale-app-serve.json")
	_, err = os.Stat(path)
	assert.True(t, os.IsNotExist(err))
}

func TestRegistry_Get_NotFound(t *testing.T) {
	dir := t.TempDir()
	reg := NewRegistry(dir)

	_, ok := reg.Get("nope", "nada")
	assert.False(t, ok)
}

func TestRegistry_CreatesDirectory(t *testing.T) {
	dir := filepath.Join(t.TempDir(), "nested", "daemons")
	reg := NewRegistry(dir)

	err := reg.Register(DaemonEntry{
		Code: "test", Daemon: "serve", PID: os.Getpid(),
	})
	require.NoError(t, err)

	info, err := os.Stat(dir)
	require.NoError(t, err)
	assert.True(t, info.IsDir())
}

func TestDefaultRegistry(t *testing.T) {
	reg := DefaultRegistry()
	assert.NotNil(t, reg)
}
```

**Step 2: Run tests to verify they fail**

```bash
cd /Users/snider/Code/core/go-process && GOWORK=/Users/snider/Code/go.work go test -run 'TestRegistry|TestDefaultRegistry' -v ./...
```

Expected: FAIL — `Registry`, `DaemonEntry`, `NewRegistry`, `DefaultRegistry` undefined.

**Step 3: Write minimal implementation**

Create `/Users/snider/Code/core/go-process/registry.go`:

```go
package process

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"
)

// DaemonEntry records a running daemon in the registry.
type DaemonEntry struct {
	Code    string    `json:"code"`
	Daemon  string    `json:"daemon"`
	PID     int       `json:"pid"`
	Health  string    `json:"health,omitempty"`
	Project string    `json:"project,omitempty"`
	Binary  string    `json:"binary,omitempty"`
	Started time.Time `json:"started"`
}

// Registry tracks running daemons via JSON files in a directory.
type Registry struct {
	dir string
}

// NewRegistry creates a registry backed by the given directory.
func NewRegistry(dir string) *Registry {
	return &Registry{dir: dir}
}

// DefaultRegistry returns a registry at ~/.core/daemons/.
func DefaultRegistry() *Registry {
	home, _ := os.UserHomeDir()
	return NewRegistry(filepath.Join(home, ".core", "daemons"))
}

// Register writes a daemon entry to the registry directory.
func (r *Registry) Register(entry DaemonEntry) error {
	if err := os.MkdirAll(r.dir, 0755); err != nil {
		return fmt.Errorf("registry: create dir: %w", err)
	}

	if entry.Started.IsZero() {
		entry.Started = time.Now()
	}

	data, err := json.MarshalIndent(entry, "", "  ")
	if err != nil {
		return fmt.Errorf("registry: marshal: %w", err)
	}

	path := r.entryPath(entry.Code, entry.Daemon)
	if err := os.WriteFile(path, data, 0644); err != nil {
		return fmt.Errorf("registry: write: %w", err)
	}

	return nil
}

// Unregister removes a daemon entry from the registry.
func (r *Registry) Unregister(code, daemon string) error {
	path := r.entryPath(code, daemon)
	if err := os.Remove(path); err != nil && !os.IsNotExist(err) {
		return fmt.Errorf("registry: remove: %w", err)
	}
	return nil
}

// Get retrieves a daemon entry. Returns false if not found or stale.
func (r *Registry) Get(code, daemon string) (*DaemonEntry, bool) {
	path := r.entryPath(code, daemon)
	entry, err := r.readEntry(path)
	if err != nil {
		return nil, false
	}

	if _, alive := ReadPID(path + ".check"); !r.isAlive(entry.PID) {
		_ = os.Remove(path)
		return nil, false
	}

	return entry, true
}

// List returns all live daemon entries, pruning stale ones.
func (r *Registry) List() ([]DaemonEntry, error) {
	files, err := filepath.Glob(filepath.Join(r.dir, "*.json"))
	if err != nil {
		return nil, fmt.Errorf("registry: glob: %w", err)
	}

	var entries []DaemonEntry
	for _, path := range files {
		entry, err := r.readEntry(path)
		if err != nil {
			_ = os.Remove(path)
			continue
		}

		if !r.isAlive(entry.PID) {
			_ = os.Remove(path)
			continue
		}

		entries = append(entries, *entry)
	}

	return entries, nil
}

func (r *Registry) readEntry(path string) (*DaemonEntry, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}

	var entry DaemonEntry
	if err := json.Unmarshal(data, &entry); err != nil {
		return nil, err
	}

	return &entry, nil
}

func (r *Registry) isAlive(pid int) bool {
	if pid <= 0 {
		return false
	}
	proc, err := os.FindProcess(pid)
	if err != nil {
		return false
	}
	return proc.Signal(os.Signal(nil)) == nil
}

func (r *Registry) entryPath(code, daemon string) string {
	name := code + "-" + daemon
	name = strings.ReplaceAll(name, "/", "-")
	return filepath.Join(r.dir, name+".json")
}
```

**Step 4: Run tests to verify they pass**

```bash
cd /Users/snider/Code/core/go-process && GOWORK=/Users/snider/Code/go.work go test -run 'TestRegistry|TestDefaultRegistry' -v ./...
```

Expected: ALL PASS.

**Step 5: Commit**

```bash
cd /Users/snider/Code/core/go-process
git add registry.go registry_test.go
git commit -m "feat: add daemon Registry for tracking running daemons"
```

---

### Task 3: Wire Registry into Daemon lifecycle

**Files:**
- Modify: `go-process/daemon.go`
- Modify: `go-process/daemon_test.go`

**Context:** When a `Registry` is set on `DaemonOptions`, `Daemon.Start()` auto-registers and `Daemon.Stop()` auto-unregisters. This is opt-in — existing consumers without a registry are unaffected.

**Step 1: Write the failing test**

Add to `/Users/snider/Code/core/go-process/daemon_test.go`:

```go
func TestDaemon_AutoRegisters(t *testing.T) {
	dir := t.TempDir()
	reg := NewRegistry(filepath.Join(dir, "daemons"))

	d := NewDaemon(DaemonOptions{
		HealthAddr: "127.0.0.1:0",
		Registry:   reg,
		RegistryEntry: DaemonEntry{
			Code:   "test-app",
			Daemon: "serve",
		},
	})

	err := d.Start()
	require.NoError(t, err)

	// Should be registered
	entry, ok := reg.Get("test-app", "serve")
	require.True(t, ok)
	assert.Equal(t, os.Getpid(), entry.PID)
	assert.NotEmpty(t, entry.Health)

	// Stop should unregister
	err = d.Stop()
	require.NoError(t, err)

	_, ok = reg.Get("test-app", "serve")
	assert.False(t, ok)
}
```

**Step 2: Run test to verify it fails**

```bash
cd /Users/snider/Code/core/go-process && GOWORK=/Users/snider/Code/go.work go test -run TestDaemon_AutoRegisters -v ./...
```

Expected: FAIL — `Registry` and `RegistryEntry` not fields on `DaemonOptions`.

**Step 3: Add Registry fields to DaemonOptions and wire into Start/Stop**

Edit `/Users/snider/Code/core/go-process/daemon.go`:

Add to `DaemonOptions` struct:
```go
	// Registry for tracking this daemon. Leave nil to skip registration.
	Registry *Registry

	// RegistryEntry provides the code and daemon name for registration.
	// PID, Health, and Started are filled automatically.
	RegistryEntry DaemonEntry
```

In `Daemon.Start()`, after `d.running = true`, add:
```go
	// Auto-register if registry is set
	if d.opts.Registry != nil {
		entry := d.opts.RegistryEntry
		entry.PID = os.Getpid()
		if d.health != nil {
			entry.Health = d.health.Addr()
		}
		_ = d.opts.Registry.Register(entry)
	}
```

In `Daemon.Stop()`, before `d.running = false`, add:
```go
	// Auto-unregister
	if d.opts.Registry != nil {
		_ = d.opts.Registry.Unregister(d.opts.RegistryEntry.Code, d.opts.RegistryEntry.Daemon)
	}
```

**Step 4: Run all daemon tests**

```bash
cd /Users/snider/Code/core/go-process && GOWORK=/Users/snider/Code/go.work go test -run TestDaemon -v ./...
```

Expected: ALL PASS (existing + new).

**Step 5: Commit**

```bash
cd /Users/snider/Code/core/go-process
git add daemon.go daemon_test.go
git commit -m "feat: auto-register/unregister daemons via optional Registry"
```

---

### Task 4: Add `core start` command in cli

**Files:**
- Create: `cli/cmd/service/cmd.go`
- Modify: `cli/main.go` — register the new commands

**Context:** `core start [daemon-name]` reads `.core/manifest.yaml` in cwd (or parent dirs), finds the daemon entry, execs the binary detached, waits for health, and registers in the global registry. Uses `os.Getwd()` + walk parents to find the manifest. Uses `go-scm/manifest` for parsing and `go-process` for the registry.

**Step 1: Write the command file**

Create `/Users/snider/Code/core/cli/cmd/service/cmd.go`:

```go
package service

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"syscall"

	"forge.lthn.ai/core/cli/pkg/cli"
	"forge.lthn.ai/core/go-process"
	"forge.lthn.ai/core/go-scm/manifest"
)

// AddServiceCommands registers core start/stop/list/restart.
func AddServiceCommands(root *cli.Command) {
	startCmd := cli.NewCommand("start", "Start a project daemon",
		"Reads .core/manifest.yaml and starts the named daemon (or the default).\n"+
			"The daemon runs detached in the background.",
		func(cmd *cli.Command, args []string) error {
			return runStart(args)
		},
	)

	stopCmd := cli.NewCommand("stop", "Stop a project daemon",
		"Stops the named daemon for the current project, or all daemons if no name given.",
		func(cmd *cli.Command, args []string) error {
			return runStop(args)
		},
	)

	listCmd := cli.NewCommand("list", "List running daemons",
		"Shows all running daemons tracked in ~/.core/daemons/.",
		func(cmd *cli.Command, args []string) error {
			return runList()
		},
	)

	restartCmd := cli.NewCommand("restart", "Restart a project daemon",
		"Stops then starts the named daemon.",
		func(cmd *cli.Command, args []string) error {
			if err := runStop(args); err != nil {
				return err
			}
			return runStart(args)
		},
	)

	root.AddCommand(startCmd, stopCmd, listCmd, restartCmd)
}

func runStart(args []string) error {
	m, projectDir, err := findManifest()
	if err != nil {
		return err
	}

	daemonName, spec, err := resolveDaemon(m, args)
	if err != nil {
		return err
	}

	reg := process.DefaultRegistry()

	// Check if already running
	if _, ok := reg.Get(m.Code, daemonName); ok {
		return fmt.Errorf("%s/%s is already running", m.Code, daemonName)
	}

	// Resolve binary
	binary := spec.Binary
	if binary == "" {
		return fmt.Errorf("daemon %q has no binary specified", daemonName)
	}

	binPath, err := exec.LookPath(binary)
	if err != nil {
		return fmt.Errorf("binary %q not found in PATH: %w", binary, err)
	}

	// Launch detached
	cmd := exec.Command(binPath, spec.Args...)
	cmd.Dir = projectDir
	cmd.Env = append(os.Environ(), "CORE_DAEMON=1")
	cmd.Stdout = nil
	cmd.Stderr = nil
	cmd.Stdin = nil
	cmd.SysProcAttr = &syscall.SysProcAttr{Setsid: true}

	if err := cmd.Start(); err != nil {
		return fmt.Errorf("failed to start %s: %w", daemonName, err)
	}

	pid := cmd.Process.Pid
	_ = cmd.Process.Release()

	// Wait for health if configured
	health := spec.Health
	if health != "" && health != "127.0.0.1:0" {
		if process.WaitForHealth(health, 5000) {
			cli.LogInfo(fmt.Sprintf("Started %s/%s (PID %d, health %s)", m.Code, daemonName, pid, health))
		} else {
			cli.LogInfo(fmt.Sprintf("Started %s/%s (PID %d, health not yet ready)", m.Code, daemonName, pid))
		}
	} else {
		cli.LogInfo(fmt.Sprintf("Started %s/%s (PID %d)", m.Code, daemonName, pid))
	}

	// Register
	_ = reg.Register(process.DaemonEntry{
		Code:    m.Code,
		Daemon:  daemonName,
		PID:     pid,
		Health:  health,
		Project: projectDir,
		Binary:  binPath,
	})

	return nil
}

func runStop(args []string) error {
	reg := process.DefaultRegistry()

	// If args given, stop specific daemon
	if len(args) > 0 {
		m, _, err := findManifest()
		if err != nil {
			return err
		}
		return stopDaemon(reg, m.Code, args[0])
	}

	// No args: stop all daemons for this project
	m, _, err := findManifest()
	if err != nil {
		return err
	}

	entries, err := reg.List()
	if err != nil {
		return err
	}

	stopped := 0
	for _, e := range entries {
		if e.Code == m.Code {
			if err := stopDaemon(reg, e.Code, e.Daemon); err != nil {
				cli.LogError(fmt.Sprintf("Failed to stop %s/%s: %v", e.Code, e.Daemon, err))
			} else {
				stopped++
			}
		}
	}

	if stopped == 0 {
		cli.LogInfo("No running daemons for " + m.Code)
	}

	return nil
}

func stopDaemon(reg *process.Registry, code, daemon string) error {
	entry, ok := reg.Get(code, daemon)
	if !ok {
		return fmt.Errorf("%s/%s is not running", code, daemon)
	}

	proc, err := os.FindProcess(entry.PID)
	if err != nil {
		return fmt.Errorf("process %d not found: %w", entry.PID, err)
	}

	if err := proc.Signal(syscall.SIGTERM); err != nil {
		return fmt.Errorf("failed to signal PID %d: %w", entry.PID, err)
	}

	cli.LogInfo(fmt.Sprintf("Stopped %s/%s (PID %d)", code, daemon, entry.PID))
	_ = reg.Unregister(code, daemon)
	return nil
}

func runList() error {
	reg := process.DefaultRegistry()
	entries, err := reg.List()
	if err != nil {
		return err
	}

	if len(entries) == 0 {
		fmt.Println("No running daemons")
		return nil
	}

	fmt.Printf("%-20s %-12s %-8s %-24s %s\n", "CODE", "DAEMON", "PID", "HEALTH", "PROJECT")
	for _, e := range entries {
		project := e.Project
		if project == "" {
			project = "-"
		}
		fmt.Printf("%-20s %-12s %-8d %-24s %s\n", e.Code, e.Daemon, e.PID, e.Health, project)
	}

	return nil
}

// findManifest walks from cwd up to / looking for .core/manifest.yaml.
func findManifest() (*manifest.Manifest, string, error) {
	dir, err := os.Getwd()
	if err != nil {
		return nil, "", err
	}

	for {
		path := filepath.Join(dir, ".core", "manifest.yaml")
		data, err := os.ReadFile(path)
		if err == nil {
			m, err := manifest.Parse(data)
			if err != nil {
				return nil, "", fmt.Errorf("invalid manifest at %s: %w", path, err)
			}
			return m, dir, nil
		}

		parent := filepath.Dir(dir)
		if parent == dir {
			break
		}
		dir = parent
	}

	return nil, "", fmt.Errorf("no .core/manifest.yaml found (checked cwd and parent directories)")
}

// resolveDaemon finds the daemon entry by name or default.
func resolveDaemon(m *manifest.Manifest, args []string) (string, manifest.DaemonSpec, error) {
	if len(args) > 0 {
		name := args[0]
		spec, ok := m.Daemons[name]
		if !ok {
			return "", manifest.DaemonSpec{}, fmt.Errorf("daemon %q not found in manifest (available: %v)", name, daemonNames(m))
		}
		return name, spec, nil
	}

	name, spec, ok := m.DefaultDaemon()
	if !ok {
		return "", manifest.DaemonSpec{}, fmt.Errorf("no default daemon in manifest (use: core start <name>)")
	}
	return name, spec, nil
}

func daemonNames(m *manifest.Manifest) []string {
	var names []string
	for name := range m.Daemons {
		names = append(names, name)
	}
	return names
}
```

**Step 2: Register in main.go**

Find the `cli.WithCommands(...)` block in `/Users/snider/Code/core/cli/main.go` and add:

```go
cli.WithCommands("service", service.AddServiceCommands),
```

With import `"forge.lthn.ai/core/cli/cmd/service"`.

**Step 3: Add go-scm dependency to cli go.mod**

```bash
cd /Users/snider/Code/core/cli && GOWORK=/Users/snider/Code/go.work go build ./...
```

If go-scm is not yet in go.mod, add it:
```
require forge.lthn.ai/core/go-scm v0.1.0
```

**Step 4: Build to verify compilation**

```bash
cd /Users/snider/Code/core/cli && GOWORK=/Users/snider/Code/go.work go build ./...
```

Expected: Success.

**Step 5: Commit**

```bash
cd /Users/snider/Code/core/cli
git add cmd/service/cmd.go main.go go.mod go.sum
git commit -m "feat: add core start/stop/list/restart for manifest-driven daemons"
```

---

### Task 5: Add core.json snapshot generation to go-devops

**Files:**
- Create: `go-devops/snapshot/snapshot.go`
- Create: `go-devops/snapshot/snapshot_test.go`

**Context:** `core build release` should call `snapshot.Generate()` to produce `core.json` from `.core/manifest.yaml`. This task creates the generation logic. Wiring it into the release pipeline is a separate step (just needs a call in the existing release code path).

**Step 1: Write the test file**

Create `/Users/snider/Code/core/go-devops/snapshot/snapshot_test.go`:

```go
package snapshot

import (
	"encoding/json"
	"testing"

	"forge.lthn.ai/core/go-scm/manifest"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestGenerate_Good(t *testing.T) {
	m := &manifest.Manifest{
		Code:        "test-app",
		Name:        "Test App",
		Version:     "1.0.0",
		Description: "A test application",
		Daemons: map[string]manifest.DaemonSpec{
			"serve": {Binary: "core-php", Args: []string{"php", "serve"}, Default: true},
		},
		Modules: []string{"core/media"},
	}

	data, err := Generate(m, "abc123def456", "v1.0.0")
	require.NoError(t, err)

	var snap Snapshot
	require.NoError(t, json.Unmarshal(data, &snap))

	assert.Equal(t, 1, snap.Schema)
	assert.Equal(t, "test-app", snap.Code)
	assert.Equal(t, "1.0.0", snap.Version)
	assert.Equal(t, "abc123def456", snap.Commit)
	assert.Equal(t, "v1.0.0", snap.Tag)
	assert.NotEmpty(t, snap.Built)
	assert.Len(t, snap.Daemons, 1)
	assert.Equal(t, "core-php", snap.Daemons["serve"].Binary)
}

func TestGenerate_Good_NoDaemons(t *testing.T) {
	m := &manifest.Manifest{
		Code:    "simple",
		Name:    "Simple",
		Version: "0.1.0",
	}

	data, err := Generate(m, "abc123", "v0.1.0")
	require.NoError(t, err)

	var snap Snapshot
	require.NoError(t, json.Unmarshal(data, &snap))

	assert.Equal(t, "simple", snap.Code)
	assert.Nil(t, snap.Daemons)
}
```

**Step 2: Run tests to verify they fail**

```bash
cd /Users/snider/Code/core/go-devops && GOWORK=/Users/snider/Code/go.work go test ./snapshot/ -v
```

Expected: FAIL — package doesn't exist.

**Step 3: Write minimal implementation**

Create `/Users/snider/Code/core/go-devops/snapshot/snapshot.go`:

```go
// Package snapshot generates frozen core.json release manifests.
package snapshot

import (
	"encoding/json"
	"time"

	"forge.lthn.ai/core/go-scm/manifest"
)

// Snapshot is the frozen release manifest written as core.json.
type Snapshot struct {
	Schema      int                          `json:"schema"`
	Code        string                       `json:"code"`
	Name        string                       `json:"name"`
	Version     string                       `json:"version"`
	Description string                       `json:"description,omitempty"`
	Commit      string                       `json:"commit"`
	Tag         string                       `json:"tag"`
	Built       string                       `json:"built"`
	Daemons     map[string]manifest.DaemonSpec `json:"daemons,omitempty"`
	Layout      string                       `json:"layout,omitempty"`
	Slots       map[string]string            `json:"slots,omitempty"`
	Permissions *manifest.Permissions        `json:"permissions,omitempty"`
	Modules     []string                     `json:"modules,omitempty"`
}

// Generate creates a core.json snapshot from a manifest.
func Generate(m *manifest.Manifest, commit, tag string) ([]byte, error) {
	snap := Snapshot{
		Schema:      1,
		Code:        m.Code,
		Name:        m.Name,
		Version:     m.Version,
		Description: m.Description,
		Commit:      commit,
		Tag:         tag,
		Built:       time.Now().UTC().Format(time.RFC3339),
		Daemons:     m.Daemons,
		Layout:      m.Layout,
		Slots:       m.Slots,
		Modules:     m.Modules,
	}

	if m.Permissions.Read != nil || m.Permissions.Write != nil ||
		m.Permissions.Net != nil || m.Permissions.Run != nil {
		snap.Permissions = &m.Permissions
	}

	return json.MarshalIndent(snap, "", "  ")
}
```

**Step 4: Run tests to verify they pass**

```bash
cd /Users/snider/Code/core/go-devops && GOWORK=/Users/snider/Code/go.work go test ./snapshot/ -v
```

Expected: PASS.

**Step 5: Commit**

```bash
cd /Users/snider/Code/core/go-devops
git add snapshot/snapshot.go snapshot/snapshot_test.go
git commit -m "feat: add core.json snapshot generation from manifest"
```

---

### Task 6: Final verification — build all affected modules

**Step 1: Build all affected modules**

```bash
cd /Users/snider/Code/core/go-scm && GOWORK=/Users/snider/Code/go.work go build ./...
cd /Users/snider/Code/core/go-process && GOWORK=/Users/snider/Code/go.work go build ./...
cd /Users/snider/Code/core/cli && GOWORK=/Users/snider/Code/go.work go build ./...
cd /Users/snider/Code/core/go-devops && GOWORK=/Users/snider/Code/go.work go build ./...
```

**Step 2: Run all tests in affected packages**

```bash
cd /Users/snider/Code/core/go-scm && GOWORK=/Users/snider/Code/go.work go test ./manifest/ -v
cd /Users/snider/Code/core/go-process && GOWORK=/Users/snider/Code/go.work go test -run 'TestRegistry|TestDaemon' -v ./...
cd /Users/snider/Code/core/go-devops && GOWORK=/Users/snider/Code/go.work go test ./snapshot/ -v
```

**Step 3: Check for stale references**

Search across `/Users/snider/Code/core/` for any remaining references to `.core/view.yml` — should all be updated to `.core/manifest.yaml`.

**Step 4: Push all repos**

```bash
cd /Users/snider/Code/core/go-scm && git push origin main
cd /Users/snider/Code/core/go-process && git push origin main
cd /Users/snider/Code/core/cli && git push origin main
cd /Users/snider/Code/core/go-devops && git push origin main
```
