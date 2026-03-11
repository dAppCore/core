# Daemon Process Management Extraction — Implementation Plan

> **For Claude:** REQUIRED SUB-SKILL: Use superpowers:executing-plans to implement this plan task-by-task.

**Goal:** Move daemon runtime types (PIDFile, HealthServer, Daemon) from `core/cli` to `core/go-process` and create a generic daemon CLI command builder in `core/cli`.

**Architecture:** The daemon runtime types (PIDFile, HealthServer, Daemon, DaemonOptions) move to `go-process` — they're process management primitives. A new generic `AddDaemonCommand()` in cli imports go-process and provides start/stop/status/run subcommands with a callback for business logic. go-ai and ide update imports.

**Tech Stack:** Go 1.26, `forge.lthn.ai/core/go-process`, `forge.lthn.ai/core/cli`, `forge.lthn.ai/core/go-ai`, `forge.lthn.ai/core/ide`

**Workspace:** `GOWORK=/Users/snider/Code/go.work` (required for all builds/tests — `go mod tidy` fails outside workspace due to stale upstream tags)

**Key source files (read these first):**
- `/Users/snider/Code/core/cli/pkg/cli/daemon.go` — types being extracted (448 LOC)
- `/Users/snider/Code/core/cli/pkg/cli/daemon_test.go` — tests being ported (235 LOC)
- `/Users/snider/Code/core/go-ai/cmd/daemon/cmd.go` — CLI commands being generalised (393 LOC)
- `/Users/snider/Code/core/go-process/types.go` — existing go-process types
- `/Users/snider/Code/core/ide/headless.go:104-113` — ide consumer of daemon types

---

### Task 1: Create `pidfile.go` in go-process

**Files:**
- Create: `go-process/pidfile.go`
- Create: `go-process/pidfile_test.go`

**Context:** PIDFile manages a process ID lock file. Currently in `cli/pkg/cli/daemon.go:76-134`. We rewrite to use stdlib (`os.ReadFile`, `os.WriteFile`, `os.MkdirAll`, `os.Remove`) instead of `go-io` to avoid adding that dependency to go-process.

Also includes `ReadPID()` from `go-ai/cmd/daemon/cmd.go:351-374` — a standalone helper that reads a PID file and checks if the process is alive.

**Step 1: Write the test file**

```go
// File: /Users/snider/Code/core/go-process/pidfile_test.go
package process

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestPIDFile_AcquireAndRelease(t *testing.T) {
	pidPath := filepath.Join(t.TempDir(), "test.pid")

	pid := NewPIDFile(pidPath)

	err := pid.Acquire()
	require.NoError(t, err)

	// File should exist with our PID
	data, err := os.ReadFile(pidPath)
	require.NoError(t, err)
	assert.NotEmpty(t, data)

	err = pid.Release()
	require.NoError(t, err)

	// File should be gone
	_, err = os.Stat(pidPath)
	assert.True(t, os.IsNotExist(err))
}

func TestPIDFile_StalePID(t *testing.T) {
	pidPath := filepath.Join(t.TempDir(), "stale.pid")

	// Write a stale PID (non-existent process)
	require.NoError(t, os.WriteFile(pidPath, []byte("999999999"), 0644))

	pid := NewPIDFile(pidPath)

	// Should acquire successfully (stale PID removed)
	err := pid.Acquire()
	require.NoError(t, err)

	err = pid.Release()
	require.NoError(t, err)
}

func TestPIDFile_CreatesParentDirectory(t *testing.T) {
	pidPath := filepath.Join(t.TempDir(), "subdir", "nested", "test.pid")

	pid := NewPIDFile(pidPath)

	err := pid.Acquire()
	require.NoError(t, err)

	err = pid.Release()
	require.NoError(t, err)
}

func TestPIDFile_Path(t *testing.T) {
	pid := NewPIDFile("/tmp/test.pid")
	assert.Equal(t, "/tmp/test.pid", pid.Path())
}

func TestReadPID_Missing(t *testing.T) {
	pid, running := ReadPID("/nonexistent/path.pid")
	assert.Equal(t, 0, pid)
	assert.False(t, running)
}

func TestReadPID_InvalidContent(t *testing.T) {
	path := filepath.Join(t.TempDir(), "bad.pid")
	require.NoError(t, os.WriteFile(path, []byte("notanumber"), 0644))

	pid, running := ReadPID(path)
	assert.Equal(t, 0, pid)
	assert.False(t, running)
}

func TestReadPID_StalePID(t *testing.T) {
	path := filepath.Join(t.TempDir(), "stale.pid")
	require.NoError(t, os.WriteFile(path, []byte("999999999"), 0644))

	pid, running := ReadPID(path)
	assert.Equal(t, 999999999, pid)
	assert.False(t, running)
}
```

**Step 2: Run tests to verify they fail**

```bash
cd /Users/snider/Code/core/go-process && GOWORK=/Users/snider/Code/go.work go test -run 'TestPIDFile|TestReadPID' -v ./...
```

Expected: FAIL — `NewPIDFile`, `ReadPID` undefined.

**Step 3: Write the implementation**

