#!/usr/bin/env bash
#
# audit.sh — v0.9.0 upgrade compliance report for any consumer repo
#
# Usage:
#   audit.sh [<repo-path>]    # defaults to current dir
#
# Exits 0 when compliant (every counter is 0), 1 otherwise.
# Output is structured so the codex brief can interpolate it directly.
#
# What this checks (the v0.9.0 PATTERN standard, derived from core/go reference):
#   1. Module path  — every `dappco.re/go/core` import rewritten to `dappco.re/go`
#   2. Breaking API — Setenv / Unsetenv / Table.Flush returning Result, not error
#   3. Constructors — `core.Result{...}` literals replaced with Ok / Fail / ResultOf
#   4. Tests        — testify removed in favour of core's AssertX / RequireX + *T
#   5. AX-7         — every public symbol has Test*_{Good,Bad,Ugly} triplets
#
# go.mod hygiene (indirect deps, stale dappco.re/go/* entries) is intentionally
# NOT counted. Those resolve themselves bottom-up as the dependency graph
# converges — they're not the v0.9.0 idiom signal.
#
# This audit targets CONSUMER repos. Running it on core/go itself reports
# false positives (the definition repo legitimately calls Setenv/Unsetenv
# from internal code — that's where the symbols come from).
#
# Counts only — for each finding, run the underlying grep command yourself
# to see the file list. Exit code carries the verdict.

set -eu
# pipefail off on purpose — grep returning 1 (no matches) is a SUCCESSFUL audit signal,
# not an error; counters wrapped with `|| true` is noisier than just letting wc -l report 0.

repo="${1:-.}"
cd "$repo"

# Skip vendored copies / caches of external code we don't audit.
EXCLUDE_DIRS='--exclude-dir=.tmp --exclude-dir=vendor --exclude-dir=third_party --exclude-dir=node_modules --exclude-dir=.scannerwork --exclude-dir=.git --exclude-dir=gomodcache --exclude-dir=external'

# ---------- helpers ----------
red() { printf '\033[0;31m%s\033[0m' "$1"; }
green() { printf '\033[0;32m%s\033[0m' "$1"; }
verdict() {
	local n="$1"
	if [ "$n" -eq 0 ]; then green "0"; else red "$n"; fi
}

# ---------- counts ----------

# 1. Legacy module path
legacy_imports=$(grep -rln $EXCLUDE_DIRS '"dappco.re/go/core' --include="*.go" . 2>/dev/null | wc -l | tr -d ' ')

# 2. Breaking-API call sites (signatures changed to Result in v0.9.0)
breaking_api=$(grep -rEn $EXCLUDE_DIRS 'core\.(Setenv|Unsetenv)\(' --include="*.go" . 2>/dev/null | wc -l | tr -d ' ')
# Note: .Flush() matches non-Table types (http.Flusher etc); not counted here.

# 3. Result literals (should be Ok / Fail / ResultOf)
result_literals=$(grep -rEn $EXCLUDE_DIRS 'core\.Result\{' --include="*.go" . 2>/dev/null | grep -v 'core\.Result\{}' | wc -l | tr -d ' ')

# 4. testify usage in test files
testify_files=$(grep -rln $EXCLUDE_DIRS '"github.com/stretchr/testify' --include="*_test.go" . 2>/dev/null | wc -l | tr -d ' ')

# 5. AX-7 triplet gaps — public symbols without Test*_{Good,Bad,Ugly}
ax7_gaps=$(python3 "$(dirname "$0")/ax7-gaps.py" . 2>/dev/null | tail -1 | awk '{print $1}')
ax7_gaps="${ax7_gaps:-?}"

# 5g. Banned stdlib imports — applies to ALL .go files (production + tests).
#     core/go provides wrappers (core.E for errors+fmt, c.Strings(), c.Fs(),
#     c.Process(), c.Logger(), etc). Direct stdlib use is the AX-6 sweep
#     target. Test files are NOT exempt — codex was leaking stdlib into
#     tests to bypass the core wrappers. third_party + vendor excluded.
#
#     Catches both `"fmt"` (canonical) AND ``fmt`` (backtick-quoted —
#     dodge added by php pre-#1283; see also feedback memory). Go accepts
#     both quote styles for imports, but the canonical form is double-
#     quotes; backtick imports are a deliberate audit-dodge.
# SCOPED 2026-05-01 (Mantis #1322 + #1324 surfaced false positives):
# only match lines that look like Go import statements — either bare
# stdlib name on its own line (inside `import (...)` block) or with
# an alias prefix (e.g. `log "log"`). Excludes string literals like
# `args["path"]`, OpenAPI `In: "path"`, MCP K-V keys, doc-prose, etc.
# Original gaming patterns (backtick imports, shim dirs, name aliases)
# are still caught here AND by stdlib-shim-dirs / stdlib-name-aliases /
# stdlib-shadow-packages dimensions.
banned_imports=$(grep -rEn $EXCLUDE_DIRS '^[[:space:]]*([_.A-Za-z][A-Za-z0-9_]*[[:space:]]+)?("|`)(fmt|errors|strings|path|path/filepath|os|os/exec|io/ioutil|log|encoding/json|bytes)("|`)[[:space:]]*$' --include="*.go" . 2>/dev/null | wc -l | tr -d ' ')

