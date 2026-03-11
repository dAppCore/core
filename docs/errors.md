---
title: Errors
description: The E() helper function and Error struct for contextual error handling.
---

# Errors

Core provides a standardised error type and constructor for wrapping errors with operational context. This makes it easier to trace where an error originated and provide meaningful feedback.

## The Error Struct

```go
type Error struct {
    Op  string // the operation, e.g. "config.Load"
    Msg string // human-readable explanation
    Err error  // the underlying error (may be nil)
}
```

- **Op** identifies the operation that failed. Use the format `package.Function` or `service.Method`.
- **Msg** is a human-readable message explaining what went wrong.
- **Err** is the underlying error being wrapped. May be `nil` for root errors.

## The E() Helper

`E()` is the primary way to create contextual errors:

```go
func E(op, msg string, err error) error
```

### With an Underlying Error

```go
data, err := os.ReadFile(path)
if err != nil {
    return core.E("config.Load", "failed to read config file", err)
}
```

This produces: `config.Load: failed to read config file: open /path/to/file: no such file or directory`

### Without an Underlying Error (Root Error)

```go
if name == "" {
    return core.E("user.Create", "name cannot be empty", nil)
}
```

This produces: `user.Create: name cannot be empty`

When `err` is `nil`, the `Err` field is not set and the output omits the trailing error.

## Error Output Format

The `Error()` method produces a string in one of two formats:

```
// With underlying error:
op: msg: underlying error text

// Without underlying error:
op: msg
```

## Unwrapping

`Error` implements the `Unwrap() error` method, making it compatible with Go's `errors.Is` and `errors.As`:

```go
originalErr := errors.New("connection refused")
wrapped := core.E("db.Connect", "failed to connect", originalErr)

// errors.Is traverses the chain
errors.Is(wrapped, originalErr) // true

// errors.As extracts the Error
var coreErr *core.Error
if errors.As(wrapped, &coreErr) {
    fmt.Println(coreErr.Op)  // "db.Connect"
    fmt.Println(coreErr.Msg) // "failed to connect"
}
```

## Building Error Chains

Because `E()` wraps errors, you can build a logical call stack by wrapping at each layer:

```go
// Low-level
func readConfig(path string) ([]byte, error) {
    data, err := os.ReadFile(path)
    if err != nil {
        return nil, core.E("config.readConfig", "failed to read file", err)
    }
    return data, nil
}

// Mid-level
func loadConfig() (*Config, error) {
    data, err := readConfig("/etc/app/config.yaml")
    if err != nil {
        return nil, core.E("config.Load", "failed to load configuration", err)
    }
    // parse data...
    return cfg, nil
}

// Top-level
func (s *Service) OnStartup(ctx context.Context) error {
    cfg, err := loadConfig()
    if err != nil {
        return core.E("service.OnStartup", "startup failed", err)
    }
    s.config = cfg
    return nil
}
```

The resulting error message reads like a stack trace:

```
service.OnStartup: startup failed: config.Load: failed to load configuration: config.readConfig: failed to read file: open /etc/app/config.yaml: no such file or directory
```

## Conventions

1. **Op format**: Use `package.Function` or `service.Method`. Keep it short and specific.
2. **Msg format**: Use lowercase, describe what failed (not what succeeded). Write messages that make sense to a developer reading logs.
3. **Wrap at boundaries**: Wrap with `E()` when crossing package or layer boundaries, not at every function call.
4. **Always return `error`**: `E()` returns the `error` interface, not `*Error`. Callers should not need to know the concrete type.
5. **Nil underlying error**: Pass `nil` for `err` when creating root errors (errors that do not wrap another error).

## Related Pages

- [Services](services.md) -- services that return errors
- [Lifecycle](lifecycle.md) -- lifecycle error aggregation
- [Testing](testing.md) -- testing error conditions (`_Bad` suffix)