```go
// File: /Users/snider/Code/core/go-process/pidfile.go
package process

import (
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"sync"
	"syscall"
)

// PIDFile manages a process ID file for single-instance enforcement.
type PIDFile struct {
	path string
	mu   sync.Mutex
}

// NewPIDFile creates a PID file manager.
func NewPIDFile(path string) *PIDFile {
	return &PIDFile{path: path}
}

// Acquire writes the current PID to the file.
// Returns error if another instance is running.
func (p *PIDFile) Acquire() error {
	p.mu.Lock()
	defer p.mu.Unlock()

	// Check for existing PID file
	if data, err := os.ReadFile(p.path); err == nil {
		pid, err := strconv.Atoi(strings.TrimSpace(string(data)))
		if err == nil && pid > 0 {
			// Check if process is still running
			if proc, err := os.FindProcess(pid); err == nil {
				if err := proc.Signal(syscall.Signal(0)); err == nil {
					return fmt.Errorf("another instance is running (PID %d)", pid)
				}
			}
		}
		// Stale PID file, remove it
		_ = os.Remove(p.path)
	}

	// Ensure directory exists
	if dir := filepath.Dir(p.path); dir != "." {
		if err := os.MkdirAll(dir, 0755); err != nil {
			return fmt.Errorf("failed to create PID directory: %w", err)
		}
	}

	// Write current PID
	pid := os.Getpid()
	if err := os.WriteFile(p.path, []byte(strconv.Itoa(pid)), 0644); err != nil {
		return fmt.Errorf("failed to write PID file: %w", err)
	}

	return nil
}

// Release removes the PID file.
func (p *PIDFile) Release() error {
	p.mu.Lock()
	defer p.mu.Unlock()
	return os.Remove(p.path)
}

// Path returns the PID file path.
func (p *PIDFile) Path() string {
	return p.path
}

// ReadPID reads a PID file and checks if the process is still running.
// Returns (pid, true) if the process is alive, (pid, false) if dead/stale,
// or (0, false) if the file doesn't exist or is invalid.
func ReadPID(path string) (int, bool) {
	data, err := os.ReadFile(path)
	if err != nil {
		return 0, false
	}

	pid, err := strconv.Atoi(strings.TrimSpace(string(data)))
	if err != nil || pid <= 0 {
		return 0, false
	}

	// Signal 0 tests if the process exists without actually sending a signal
	proc, err := os.FindProcess(pid)
	if err != nil {
		return pid, false
	}

	if err := proc.Signal(syscall.Signal(0)); err != nil {
		return pid, false
	}

	return pid, true
}
```

**Step 4: Run tests to verify they pass**

```bash
cd /Users/snider/Code/core/go-process && GOWORK=/Users/snider/Code/go.work go test -run 'TestPIDFile|TestReadPID' -v ./...
```

Expected: PASS

**Step 5: Commit**

```bash
cd /Users/snider/Code/core/go-process
git add pidfile.go pidfile_test.go
git commit -m "feat: add PIDFile and ReadPID for daemon process management"
```

---

### Task 2: Create `health.go` in go-process

**Files:**
- Create: `go-process/health.go`
- Create: `go-process/health_test.go`

**Context:** HealthServer provides HTTP `/health` and `/ready` endpoints for daemon liveness/readiness probes. Currently in `cli/pkg/cli/daemon.go:138-244`. Also includes `WaitForHealth()` from `go-ai/cmd/daemon/cmd.go:377-393`.

**Step 1: Write the test file**

```go
// File: /Users/snider/Code/core/go-process/health_test.go
package process

import (
	"context"
	"net/http"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestHealthServer_Endpoints(t *testing.T) {
	hs := NewHealthServer("127.0.0.1:0") // Random port

	err := hs.Start()
	require.NoError(t, err)
	defer func() { _ = hs.Stop(context.Background()) }()

	addr := hs.Addr()
	require.NotEmpty(t, addr)

	// Health should be OK
	resp, err := http.Get("http://" + addr + "/health")
	require.NoError(t, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
	_ = resp.Body.Close()

	// Ready should be OK by default
	resp, err = http.Get("http://" + addr + "/ready")
	require.NoError(t, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
	_ = resp.Body.Close()

	// Set not ready
	hs.SetReady(false)

	resp, err = http.Get("http://" + addr + "/ready")
	require.NoError(t, err)
	assert.Equal(t, http.StatusServiceUnavailable, resp.StatusCode)
	_ = resp.Body.Close()
}

func TestHealthServer_WithChecks(t *testing.T) {
	hs := NewHealthServer("127.0.0.1:0")

	healthy := true
	hs.AddCheck(func() error {
		if !healthy {
			return assert.AnError
		}
		return nil
	})

	err := hs.Start()
	require.NoError(t, err)
	defer func() { _ = hs.Stop(context.Background()) }()

	addr := hs.Addr()

	// Should be healthy
	resp, err := http.Get("http://" + addr + "/health")
	require.NoError(t, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
	_ = resp.Body.Close()

	// Make unhealthy
	healthy = false

	resp, err = http.Get("http://" + addr + "/health")
	require.NoError(t, err)
	assert.Equal(t, http.StatusServiceUnavailable, resp.StatusCode)
	_ = resp.Body.Close()
}

func TestWaitForHealth_Reachable(t *testing.T) {
	hs := NewHealthServer("127.0.0.1:0")
	require.NoError(t, hs.Start())
	defer func() { _ = hs.Stop(context.Background()) }()

	ok := WaitForHealth(hs.Addr(), 2_000)
	assert.True(t, ok)
}

func TestWaitForHealth_Unreachable(t *testing.T) {
	// Port that nothing is listening on
	ok := WaitForHealth("127.0.0.1:19999", 500)
	assert.False(t, ok)
}
```

**Step 2: Run tests to verify they fail**

```bash
cd /Users/snider/Code/core/go-process && GOWORK=/Users/snider/Code/go.work go test -run 'TestHealthServer|TestWaitForHealth' -v ./...
```

Expected: FAIL — `NewHealthServer`, `WaitForHealth` undefined.

**Step 3: Write the implementation**