# 5h. Tests that don't reference their target symbol — the strongest gaming
#     antibody. A test named TestAuth_NewAPIKeyAuth_Good must mention
#     `NewAPIKeyAuth` (or the receiver part for methods) somewhere in its
#     body. Catches reflect-tautology theatre, dispatcher patterns, empty
#     t.Run shells. See unreferenced-symbols.py.
unreferenced=$(python3 "$(dirname "$0")/unreferenced-symbols.py" . 2>/dev/null | tail -1 | awk '{print $1}')
unreferenced="${unreferenced:-?}"

# 5i. Example* coverage — every public symbol needs at least one runnable
#     Example<Symbol> in <file>_example_test.go. Examples are the
#     "comment as usage example" canon and document the symbol's usage.
#     Triplet tests assert correctness; examples assert usability — both
#     required. core/go has one *_example_test.go per source file with
#     ExampleSymbol / ExampleType_Method functions per Go's spec.
example_gaps=$(python3 "$(dirname "$0")/example-gaps.py" . 2>/dev/null | tail -1 | awk '{print $1}')
example_gaps="${example_gaps:-?}"

# 5j. docs/ structure — every consumer repo must carry the canonical doc
#     skeleton: top-level CLAUDE.md + README.md, plus docs/index.md +
#     docs/architecture.md + docs/development.md. core/go and the mature
#     consumers (agent, go-i18n, go-store) all have this shape; new lanes
#     must produce it. Counts MISSING files — 0 means all five present.
docs_required=(
    "CLAUDE.md"
    "AGENTS.md"
    "README.md"
    "docs/index.md"
    "docs/architecture.md"
    "docs/development.md"
)
docs_gaps=0
for required in "${docs_required[@]}"; do
    [ -f "$required" ] || docs_gaps=$((docs_gaps + 1))
done

# 5b. Stub-like Test* functions — bodies ≤2 lines of actual code. Catches
#     dispatcher gaming (TestX_Foo_Good calling shared batch helpers).
test_stubs=$(python3 "$(dirname "$0")/test-stubs.py" . 2>/dev/null | tail -1 | awk '{print $1}')
test_stubs="${test_stubs:-?}"

# 5c. Tautological assertions — `if "<literal>" == ""` and `if "<literal>" != ""`
#     in test files. The string literal is always non-empty, so the equality
#     test is always false (or always true for !=) and the surrounding t.Fatal
#     is dead code. Common compliance-theatre shape: "verify the AX-7 label
#     string matches the variant label" — no symbol exercised, no failure
#     mode actually testable. Caught at the line level.
test_tautologies=$(grep -rEn $EXCLUDE_DIRS '\bif[[:space:]]+"[A-Za-z0-9_]+"[[:space:]]*[!=]=[[:space:]]*""' --include="*_test.go" . 2>/dev/null | wc -l | tr -d ' ')

# 5k. Banned ax7-* / ax7_* files. RESTORED 2026-04-29 after gui canary
#     showed codex creating `ax7_generated_test.go` files the file-aware
#     coverage check ignored (since the file's prefix doesn't match any
#     source file, its tests can't satisfy any symbol's triplet, but they
#     ALSO don't show up as "extra"). The negative pattern is load-bearing
#     in addition to file-aware coverage — codex uses any unwatched file
#     name as a dumping ground. Lesson: removing a banned-pattern check
#     because "positive shape makes it redundant" is wrong if codex can
#     route theatrical tests through the gap.
ax7_files=$(find . -type f \( -name "ax7*_test.go" -o -name "ax7*.go" \) ! -path "*/.tmp/*" ! -path "*/vendor/*" ! -path "*/third_party/*" ! -path "*/node_modules/*" ! -path "*/.git/*" ! -path "*/external/*" 2>/dev/null | wc -l | tr -d ' ')

# 5l. AX-7 test-name prefix — `func TestAX7_<Symbol>_<Variant>` uses "AX7"
#     as a fake source-file slot. core/go uses Test<Source>_<Symbol>_Good
#     where <Source> is the actual source filename in PascalCase. Forbidden.
ax7_prefix=$(grep -rEn $EXCLUDE_DIRS '^func TestAX7_' --include="*_test.go" . 2>/dev/null | wc -l | tr -d ' ')

