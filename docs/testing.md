---
title: Testing
description: Test naming and testing patterns used by CoreGO.
---

# Testing

The repository uses `github.com/stretchr/testify/assert` and a simple AX-friendly naming pattern.

## Test Names

Use:

- `_Good` for expected success
- `_Bad` for expected failure
- `_Ugly` for panics, degenerate input, and edge behavior

Examples from this repository:

```go
func TestNew_Good(t *testing.T) {}
func TestService_Register_Duplicate_Bad(t *testing.T) {}
func TestCore_Must_Ugly(t *testing.T) {}
```

## Start with a Small Core

```go
c := core.New(core.Options{
	{Key: "name", Value: "test-core"},
})
```

Then register only the pieces your test needs.

## Test a Service

```go
started := false

c.Service("audit", core.Service{
	OnStart: func() core.Result {
		started = true
		return core.Result{OK: true}
	},
})

r := c.ServiceStartup(context.Background(), nil)
assert.True(t, r.OK)
assert.True(t, started)
```

## Test a Command

```go
c.Command("greet", core.Command{
	Action: func(opts core.Options) core.Result {
		return core.Result{Value: "hello " + opts.String("name"), OK: true}
	},
})

r := c.Cli().Run("greet", "--name=world")
assert.True(t, r.OK)
assert.Equal(t, "hello world", r.Value)
```

## Test a Query or Task

```go
c.RegisterQuery(func(_ *core.Core, q core.Query) core.Result {
	if q == "ping" {
		return core.Result{Value: "pong", OK: true}
	}
	return core.Result{}
})

assert.Equal(t, "pong", c.QUERY("ping").Value)
```

```go
c.RegisterTask(func(_ *core.Core, t core.Task) core.Result {
	if t == "compute" {
		return core.Result{Value: 42, OK: true}
	}
	return core.Result{}
})

assert.Equal(t, 42, c.PERFORM("compute").Value)
```

## Test Async Work

For `PerformAsync`, observe completion through the action bus.

```go
completed := make(chan core.ActionTaskCompleted, 1)

c.RegisterAction(func(_ *core.Core, msg core.Message) core.Result {
	if event, ok := msg.(core.ActionTaskCompleted); ok {
		completed <- event
	}
	return core.Result{OK: true}
})
```

Then wait with normal Go test tools such as channels, timers, or `assert.Eventually`.

## Use Real Temporary Paths

When testing `Fs`, `Data.Extract`, or other I/O helpers, use `t.TempDir()` and create realistic paths instead of mocking the filesystem by default.

## Repository Commands

```bash
core go test
core go test --run TestPerformAsync_Good
go test ./...
```