```go
// File: /Users/snider/Code/core/go-process/health.go
package process

import (
	"context"
	"fmt"
	"net"
	"net/http"
	"sync"
	"time"
)

// HealthCheck is a function that returns nil if healthy.
type HealthCheck func() error

// HealthServer provides HTTP /health and /ready endpoints for process monitoring.
type HealthServer struct {
	addr     string
	server   *http.Server
	listener net.Listener
	mu       sync.Mutex
	ready    bool
	checks   []HealthCheck
}

// NewHealthServer creates a health check server on the given address.
// Use "127.0.0.1:0" for a random port.
func NewHealthServer(addr string) *HealthServer {
	return &HealthServer{
		addr:  addr,
		ready: true,
	}
}

// AddCheck registers a health check function.
func (h *HealthServer) AddCheck(check HealthCheck) {
	h.mu.Lock()
	h.checks = append(h.checks, check)
	h.mu.Unlock()
}

// SetReady sets the readiness status.
func (h *HealthServer) SetReady(ready bool) {
	h.mu.Lock()
	h.ready = ready
	h.mu.Unlock()
}

// Start begins serving health check endpoints.
//   - /health — liveness probe (200 if all checks pass, 503 if any fail)
//   - /ready  — readiness probe (200 if ready, 503 if not)
func (h *HealthServer) Start() error {
	mux := http.NewServeMux()

	mux.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		h.mu.Lock()
		checks := h.checks
		h.mu.Unlock()

		for _, check := range checks {
			if err := check(); err != nil {
				w.WriteHeader(http.StatusServiceUnavailable)
				_, _ = fmt.Fprintf(w, "unhealthy: %v\n", err)
				return
			}
		}

		w.WriteHeader(http.StatusOK)
		_, _ = fmt.Fprintln(w, "ok")
	})

	mux.HandleFunc("/ready", func(w http.ResponseWriter, r *http.Request) {
		h.mu.Lock()
		ready := h.ready
		h.mu.Unlock()

		if !ready {
			w.WriteHeader(http.StatusServiceUnavailable)
			_, _ = fmt.Fprintln(w, "not ready")
			return
		}

		w.WriteHeader(http.StatusOK)
		_, _ = fmt.Fprintln(w, "ready")
	})

	listener, err := net.Listen("tcp", h.addr)
	if err != nil {
		return fmt.Errorf("failed to listen on %s: %w", h.addr, err)
	}

	h.listener = listener
	h.server = &http.Server{Handler: mux}

	go func() {
		_ = h.server.Serve(listener)
	}()

	return nil
}

// Stop gracefully shuts down the health server.
func (h *HealthServer) Stop(ctx context.Context) error {
	if h.server == nil {
		return nil
	}
	return h.server.Shutdown(ctx)
}

// Addr returns the actual address the server is listening on.
// Useful when using port 0 for dynamic port assignment.
func (h *HealthServer) Addr() string {
	if h.listener != nil {
		return h.listener.Addr().String()
	}
	return h.addr
}

// WaitForHealth polls a health endpoint until it responds 200 or the timeout
// (in milliseconds) expires. Returns true if healthy, false on timeout.
func WaitForHealth(addr string, timeoutMs int) bool {
	deadline := time.Now().Add(time.Duration(timeoutMs) * time.Millisecond)
	url := fmt.Sprintf("http://%s/health", addr)

	for time.Now().Before(deadline) {
		resp, err := http.Get(url)
		if err == nil {
			resp.Body.Close()
			if resp.StatusCode == http.StatusOK {
				return true
			}
		}
		time.Sleep(200 * time.Millisecond)
	}

	return false
}
```

**Step 4: Run tests to verify they pass**

```bash
cd /Users/snider/Code/core/go-process && GOWORK=/Users/snider/Code/go.work go test -run 'TestHealthServer|TestWaitForHealth' -v ./...
```

Expected: PASS

**Step 5: Commit**

```bash
cd /Users/snider/Code/core/go-process
git add health.go health_test.go
git commit -m "feat: add HealthServer and WaitForHealth for daemon monitoring"
```

---

### Task 3: Create `daemon.go` in go-process (Daemon orchestrator)

**Files:**
- Create: `go-process/daemon.go`
- Create: `go-process/daemon_test.go`

**Context:** Daemon wraps PIDFile + HealthServer + signal handling into a managed lifecycle. Currently in `cli/pkg/cli/daemon.go:248-404`. This is the orchestrator — call `Start()` to acquire PID + start health, `Run(ctx)` to block until context cancellation, `Stop()` for graceful shutdown.

**Step 1: Write the test file**

```go
// File: /Users/snider/Code/core/go-process/daemon_test.go
package process

import (
	"context"
	"net/http"
	"path/filepath"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestDaemon_StartAndStop(t *testing.T) {
	pidPath := filepath.Join(t.TempDir(), "test.pid")

	d := NewDaemon(DaemonOptions{
		PIDFile:         pidPath,
		HealthAddr:      "127.0.0.1:0",
		ShutdownTimeout: 5 * time.Second,
	})

	err := d.Start()
	require.NoError(t, err)

	// Health server should be running
	addr := d.HealthAddr()
	require.NotEmpty(t, addr)

	resp, err := http.Get("http://" + addr + "/health")
	require.NoError(t, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
	_ = resp.Body.Close()

	// Stop should succeed
	err = d.Stop()
	require.NoError(t, err)
}

func TestDaemon_DoubleStartFails(t *testing.T) {
	d := NewDaemon(DaemonOptions{
		HealthAddr: "127.0.0.1:0",
	})

	err := d.Start()
	require.NoError(t, err)
	defer func() { _ = d.Stop() }()

	err = d.Start()
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "already running")
}

func TestDaemon_RunWithoutStartFails(t *testing.T) {
	d := NewDaemon(DaemonOptions{})

	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	err := d.Run(ctx)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "not started")
}

func TestDaemon_SetReady(t *testing.T) {
	d := NewDaemon(DaemonOptions{
		HealthAddr: "127.0.0.1:0",
	})

	err := d.Start()
	require.NoError(t, err)
	defer func() { _ = d.Stop() }()

	addr := d.HealthAddr()

	// Initially ready
	resp, _ := http.Get("http://" + addr + "/ready")
	assert.Equal(t, http.StatusOK, resp.StatusCode)
	_ = resp.Body.Close()

	// Set not ready
	d.SetReady(false)

	resp, _ = http.Get("http://" + addr + "/ready")
	assert.Equal(t, http.StatusServiceUnavailable, resp.StatusCode)
	_ = resp.Body.Close()
}

func TestDaemon_NoHealthAddrReturnsEmpty(t *testing.T) {
	d := NewDaemon(DaemonOptions{})
	assert.Empty(t, d.HealthAddr())
}

func TestDaemon_DefaultShutdownTimeout(t *testing.T) {
	d := NewDaemon(DaemonOptions{})
	assert.Equal(t, 30*time.Second, d.opts.ShutdownTimeout)
}
```