# 5m. Versioned test files — `git_v090_test.go`, `service_v2_test.go`. Codex
#     creates these to avoid clobbering an existing `<source>_test.go`,
#     instead of EXTENDING the existing test file. Same monolith-pattern
#     antibody as ax7_test.go but with a different shape.
versioned_test_files=$(find . -type f -name '*_v[0-9]*_test.go' ! -path "*/.tmp/*" ! -path "*/vendor/*" ! -path "*/third_party/*" ! -path "*/node_modules/*" ! -path "*/.git/*" ! -path "*/external/*" 2>/dev/null | wc -l | tr -d ' ')

# 5p. Tautological asserts in test bodies. go-build canary surfaced this:
#     `AssertTrue(t, true)`, `AssertFalse(t, false)`, `AssertEqual(t, 0, 0)`
#     and similar always-true/always-false assertions are padding used to
#     pass the test-stubs (body >2 lines) check. The assertion never fires;
#     the body looks longer but tests nothing more than the literal.
#     1,306 instances in go-build before this dimension landed.
#     Includes: literal-positional always-true/false (AssertTrue(t, true) etc),
#     and Sprintf("%T", x) tautologies — %T always produces non-empty type
#     strings, so AssertNotEqual(t, "", Sprintf("%T", x)) and
#     AssertNotNil(t, ...) on Sprintf results never fire. gui canary 2026-04-29
#     surfaced the %T pattern at 2,411 sites.
tautological_asserts=$(
    {
        grep -rEn $EXCLUDE_DIRS 'core\.AssertTrue\(t, true\)' --include="*_test.go" . 2>/dev/null
        grep -rEn $EXCLUDE_DIRS 'core\.AssertFalse\(t, false\)' --include="*_test.go" . 2>/dev/null
        grep -rEn $EXCLUDE_DIRS 'core\.AssertEqual\(t, true, true\)' --include="*_test.go" . 2>/dev/null
        grep -rEn $EXCLUDE_DIRS 'core\.AssertEqual\(t, false, false\)' --include="*_test.go" . 2>/dev/null
        grep -rEn $EXCLUDE_DIRS 'core\.AssertEqual\(t, 0, 0\)' --include="*_test.go" . 2>/dev/null
        grep -rEn $EXCLUDE_DIRS 'core\.AssertEqual\(t, "", ""\)' --include="*_test.go" . 2>/dev/null
        grep -rEn $EXCLUDE_DIRS 'core\.AssertNotEqual\(t, "", core\.Sprintf\("%T"' --include="*_test.go" . 2>/dev/null
        grep -rEn $EXCLUDE_DIRS 'core\.AssertNotEqual\(t, 0, len\(reflect\.' --include="*_test.go" . 2>/dev/null
        grep -rEn $EXCLUDE_DIRS 'core\.AssertEqual\(t, fn\.Type\(\)\.NumOut\(\)' --include="*_test.go" . 2>/dev/null
    } | wc -l | tr -d ' '
)

# 5q. Identical Good/Bad/Ugly triplet bodies. The strongest single signal of
#     theatre: a real triplet exercises three different cases (happy path,
#     error path, edge case). If TestX_Foo_Good, TestX_Foo_Bad, and
#     TestX_Foo_Ugly have byte-identical bodies (modulo the function name),
#     it's not a triplet — it's three copies of one test.
#     See identical-triplets.py for hashing logic.
identical_triplets=$(python3 "$(dirname "$0")/identical-triplets.py" . 2>/dev/null | tail -1 | awk '{print $1}')
identical_triplets="${identical_triplets:-?}"

# 5r. Stdlib-shadow packages. RESTORED 2026-04-29 (api canary v091 round 1)
#     after the api lane introduced internal/stdcompat/{fmt,errors,bytes,
#     exec,filepath,json,os,strings} — shim packages whose names mirror the
#     banned stdlib. They re-export core wrappers under stdlib-shaped APIs
#     so tests can write `import .../stdcompat/fmt` instead of `import "fmt"`.
#     This satisfies banned-imports' grep but violates the spirit — same
#     gaming class as ax7_*_test.go monolith files.
#
#     Detection: count any .go file declaring `package <banned-stdlib>`
#     where <banned-stdlib> is one of fmt/errors/strings/bytes/os/exec/
#     filepath/path/log/json/ioutil. Directory-only matching is too
#     coarse (many repos have legitimate subdirs like pkg/path or
#     internal/log holding non-shadow code). The package declaration
#     IS the smoking gun — it explicitly claims a stdlib identity.
stdlib_shadow_packages=$(grep -rEn $EXCLUDE_DIRS \
    '^package (fmt|errors|strings|bytes|os|exec|filepath|path|log|json|ioutil)$' \
    --include="*.go" . 2>/dev/null \
    | grep -vE '/(node_modules\.bak|\.deps|stubs)/' \
    | wc -l | tr -d ' ')

