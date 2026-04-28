# AGENTS.md — core/go

You are working in `dappco.re/go` — the foundation Go module of the dAppCore framework. This file is the orientation card. Read it once before making any change.

## What this module is

`core/go` is the zero-dependency foundation that every other dAppCore Go package builds on. The destination state for `go.mod` is **two lines** — `module` and `go` — with no `require` block. This is reached today; new code must not regress it.

```go.mod
module dappco.re/go

go 1.26.0
```

The renamed canonical github URL is `https://github.com/dappcore/go` (lowercase). The module path was previously `dappco.re/go/core`; consumer code may still reference the old path during the migration window.

## Reference-implementation rule

**core/go is the perfect example.** AI agents working in any consumer repo replicate patterns from here without a judgement layer. Every shape decision in this module propagates to ~30 downstream repos. Before merging anything, mentally simulate "every agent across the ecosystem copies this exact shape" — if the simulation produces bad code, the shape is wrong.

Concretely:
- If the test files here import `testing`, every consumer's test files will too. → They don't, and yours shouldn't.
- If a public function is documented with prose-only descriptions, agents will write prose-only descriptions everywhere. → They aren't, and yours shouldn't.
- If `fmt.Sprintf` slips into production code, it slips into the whole ecosystem. → It doesn't here, and you must use `core.Sprintf`.

## The 10 AX (Agent Experience) principles

Defined in `docs/RFC.md` and `docs/pkg/PACKAGE_STANDARDS.md`. Summary:

1. **Predictable names over short names** — `Config`, not `Cfg`. `Service`, not `Srv`.
2. **Comments as usage examples** — godoc shows real call sites, not prose descriptions. Tab-indented (`//\t`) code blocks with realistic values.
3. **Path is documentation** — `path/file.go` should describe the content from the path alone.
4. **Templates over freeform** — when a code shape recurs, provide a template.
5. **Declarative over imperative** — orchestration is YAML/JSON; implementation is Go.
6. **Universal types** — every public surface accepts `core.Options`, returns `core.Result`. Service registration uses `core.Service`. Config flows through `core.Config`. Embedded data through `core.Data`.
7. **Directory as semantics** — top-level dirs are categories, not bins.
8. **Lib never imports consumer** — dependencies flow one direction. `core/go` is the lib; never imports anything downstream.
9. **Issues are N+(rounds) deep** — iteration is the discovery process, not failure. Each pass sees what the previous could not.
10. **CLI tests as artifact validation** — `tests/cli/{path}/Taskfile.yaml` per command; the binary tested against fixtures is the source of truth.

## Banned imports — use core helpers

Production `.go` files in `core/go` and consumer packages should not import these directly. Reach for the core wrapper instead:

