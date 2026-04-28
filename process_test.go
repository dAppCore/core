package core_test

import (
	"context"
	"testing"

	. "dappco.re/go/core"
)

// --- Process.Run ---

func TestProcess_Run_Good(t *testing.T) {
	c := New()
	// Register a mock process handler
	c.Action("process.run", func(_ context.Context, opts Options) Result {
		cmd := opts.String("command")
		return Result{Value: Concat("output of ", cmd), OK: true}
	})

	r := c.Process().Run(context.Background(), "git", "log")
	AssertTrue(t, r.OK)
	AssertEqual(t, "output of git", r.Value)
}

func TestProcess_Run_Bad_NotRegistered(t *testing.T) {
	c := New()
	// No process service registered — sandboxed Core
	r := c.Process().Run(context.Background(), "git", "log")
	AssertFalse(t, r.OK, "sandboxed Core must not execute commands")
}

func TestProcess_Run_Ugly_HandlerPanics(t *testing.T) {
	c := New()
	c.Action("process.run", func(_ context.Context, _ Options) Result {
		panic("segfault")
	})
	r := c.Process().Run(context.Background(), "test")
	AssertFalse(t, r.OK, "panicking handler must not crash")
}

// --- Process.RunIn ---

func TestProcess_RunIn_Good(t *testing.T) {
	c := New()
	c.Action("process.run", func(_ context.Context, opts Options) Result {
		dir := opts.String("dir")
		cmd := opts.String("command")
		return Result{Value: Concat(cmd, " in ", dir), OK: true}
	})

	r := c.Process().RunIn(context.Background(), "/repo", "go", "test")
	AssertTrue(t, r.OK)
	AssertEqual(t, "go in /repo", r.Value)
}

// --- Process.RunWithEnv ---

func TestProcess_RunWithEnv_Good(t *testing.T) {
	c := New()
	c.Action("process.run", func(_ context.Context, opts Options) Result {
		r := opts.Get("env")
		if !r.OK {
			return Result{Value: "no env", OK: true}
		}
		env := r.Value.([]string)
		return Result{Value: env[0], OK: true}
	})

	r := c.Process().RunWithEnv(context.Background(), "/repo", []string{"GOWORK=off"}, "go", "test")
	AssertTrue(t, r.OK)
	AssertEqual(t, "GOWORK=off", r.Value)
}

// --- Process.Start ---

func TestProcess_Start_Good(t *testing.T) {
	c := New()
	c.Action("process.start", func(_ context.Context, opts Options) Result {
		return Result{Value: "proc-1", OK: true}
	})

	r := c.Process().Start(context.Background(), NewOptions(
		Option{Key: "command", Value: "docker"},
		Option{Key: "args", Value: []string{"run", "nginx"}},
	))
	AssertTrue(t, r.OK)
	AssertEqual(t, "proc-1", r.Value)
}

func TestProcess_Start_Bad_NotRegistered(t *testing.T) {
	c := New()
	r := c.Process().Start(context.Background(), NewOptions())
	AssertFalse(t, r.OK)
}

// --- Process.Kill ---

func TestProcess_Kill_Good(t *testing.T) {
	c := New()
	c.Action("process.kill", func(_ context.Context, opts Options) Result {
		return Result{OK: true}
	})

	r := c.Process().Kill(context.Background(), NewOptions(
		Option{Key: "id", Value: "proc-1"},
	))
	AssertTrue(t, r.OK)
}

// --- Process.Exists ---

func TestProcess_Exists_Good(t *testing.T) {
	c := New()
	AssertFalse(t, c.Process().Exists(), "no process service = no capability")

	c.Action("process.run", func(_ context.Context, _ Options) Result {
		return Result{OK: true}
	})
	AssertTrue(t, c.Process().Exists(), "process.run registered = capability exists")
}

// --- Permission model ---

func TestProcess_Ugly_PermissionByRegistration(t *testing.T) {
	// Full Core
	full := New()
	full.Action("process.run", func(_ context.Context, opts Options) Result {
		return Result{Value: Concat("executed ", opts.String("command")), OK: true}
	})

	// Sandboxed Core
	sandboxed := New()

	// Full can execute
	AssertTrue(t, full.Process().Exists())
	r := full.Process().Run(context.Background(), "whoami")
	AssertTrue(t, r.OK)

	// Sandboxed cannot
	AssertFalse(t, sandboxed.Process().Exists())
	r = sandboxed.Process().Run(context.Background(), "whoami")
	AssertFalse(t, r.OK)
}
