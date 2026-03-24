# AX Package Standards

This page describes how to build packages on top of CoreGO in the style described by RFC-025.

## 1. Prefer Predictable Names

Use names that tell an agent what the thing is without translation.

Good:

- `RepositoryService`
- `RepositoryServiceOptions`
- `WorkspaceCountQuery`
- `SyncRepositoryTask`

Avoid shortening names unless the abbreviation is already universal.

## 2. Put Real Usage in Comments

Write comments that show a real call with realistic values.

Good:

```go
// Sync a repository into the local workspace cache.
// svc.SyncRepository("core-go", "/srv/repos/core-go")
```

Avoid comments that only repeat the signature.

## 3. Keep Paths Semantic

If a command or template lives at a path, let the path explain the intent.

Good:

```text
deploy/to/homelab
workspace/create
template/workspace/go
```

That keeps the CLI, tests, docs, and message vocabulary aligned.

## 4. Reuse CoreGO Primitives

At Core boundaries, prefer the shared shapes:

- `core.Options` for lightweight input
- `core.Result` for output
- `core.Service` for lifecycle registration
- `core.Message`, `core.Query`, `core.Task` for bus protocols

Inside your package, typed structs are still good. Use `ServiceRuntime[T]` when you want typed package options plus a `Core` reference.

```go
type repositoryServiceOptions struct {
	BaseDirectory string
}

type repositoryService struct {
	*core.ServiceRuntime[repositoryServiceOptions]
}
```

## 5. Prefer Explicit Registration

Register services and commands with names and paths that stay readable in grep results.

```go
c.Service("repository", core.Service{...})
c.Command("repository/sync", core.Command{...})
```

## 6. Use the Bus for Decoupling

When one package needs another package’s behavior, prefer queries and tasks over tight package coupling.

```go
type repositoryCountQuery struct{}
type syncRepositoryTask struct {
	Name string
}
```

That keeps the protocol visible in code and easy for agents to follow.

## 7. Use Structured Errors

Use `core.E`, `core.Wrap`, and `core.WrapCode`.

```go
return core.Result{
	Value: core.E("repository.Sync", "git fetch failed", err),
	OK:    false,
}
```

Do not introduce free-form `fmt.Errorf` chains in framework code.

## 8. Keep Testing Names Predictable

Follow the repository pattern:

- `_Good`
- `_Bad`
- `_Ugly`

Example:

```go
func TestRepositorySync_Good(t *testing.T) {}
func TestRepositorySync_Bad(t *testing.T) {}
func TestRepositorySync_Ugly(t *testing.T) {}
```

## 9. Prefer Stable Shapes Over Clever APIs

For package APIs, avoid patterns that force an agent to infer too much hidden control flow.

Prefer:

- clear structs
- explicit names
- path-based commands
- visible message types

Avoid:

- implicit global state unless it is truly a default service
- panic-hiding constructors
- dense option chains when a small explicit struct would do

## 10. Document the Current Reality

If the implementation is in transition, document what the code does now, not the API shape you plan to have later.

That keeps agents correct on first pass, which is the real AX metric.