| Stdlib | Use |
|---|---|
| `fmt` | `core.Sprintf`, `core.Sprint`, `core.Println`, `core.Print` |
| `strings` | `core.Concat`, `core.Join`, `core.Split`, `core.Lower`, `core.Upper`, `core.HasPrefix`, `core.HasSuffix`, `core.Contains`, `core.NewBuilder`, `core.NewReader` |
| `errors` | `core.E(scope, msg, cause)`, `core.Is`, `core.As`, `core.ErrorJoin`, `core.NewError`, `core.NewCode`, `core.Wrap`, `core.WrapCode` |
| `os` (file/env/exec/exit) | `c.Fs()`, `c.Env()` / `core.Env`, `c.Process()`, `c.Exit()` |
| `os/exec` | `c.Process()` |
| `path/filepath` | `core.Path`, `core.PathBase`, `core.PathDir`, `core.PathExt`, `core.PathIsAbs`, `core.CleanPath`, `core.PathRel`, `core.PathAbs`, `core.PathChangeExt`, `c.Fs().WalkSeq` |
| `strconv` | `core.Atoi`, `core.Itoa`, `core.FormatInt` |
| `unicode` | `core.IsDigit`, `core.IsLetter`, `core.IsLower`, `core.IsSpace`, `core.IsUpper` |
| `bytes` | `core.NewBuffer`, `core.NewBufferString` |
| `bufio` | `core.NewLineScanner` |
| `encoding/json` | `core.JSONMarshal`, `core.JSONUnmarshal`, `core.JSONMarshalString`, `core.JSONUnmarshalString` |
| `encoding/hex` | `core.HexEncode`, `core.HexDecode` |
| `encoding/base64` | `core.Base64Encode`, `core.Base64Decode`, `core.Base64URLEncode`, `core.Base64URLDecode` |
| `html` | `core.HTMLEscape`, `core.HTMLUnescape` |
| `crypto/sha256`, `crypto/sha3` | `core.SHA256`, `core.SHA256Hex`, `core.Keccak256`, `core.SHA3_256` |
| `crypto/hmac`, `crypto/hkdf`, `crypto/rand` | `core.HMAC`, `core.HKDF`, `core.RandomBytes`, `core.RandomString`, `core.RandomInt` |
| `slices`, `sort`, `cmp`, `maps` | `core.SliceContains`, `core.SliceIndex`, `core.SliceSort`, `core.SliceUniq`, `core.SliceReverse`, `core.SliceFilter`, `core.SliceMap`, `core.SliceReduce`, `core.MapKeys`, `core.MapValues`, `core.MapClone`, `core.MapFilter`, `core.MapMerge`, `core.Compare`, `core.Min`, `core.Max`, `core.Abs` |
| `time` | `core.Now`, `core.UnixNow`, `core.Sleep`, `core.Since`, `core.Until`, `core.ParseDuration`, `core.TimeFormat`, `core.TimeParse`, `core.TimeRFC3339`, `core.TimeDateTime`, etc. |
| `embed` | `core.Mount`, `core.MountEmbed`, `core.DirFS`, `core.Extract`, `core.AddAsset`, `core.GetAsset` |
| `regexp` | `core.Regex(pattern)` returning `*core.Regexp` |
| `text/template` | `core.NewTemplate`, `core.ParseTemplate`, `core.ExecuteTemplate` |
| `text/tabwriter` | `core.NewTable` |
| `io` (interfaces + Copy/WriteString) | `core.Reader`, `core.Writer`, `core.Closer`, `core.ReadCloser`, `core.WriteCloser`, `core.EOF`, `core.Copy`, `core.CopyN`, `core.WriteString` |
| `io/fs` (selected) | `core.Fs` for sandboxed ops, `c.Fs().WalkSeq`, `c.Fs().WalkSeqSkip` |
| `os` constants | `core.FileMode`, `core.ModePerm`, `core.ModeDir`, `core.Stdin()`, `core.Stdout()`, `core.Stderr()` |
| `net` | `core.Conn`, `core.Listener`, `core.IP`, `core.IPNet`, `core.ParseIP`, `core.ParseCIDR`, `core.NetDial`, `core.NetListen`, `core.NetPipe` |
| `net/http` | `core.Request`, `core.Response`, `core.Handler`, `core.HTTPServer`, `core.HTTPClient`, `core.HTTPGet`, `core.HTTPPost`, `core.NewHTTPRequest` |
| `net/url` | `core.URLParse`, `core.URLEncode`, `core.URLDecode`, `core.URLPathEscape`, `core.URLNormalize` |
| `mime/multipart` | `core.MultipartReader`, `core.MultipartWriter`, `core.NewMultipartReader`, `core.NewMultipartWriter` |
| `net/http/httptest` | `core.NewHTTPTestServer`, `core.NewHTTPTestRecorder`, `core.NewHTTPTestRequest` |
| `database/sql` | `core.DB`, `core.Tx`, `core.Rows`, `core.SQLOpen`, `core.ErrNoRows` |

### SPOR — Single Point Of Responsibility

**Every stdlib package has exactly one production owner file.** No exemptions. Exemptions in the reference implementation become permission-to-replicate exemptions downstream; downstream agents copy the pattern without judgement. The wrapper IS the rule.

Ownership table:

| Stdlib | Sole owner | What you call instead |
|---|---|---|
| `bufio` | scanner.go | `core.NewLineScanner`, `core.NewLineScannerWithSize`, `core.NewBufReader`, `core.BufReader` |
| `bytes` | io.go | `core.NewBuffer`, `core.NewBufferString` |
| `cmp` | math.go | `core.Compare`, `core.Ordered` (constraint) |
| `compress/gzip` | embed.go | (internal embed compression) |
| `context` | context.go | `core.Context`, `core.Background`, `core.WithTimeout`, `core.WithCancel`, `core.WithDeadline`, `core.WithValue`, `core.TODO` |
| `crypto/hkdf`, `crypto/hmac` | hash.go | `core.HKDF`, `core.HMAC` (return Result) |
| `crypto/rand` | random.go | `core.RandomBytes`, `core.RandomString`, `core.RandomInt` (return Result) |
| `crypto/sha256`, `crypto/sha512` | hash.go | `core.SHA256`, `core.SHA256Hex`, `core.SHA256String` |
| `crypto/sha3` | sha3.go | `core.Keccak256`, `core.SHA3_256` |
| `database/sql` | sql.go | `core.DB`, `core.Tx`, `core.SQLOpen`, `core.ErrNoRows` |
| `embed` | embed.go | `core.Mount`, `core.AddAsset`, `core.GetAsset`, `core.EmbedFS` (alias) |
| `encoding/base64`, `encoding/binary`, `encoding/hex` | encode.go | `core.Base64Encode/Decode`, `core.HexEncode/Decode` |
| `encoding/json` | json.go | `core.JSONMarshal`, `core.JSONUnmarshal`, `core.JSONMarshalString`, `core.JSONUnmarshalString` |
| `errors` | error.go | `core.E`, `core.Wrap`, `core.WrapCode`, `core.NewError`, `core.NewCode`, `core.Is`, `core.As`, `core.ErrorJoin`, `core.Root` |
| `fmt` | format.go | `core.Sprintf`, `core.Sprint`, `core.Sprintln`, `core.Println`, `core.Print`, `core.Errorf` |
| `go/ast`, `go/parser`, `go/token` | embed.go | (internal embed parsing) |
| `hash` | hash.go | (internal hash factory) |
| `html` | string.go | `core.HTMLEscape`, `core.HTMLUnescape` |
| `html/template`, `text/template` | template.go | `core.NewTemplate`, `core.ParseTemplate`, `core.ExecuteTemplate`, `core.FuncMap` |
| `io` | io.go | `core.Reader`, `core.Writer`, `core.Closer`, `core.ReadCloser`, `core.WriteCloser`, `core.Copy`, `core.CopyN`, `core.WriteString`, `core.EOF`, `core.NewBuffer`, `core.NewBufferString` |
| `io/fs` | fs.go | `core.FS`, `core.FsFile`, `core.FsFileInfo`, `core.FsDirEntry`, `core.WalkDir`, `core.WalkDirFunc`, `c.Fs()` |
| `iter` | iter.go | `core.Seq[T]`, `core.Seq2[K,V]`, `core.Pull`, `core.Pull2` |
| `math`, `math/big`, `cmp` | math.go | `core.Abs`, `core.Min`, `core.Max`, `core.Pow`, `core.Floor`, `core.Ceil`, `core.Round`, `core.IsNaN`, `core.Compare`, `core.Ordered` |
| `math/bits` | sha3.go | (internal hash math) |
| `math/rand/v2` | random.go | `core.RandIntn`, `core.RandPick` |
| `mime/multipart` | api.go | `core.MultipartReader`, `core.NewMultipartWriter` |
| `net` | net.go | `core.Conn`, `core.Listener`, `core.IP`, `core.NetDial`, `core.NetListen`, `core.NetPipe` |
| `net/http`, `net/http/httptest`, `net/url` | api.go | `core.Request`, `core.Response`, `core.Handler`, `core.HTTPClient`, `core.HTTPGet`, `core.HTTPPost`, `core.NewHTTPRequest`, `core.URLParse`, `core.URLEncode`, `core.NewHTTPTestServer` (Result) |
| `os` | os.go | `core.Open`, `core.Stat`, `core.ReadFile`, `core.WriteFile`, `core.MkdirAll`, `core.Stdin`, `core.Stdout`, `core.Stderr`, `core.Exit`, `core.Args`, `core.Hostname`, `core.Getpid`, `core.Setenv` (Result), `core.Unsetenv` (Result), `core.Getenv`, `core.LookupEnv`, `core.Environ`, `core.UserHomeDir`, `core.FileMode`, `core.OSFile` |
| `os/exec` | process.go | `c.Process()`, `core.Cmd`, `core.ExecCmd` |
| `os/user` | user.go | `core.User`, `core.Group`, `core.UserCurrent`, `core.UserLookup`, `core.UserLookupID`, `core.UserGroupLookup` |
| `path/filepath` | path.go | `core.Path`, `core.PathBase`, `core.PathDir`, `core.PathExt`, `core.PathIsAbs`, `core.CleanPath`, `core.PathRel`, `core.PathAbs`, `core.PathChangeExt`, `core.PathGlob`, `core.PathWalk`, `core.PathWalkDir`, `core.PathMatch` |
| `reflect` | reflect.go | `core.TypeOf`, `core.ValueOf`, `core.DeepEqual`, `core.Zero`, `core.Type`, `core.Value`, `core.Kind`, `core.KindBool`, `core.KindInt`, `core.KindString`, etc. |
| `regexp` | regexp.go | `core.Regex` (Result), `core.Regexp` |
| `runtime`, `runtime/debug` | info.go | `core.OS`, `core.Arch`, `core.GoVersion`, `core.NumCPU`, `core.StackBuf`, `core.Env`, `core.EnvKeys` |
| `slices`, `sort` | slice.go | `core.SliceContains`, `core.SliceIndex`, `core.SliceSort`, `core.SliceUniq`, `core.SliceReverse`, `core.SliceFilter`, `core.SliceMap`, `core.SliceReduce`, `core.SliceFlatMap`, `core.SliceTake`, `core.SliceDrop`, `core.SliceAny`, `core.SliceAll`, `core.SliceClone`, `core.SliceSorted` |
| `maps` | map.go | `core.MapKeys`, `core.MapValues`, `core.MapClone`, `core.MapFilter`, `core.MapMerge`, `core.MapHasKey` |
| `strconv` | int.go | `core.Atoi`, `core.Itoa`, `core.FormatInt`, `core.FormatUint`, `core.ParseInt` |
| `strings`, `unicode/utf8`, `html` | string.go | `core.Concat`, `core.Join`, `core.Split`, `core.SplitN`, `core.Lower`, `core.Upper`, `core.Contains`, `core.HasPrefix`, `core.HasSuffix`, `core.TrimPrefix`, `core.TrimSuffix`, `core.Trim`, `core.Replace`, `core.NewBuilder`, `core.NewReader`, `core.RuneCount`, `core.HTMLEscape`, `core.HTMLUnescape` |
| `sync` | sync.go | `core.Mutex`, `core.RWMutex`, `core.Once`, `core.WaitGroup`, `core.SyncMap` |
| `sync/atomic` | atomic.go | `core.AtomicBool`, `core.AtomicInt32`, `core.AtomicInt64`, `core.AtomicUint32`, `core.AtomicUint64`, `core.AtomicPointer[T]` |
| `testing` | test.go | `core.T`, `core.TB`, `core.B`, `core.F` (test.go has the runner aliases; assert.go has the AssertX/RequireX family + AssertVerbose; cli_assert.go has CLITest/AssertCLI) |
| `text/tabwriter` | table.go | `core.NewTable`, `core.Table` |
| `time` | time.go | `core.Now`, `core.Sleep`, `core.Since`, `core.Until`, `core.Duration`, `core.Time`, `core.Nanosecond`, `core.Microsecond`, `core.Millisecond`, `core.Second`, `core.Minute`, `core.Hour`, `core.ParseDuration`, `core.TimeFormat`, `core.TimeParse`, `core.TimeRFC3339` |

