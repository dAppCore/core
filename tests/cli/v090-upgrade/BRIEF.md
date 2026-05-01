# Compliance against the core/go reference

`core/go` is the reference implementation for every shape decision in
this codebase. **The audit script at
`/Users/snider/Code/core/go/tests/cli/v090-upgrade/audit.sh` IS THE
PROVIDER OF WORK.** When the audit reports `verdict: COMPLIANT` with
every counter at `0`, the work is done. Until then, the work is not
done.

There is no other definition of "done". Not "tests pass". Not "the
diff looks reasonable". Not "I refactored a lot of files". The audit
script's exit code and counter table are the contract. Anything short
of every-counter-zero is "not done yet" — not negotiable, not a
judgement call, not subject to "this dimension is unfair for cgo".

## The gate

```
bash /tmp/v090/audit.sh .
```

Run it, fix what it flags, run again, repeat. When every counter is `0`
and the verdict reads `COMPLIANT`, you're ready to commit.

## The reference

Every shape decision (file naming, test naming, body structure,
assertion style, error construction, Result/Options/Action shape) is
set by `core/go`. Before writing any new test, read these files
verbatim and match the shape:

- `/Users/snider/Code/core/go/result_test.go`
- `/Users/snider/Code/core/go/options_test.go`
- `/Users/snider/Code/core/go/error_test.go`
- `/Users/snider/Code/core/go/string_test.go`
- `/Users/snider/Code/core/go/action_test.go`

Match: same package shape, same imports, same naming, same depth.

## Test file convention (REQUIRED — read carefully)

For every production source file `<file>.go`, tests live in the
matching `<file>_test.go` next to it. Extend the existing test file
if it exists; create it if it doesn't. **Do not** create sibling files
with version suffixes, prefix words, or any other discriminator
(`<file>_extra_test.go`, `<file>_v0_test.go`, `<file>_compliance_test.go`,
etc.) — those break the reader's ability to find a symbol's tests by
opening the matching `_test.go`. One source file, one test file.

Test functions follow the pattern:

```
func Test<File>_<Symbol>_<Variant>(t *T) { ... }
```

where `<File>` is the source file's stem with the first letter
uppercased, `<Symbol>` is the public function or method being tested,
and `<Variant>` is one of `Good` (happy path), `Bad` (failure case
with explicit error), or `Ugly` (edge case — empty input, boundary,
panic recovery).

Concrete examples from `core/go`:

| Source | Test file | Test function |
|---|---|---|
| `result.go` | `result_test.go` | `TestResult_New_Good`, `TestResult_Ok_Good`, `TestResult_Fail_Good` |
| `options.go` | `options_test.go` | `TestOptions_NewOptions_Good`, `TestOptions_Set_Good`, `TestOptions_Get_Bad` |
| `string.go` | `string_test.go` | `TestString_Concat_Good`, `TestString_Sprintf_Good` |

There is **no** `ax7_test.go`, no `v090_test.go`, no `compliance_test.go`
anywhere in `core/go`. Tests sit beside the symbol they exercise,
named after the source file.

## Each test exercises one symbol with its own assertions

```go
// REQUIRED — each test invokes the named symbol with arguments
// and asserts on the returned value.
func TestFoo_Foo_Good(t *T) {
    got := Foo("input")
    AssertEqual(t, "expected", got)
}

func TestFoo_Foo_Bad(t *T) {
    got := Foo("")
    AssertEqual(t, "", got)
}

func TestFoo_Foo_Ugly(t *T) {
    AssertNotPanics(t, func() { Foo(strings.Repeat("x", 1<<20)) })
}
```

## Forbidden patterns

The audit catches these mechanically. Each is a real shape codex has
emitted in the past — the audit dimensions exist because of them.

### 1. One test calls a shared dispatcher — FORBIDDEN

```go
// FORBIDDEN — body is a one-liner delegating to a shared helper.
func TestX_Foo_Good(t *T) { dispatch(t, "Foo", "Good") }
func TestX_Bar_Good(t *T) { dispatch(t, "Bar", "Good") }
```

If one assertion in `dispatch` breaks, every test in the family
fails. You can't isolate which symbol broke. Inline real assertions
into each test.

### 2. One file with every package's tests — FORBIDDEN

```
pkg/foo/
├── foo.go
├── bar.go
├── foo_test.go         (existing — extend this)
├── bar_test.go         (existing — extend this)
└── everything_test.go  FORBIDDEN — monolith of all triplets
```