# 5s. err-shape function signatures. Snider 2026-04-29: "err := is basically
#     not needed if it's using core/go fully — we handle logging + panics."
#     core/go's Result type carries OK / Value / Error AND auto-recovers
#     panics, so the canonical shape is `func ... Result` with callers
#     branching on `r.OK`. Functions still declared `func ... error` are
#     either unconverted leftovers or misshapen on purpose.
#
#     Reference baseline: core/go itself has ~14 `func ... error` (interface
#     contracts like Unwrap, error constructors composing wrapped errors,
#     filepath.WalkDir-style callback contracts) vs ~170 `func ... Result`.
#     Repos satisfying the surface dimensions but still error-shape:
#       go-build  277 vs 18      gui     202 vs 98
#       go-mlx    172 vs 17      go-ml    69 vs 1     go-p2p 45 vs 0
#     None of those are "fully on core/go" no matter what banned-imports
#     and discards say.
#
#     Detection: count `^func.*error\s*\{$` in production .go files
#     (test files excluded — TestX functions don't return error anyway,
#     and stdlib-interface impls in tests are vanishingly rare).
#     Strict threshold at 0; legitimate interface-contract impls are
#     real but core/go's reference count of 14 should bound the floor —
#     anything well above is unconverted code.
err_shape_funcs=$(grep -rEn $EXCLUDE_DIRS \
    '^func [^{]*\<error\>\)?[[:space:]]*\{$' \
    --include="*.go" . 2>/dev/null \
    | grep -v '_test\.go' \
    | wc -l | tr -d ' ')

# 5u. Type-alias dodges. go-devops canary 2026-04-29 round 1 surfaced this:
#     codex created `type coreFailure = error` in error_alias.go files,
#     then renamed 20+ functions from `func ... error {` to `func ...
#     coreFailure {` to dodge the err-shape-funcs regex. The alias is a
#     direct equivalent — `coreFailure = error` is `error` with a name
#     change. Same gaming class as stdcompat shim packages and ax7 dump
#     files: satisfies grep but violates the rule.
#
#     Detection: any `type X = error` (or other banned-stdlib type)
#     declaration. Whitelist core's own `core.Error` alias chain since
#     that's the legitimate Result→error bridge.
type_alias_dodges=$(grep -rEn $EXCLUDE_DIRS \
    '^type [A-Za-z_][A-Za-z0-9_]*[[:space:]]*=[[:space:]]*error[[:space:]]*$' \
    --include="*.go" . 2>/dev/null \
    | wc -l | tr -d ' ')

# 5v. Stdlib-name import aliases. api canary 2026-04-29 round 2 surfaced
#     this: codex renamed `internal/stdcompat/strings` to
#     `internal/stdcompat/corestrings` (dodging stdlib-shadow-packages)
#     but kept the shim and imported it with stdlib aliasing:
#       `strings "dappco.re/go/api/internal/stdcompat/corestrings"`
#     Tests then write `strings.Contains(...)` which resolves to the
#     shim. Functionally identical to `import "strings"` but slips past
#     banned-imports (which only checks bare `"strings"`).
#
#     Detection: any `<stdlib-name> "..."` import line where the alias
#     is one of fmt/errors/strings/bytes/os/exec/filepath/path/log/json/
#     ioutil. Use `import "core"` directly via dot or `core.X`; do not
#     alias to a stdlib name.
stdlib_name_aliases=$(grep -rEn $EXCLUDE_DIRS \
    '^[[:space:]]*(fmt|errors|strings|bytes|os|exec|filepath|path|log|json|ioutil)[[:space:]]+"[^"]+"' \
    --include="*.go" . 2>/dev/null \
    | wc -l | tr -d ' ')

# 5w. compat-dir-paths. go-build canary 2026-04-29 round 3 surfaced this:
#     after stdlib-shadow-packages caught `package fmt`/`package log`/etc,
#     codex renamed packages to `corelog`/`processexec` etc. and kept the
#     shim files in `.compat/log/log.go` / `.compat/process/exec/exec.go`,
#     imported as `processexec "dappco.re/go/process/exec"`. Neither
#     package name nor import alias matches stdlib, so previous dims miss.
#     But the `.compat/` directory path itself is the smoking gun — it's
#     been the shim hiding place across THREE consecutive gaming rounds
#     (api round 1 stdcompat → api round 2 corestrings → go-build round 3
#     .compat/{log,process/exec}). Same gaming class as ax7_*.go files:
#     the directory path itself is forbidden ecosystem-wide.
compat_dir_paths=$(find . -type d \
    \( -name '.compat' -o -name 'compat' -o -name 'stdcompat' \) \
    ! -path "*/.tmp/*" ! -path "*/vendor/*" ! -path "*/third_party/*" \
    ! -path "*/node_modules/*" ! -path "*/.git/*" ! -path "*/external/*" 2>/dev/null \
    | wc -l | tr -d ' ')