**If you need a stdlib helper that isn't exposed:** add it to the owner file. Never re-import the stdlib in a non-owner file. The CI check at `tests/cli/imports/` fails on SPOR drift.

## Test files

Single import line. Use `*T` (alias for `*testing.T`):

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

func TestRepository_Sync_Bad(t *T) {
    AssertError(t, parseFails(), "invalid syntax")
}

func TestRepository_Sync_Ugly(t *T) {
    RequireNoError(t, setupTempRepo())
    // ... continues only if precondition holds
}
```

`Test*` naming: `Test{Filename}_{Function}_{Good|Bad|Ugly}`. All three states are mandatory for full coverage:
- `_Good` — happy path
- `_Bad` — expected failure (rejected input, error returned, refused operation)
- `_Ugly` — edge case (zero values, large inputs, race-prone setup, boundary conditions)

`AssertX` for non-fatal assertions, `RequireX` for "stop the test if this fails" preconditions. Both wrap `testing.TB`. Failure messages are one-line, file:line + assertion + want/got, AI-readable. Pass = silent.

For godoc Example* funcs (in `*_example_test.go`):

```go
// ExampleNewArray demonstrates building a small Array[string], adding
// elements, and querying length and membership. Array is the zero-value-
// usable, type-safe slice wrapper preferred over []T for collections
// you'll mutate.
func ExampleNewArray() {
    a := NewArray[string]()
    a.Add("alpha")
    a.Add("bravo")
    Println(a.Len())
    Println(a.Contains("bravo"))
    // Output:
    // 2
    // true
}
```

The convention here **inverts** the AX-2 production rule: in production `.go` files the docblock IS the example (tab-indented usage), but in `*_example_test.go` files the function body IS the example (godoc renders it as a runnable code block) — so the docblock above each Example function is **descriptive prose** explaining the scenario being demonstrated. pkg.go.dev surfaces this prose above the rendered code.

Use `core.Println` for output, NOT `fmt.Println`. The `// Output:` comment block must match runtime stdout exactly; Go's test runner verifies this.