```go
// FORBIDDEN — fake source-file slot in the prefix
func TestEverything_Foo_Good(t *T)   { ... }
func TestEverything_Bar_Good(t *T)   { ... }
```

If `foo.go` already has `foo_test.go`, extend it. Don't make a sibling
monolith.

### 3. Tautological assertions — FORBIDDEN

```go
// FORBIDDEN — string literal is always non-empty, the test cannot fail.
func TestX_Foo_Good(t *testing.T) {
    symbol := (*Client).Foo
    if symbol == nil { t.Fatal("expected symbol linked") }
    if "Foo_Good" == "" { t.Fatal("expected non-empty label") }
}
```

The test does not invoke the symbol — only references it as a method
value, then compares a literal to "". Every branch is unreachable. A
real test exercises the symbol with arguments and asserts on the
returned value.

### 4. Tests that don't reference their target — FORBIDDEN

```go
// FORBIDDEN — names a test for WithErrno but body never invokes WithErrno.
// Calls a test helper instead, leaving the public symbol uncovered.
func TestStringConversion_WithErrno_Good(t *T) {
    rc, err := testWithErrno(0)
    AssertNoError(t, err)
    AssertEqual(t, 0, rc)
}

// FORBIDDEN — reflect-tautology. The Kind() check passes for every
// function-typed value; the symbol is named but never invoked.
func TestSigner_NLSAGSigner_Version_Bad(t *T) {
    subject := any((*NLSAGSigner).Version)
    rv := reflect.ValueOf(subject)
    AssertEqual(t, reflect.Func, rv.Kind())
}

// REQUIRED — the body INVOKES the named symbol with arguments and
// asserts on its returned value. The bare token must appear in the body.
func TestStringConversion_WithErrno_Good(t *T) {
    s, err := WithErrno(0)
    AssertNoError(t, err)
    AssertEqual(t, "", s)
}
```

### 5. Banned stdlib imports (incl. test files) — FORBIDDEN

`fmt`, `errors`, `strings`, `path`, `path/filepath`, `os`, `os/exec`,
`io/ioutil`, `log`, `encoding/json`, `bytes` — all have core/go wrappers
and the audit catches them in ALL `.go` files including `_test.go`.

```go
// FORBIDDEN — direct stdlib in test file
import (
    "fmt"
    "os"
    "strings"
    "testing"
)

// REQUIRED — core/go wrappers
import (
    . "dappco.re/go"
)
// Use core.E for error construction, c.Strings()/c.Fs()/c.Process() at
// runtime, *T as the test signature, AssertX/RequireX for assertions.
```

The audit's `banned-imports` dimension counts every offending import
line. Every one needs to be replaced — no exceptions for "this is just
a test", "this is just convenience formatting", or "core doesn't have
this helper" (file an issue if a wrapper is genuinely missing).

Do NOT create local packages whose names mirror the banned stdlib
(`internal/stdcompat/fmt`, `pkg/util/strings`, `internal/io/os`, etc).
Shadowing the stdlib name to write `import .../stdcompat/fmt` instead
of `import "fmt"` satisfies the grep but violates the audit's spirit —
the new `stdlib-shadow-packages` dimension catches the directory layout
and the `package <stdlib-name>` declaration. Use core wrappers directly
via the `core` alias or dot-import.

### 6. Missing Example* / docs/ — FORBIDDEN

Compliance is not just tests. Every public symbol needs:

```go
// In <file>_example_test.go (file-aware, alongside <file>.go)
func ExampleFoo() {
    out := Foo("hello")
    fmt.Println(out)        // legitimate fmt — example output
    // Output: HELLO
}

// Methods: Example<Receiver>_<Method>
func ExampleClient_Send() {
    c := NewClient()
    _, _ = c.Send("ping")
    // Output: pong
}
```

Top-level structural docs are required at every consumer repo:

```
README.md                     overview, install, quick start
CLAUDE.md                     conventions for Claude Code agents
AGENTS.md                     code structure overview for new agents
docs/index.md                 canonical docs landing page
docs/architecture.md          high-level shape, primitives used
docs/development.md           how to build, test, contribute locally
```

The audit dimensions `example-gaps` and `docs-gaps` count missing
items; both must hit zero for COMPLIANT.

### 7. Result discards — FORBIDDEN

`_ = func()` in production code throws away a Result/error.

