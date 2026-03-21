---
title: Errors
description: Structured errors, logging helpers, and panic recovery.
---

# Errors

CoreGO treats failures as structured operational data.

Repository convention: use `E()` instead of `fmt.Errorf` for framework and service errors.

## `Err`

The structured error type is:

```go
type Err struct {
	Operation string
	Message   string
	Cause     error
	Code      string
}
```

## Create Errors

### `E`

```go
err := core.E("workspace.Load", "failed to read workspace manifest", cause)
```

### `Wrap`

```go
err := core.Wrap(cause, "workspace.Load", "manifest parse failed")
```

### `WrapCode`

```go
err := core.WrapCode(cause, "WORKSPACE_INVALID", "workspace.Load", "manifest parse failed")
```

### `NewCode`

```go
err := core.NewCode("NOT_FOUND", "workspace not found")
```

## Inspect Errors

```go
op := core.Operation(err)
code := core.ErrorCode(err)
msg := core.ErrorMessage(err)
root := core.Root(err)
stack := core.StackTrace(err)
pretty := core.FormatStackTrace(err)
```

These helpers keep the operational chain visible without extra type assertions.

## Join and Standard Wrappers

```go
combined := core.ErrorJoin(err1, err2)
same := core.Is(combined, err1)
```

`core.As` and `core.NewError` mirror the standard library for convenience.

## Log-and-Return Helpers

`Core` exposes two convenience wrappers:

```go
r1 := c.LogError(err, "workspace.Load", "workspace load failed")
r2 := c.LogWarn(err, "workspace.Load", "workspace load degraded")
```

These log through the default logger and return `core.Result`.

You can also use the underlying `ErrorLog` directly:

```go
r := c.Log().Error(err, "workspace.Load", "workspace load failed")
```

`Must` logs and then panics when the error is non-nil:

```go
c.Must(err, "workspace.Load", "workspace load failed")
```

## Panic Recovery

`ErrorPanic` handles process-safe panic capture.

```go
defer c.Error().Recover()
```

Run background work with recovery:

```go
c.Error().SafeGo(func() {
	panic("captured")
})
```

If `ErrorPanic` has a configured crash file path, it appends JSON crash reports and `Reports(n)` reads them back.

That crash file path is currently internal state on `ErrorPanic`, not a public constructor option on `Core.New()`.

## Logging and Error Context

The logging subsystem automatically extracts `op` and logical stack information from structured errors when those values are present in the key-value list.

That makes errors created with `E`, `Wrap`, or `WrapCode` much easier to follow in logs.
