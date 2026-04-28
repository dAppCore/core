// SPDX-License-Identifier: EUPL-1.2

// Process is the Core primitive for managed execution.
// Methods emit via named Actions — actual execution is handled by
// whichever service registers the "process.*" actions (typically go-process).
//
// If go-process is NOT registered, all methods return Result{OK: false}.
// This is permission-by-registration: no handler = no capability.
//
// Usage:
//
//	r := c.Process().Run(ctx, "git", "log", "--oneline")
//	if r.OK { output := r.Value.(string) }
//
//	r := c.Process().RunIn(ctx, "/path/to/repo", "go", "test", "./...")
//
// Permission model:
//
//	// Full Core — process registered:
//	c := core.New(core.WithService(process.Register))
//	c.Process().Run(ctx, "git", "log")  // works
//
//	// Sandboxed Core — no process:
//	c := core.New()
//	c.Process().Run(ctx, "git", "log")  // Result{OK: false}
package core

// Process is the Core primitive for process management.
// Zero dependencies — delegates to named Actions.
//
//	c := core.New()
//	proc := c.Process()
//	if proc.Exists() {
//	    _ = proc.Run(Background(), "git", "status", "--short")
//	}
type Process struct {
	core *Core
}

// Process returns the process management primitive.
//
//	c.Process().Run(ctx, "git", "log")
func (c *Core) Process() *Process {
	return &Process{core: c}
}

// Run executes a command synchronously and returns the output.
//
//	r := c.Process().Run(ctx, "git", "log", "--oneline")
//	if r.OK { output := r.Value.(string) }
func (p *Process) Run(ctx Context, command string, args ...string) Result {
	return p.core.Action("process.run").Run(ctx, NewOptions(
		Option{Key: "command", Value: command},
		Option{Key: "args", Value: args},
	))
}

// RunIn executes a command in a specific directory.
//
//	r := c.Process().RunIn(ctx, "/repo", "go", "test", "./...")
func (p *Process) RunIn(ctx Context, dir string, command string, args ...string) Result {
	return p.core.Action("process.run").Run(ctx, NewOptions(
		Option{Key: "command", Value: command},
		Option{Key: "args", Value: args},
		Option{Key: "dir", Value: dir},
	))
}

// RunWithEnv executes with additional environment variables.
//
//	r := c.Process().RunWithEnv(ctx, dir, []string{"GOWORK=off"}, "go", "test")
func (p *Process) RunWithEnv(ctx Context, dir string, env []string, command string, args ...string) Result {
	return p.core.Action("process.run").Run(ctx, NewOptions(
		Option{Key: "command", Value: command},
		Option{Key: "args", Value: args},
		Option{Key: "dir", Value: dir},
		Option{Key: "env", Value: env},
	))
}

// Start spawns a detached/background process.
//
//	r := c.Process().Start(ctx, ProcessStartOptions{Command: "docker", Args: []string{"run", "..."}})
func (p *Process) Start(ctx Context, opts Options) Result {
	return p.core.Action("process.start").Run(ctx, opts)
}

// Kill terminates a managed process by ID or PID.
//
//	c.Process().Kill(ctx, core.NewOptions(core.Option{Key: "id", Value: processID}))
func (p *Process) Kill(ctx Context, opts Options) Result {
	return p.core.Action("process.kill").Run(ctx, opts)
}

// Exists returns true if any process execution capability is registered.
//
//	if c.Process().Exists() { /* can run commands */ }
func (p *Process) Exists() bool {
	return p.core.Action("process.run").Exists()
}