```go
// FORBIDDEN
_ = core.Setenv("X", "Y")
_ = service.Shutdown(ctx)

// REQUIRED — branch, log, or comment-justified
if r := core.Setenv("X", "Y"); !r.OK { return r }

defer func() {
    if err := service.Shutdown(ctx); err != nil {
        core.LogWarn("shutdown failed", "err", err)
    }
}()

// Genuinely intentional — KEEP the discard but justify it
_ = b.Reset()  // note: Reset cannot fail; only present to clear state
```

### 8. `func ... error` signatures — FORBIDDEN in production

core/go's `Result` carries `OK / Value / Error` AND auto-recovers panics.
The canonical Core signature is `func ... Result`; callers branch on
`r.OK`. Functions still declared `func ... error` are old-shape and the
audit's `err-shape-funcs` dimension counts them.

```go
// FORBIDDEN — old Go shape
func LoadConfig(path string) (*Config, error) {
    data, err := os.ReadFile(path)
    if err != nil {
        return nil, err
    }
    cfg := &Config{}
    if err := json.Unmarshal(data, cfg); err != nil {
        return nil, err
    }
    return cfg, nil
}

// REQUIRED — Result shape, panic-safe, no err propagation
func LoadConfig(path string) Result {
    r := core.ReadFile(path)
    if !r.OK { return r }
    cfg := &Config{}
    return core.JSONUnmarshal(r.Bytes(), cfg)
}

// Caller side
r := LoadConfig("./app.yaml")
if !r.OK { return r }
cfg := r.Value.(*Config)
```

Construct Results with `Ok(value)`, `Fail(err)`, or `ResultOf(value, err)`
when bridging from an existing stdlib-shaped call. But: **there are no
legitimate stdlib imports in a consumer repo** — core/go provides the
wrappers for everything (file IO, JSON, exec, formatting, paths, regexp,
HTTP, crypto, networking). If you find yourself writing `Fail(err)` for
a stdlib call, replace the stdlib call with the core wrapper that
already exists. The bridge is a one-way conversion at the language
boundary, not a place to host new error-shape code.

The audit's `err-shape-funcs` target for a consumer repo is 0 — every
`func ... error` should be `func ... Result`. Callers stop needing
`if err != nil { return err }` chains entirely; they become `if !r.OK
{ return r }`. The chain disappearing is the diagnostic: any `err :=`
left in production is an unconverted call site or an old-shape function
still pulling errors back into the codebase.

## Migration order

Run the audit to see what's outstanding. Fix in the order the audit
prints — simplest first so each pass narrows the surface:

1. Module path imports → mechanical sed rewrite, `go mod tidy`
2. Breaking-API call sites → rewrite per the `Result` template
3. `core.Result{...}` literals → `Ok` / `Fail` / `ResultOf`
4. testify in `*_test.go` → `*T` + `AssertX` + dot-import `dappco.re/go`
5. `_ = func()` discards → handle each per Forbidden #4 above
6. Public-symbol triplet gaps → author missing tests in the matching
   `<source>_test.go` per the convention above
7. Stub-like test bodies → inline real assertions
8. Tautological literal-equality → exercise the symbol with arguments
9. Sibling monolith files → merge into per-source `<source>_test.go`,
   delete the sibling
10. Versioned/discriminator-suffix test files → same as 9

Test rules:

- `package <pkg>_test` (or matching the existing convention in the file)
- `import . "dappco.re/go"` — single dot-import; nothing else
- `*T` not `*testing.T`
- `AssertX` / `RequireX` from core
- Fast — `t.TempDir()` for filesystem, no network, no external binaries

Symbols genuinely without testable behaviour (sentinel constants,
type aliases): skip them and list at the end of your commit body. Do
not write stub tests with `t.Skip()` or `_ = X`.

## Verification before commit

The single source of truth:

```
bash /tmp/v090/audit.sh .
```

When it prints `verdict: COMPLIANT`, also confirm:

```
GOWORK=off go test -count=1 ./...
GOWORK=off go vet ./...
gofmt -l .
```

All green → commit. Never commit a partial state.

There is no skipped-symbol category. Every public symbol in `<file>.go`
has a concrete `Test<File>_<Symbol>_{Good,Bad,Ugly}` triplet in
`<file>_test.go` plus an `Example<Symbol>` in `<file>_example_test.go`.
"No testable behaviour" is not a defence — every public symbol exposes
a contract (return values, panics, side effects, type satisfaction) and
the test exercises that contract.

Conventional commit subject:

```
refactor(core): align with core/go reference shape
```

Body should report what changed across each audit dimension.

Do NOT push.