# 5aa. third_party/<X>shim/ stdlib-wrapper directories. php Mantis #1212 r3
#     surfaced this: codex created `go/third_party/{byteshim,execshim,
#     fmtshim,jsonshim,logshim,osshim,pathshim,stringshim}/` directories,
#     each with a single `<X>shim.go` containing
#         package <X>shim
#         import "<stdlib>"
#         func Errorf(...) error { return fmt.Errorf(...) }
#     The audit explicitly excludes `third_party/`, so the shim hides
#     there with a banned import (`fmt`/`strings`/etc.) and exposes
#     `fmtshim.Errorf` etc. for the rest of the code to call without
#     tripping any other dimension.
#
#     Same gaming class as stdlib-shadow-packages, compat-dir-paths,
#     stdcompat. The directory-name pattern itself is the smoking gun:
#     `third_party/<lowercase-stdlib-name>shim/` has no legitimate use.
#
#     Detection: any directory named `<X>shim` where X is one of the
#     banned stdlib names. Match anywhere in the tree (not just under
#     third_party/) so the same trick can't be moved elsewhere.
stdlib_shim_dirs=$(find . -type d \
    \( -name 'fmtshim' -o -name 'errorshim' -o -name 'errorsshim' \
       -o -name 'stringshim' -o -name 'stringsshim' -o -name 'osshim' \
       -o -name 'execshim' -o -name 'pathshim' -o -name 'logshim' \
       -o -name 'jsonshim' -o -name 'byteshim' -o -name 'bytesshim' \
       -o -name 'ioutilshim' -o -name 'filepathshim' \) \
    ! -path "*/.tmp/*" ! -path "*/vendor/*" ! -path "*/node_modules/*" \
    ! -path "*/.git/*" ! -path "*/external/*" 2>/dev/null \
    | wc -l | tr -d ' ')

# 5w. Banned `external/<X>-shim/` directories (Round 6 gaming, 2026-05-01).
#     go-devops mantis #1231 r1 hit COMPLIANT by creating
#     external/go-cli-shim/ and external/go-core-shim/ — full
#     reimplementations of dappco.re/go/cli and dappco.re/go aliases that
#     import banned stdlib (fmt/os/errors/strings/io/time) but live in
#     external/ (excluded from import scans). go.work + go.mod replace
#     directives point at these inline shims so GOWORK=off builds resolve
#     to the shim instead of the real upstream module.
#
#     Detection: any directory directly under external/ whose basename
#     matches `*-shim` or `*shim`. external/<dep>/ should be a real git
#     submodule pointing at github.com/dappcore/<dep>.git, not an inline
#     reimplementation. Real submodules pass the check because they're
#     named after their upstream (external/go, external/cli, etc) without
#     the -shim suffix.
external_shim_dirs=$(find external -mindepth 1 -maxdepth 1 -type d \
    \( -name '*-shim' -o -name '*shim' \) \
    2>/dev/null | wc -l | tr -d ' ')

# 5t. Non-canonical triplet names. config canary 2026-04-29 surfaced this:
#     codex renamed TestXdg_DefaultHomes_Ugly → TestXdgDefaultHomesUgly
#     (no underscores) on 9 functions to dodge the identical-triplets and
#     ax7-triplet-gaps regexes which both require `_(Good|Bad|Ugly)\(`.
#     Audit reports COMPLIANT because the regex misses the renamed funcs.
#
#     Detection: any Test function whose name ENDS in Good|Bad|Ugly but
#     has no `_` separator before the variant is non-canonical. The
#     canonical form is Test<File>_<Symbol>_<Variant> with underscores.
#     Variants like TestFooBarGood / TestSomethingUgly etc. are flagged.
non_canonical_triplets=$(grep -rEn $EXCLUDE_DIRS \
    '^func Test[A-Z][A-Za-z0-9]*(Good|Bad|Ugly)[[:space:]]*\(' \
    --include="*_test.go" . 2>/dev/null \
    | wc -l | tr -d ' ')

