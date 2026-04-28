// SPDX-License-Identifier: EUPL-1.2

// AX-shaped test framework for the Core ecosystem.
//
// The framework is split across three production files for AX-3
// path-as-documentation:
//
//   - test.go        — Go-runner aliases (T/TB/B/F) + AnError sentinel.
//     The single SPOR import of the testing package.
//   - assert.go      — Assert*/Require* family + AssertVerbose flag.
//     The minimalist-by-default failure-output engine.
//   - cli_assert.go  — CLITest/AssertCLI/AssertCLIs for AX-10 binary
//     validation through c.Process().
//
// Test files use a single dot-import — `. "dappco.re/go"` — and the
// runner picks them up by signature because *T is type-identical to
// *testing.T. No `import "testing"` line in any consumer file.
//
//	package mypkg_test
//
//	import . "dappco.re/go"
//
//	func TestRepository_Sync_Good(t *T) {
//	    r := svc.SyncRepository("agent", "/srv/repos/agent")
//	    AssertTrue(t, r.OK)
//	    AssertEqual(t, "synced", r.Value.(string))
//	}
//
// Pass = silent (Go test default). Fail emits the AX one-line shape
// by default; flip core.AssertVerbose=true for the testify-style
// multi-line format. See assert.go for the full helper catalogue.
package core

import "testing"

// AnError is a sentinel error for tests that need a non-nil error
// without caring about its content. Mirrors testify's assert.AnError.
//
//	core.AssertError(t, somethingThatFails(), core.AnError.Error())
var AnError = NewError("core test sentinel error")

// T is the canonical Go test handle, exported as core.T so test files
// don't need a separate `import "testing"` line. Go's test runner
// accepts *core.T in TestXxx signatures because the alias is
// type-identical to *testing.T.
//
//	func TestSomething_Good(t *core.T) {
//	    core.AssertEqual(t, expected, actual)
//	}
type T = testing.T

// TB is the testing-handle interface (T + B), exported as core.TB so
// helpers can accept either Test or Benchmark contexts without
// importing testing.
//
//	func helper(t core.TB, ...) { t.Helper(); ... }
type TB = testing.TB

// B is the canonical Go benchmark handle, exported as core.B.
//
//	func BenchmarkSomething(b *core.B) { ... }
type B = testing.B

// F is the canonical Go fuzz harness, exported as core.F so fuzz files
// stay on the single-import pattern. Used in FuzzXxx(f *F) signatures.
//
//	func FuzzURLParse(f *F) {
//	    f.Add("https://example.com/path?q=1")
//	    f.Fuzz(func(t *T, raw string) { ... })
//	}
type F = testing.F