**Step 2: Run tests to verify they fail**

```bash
cd /Users/snider/Code/core/go-process && GOWORK=/Users/snider/Code/go.work go test -run 'TestDaemon' -v ./...
```

Expected: FAIL — `NewDaemon`, `DaemonOptions` undefined.

**Step 3: Write the implementation**

```go
// File: /Users/snider/Code/core/go-process/daemon.go
package process

import (
	"context"
	"errors"
	"fmt"
	"os"
	"sync"
	"time"
)

// DaemonOptions configures daemon mode execution.
type DaemonOptions struct {
	// PIDFile path for single-instance enforcement.
	// Leave empty to skip PID file management.
	PIDFile string

	// ShutdownTimeout is the maximum time to wait for graceful shutdown.
	// Default: 30 seconds.
	ShutdownTimeout time.Duration

	// HealthAddr is the address for health check endpoints.
	// Example: ":8080", "127.0.0.1:9000"
	// Leave empty to disable health checks.
	HealthAddr string

	// HealthChecks are additional health check functions.
	HealthChecks []HealthCheck

	// OnReload is called when SIGHUP is received.
	// Use for config reloading. Leave nil to ignore SIGHUP.
	OnReload func() error
}

// Daemon manages daemon lifecycle: PID file, health server, graceful shutdown.
type Daemon struct {
	opts    DaemonOptions
	pid     *PIDFile
	health  *HealthServer
	running bool
	mu      sync.Mutex
}

// NewDaemon creates a daemon runner with the given options.
func NewDaemon(opts DaemonOptions) *Daemon {
	if opts.ShutdownTimeout == 0 {
		opts.ShutdownTimeout = 30 * time.Second
	}

	d := &Daemon{opts: opts}

	if opts.PIDFile != "" {
		d.pid = NewPIDFile(opts.PIDFile)
	}

	if opts.HealthAddr != "" {
		d.health = NewHealthServer(opts.HealthAddr)
		for _, check := range opts.HealthChecks {
			d.health.AddCheck(check)
		}
	}

	return d
}

// Start initialises the daemon (PID file, health server).
func (d *Daemon) Start() error {
	d.mu.Lock()
	defer d.mu.Unlock()

	if d.running {
		return errors.New("daemon already running")
	}

	// Acquire PID file
	if d.pid != nil {
		if err := d.pid.Acquire(); err != nil {
			return err
		}
	}

	// Start health server
	if d.health != nil {
		if err := d.health.Start(); err != nil {
			if d.pid != nil {
				_ = d.pid.Release()
			}
			return err
		}
	}

	d.running = true
	return nil
}

// Run blocks until the context is cancelled.
// Handles graceful shutdown with the configured timeout.
func (d *Daemon) Run(ctx context.Context) error {
	d.mu.Lock()
	if !d.running {
		d.mu.Unlock()
		return errors.New("daemon not started - call Start() first")
	}
	d.mu.Unlock()

	// Wait for context cancellation (from signal handler)
	<-ctx.Done()

	return d.Stop()
}

// Stop performs graceful shutdown.
func (d *Daemon) Stop() error {
	d.mu.Lock()
	defer d.mu.Unlock()

	if !d.running {
		return nil
	}

	var errs []error

	// Create shutdown context with timeout
	shutdownCtx, cancel := context.WithTimeout(context.Background(), d.opts.ShutdownTimeout)
	defer cancel()

	// Stop health server
	if d.health != nil {
		d.health.SetReady(false)
		if err := d.health.Stop(shutdownCtx); err != nil {
			errs = append(errs, fmt.Errorf("health server: %w", err))
		}
	}

	// Release PID file
	if d.pid != nil {
		if err := d.pid.Release(); err != nil && !os.IsNotExist(err) {
			errs = append(errs, fmt.Errorf("pid file: %w", err))
		}
	}

	d.running = false

	if len(errs) > 0 {
		return fmt.Errorf("shutdown errors: %v", errs)
	}
	return nil
}

// SetReady sets the daemon readiness status for health checks.
func (d *Daemon) SetReady(ready bool) {
	if d.health != nil {
		d.health.SetReady(ready)
	}
}

// HealthAddr returns the health server address, or empty if disabled.
func (d *Daemon) HealthAddr() string {
	if d.health != nil {
		return d.health.Addr()
	}
	return ""
}
```

**Step 4: Run tests to verify they pass**

```bash
cd /Users/snider/Code/core/go-process && GOWORK=/Users/snider/Code/go.work go test -run 'TestDaemon' -v ./...
```

Expected: PASS

**Step 5: Run ALL go-process tests to check for regressions**

```bash
cd /Users/snider/Code/core/go-process && GOWORK=/Users/snider/Code/go.work go test -v ./...
```

Expected: All existing + new tests PASS.

**Step 6: Commit**

```bash
cd /Users/snider/Code/core/go-process
git add daemon.go daemon_test.go
git commit -m "feat: add Daemon orchestrator for managed process lifecycle"
```

---

### Task 4: Create generic `AddDaemonCommand` in core/cli

**Files:**
- Create: `cli/pkg/cli/daemon_cmd.go`
- Create: `cli/pkg/cli/daemon_cmd_test.go`

