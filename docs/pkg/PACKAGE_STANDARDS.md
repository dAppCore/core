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

CoreGO also re-exports the stdlib types your package would otherwise reach for. Prefer the core surface over importing stdlib directly:

- I/O primitives: `core.Reader`, `core.Writer`, `core.Closer`, `core.EOF`, `core.Copy`, `core.WriteString`
- OS surface: `core.FileMode`, `core.ModePerm`, `core.Stdin`, `core.Stdout`, `core.Stderr`
- Networking: `core.Conn`, `core.Listener`, `core.IP`, `core.ParseIP`, `core.NetDial`, `core.NetListen`, `core.NetPipe`
- HTTP: `core.Request`, `core.Response`, `core.Handler`, `core.HTTPServer`, `core.HTTPClient`, `core.HTTPGet`, `core.HTTPPost`, `core.NewHTTPTestServer`
- Templating: `core.Template`, `core.NewTemplate`, `core.ParseTemplate`, `core.ExecuteTemplate`
- SQL: `core.DB`, `core.Tx`, `core.Rows`, `core.ErrNoRows`, `core.SQLOpen`
- Regex: `core.Regexp`, `core.Regex(pattern)`
- Time: `core.Now`, `core.Sleep`, `core.Since`, `core.TimeFormat`, `core.TimeParse`, `core.TimeRFC3339`
- Path: `core.Path`, `core.PathBase`, `core.PathDir`, `core.PathExt`, `core.PathRel`, `core.PathAbs`, `core.PathChangeExt`
- Filesystem: `c.Fs().WalkSeq`, `c.Fs().WalkSeqSkip` for directory traversal
- Slices/maps: `core.SliceContains`, `core.SliceFilter`, `core.SliceMap`, `core.SliceReduce`, `core.MapKeys`, `core.MapFilter`, `core.MapMerge`

The destination state for `go.mod` in any package built on CoreGO is two lines — `module` and `go` — with no `require` block at all. Reach for stdlib only when CoreGO genuinely lacks the primitive; if your package needs something universal that's missing, the answer is to add it to CoreGO, not to import stdlib in your consumer.

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

- `_Good`   — happy path
- `_Bad`    — expected failure (rejected input, returned error, refused operation)
- `_Ugly`   — edge case (zero values, large inputs, boundary conditions, race-prone setup)

Test files import only the core package and use the `*T` alias for the test handle:

```go
package mypkg_test

import (
	. "dappco.re/go/core"
)

func TestRepositorySync_Good(t *T) {
	r := svc.SyncRepository("core-go", "/srv/repos/core-go")
	AssertTrue(t, r.OK)
	AssertEqual(t, "synced", r.Value.(string))
}

func TestRepositorySync_Bad(t *T) {
	r := svc.SyncRepository("", "")
	AssertError(t, r.Value.(error), "name required")
}

func TestRepositorySync_Ugly(t *T) {
	RequireNoError(t, setupTempRepo())
	// ... continues with Assert* on the happy path
}
```

`*T` is `core.T` — an alias for `testing.T` so tests no longer need a separate `import "testing"` line. Go's test runner discovers `Test*` functions by signature; the alias is type-identical so discovery still works.

Use `Assert*` for non-fatal assertions and `Require*` for "stop the test if this fails" preconditions. Both wrap `testing.TB` and emit one-line, AI-readable failure output:

```
fs_test.go:144: AssertEqual want="hello" got="world"
```

The full assertion family is in `test.go`:

- Comparison: `AssertEqual`, `AssertNotEqual`, `AssertTrue`, `AssertFalse`
- Nil: `AssertNil`, `AssertNotNil`
- Errors: `AssertNoError`, `AssertError`, `AssertErrorIs`
- Containers: `AssertContains`, `AssertNotContains`, `AssertLen`, `AssertEmpty`, `AssertNotEmpty`
- Ordered: `AssertGreater`, `AssertGreaterOrEqual`, `AssertLess`, `AssertLessOrEqual`, `AssertInDelta`
- Identity: `AssertSame`, `AssertElementsMatch`
- Panics: `AssertPanics`, `AssertNotPanics`, `AssertPanicsWithError`
- Fail-fast: `RequireNoError`, `RequireTrue`, `RequireNotEmpty`

`core.AnError` is a sentinel error for tests that need a non-nil error without caring about its content.

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