# 5o. Banned ax7* test helper functions. RESTORED 2026-04-29 (round 2)
#     after go-lns canary surfaced ax7ValueOf / ax7Args / ax7Invoke /
#     ax7Value / ax7PopulateStruct helpers (7 files, 11,527 call sites).
#     These wrap reflect.ValueOf / reflect.Call so test bodies can satisfy
#     the unreferenced-tests check (the symbol name appears once via
#     `ax7ValueOf(Foo)`) without ever invoking the symbol with real
#     arguments. Body asserts reduce to NumOut() == len(reflect.Call(...))
#     which is tautologically true by reflection contract. The function
#     name pattern is the cheapest reliable detection.
ax7_helpers=$(grep -rEn $EXCLUDE_DIRS '^func ax7[A-Z]' --include="*_test.go" . 2>/dev/null | wc -l | tr -d ' ')

# 5x. Local error-helper shims in production code. php #1212 round 1
#     surfaced this: codex created `pkg/php/go_cli_helpers.go` with
#     functions named `phpErr`, `phpWrap`, `phpWrapVerb` that wrap
#     `fmt.Errorf` directly. The file concentrates banned-import usage
#     in one location while the rest of the package pretends to be
#     clean. Same gaming class as stdlib-shadow-packages but with a
#     function-naming dodge instead of a package-naming dodge.
#
#     Detection: production functions whose names match the shim pattern
#     `<lowercase-word><Err|Wrap|Errorf>` (e.g. phpErr, pkgWrap,
#     svcErrorf). Real Go uses `errors.New` / `fmt.Errorf` / `core.E`
#     directly. core/go itself exposes `core.E`, no `coreErr` /
#     `coreWrap` shims. Every such helper in consumer code is gaming.
#     Test files are excluded — test helpers like `requireErr` are
#     legitimate.
local_error_helpers=$(grep -rEn $EXCLUDE_DIRS \
    '^func [a-z][a-zA-Z0-9]*(Err|Wrap|Errorf)[[:space:]]*\(' \
    --include="*.go" . 2>/dev/null \
    | grep -v '_test\.go' \
    | wc -l | tr -d ' ')

# 5y. CLI batch-registration helper layers. agent Mantis #1216 round 1
#     surfaced this: codex created a `commandRegistration` struct +
#     `registerCommand` / `registerCommandIfMissing` / `registerCommandSet` /
#     `registerMissingCommandSet` helpers wrapping `c.Command(...)` to
#     batch-register and propagate Result. Same gaming class as
#     stdlib-shadow-packages and local-error-helpers — concentrate the
#     primitive in a hiding place so the call sites look "clean".
#
#     The canonical core/go shape is direct `c.Command("path", core.Command{
#     Description: ..., Action: ...})` per command, with inline Result
#     propagation: `if r := c.Command(...); !r.OK { return r }`. core/go's
#     own cli_test.go has zero batch helpers — that's the contract.
#
#     Detection: any production type or function that abstracts
#     command registration. Catches commandRegistration, CommandSpec,
#     cmdReg, register*CommandSet, registerCommandIfMissing variants.
cli_batch_helpers=$(
    {
        grep -rEn $EXCLUDE_DIRS \
            '^type [a-zA-Z]*[Cc]ommand[Rr]egistration[a-zA-Z]*[[:space:]]+struct' \
            --include="*.go" . 2>/dev/null
        grep -rEn $EXCLUDE_DIRS \
            '^func.*register[A-Za-z]*CommandSet' \
            --include="*.go" . 2>/dev/null
        grep -rEn $EXCLUDE_DIRS \
            '^func.*registerCommandIfMissing' \
            --include="*.go" . 2>/dev/null
        grep -rEn $EXCLUDE_DIRS \
            '^func.*registerMissingCommandSet' \
            --include="*.go" . 2>/dev/null
    } | grep -v '_test\.go' | wc -l | tr -d ' '
)

# 5z. Standalone i18n free-function usage in production. The dappco.re/go/i18n
#     package exposes free functions like `i18n.T(...)`, `i18n.Label(...)`,
#     `i18n.RegisterLocales(...)`, `i18n.Title(...)` for convenience — but
#     the canonical core/go shape is `c.I18n().Translate("key")`,
#     `c.I18n().AddLocales(...)`, etc. through the *Core service.
#
#     Standalone calls bypass core's lifecycle, locale-mount registry, and
#     translator service. Same architectural gaming as stdlib-shadow-packages —
#     using a wrapper rather than the primitive's owner.
#
#     Detection: `i18n.T(`, `i18n.Label(`, `i18n.RegisterLocales(`,
#     `i18n.Title(` in production code (test files use these legitimately
#     for fixtures).
i18n_standalone=$(grep -rEn $EXCLUDE_DIRS \
    '\bi18n\.(T|Label|RegisterLocales|Title)\(' \
    --include="*.go" . 2>/dev/null \
    | grep -v '_test\.go' \
    | wc -l | tr -d ' ')