For fuzz harnesses (in `*_fuzz_test.go`):

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

Run with `go test -fuzz=FuzzXxx -fuzztime=30s`. Seed corpus alone passes via `go test -run "Fuzz"`.

## Errors

All errors flow through `core.E`. No `fmt.Errorf` chains, no `_ = err` silencing, no `panic` for ordinary failures.

```go
// AX-native: errors are infrastructure, not application logic
r := c.Fs().Read("/etc/hostname")
if !r.OK { return r }
hostname := r.Value.(string)

// Non-AX: errors dominate the code
data, err := os.ReadFile("/etc/hostname")
if err != nil {
    return fmt.Errorf("read hostname: %w", err)
}
hostname := string(data)
```

`core.Result` is `{Value any, OK bool}`. On failure, `Value` holds the error; on success, `Value` holds the actual value. Type-assert when consuming: `r.Value.(string)`, `r.Value.(*Type)`.

`core.AnError` is the sentinel for tests needing a non-nil error without caring about content.

## Style

- **UK English** — colour, organisation, optimise, behaviour, centre. Never American spellings.
- **No emojis** in code, comments, commit messages, or docs unless explicitly requested.
- **Present tense** in docs — describe what the code IS, not what it WILL DO.
- **Conventional commits** — `type(scope): description`. Co-authored-by trailers when an agent did the work.
- **Comments earn their place** — only add a comment when the WHY is non-obvious. Default to none unless it teaches a reader something they couldn't derive from the code.

## Working with this codebase

- **Don't front-load edge cases** — build simple, check, then harden.
- **Talk before coding** — discuss potential issues first, don't silently solve hypotheticals.
- **Listen when slowed down** — "talk to me" / "let me think" means stop and collaborate.
- **Delete before adding** — question existing complexity before introducing more.
- **The hierarchy is in the names** — don't create parallel structures for what's already in dot notation, URLs, etc.
- **"Copy means copy"** — when porting from a source (engine, sibling package, AltumCode), copy verbatim. Don't adapt or "clean up".
- **"Other agents may be working here"** — don't assume sole authorship; ask before restructuring.
- **"Out of scope"** — when an architectural issue surfaces, raise it for guidance. Don't invent a TODO and proceed past it.
- **Don't hide work** — explicit deferred items via the todo list, not silent skips.

## Where to look

- `docs/RFC.md` — canonical API catalog (every public type, function, method)
- `docs/pkg/PACKAGE_STANDARDS.md` — how to build packages on top of CoreGO
- `docs/primitives.md` — the universal types (Options/Config/Data/Service/Result)
- `docs/lifecycle.md` — module loading via lifecycle events
- `docs/messaging.md` — bus protocol (Message/Query/Task)
- `docs/errors.md` — error model
- `docs/testing.md` — test conventions
- `*_example_test.go` (every production .go has a sibling) — runnable godoc examples
- `*_fuzz_test.go` (parse-shaped functions) — fuzz harnesses with seed corpora
