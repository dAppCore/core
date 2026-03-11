---
title: Testing
description: Test naming conventions, test helpers, and patterns for Core applications.
---

# Testing

Core uses `github.com/stretchr/testify` for assertions and follows a structured test naming convention. This page covers the patterns used in the framework itself and recommended for services built on it.

## Naming Convention

Tests use a `_Good`, `_Bad`, `_Ugly` suffix pattern:

| Suffix | Purpose | Example |
|--------|---------|---------|
| `_Good` | Happy path -- expected behaviour | `TestCore_New_Good` |
| `_Bad` | Expected error conditions | `TestCore_WithService_Bad` |
| `_Ugly` | Panics, edge cases, degenerate input | `TestCore_MustServiceFor_Ugly` |

The format is `Test{Component}_{Method}_{Suffix}`:

```go
func TestCore_New_Good(t *testing.T) {
    c, err := New()
    assert.NoError(t, err)
    assert.NotNil(t, c)
}

func TestCore_WithService_Bad(t *testing.T) {
    factory := func(c *Core) (any, error) {
        return nil, assert.AnError
    }
    _, err := New(WithService(factory))
    assert.Error(t, err)
    assert.ErrorIs(t, err, assert.AnError)
}

func TestCore_MustServiceFor_Ugly(t *testing.T) {
    c, _ := New()
    assert.Panics(t, func() {
        MustServiceFor[*MockService](c, "nonexistent")
    })
}
```

## Creating a Test Core

For unit tests, create a minimal Core with only the services needed:

```go
func TestMyFeature(t *testing.T) {
    c, err := core.New()
    assert.NoError(t, err)

    // Register only what the test needs
    err = c.RegisterService("my-service", &MyService{})
    assert.NoError(t, err)
}
```

## Mock Services

Define mock services as test-local structs. Core's interface-based design makes this straightforward:

```go
// Mock a Startable service
type MockStartable struct {
    started bool
    err     error
}

func (m *MockStartable) OnStartup(ctx context.Context) error {
    m.started = true
    return m.err
}

// Mock a Stoppable service
type MockStoppable struct {
    stopped bool
    err     error
}

func (m *MockStoppable) OnShutdown(ctx context.Context) error {
    m.stopped = true
    return m.err
}
```

For services implementing both lifecycle interfaces:

```go
type MockLifecycle struct {
    MockStartable
    MockStoppable
}
```

## Testing Lifecycle

Verify that startup and shutdown are called in the correct order:

```go
func TestLifecycleOrder(t *testing.T) {
    c, _ := core.New()
    var callOrder []string

    s1 := &OrderTracker{id: "1", log: &callOrder}
    s2 := &OrderTracker{id: "2", log: &callOrder}

    _ = c.RegisterService("s1", s1)
    _ = c.RegisterService("s2", s2)

    _ = c.ServiceStartup(context.Background(), nil)
    assert.Equal(t, []string{"start-1", "start-2"}, callOrder)

    callOrder = nil
    _ = c.ServiceShutdown(context.Background())
    assert.Equal(t, []string{"stop-2", "stop-1"}, callOrder) // reverse order
}
```

## Testing Message Handlers

### Actions

Register an action handler and verify it receives the expected message:

```go
func TestAction(t *testing.T) {
    c, _ := core.New()
    var received core.Message

    c.RegisterAction(func(c *core.Core, msg core.Message) error {
        received = msg
        return nil
    })

    _ = c.ACTION(MyEvent{Data: "test"})
    event, ok := received.(MyEvent)
    assert.True(t, ok)
    assert.Equal(t, "test", event.Data)
}
```

### Queries

```go
func TestQuery(t *testing.T) {
    c, _ := core.New()

    c.RegisterQuery(func(c *core.Core, q core.Query) (any, bool, error) {
        if _, ok := q.(GetStatus); ok {
            return "healthy", true, nil
        }
        return nil, false, nil
    })

    result, handled, err := c.QUERY(GetStatus{})
    assert.NoError(t, err)
    assert.True(t, handled)
    assert.Equal(t, "healthy", result)
}
```

### Tasks