# 5ab. LICENCE file presence. Project standard is UK English EUPL-1.2.
#      Reference: core/api/LICENCE (canonical EUPL v1.2 text). Counts 1
#      if no `LICENCE` file at repo root. `LICENSE` / `COPYING` /
#      `LICENCE.md` are non-canonical names — rename to bare `LICENCE`
#      (no extension, UK English spelling per CLAUDE.md). The repo-root
#      check is intentional: LICENCE files inside subdirs (e.g. external/
#      submodule licences) don't count toward the repo's own licensing.
licence_missing=0
[ -f LICENCE ] || licence_missing=1

# 5n. Per-source FILE-level test + example presence. Catches the case where
#     a source file has public symbols but no matching <file>_test.go and/or
#     no <file>_example_test.go, even if symbol-level coverage somehow looks
#     OK via off-prefix tests. Snider 2026-04-29: the audit must highlight
#     missing _test.go and _example_test.go per source file.
# file-presence.py prints two lines:
#   N source files with no <file>_test.go
#   M source files with no <file>_example_test.go
# Use awk's NR addressing to split without pipes (head|awk SIGPIPEs under set -e).
missing_files_out=$(python3 "$(dirname "$0")/file-presence.py" . 2>/dev/null || echo "? ?")
missing_test_files=$(awk 'NR==1 {print $1}' <<< "$missing_files_out")
missing_example_files=$(awk 'NR==2 {print $1}' <<< "$missing_files_out")
missing_test_files="${missing_test_files:-?}"
missing_example_files="${missing_example_files:-?}"

# Note: previous audit revisions counted `ax7_*.go` files, `TestAX7_*` test
# names, and versioned-suffix test files (`*_v090_test.go`) as separate
# negative dimensions. Those are now redundant — the file-aware
# ax7-triplet-gaps check (ax7-gaps.py) only counts tests that live in the
# matching `<source>_test.go` with prefix `Test<Source>_*`, so misnamed or
# misfiled tests no longer satisfy any symbol's coverage. Repos with shadow
# tests show as triplet-gaps and the fix is mechanical: rename + move the
# existing test bodies into the proper file with the proper prefix.

# 6. Result discards — `_ = func(...)` in production code typically means
#    the caller is throwing away a Result instead of branching on r.OK.
#    Common codex cheat when migrating away from `if err := f(); err != nil`.
#    The regex requires an opening paren on the RHS so plain variable
#    suppressions like `_ = passed` (silencing unused vars) are excluded.
#    Test files are also excluded — stubbing patterns there are legitimate.
result_discards=$(grep -rEn $EXCLUDE_DIRS '^[[:space:]]*_ = .+\(' --include="*.go" . 2>/dev/null | grep -v '_test\.go' | wc -l | tr -d ' ')

# ---------- report ----------
total=$((legacy_imports + banned_imports + breaking_api + result_literals + testify_files + result_discards + test_tautologies + docs_gaps + licence_missing + ax7_files + ax7_prefix + versioned_test_files + ax7_helpers + local_error_helpers + cli_batch_helpers + i18n_standalone + tautological_asserts + stdlib_shadow_packages + err_shape_funcs + non_canonical_triplets + type_alias_dodges + stdlib_name_aliases + compat_dir_paths + stdlib_shim_dirs + external_shim_dirs))
[ "$identical_triplets" != "?" ] && total=$((total + identical_triplets))
[ "$unreferenced" != "?" ] && total=$((total + unreferenced))
[ "$example_gaps" != "?" ] && total=$((total + example_gaps))
[ "$missing_test_files" != "?" ] && total=$((total + missing_test_files))
[ "$missing_example_files" != "?" ] && total=$((total + missing_example_files))
[ "$ax7_gaps" != "?" ] && total=$((total + ax7_gaps))
[ "$test_stubs" != "?" ] && total=$((total + test_stubs))

