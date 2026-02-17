package coredeno

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"sync"
)

// Options configures the CoreDeno sidecar.
type Options struct {
	DenoPath   string // path to deno binary (default: "deno")
	SocketPath string // Unix socket path for gRPC
}

// Permissions declares per-module Deno permission flags.
type Permissions struct {
	Read  []string
	Write []string
	Net   []string
	Run   []string
}

// Flags converts permissions to Deno --allow-* CLI flags.
func (p Permissions) Flags() []string {
	var flags []string
	if len(p.Read) > 0 {
		flags = append(flags, fmt.Sprintf("--allow-read=%s", strings.Join(p.Read, ",")))
	}
	if len(p.Write) > 0 {
		flags = append(flags, fmt.Sprintf("--allow-write=%s", strings.Join(p.Write, ",")))
	}
	if len(p.Net) > 0 {
		flags = append(flags, fmt.Sprintf("--allow-net=%s", strings.Join(p.Net, ",")))
	}
	if len(p.Run) > 0 {
		flags = append(flags, fmt.Sprintf("--allow-run=%s", strings.Join(p.Run, ",")))
	}
	return flags
}

// DefaultSocketPath returns the default Unix socket path.
func DefaultSocketPath() string {
	xdg := os.Getenv("XDG_RUNTIME_DIR")
	if xdg == "" {
		xdg = "/tmp"
	}
	return filepath.Join(xdg, "core", "deno.sock")
}

// Sidecar manages a Deno child process.
type Sidecar struct {
	opts   Options
	mu     sync.RWMutex
	cmd    *exec.Cmd
	ctx    context.Context
	cancel context.CancelFunc
}

// NewSidecar creates a Sidecar with the given options.
func NewSidecar(opts Options) *Sidecar {
	if opts.DenoPath == "" {
		opts.DenoPath = "deno"
	}
	if opts.SocketPath == "" {
		opts.SocketPath = DefaultSocketPath()
	}
	return &Sidecar{opts: opts}
}