```go
func TestTask(t *testing.T) {
    c, _ := core.New()

    c.RegisterTask(func(c *core.Core, t core.Task) (any, bool, error) {
        if m, ok := t.(ProcessItem); ok {
            return "processed-" + m.ID, true, nil
        }
        return nil, false, nil
    })

    result, handled, err := c.PERFORM(ProcessItem{ID: "42"})
    assert.NoError(t, err)
    assert.True(t, handled)
    assert.Equal(t, "processed-42", result)
}
```

### Async Tasks

Use `assert.Eventually` to wait for background task completion:

```go
func TestAsyncTask(t *testing.T) {
    c, _ := core.New()

    var completed atomic.Bool
    var resultReceived any

    c.RegisterAction(func(c *core.Core, msg core.Message) error {
        if tc, ok := msg.(core.ActionTaskCompleted); ok {
            resultReceived = tc.Result
            completed.Store(true)
        }
        return nil
    })

    c.RegisterTask(func(c *core.Core, task core.Task) (any, bool, error) {
        return "async-result", true, nil
    })

    taskID := c.PerformAsync(MyTask{})
    assert.NotEmpty(t, taskID)

    assert.Eventually(t, func() bool {
        return completed.Load()
    }, 1*time.Second, 10*time.Millisecond)

    assert.Equal(t, "async-result", resultReceived)
}
```

## Testing with Context Cancellation

Verify that lifecycle methods respect context cancellation:

```go
func TestStartupCancellation(t *testing.T) {
    c, _ := core.New()
    ctx, cancel := context.WithCancel(context.Background())
    cancel() // cancel immediately

    s := &MockStartable{}
    _ = c.RegisterService("s1", s)

    err := c.ServiceStartup(ctx, nil)
    assert.Error(t, err)
    assert.ErrorIs(t, err, context.Canceled)
    assert.False(t, s.started)
}
```

## Global Instance in Tests

If your code under test uses `core.App()` or `core.GetInstance()`, save and restore the global instance:

```go
func TestWithGlobalInstance(t *testing.T) {
    original := core.GetInstance()
    defer core.SetInstance(original)

    c, _ := core.New(core.WithApp(&mockApp{}))
    core.SetInstance(c)

    // Test code that calls core.App()
    assert.NotNil(t, core.App())
}
```

Or use `ClearInstance()` to ensure a clean state:

```go
func TestAppPanicsWhenNotSet(t *testing.T) {
    original := core.GetInstance()
    core.ClearInstance()
    defer core.SetInstance(original)

    assert.Panics(t, func() {
        core.App()
    })
}
```

## Fuzz Testing

Core includes fuzz tests for critical paths. The pattern is to exercise constructors and registries with arbitrary input:

```go
func FuzzE(f *testing.F) {
    f.Add("svc.Method", "something broke", true)
    f.Add("", "", false)

    f.Fuzz(func(t *testing.T, op, msg string, withErr bool) {
        var underlying error
        if withErr {
            underlying = errors.New("wrapped")
        }
        e := core.E(op, msg, underlying)
        if e == nil {
            t.Fatal("E() returned nil")
        }
    })
}
```

Run fuzz tests with:

```bash
core go test --run Fuzz --fuzz FuzzE
```

Or directly with `go test`:

```bash
go test -fuzz FuzzE ./pkg/core/
```

## Benchmarks

Core includes benchmarks for the message bus. Run them with:

```bash
go test -bench . ./pkg/core/
```

Available benchmarks:

- `BenchmarkMessageBus_Action` -- ACTION dispatch throughput
- `BenchmarkMessageBus_Query` -- QUERY dispatch throughput
- `BenchmarkMessageBus_Perform` -- PERFORM dispatch throughput

## Running Tests

```bash
# All tests
core go test

# Single test
core go test --run TestCore_New_Good

# With race detector
go test -race ./pkg/core/

# Coverage
core go cov
core go cov --open  # opens HTML report in browser
```

## Related Pages

- [Services](services.md) -- what you are testing
- [Lifecycle](lifecycle.md) -- startup/shutdown behaviour
- [Messaging](messaging.md) -- ACTION/QUERY/PERFORM
- [Errors](errors.md) -- the `E()` helper used in tests