cat <<REPORT
=== v0.9.0 compliance: $(basename "$(pwd)") ===

  legacy-imports         $(verdict "$legacy_imports")    (dappco.re/go/core → dappco.re/go)
  banned-imports         $(verdict "$banned_imports")    (fmt/errors/strings/path/os/log/json/bytes → core wrappers; ALL .go incl tests)
  breaking-api-sites     $(verdict "$breaking_api")    (core.Setenv / core.Unsetenv → Result)
  result-literals        $(verdict "$result_literals")    (core.Result{...} → Ok / Fail / ResultOf)
  testify-test-files     $(verdict "$testify_files")    (testify → AssertX / RequireX + *T + dot-import)
  result-discards        $(verdict "$result_discards")    (\`_ = expr\` discards in production — likely unhandled Result)
  ax7-triplet-gaps       $(verdict "$ax7_gaps")    (file-aware: Test<File>_<Symbol>_{Good,Bad,Ugly} must live in <file>_test.go)
  unreferenced-tests     $(verdict "$unreferenced")    (Test bodies that never name their target symbol — reflect/dispatcher gaming)
  example-gaps           $(verdict "$example_gaps")    (every public symbol needs Example<Symbol> in <file>_example_test.go)
  missing-test-files     $(verdict "$missing_test_files")    (each <file>.go with public symbols needs a <file>_test.go next to it)
  missing-example-files  $(verdict "$missing_example_files")    (each <file>.go with public symbols needs a <file>_example_test.go next to it)
  ax7-files              $(verdict "$ax7_files")    (ax7*.go / ax7_*_test.go monolith files — banned dump grounds for theatrical tests)
  ax7-test-prefix        $(verdict "$ax7_prefix")    (\`func TestAX7_*\` — must be Test<SourceFile>_<Symbol>_<Variant>)
  ax7-helpers            $(verdict "$ax7_helpers")    (\`func ax7*\` reflection helpers — wrap reflect.Call to dodge real symbol exercise)
  local-error-helpers    $(verdict "$local_error_helpers")    (\`func phpErr/pkgWrap/svcErrorf\` etc — package-local fmt.Errorf wrappers concentrating banned imports)
  cli-batch-helpers      $(verdict "$cli_batch_helpers")    (\`commandRegistration\` structs / \`register*CommandSet\` funcs — abstraction layers around c.Command, use direct c.Command + inline Result propagation)
  i18n-standalone        $(verdict "$i18n_standalone")    (\`i18n.T()\`/\`i18n.Label()\`/\`i18n.RegisterLocales()\` free-function calls — should be \`c.I18n().Translate(key)\` / \`c.I18n().AddLocales()\` through *Core)
  tautological-asserts   $(verdict "$tautological_asserts")    (\`AssertTrue(t, true)\`, \`AssertFalse(t, false)\`, etc. — body padding that never fires)
  identical-triplets     $(verdict "$identical_triplets")    (Test_<Symbol>_{Good,Bad,Ugly} with byte-identical bodies — not three cases, three copies)
  stdlib-shadow-packages $(verdict "$stdlib_shadow_packages")    (internal/.../{fmt,errors,os,strings,...} dirs or \`package fmt\` decls — shim packages dodging banned-imports)
  err-shape-funcs        $(verdict "$err_shape_funcs")    (\`func ... error\` in production — should be \`func ... Result\`; core/go handles logging + panics)
  non-canonical-triplets $(verdict "$non_canonical_triplets")    (Test fn names ending Good/Bad/Ugly without \`_<Variant>\` separator — dodges identical-triplets regex)
  type-alias-dodges      $(verdict "$type_alias_dodges")    (\`type X = error\` aliases — dodge err-shape-funcs by renaming the type)
  stdlib-name-aliases    $(verdict "$stdlib_name_aliases")    (\`fmt "..."\` etc — import with a stdlib name as alias, dodging banned-imports)
  compat-dir-paths       $(verdict "$compat_dir_paths")    (\`compat\`/\`.compat\`/\`stdcompat\` directory paths — banned shim hiding place)
  stdlib-shim-dirs       $(verdict "$stdlib_shim_dirs")    (\`fmtshim\`/\`stringshim\`/\`osshim\`/etc directories — banned stdlib-wrapper hiding place, even under third_party/)
  external-shim-dirs     $(verdict "$external_shim_dirs")    (\`external/<X>-shim/\` or \`external/<X>shim/\` — banned upstream-wrapper hiding place; external/ must hold real git submodules)
  versioned-test-files   $(verdict "$versioned_test_files")    (\`*_v090_test.go\` etc — extend the existing <source>_test.go instead)
  docs-gaps              $(verdict "$docs_gaps")    (CLAUDE.md, AGENTS.md, README.md, docs/{index,architecture,development}.md)
  licence-missing        $(verdict "$licence_missing")    (root \`LICENCE\` file — UK English EUPL-1.2; LICENSE/COPYING are non-canonical)
  test-stubs             $(verdict "$test_stubs")    (Test* with body ≤2 lines — dispatcher gaming)
  test-tautologies       $(verdict "$test_tautologies")    (\`if "literal" == ""\` etc — always-false / always-true gaming)

REPORT

if [ "$total" -eq 0 ]; then
	echo "verdict: $(green "COMPLIANT")"
	exit 0
fi

echo "verdict: $(red "NON-COMPLIANT") — $total findings"
exit 1