**Context:** This is the generic CLI command builder that replaces `go-ai/cmd/daemon/cmd.go`. It provides start/stop/status/run subcommands and accepts a `RunForeground` callback for business logic. The caller (go-ai, ide, etc.) provides their own foreground function.

The `start` command re-execs the binary as a detached process. It needs to pass flags through to the `run` subcommand. The config provides `ExtraStartArgs` for app-specific flags.

**Step 1: Write the test file**

```go
// File: /Users/snider/Code/core/cli/pkg/cli/daemon_cmd_test.go
package cli

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestAddDaemonCommand_RegistersSubcommands(t *testing.T) {
	root := &Command{Use: "test"}

	AddDaemonCommand(root, DaemonCommandConfig{
		Name:       "daemon",
		PIDFile:    "/tmp/test-daemon.pid",
		HealthAddr: "127.0.0.1:0",
	})

	// Should have the daemon command
	daemonCmd, _, err := root.Find([]string{"daemon"})
	require.NoError(t, err)
	require.NotNil(t, daemonCmd)

	// Should have subcommands
	var subNames []string
	for _, sub := range daemonCmd.Commands() {
		subNames = append(subNames, sub.Name())
	}
	assert.Contains(t, subNames, "start")
	assert.Contains(t, subNames, "stop")
	assert.Contains(t, subNames, "status")
	assert.Contains(t, subNames, "run")
}

func TestDaemonCommandConfig_Defaults(t *testing.T) {
	cfg := DaemonCommandConfig{}

	if cfg.Name == "" {
		cfg.Name = "daemon"
	}
	assert.Equal(t, "daemon", cfg.Name)
}
```

**Step 2: Run tests to verify they fail**

```bash
cd /Users/snider/Code/core/cli && GOWORK=/Users/snider/Code/go.work go test -run 'TestAddDaemonCommand|TestDaemonCommandConfig' -v ./pkg/cli/
```

Expected: FAIL — `AddDaemonCommand`, `DaemonCommandConfig` undefined.

**Step 3: Write the implementation**

