---
title: Testing
description: Test naming, AX assertions, fuzzing, and CLI tests in CoreGO.
---

# Testing

CoreGO ships its own AX-shaped test framework in `test.go`. Test files use a single dot-import — never `import "testing"` — and consume `*T`, `AssertX`, `RequireX`, `CLITest`, `AssertCLI` directly.

```go
package mypkg_test

import (
	. "dappco.re/go"
)

func TestRepository_Sync_Good(t *T) {
	r := svc.SyncRepository("agent", "/srv/repos/agent")
	AssertTrue(t, r.OK)
	AssertEqual(t, "synced", r.Value.(string))
}
```

`*T` is `core.T`, a type-identical alias for `*testing.T`, so Go's runner discovers `Test*(t *T)` functions exactly as it does `Test*(t *testing.T)`.

## Test naming

`Test{Filename}_{Function}_{Good|Bad|Ugly}`. All three states are mandatory for full coverage:

- `_Good` — happy path
- `_Bad` — expected failure (rejected input, error returned, refused operation)
- `_Ugly` — edge case (zero values, large inputs, race-prone setup, boundary conditions)

```go
func TestNew_Good(t *T) {}
func TestService_Register_Duplicate_Bad(t *T) {}
func TestCore_Must_Ugly(t *T) {}
```

## Assert and Require

`AssertX` records a failure and continues; `RequireX` records and stops the test. Both wrap `testing.TB`, so helpers can accept either Test or Benchmark contexts.

| Class | Helpers |
|---|---|
| Comparison | `AssertEqual`, `AssertNotEqual`, `AssertTrue`, `AssertFalse` |
| Nil | `AssertNil`, `AssertNotNil` |
| Errors | `AssertNoError`, `AssertError`, `AssertErrorIs` |
| Containers | `AssertContains`, `AssertNotContains`, `AssertLen`, `AssertEmpty`, `AssertNotEmpty` |
| Ordered | `AssertGreater`, `AssertGreaterOrEqual`, `AssertLess`, `AssertLessOrEqual`, `AssertInDelta` |
| Identity | `AssertSame`, `AssertElementsMatch` |
| Panics | `AssertPanics`, `AssertNotPanics`, `AssertPanicsWithError` |
| Fail-fast | `RequireNoError`, `RequireTrue`, `RequireNotEmpty` |

`AnError` is the sentinel for tests that need a non-nil error without caring about content.

Failure messages are one-line, file:line + assertion + want/got — AI-readable. Pass = silent.

## Start with a small Core

```go
c := New(WithName("test-core", nil))
```

Then register only the pieces your test needs.

## Test a service

```go
started := false

c := New(WithService(func(c *Core) Result {
	c.Action("audit.startup", func(ctx Context, opts Options) Result {
		started = true
		return Result{OK: true}
	})
	return Result{OK: true}
}))

r := c.ServiceStartup(Background(), nil)
AssertTrue(t, r.OK)
AssertTrue(t, started)
```

## Test a query or action

```go
c := New()
c.Action("compute", func(_ Context, _ Options) Result {
	return Result{Value: 42, OK: true}
})

r := c.Action("compute").Run(Background(), NewOptions())
AssertEqual(t, 42, r.Value)
```

## Test async work

```go
completed := make(chan ActionTaskCompleted, 1)

c.RegisterAction(func(_ *Core, msg Message) Result {
	if event, ok := msg.(ActionTaskCompleted); ok {
		completed <- event
	}
	return Result{OK: true}
})
```

## Real temporary paths

When testing `Fs`, `Data.Extract`, or any I/O helper, use `t.TempDir()` and create realistic paths instead of mocking the filesystem.

## Godoc examples

Example functions live in `*_example_test.go` files. The function body IS the runnable example; pkg.go.dev renders it as a code block. The DOCBLOCK above each Example is descriptive prose explaining the scenario being demonstrated — convention inverted from production `.go` files (where the docblock IS the usage example).

```go
// ExampleHKDF demonstrates deriving a session key from a master secret
// and salt using HKDF-SHA256. The 32-byte derived key is suitable for
// symmetric AEAD construction.
func ExampleHKDF() {
	r := HKDF("sha256", []byte("secret"), []byte("salt"), []byte("session"), 32)
	if r.OK {
		Println(len(r.Value.([]byte)))
	}
	// Output: 32
}
```

Use `core.Println` for output, NOT `fmt.Println`. The `// Output:` comment block must match runtime stdout exactly.

## Fuzz harnesses

Fuzz harnesses live in `*_fuzz_test.go` files. Use `*F` (alias for `*testing.F`):

```go
func FuzzURLParse(f *F) {
	f.Add("https://example.com/path?q=1")
	f.Add("")
	f.Fuzz(func(t *T, raw string) {
		r := URLParse(raw)
		if r.OK {
			// exercise round-trip / invariants
		}
	})
}
```

Run with `go test -fuzz=FuzzXxx -fuzztime=30s`. The seed corpus alone runs via `go test -run "Fuzz"`.

## CLI tests (AX-10)

Per AX-10 (CLI tests as artifact validation), each command path has a `tests/cli/{path}/Taskfile.yaml` that drives binary-level scenarios. Inside Go tests, the `CLITest` shape + `AssertCLI`/`AssertCLIs` helpers dispatch through `c.Process()`, so the caller registers a process service (typically `dappco.re/go-process`) on the Core before invoking.

```go
c := New(WithService(process.Register))

AssertCLI(t, c, CLITest{
	Cmd:      "go",
	Args:     []string{"version"},
	WantOK:   true,
	Contains: "go1.",
})

AssertCLIs(t, c, []CLITest{
	{Name: "version", Cmd: "go", Args: []string{"version"}, WantOK: true, Contains: "go1."},
	{Name: "vet",     Cmd: "go", Args: []string{"vet", "./..."}, WantOK: true},
})
```

## Test data

Fixtures live under `tests/data/`. Files with names beginning with `_` (e.g. `tests/data/_scantest/sample.go`) are excluded from compilation by Go's tool, so source-file fixtures don't pollute the module's compiled package set.

The canonical embed pattern is `//go:embed all:tests/data` — the `all:` prefix includes `_`-prefixed entries that would otherwise be excluded.

## SPOR drift guard

`tests/cli/imports/check.sh` enforces the SPOR ownership rule mechanically — every stdlib package has exactly one production owner file. CI fails on any package imported by more than one non-owner. See `AGENTS.md` for the full ownership table.

## Repository commands

```bash
go test ./...
go test ./... -run TestPerformAsync_Good
go test -fuzz=FuzzURLParse -fuzztime=30s
bash tests/cli/imports/check.sh
```
