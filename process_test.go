package core_test

import (
	. "dappco.re/go"
)

func TestProcess_Core_Process_Good(t *T) {
	c := New()

	AssertNotNil(t, c.Process())
}

func TestProcess_Core_Process_Bad(t *T) {
	var c *Core
	var p *Process

	AssertNotPanics(t, func() { p = c.Process() })
	AssertNotNil(t, p)
}

func TestProcess_Core_Process_Ugly(t *T) {
	c := New()

	AssertFalse(t, c.Process() == c.Process())
}

func TestProcess_Process_Exists_Good(t *T) {
	c := New()
	c.Action("process.run", func(Context, Options) Result { return Result{OK: true} })

	AssertTrue(t, c.Process().Exists())
}

func TestProcess_Process_Exists_Bad(t *T) {
	c := New()

	AssertFalse(t, c.Process().Exists())
}

func TestProcess_Process_Exists_Ugly(t *T) {
	c := New()
	c.Action("process.start", func(Context, Options) Result { return Result{OK: true} })

	AssertFalse(t, c.Process().Exists())
}

func TestProcess_Process_Kill_Good(t *T) {
	c := New()
	c.Action("process.kill", func(_ Context, opts Options) Result {
		return Result{Value: opts.String("id"), OK: true}
	})

	r := c.Process().Kill(Background(), NewOptions(Option{Key: "id", Value: "agent-42"}))

	AssertTrue(t, r.OK)
	AssertEqual(t, "agent-42", r.Value)
}

func TestProcess_Process_Kill_Bad(t *T) {
	c := New()
	r := c.Process().Kill(Background(), NewOptions(Option{Key: "id", Value: "agent-42"}))

	AssertFalse(t, r.OK)
}

func TestProcess_Process_Kill_Ugly(t *T) {
	c := New()
	c.Action("process.kill", func(_ Context, opts Options) Result {
		return Result{Value: opts.Len(), OK: true}
	})

	r := c.Process().Kill(Background(), NewOptions())

	AssertTrue(t, r.OK)
	AssertEqual(t, 0, r.Value)
}

func TestProcess_Process_Run_Good(t *T) {
	c := New()
	c.Action("process.run", func(_ Context, opts Options) Result {
		return Result{Value: Join(" ", opts.String("command"), Join(" ", opts.Get("args").Value.([]string)...)), OK: true}
	})

	r := c.Process().Run(Background(), "agentctl", "dispatch", "codex")

	AssertTrue(t, r.OK)
	AssertEqual(t, "agentctl dispatch codex", r.Value)
}

func TestProcess_Process_Run_Bad(t *T) {
	c := New()
	r := c.Process().Run(Background(), "agentctl", "dispatch")

	AssertFalse(t, r.OK)
}

func TestProcess_Process_Run_Ugly(t *T) {
	c := New()
	c.Action("process.run", func(Context, Options) Result {
		panic("agent runner crashed")
	})

	r := c.Process().Run(Background(), "agentctl")

	AssertFalse(t, r.OK)
	AssertError(t, r.Value.(error), "panic in action")
}

func TestProcess_Process_RunIn_Good(t *T) {
	c := New()
	c.Action("process.run", func(_ Context, opts Options) Result {
		return Result{Value: Join("@", opts.String("command"), opts.String("dir")), OK: true}
	})

	r := c.Process().RunIn(Background(), "/srv/dappcore", "go", "test")

	AssertTrue(t, r.OK)
	AssertEqual(t, "go@/srv/dappcore", r.Value)
}

func TestProcess_Process_RunIn_Bad(t *T) {
	c := New()
	r := c.Process().RunIn(Background(), "/srv/dappcore", "go", "test")

	AssertFalse(t, r.OK)
}

func TestProcess_Process_RunIn_Ugly(t *T) {
	c := New()
	c.Action("process.run", func(_ Context, opts Options) Result {
		return Result{Value: opts.String("dir"), OK: true}
	})

	r := c.Process().RunIn(Background(), "", "agentctl")

	AssertTrue(t, r.OK)
	AssertEqual(t, "", r.Value)
}

func TestProcess_Process_RunWithEnv_Good(t *T) {
	c := New()
	c.Action("process.run", func(_ Context, opts Options) Result {
		env := opts.Get("env").Value.([]string)
		return Result{Value: env[0], OK: true}
	})

	r := c.Process().RunWithEnv(Background(), "/srv/dappcore", []string{"GOWORK=off"}, "go", "test")

	AssertTrue(t, r.OK)
	AssertEqual(t, "GOWORK=off", r.Value)
}

func TestProcess_Process_RunWithEnv_Bad(t *T) {
	c := New()
	r := c.Process().RunWithEnv(Background(), "/srv/dappcore", []string{"GOWORK=off"}, "go", "test")

	AssertFalse(t, r.OK)
}

func TestProcess_Process_RunWithEnv_Ugly(t *T) {
	c := New()
	c.Action("process.run", func(_ Context, opts Options) Result {
		return Result{Value: opts.Get("env").Value, OK: true}
	})

	r := c.Process().RunWithEnv(Background(), "", nil, "agentctl")

	AssertTrue(t, r.OK)
	AssertNil(t, r.Value)
}

func TestProcess_Process_Start_Good(t *T) {
	c := New()
	c.Action("process.start", func(_ Context, opts Options) Result {
		return Result{Value: opts.String("command"), OK: true}
	})

	r := c.Process().Start(Background(), NewOptions(Option{Key: "command", Value: "agentd"}))

	AssertTrue(t, r.OK)
	AssertEqual(t, "agentd", r.Value)
}

func TestProcess_Process_Start_Bad(t *T) {
	c := New()
	r := c.Process().Start(Background(), NewOptions(Option{Key: "command", Value: "agentd"}))

	AssertFalse(t, r.OK)
}

func TestProcess_Process_Start_Ugly(t *T) {
	c := New()
	c.Action("process.start", func(_ Context, opts Options) Result {
		return Result{Value: opts.Len(), OK: true}
	})

	r := c.Process().Start(Background(), NewOptions())

	AssertTrue(t, r.OK)
	AssertEqual(t, 0, r.Value)
}