```go
// File: /Users/snider/Code/core/cli/pkg/cli/daemon_cmd.go
package cli

import (
	"fmt"
	"net/http"
	"os"
	"os/exec"
	"syscall"
	"time"

	"forge.lthn.ai/core/go-process"
)

// DaemonCommandConfig configures the generic daemon CLI command group.
type DaemonCommandConfig struct {
	// Name is the command group name (default: "daemon").
	Name string

	// Description is the short description for the command group.
	Description string

	// RunForeground is called when the daemon runs in foreground mode.
	// Receives context (cancelled on SIGINT/SIGTERM) and the started Daemon.
	// This is where business logic (MCP, HTTP server, etc.) goes.
	// If nil, the run command just blocks until signal.
	RunForeground func(ctx Context, daemon *process.Daemon) error

	// PIDFile default path.
	PIDFile string

	// HealthAddr default address.
	HealthAddr string

	// ExtraStartArgs returns additional CLI args to pass when re-execing
	// the binary as a background daemon. These are appended after the
	// standard --pid-file and --health-addr flags.
	ExtraStartArgs func() []string

	// Flags registers custom persistent flags on the daemon command group.
	// Called after standard flags (--pid-file, --health-addr) are registered.
	Flags func(cmd *Command)
}

// AddDaemonCommand registers start/stop/status/run subcommands on root.
func AddDaemonCommand(root *Command, cfg DaemonCommandConfig) {
	if cfg.Name == "" {
		cfg.Name = "daemon"
	}
	if cfg.Description == "" {
		cfg.Description = "Manage the background daemon"
	}

	daemonCmd := NewGroup(
		cfg.Name,
		cfg.Description,
		fmt.Sprintf("Manage the background daemon process.\n\n"+
			"Subcommands:\n"+
			"  start   - Start the daemon in the background\n"+
			"  stop    - Stop the running daemon\n"+
			"  status  - Show daemon status\n"+
			"  run     - Run in foreground (for development/debugging)"),
	)

	// Standard persistent flags
	PersistentStringFlag(daemonCmd, &cfg.HealthAddr, "health-addr", "", cfg.HealthAddr,
		"Health check endpoint address (empty to disable)")
	PersistentStringFlag(daemonCmd, &cfg.PIDFile, "pid-file", "", cfg.PIDFile,
		"PID file path (empty to disable)")

	// Custom flags from caller
	if cfg.Flags != nil {
		cfg.Flags(daemonCmd)
	}

	startCmd := NewCommand("start", "Start the daemon in the background",
		"Re-executes the binary as a background daemon process.\n"+
			"The daemon PID is written to the PID file for later management.",
		func(cmd *Command, args []string) error {
			return daemonRunStart(cfg)
		},
	)

	stopCmd := NewCommand("stop", "Stop the running daemon",
		"Sends SIGTERM to the daemon process identified by the PID file.\n"+
			"Waits for graceful shutdown before returning.",
		func(cmd *Command, args []string) error {
			return daemonRunStop(cfg)
		},
	)

	statusCmd := NewCommand("status", "Show daemon status",
		"Checks if the daemon is running and queries its health endpoint.",
		func(cmd *Command, args []string) error {
			return daemonRunStatus(cfg)
		},
	)

	runCmd := NewCommand("run", "Run the daemon in the foreground",
		"Runs the daemon in the current terminal (blocks until SIGINT/SIGTERM).\n"+
			"Useful for development, debugging, or running under a process manager.",
		func(cmd *Command, args []string) error {
			return daemonRunForeground(cfg)
		},
	)

	daemonCmd.AddCommand(startCmd, stopCmd, statusCmd, runCmd)
	root.AddCommand(daemonCmd)
}

// daemonRunStart re-execs the current binary as a detached daemon process.
func daemonRunStart(cfg DaemonCommandConfig) error {
	// Check if already running
	if pid, running := process.ReadPID(cfg.PIDFile); running {
		return fmt.Errorf("daemon already running (PID %d)", pid)
	}

	// Find the current binary
	exePath, err := os.Executable()
	if err != nil {
		return fmt.Errorf("failed to find executable: %w", err)
	}

	// Build args for the foreground run command
	args := []string{cfg.Name, "run",
		"--health-addr", cfg.HealthAddr,
		"--pid-file", cfg.PIDFile,
	}

	// Append app-specific flags
	if cfg.ExtraStartArgs != nil {
		args = append(args, cfg.ExtraStartArgs()...)
	}

	// Launch detached child with CORE_DAEMON=1
	cmd := exec.Command(exePath, args...)
	cmd.Env = append(os.Environ(), "CORE_DAEMON=1")
	cmd.Stdout = nil
	cmd.Stderr = nil
	cmd.Stdin = nil
	cmd.SysProcAttr = &syscall.SysProcAttr{
		Setsid: true,
	}

	if err := cmd.Start(); err != nil {
		return fmt.Errorf("failed to start daemon: %w", err)
	}

	pid := cmd.Process.Pid
	_ = cmd.Process.Release()

	// Wait briefly for the health endpoint
	if cfg.HealthAddr != "" {
		if process.WaitForHealth(cfg.HealthAddr, 5_000) {
			LogInfo(fmt.Sprintf("Daemon started (PID %d, health %s)", pid, cfg.HealthAddr))
		} else {
			LogInfo(fmt.Sprintf("Daemon started (PID %d, health not yet ready)", pid))
		}
	} else {
		LogInfo(fmt.Sprintf("Daemon started (PID %d)", pid))
	}

	return nil
}

// daemonRunStop sends SIGTERM to the daemon process.
func daemonRunStop(cfg DaemonCommandConfig) error {
	pid, running := process.ReadPID(cfg.PIDFile)
	if !running {
		LogInfo("Daemon is not running")
		return nil
	}

	proc, err := os.FindProcess(pid)
	if err != nil {
		return fmt.Errorf("failed to find process %d: %w", pid, err)
	}

	LogInfo(fmt.Sprintf("Stopping daemon (PID %d)", pid))
	if err := proc.Signal(syscall.SIGTERM); err != nil {
		return fmt.Errorf("failed to send SIGTERM to PID %d: %w", pid, err)
	}

	// Wait for the process to exit (poll PID file removal)
	deadline := time.Now().Add(30 * time.Second)
	for time.Now().Before(deadline) {
		if _, still := process.ReadPID(cfg.PIDFile); !still {
			LogInfo("Daemon stopped")
			return nil
		}
		time.Sleep(250 * time.Millisecond)
	}

	LogWarn("Daemon did not stop within 30s, sending SIGKILL")
	_ = proc.Signal(syscall.SIGKILL)
	_ = os.Remove(cfg.PIDFile)
	LogInfo("Daemon killed")
	return nil
}

// daemonRunStatus checks daemon status via PID and health endpoint.
func daemonRunStatus(cfg DaemonCommandConfig) error {
	pid, running := process.ReadPID(cfg.PIDFile)
	if !running {
		fmt.Println("Daemon is not running")
		return nil
	}

	fmt.Printf("Daemon is running (PID %d)\n", pid)

	if cfg.HealthAddr != "" {
		healthURL := fmt.Sprintf("http://%s/health", cfg.HealthAddr)
		resp, err := http.Get(healthURL)
		if err != nil {
			fmt.Printf("Health: unreachable (%v)\n", err)
			return nil
		}
		defer resp.Body.Close()

		if resp.StatusCode == http.StatusOK {
			fmt.Println("Health: ok")
		} else {
			fmt.Printf("Health: unhealthy (HTTP %d)\n", resp.StatusCode)
		}

		readyURL := fmt.Sprintf("http://%s/ready", cfg.HealthAddr)
		resp2, err := http.Get(readyURL)
		if err == nil {
			defer resp2.Body.Close()
			if resp2.StatusCode == http.StatusOK {
				fmt.Println("Ready:  yes")
			} else {
				fmt.Println("Ready:  no")
			}
		}
	}

	return nil
}

// daemonRunForeground runs the daemon in the current process (blocking).
func daemonRunForeground(cfg DaemonCommandConfig) error {
	os.Setenv("CORE_DAEMON", "1")

	daemon := process.NewDaemon(process.DaemonOptions{
		PIDFile:         cfg.PIDFile,
		HealthAddr:      cfg.HealthAddr,
		ShutdownTimeout: 30 * time.Second,
	})

	if err := daemon.Start(); err != nil {
		return fmt.Errorf("failed to start daemon: %w", err)
	}

	daemon.SetReady(true)

	ctx := Context()

	if cfg.RunForeground != nil {
		// Run the business logic — if it returns, trigger shutdown
		svcErr := make(chan error, 1)
		go func() {
			svcErr <- cfg.RunForeground(ctx, daemon)
		}()

		select {
		case <-ctx.Done():
			LogInfo("Shutting down daemon")
		case err := <-svcErr:
			if err != nil {
				LogError(fmt.Sprintf("Service exited with error: %v", err))
			}
		}
	} else {
		// No business logic — just block until signal
		<-ctx.Done()
	}

	return daemon.Stop()
}
```

**Step 4: Add `go-process` to cli's go.mod**

```bash
cd /Users/snider/Code/core/cli && GOWORK=/Users/snider/Code/go.work go get forge.lthn.ai/core/go-process
```

**Note:** This may not update go.mod cleanly outside workspace. If it fails, manually add `forge.lthn.ai/core/go-process v0.1.2` to the `require` block in `go.mod` and let the workspace resolve it.

**Step 5: Run tests to verify they pass**

