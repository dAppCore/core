# Package Reference: `core`

Import path:

```go
import "dappco.re/go"
```

This repository exposes one root package. The main areas are:

## Constructors and Accessors

| Name | Purpose |
|------|---------|
| `New` | Create a `*Core` |
| `NewRuntime` | Create an empty runtime wrapper |
| `NewWithFactories` | Create a runtime wrapper from named service factories |
| `Options`, `App`, `Data`, `Drive`, `Fs`, `Config`, `Error`, `Log`, `Cli`, `IPC`, `I18n`, `Context` | Access the built-in subsystems |

## Core Primitives

| Name | Purpose |
|------|---------|
| `Option`, `Options` | Input configuration and metadata |
| `Result` | Shared output shape |
| `Service` | Lifecycle DTO |
| `Command` | Command tree node |
| `Message`, `Query`, `Task` | Message bus payload types |

## Service and Runtime APIs

| Name | Purpose |
|------|---------|
| `Service` | Register or read a named service |
| `Services` | List registered service names |
| `Startables`, `Stoppables` | Snapshot lifecycle-capable services |
| `LockEnable`, `LockApply` | Activate the service registry lock |
| `ServiceRuntime[T]` | Helper for package authors |

## Command and CLI APIs

| Name | Purpose |
|------|---------|
| `Command` | Register or read a command by path |
| `Commands` | List command paths |
| `Cli().Run` | Resolve arguments to a command and execute it |
| `Cli().PrintHelp` | Show executable commands |

## Messaging APIs

| Name | Purpose |
|------|---------|
| `ACTION`, `Action` | Broadcast a message |
| `QUERY`, `Query` | Return the first successful query result |
| `QUERYALL`, `QueryAll` | Collect all successful query results |
| `PERFORM`, `Perform` | Run the first task handler that accepts the task |
| `PerformAsync` | Run a task in the background |
| `Progress` | Broadcast async task progress |
| `RegisterAction`, `RegisterActions`, `RegisterQuery`, `RegisterTask` | Register bus handlers |

## Subsystems

| Name | Purpose |
|------|---------|
| `Config` | Runtime settings and feature flags |
| `Data` | Embedded filesystem mounts |
| `Drive` | Named transport handles |
| `Fs` | Local filesystem operations |
| `I18n` | Locale collection and translation delegation |
| `App`, `Find` | Application identity and executable lookup |

## Errors and Logging

| Name | Purpose |
|------|---------|
| `E`, `Wrap`, `WrapCode`, `NewCode` | Structured error creation |
| `Operation`, `ErrorCode`, `ErrorMessage`, `Root`, `StackTrace`, `FormatStackTrace` | Error inspection |
| `NewLog`, `Default`, `SetDefault`, `SetLevel`, `SetRedactKeys` | Logger creation and defaults |
| `LogErr`, `LogPanic`, `ErrorLog`, `ErrorPanic` | Error-aware logging and panic recovery |

Use the top-level docs in `docs/` for task-oriented guidance, then use this page as a compact reference.
