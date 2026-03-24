# Logging Reference

Logging lives in the root `core` package in this repository. There is no separate `pkg/log` import path here.

## Create a Logger

```go
logger := core.NewLog(core.LogOptions{
	Level: core.LevelInfo,
})
```

## Levels

| Level | Meaning |
|-------|---------|
| `LevelQuiet` | no output |
| `LevelError` | errors and security events |
| `LevelWarn` | warnings, errors, security events |
| `LevelInfo` | informational, warnings, errors, security events |
| `LevelDebug` | everything |

## Log Methods

```go
logger.Debug("workspace discovered", "path", "/srv/workspaces")
logger.Info("service started", "service", "audit")
logger.Warn("retrying fetch", "attempt", 2)
logger.Error("fetch failed", "err", err)
logger.Security("sandbox escape detected", "path", attemptedPath)
```

## Default Logger

The package owns a default logger.

```go
core.SetLevel(core.LevelDebug)
core.SetRedactKeys("token", "password")

core.Info("service started", "service", "audit")
```

## Redaction

Values for keys listed in `RedactKeys` are replaced with `[REDACTED]`.

```go
logger.SetRedactKeys("token")
logger.Info("login", "user", "cladius", "token", "secret-value")
```

## Output and Rotation

```go
logger := core.NewLog(core.LogOptions{
	Level:  core.LevelInfo,
	Output: os.Stderr,
})
```

If you provide `Rotation` and set `RotationWriterFactory`, the logger writes to the rotating writer instead of the plain output stream.

## Error-Aware Logging

`LogErr` extracts structured error context before logging:

```go
le := core.NewLogErr(logger)
le.Log(err)
```

`ErrorLog` is the log-and-return wrapper exposed through `c.Log()`.

## Panic-Aware Logging

`LogPanic` is the lightweight panic logger:

```go
defer core.NewLogPanic(logger).Recover()
```

It logs the recovered panic but does not manage crash files. For crash reports, use `c.Error().Recover()`.