```bash
cd /Users/snider/Code/core/cli && GOWORK=/Users/snider/Code/go.work go test -run 'TestAddDaemonCommand|TestDaemonCommandConfig' -v ./pkg/cli/
```

Expected: PASS

**Step 6: Build to verify compilation**

```bash
cd /Users/snider/Code/core/cli && GOWORK=/Users/snider/Code/go.work go build ./...
```

Expected: Success

**Step 7: Commit**

```bash
cd /Users/snider/Code/core/cli
git add pkg/cli/daemon_cmd.go pkg/cli/daemon_cmd_test.go go.mod go.sum
git commit -m "feat: add generic AddDaemonCommand with go-process daemon types"
```

---

### Task 5: Update go-ai to use new daemon commands

**Files:**
- Modify: `go-ai/cmd/daemon/cmd.go` — replace with thin wrapper
- Modify: `go-ai/go.mod` — may need go-process dep (likely already via cli)

**Context:** go-ai's `cmd/daemon/cmd.go` (393 LOC) becomes a thin wrapper that passes its MCP-specific `runForeground` callback to `cli.AddDaemonCommand()`. Transport constants and MCP config stay. The generic daemon lifecycle moves out.

**Step 1: Read current file to understand MCP-specific parts**

Read `/Users/snider/Code/core/go-ai/cmd/daemon/cmd.go` — specifically the `Config`, `ConfigFromEnv`, `startMCP`, and transport constants. These stay.

**Step 2: Rewrite cmd.go**

```go
// File: /Users/snider/Code/core/go-ai/cmd/daemon/cmd.go
// Package daemon provides the `core daemon` command for running as a background service.
package daemon

import (
	"context"
	"fmt"
	"os"
	"path/filepath"

	"forge.lthn.ai/core/cli/pkg/cli"
	"forge.lthn.ai/core/go-ai/mcp"
	"forge.lthn.ai/core/go-log"
	"forge.lthn.ai/core/go-process"
)

// Transport types for MCP server.
const (
	TransportStdio  = "stdio"
	TransportTCP    = "tcp"
	TransportSocket = "socket"
)

// Config holds daemon configuration.
type Config struct {
	// MCPTransport is the MCP server transport type (stdio, tcp, socket).
	MCPTransport string
	// MCPAddr is the address/path for tcp or socket transports.
	MCPAddr string
	// HealthAddr is the address for health check endpoints.
	HealthAddr string
	// PIDFile is the path for the PID file.
	PIDFile string
}

// DefaultConfig returns the default daemon configuration.
func DefaultConfig() Config {
	home, _ := os.UserHomeDir()
	return Config{
		MCPTransport: TransportTCP,
		MCPAddr:      mcp.DefaultTCPAddr,
		HealthAddr:   "127.0.0.1:9101",
		PIDFile:      filepath.Join(home, ".core", "daemon.pid"),
	}
}

// ConfigFromEnv loads configuration from environment variables.
func ConfigFromEnv() Config {
	cfg := DefaultConfig()

	if v := os.Getenv("CORE_MCP_TRANSPORT"); v != "" {
		cfg.MCPTransport = v
	}
	if v := os.Getenv("CORE_MCP_ADDR"); v != "" {
		cfg.MCPAddr = v
	}
	if v := os.Getenv("CORE_HEALTH_ADDR"); v != "" {
		cfg.HealthAddr = v
	}
	if v := os.Getenv("CORE_PID_FILE"); v != "" {
		cfg.PIDFile = v
	}

	return cfg
}

// AddDaemonCommand adds the 'daemon' command group to the root.
func AddDaemonCommand(root *cli.Command) {
	cfg := ConfigFromEnv()

	cli.AddDaemonCommand(root, cli.DaemonCommandConfig{
		Name:        "daemon",
		Description: "Manage the core daemon",
		PIDFile:     cfg.PIDFile,
		HealthAddr:  cfg.HealthAddr,
		RunForeground: func(ctx cli.Context, daemon *process.Daemon) error {
			log.Info("Starting MCP service",
				"transport", cfg.MCPTransport,
				"addr", cfg.MCPAddr,
			)

			mcpSvc, err := mcp.New()
			if err != nil {
				return fmt.Errorf("failed to create MCP service: %w", err)
			}

			return startMCP(ctx, mcpSvc, cfg)
		},
		ExtraStartArgs: func() []string {
			return []string{
				"--mcp-transport", cfg.MCPTransport,
				"--mcp-addr", cfg.MCPAddr,
			}
		},
		Flags: func(cmd *cli.Command) {
			cli.PersistentStringFlag(cmd, &cfg.MCPTransport, "mcp-transport", "t", cfg.MCPTransport,
				"MCP transport type (stdio, tcp, socket)")
			cli.PersistentStringFlag(cmd, &cfg.MCPAddr, "mcp-addr", "a", cfg.MCPAddr,
				"MCP listen address (e.g., :9100 or /tmp/mcp.sock)")
		},
	})
}

// startMCP starts the MCP server with the configured transport.
func startMCP(ctx context.Context, svc *mcp.Service, cfg Config) error {
	switch cfg.MCPTransport {
	case TransportStdio:
		log.Info("Starting MCP server", "transport", "stdio")
		return svc.ServeStdio(ctx)

	case TransportTCP:
		log.Info("Starting MCP server", "transport", "tcp", "addr", cfg.MCPAddr)
		return svc.ServeTCP(ctx, cfg.MCPAddr)

	case TransportSocket:
		log.Info("Starting MCP server", "transport", "unix", "path", cfg.MCPAddr)
		return svc.ServeUnix(ctx, cfg.MCPAddr)

	default:
		return fmt.Errorf("unknown MCP transport: %s (valid: stdio, tcp, socket)", cfg.MCPTransport)
	}
}
```

**Step 3: Build to verify compilation**

```bash
cd /Users/snider/Code/core/go-ai && GOWORK=/Users/snider/Code/go.work go build ./...
```

Expected: Success

**Step 4: Commit**

```bash
cd /Users/snider/Code/core/go-ai
git add cmd/daemon/cmd.go go.mod go.sum
git commit -m "refactor: use generic cli.AddDaemonCommand, remove duplicated daemon lifecycle"
```

---

### Task 6: Update core/ide to use go-process daemon types

**Files:**
- Modify: `ide/headless.go:104-113` — change `cli.NewDaemon`/`cli.DaemonOptions` to `process.NewDaemon`/`process.DaemonOptions`

**Context:** `ide/headless.go` uses `cli.NewDaemon(cli.DaemonOptions{...})`. After the move, it should use `process.NewDaemon(process.DaemonOptions{...})`. The API is identical — only the import path changes.

**Step 1: Read the file for full context around the daemon usage**

Read `/Users/snider/Code/core/ide/headless.go` — find all references to `cli.NewDaemon`, `cli.DaemonOptions`, `cli.Daemon`.

**Step 2: Update imports**

Add `"forge.lthn.ai/core/go-process"` to imports. Remove `cli` import only if no other cli usage remains (unlikely — check first).

**Step 3: Replace type references**

Change:
```go
daemon := cli.NewDaemon(cli.DaemonOptions{
    PIDFile:    filepath.Join(os.Getenv("HOME"), ".core", "core-ide.pid"),
    HealthAddr: "127.0.0.1:9878",
})
```

To:
```go
daemon := process.NewDaemon(process.DaemonOptions{
    PIDFile:    filepath.Join(os.Getenv("HOME"), ".core", "core-ide.pid"),
    HealthAddr: "127.0.0.1:9878",
})
```

**Step 4: Build to verify compilation**

```bash
cd /Users/snider/Code/core/ide && GOWORK=/Users/snider/Code/go.work go build ./...
```

Expected: Success

**Step 5: Commit**

```bash
cd /Users/snider/Code/core/ide
git add headless.go go.mod go.sum
git commit -m "refactor: use process.Daemon from go-process instead of cli.Daemon"
```

---

### Task 7: Clean up cli/daemon.go — remove extracted types

**Files:**
- Modify: `cli/pkg/cli/daemon.go` — remove PIDFile, HealthServer, Daemon, DaemonOptions, NewPIDFile, NewHealthServer, NewDaemon (moved to go-process)
- Modify: `cli/pkg/cli/daemon_test.go` — remove tests for moved types

**Context:** After Tasks 4-6, the moved types in cli are unused. Keep only CLI-specific code:
- `Mode`, `DetectMode()` — execution mode detection
- `IsTTY()`, `IsStdinTTY()`, `IsStderrTTY()` — terminal helpers
- `Run()`, `RunWithTimeout()` — CLI convenience functions

**Step 1: Reduce daemon.go to CLI-specific code only**

The file should contain only lines 1-72 (Mode/DetectMode/IsTTY) and lines 406-448 (Run/RunWithTimeout) from the original. Everything in between (PIDFile, HealthServer, Daemon) is deleted.

**Step 2: Reduce daemon_test.go to kept tests only**

Keep only `TestDetectMode` and `TestRunWithTimeout`. Remove `TestPIDFile`, `TestHealthServer`, `TestDaemon` (now tested in go-process).

**Step 3: Build and test**

```bash
cd /Users/snider/Code/core/cli && GOWORK=/Users/snider/Code/go.work go build ./... && go test -v ./pkg/cli/ -run 'TestDetectMode|TestRunWithTimeout'
```

Expected: Success — no compilation errors, remaining tests pass.

**Step 4: Run full cli test suite**

```bash
cd /Users/snider/Code/core/cli && GOWORK=/Users/snider/Code/go.work go test ./...
```

Expected: All tests PASS. If any test references removed types, update it to import from go-process.

**Step 5: Commit**

```bash
cd /Users/snider/Code/core/cli
git add pkg/cli/daemon.go pkg/cli/daemon_test.go
git commit -m "refactor: remove daemon types moved to go-process, keep CLI-specific helpers"
```

---

### Task 8: Final verification — build all affected modules

**Files:** None (verification only)

**Step 1: Build all affected modules**

```bash
cd /Users/snider/Code/core/go-process && GOWORK=/Users/snider/Code/go.work go test ./...
cd /Users/snider/Code/core/cli && GOWORK=/Users/snider/Code/go.work go test ./...
cd /Users/snider/Code/core/go-ai && GOWORK=/Users/snider/Code/go.work go build ./...
cd /Users/snider/Code/core/ide && GOWORK=/Users/snider/Code/go.work go build ./...
```

Expected: All build and test successfully.

**Step 2: Verify no stale references to cli.NewDaemon or cli.DaemonOptions**

Search the workspace for old import patterns:

```bash
cd /Users/snider/Code && grep -r 'cli\.NewDaemon\|cli\.DaemonOptions\|cli\.PIDFile\|cli\.HealthServer\|cli\.HealthCheck' --include='*.go' core/
```

Expected: No matches (except possibly in core-org/ which is archived/old).

**Step 3: Verify go-process exports are correct**

```bash
cd /Users/snider/Code/core/go-process && GOWORK=/Users/snider/Code/go.work go doc .
```

Expected: Shows `Daemon`, `DaemonOptions`, `PIDFile`, `HealthServer`, `HealthCheck`, `ReadPID`, `WaitForHealth` alongside existing `Process`, `Runner`, `RunSpec`, etc.

**Step 4: Tag go-process if all is well**

Once verified, tag go-process with the next version:

```bash
cd /Users/snider/Code/core/go-process
git tag v0.2.0
git push origin main --tags
```

**Note:** Confirm the current version first (`git tag --list`). If latest is v0.1.x, bump to v0.2.0 for the new daemon feature.
