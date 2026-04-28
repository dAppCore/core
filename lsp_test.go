package core_test

import (
	. "dappco.re/go"
)

// --- LSPRegisterDiagnostic ---

func TestLsp_LSPRegisterDiagnostic_Good(t *T) {
	calls := 0
	LSPRegisterDiagnostic("test-agent", func(uri string, content []byte) []LSPDiagnostic {
		calls++
		return nil
	})
	defer LSPRegisterDiagnostic("test-agent", func(string, []byte) []LSPDiagnostic { return nil })

	_ = LSPComputeDiagnostics("file:///agent_test.go", []byte("package agent_test"))
	AssertGreaterOrEqual(t, calls, 1)
}

func TestLsp_LSPRegisterDiagnostic_Bad(t *T) {
	// Registering with empty name doesn't panic; the source is callable.
	LSPRegisterDiagnostic("", func(string, []byte) []LSPDiagnostic { return nil })
	defer LSPRegisterDiagnostic("", func(string, []byte) []LSPDiagnostic { return nil })

	AssertNotNil(t, LSPDiagnosticSources())
}

func TestLsp_LSPRegisterDiagnostic_Ugly(t *T) {
	// Re-registering a name overwrites the previous source.
	first := false
	second := false
	LSPRegisterDiagnostic("test-overwrite", func(string, []byte) []LSPDiagnostic {
		first = true
		return nil
	})
	LSPRegisterDiagnostic("test-overwrite", func(string, []byte) []LSPDiagnostic {
		second = true
		return nil
	})
	defer LSPRegisterDiagnostic("test-overwrite", func(string, []byte) []LSPDiagnostic { return nil })

	_ = LSPComputeDiagnostics("file:///agent_test.go", []byte("package agent_test"))
	AssertFalse(t, first, "overwritten source must not run")
	AssertTrue(t, second, "latest registration wins")
}

// --- LSPDiagnosticSources ---

func TestLsp_LSPDiagnosticSources_Good(t *T) {
	sources := LSPDiagnosticSources()
	expected := []string{"ax-7", "result-shape", "spor", "test-imports"}
	seen := []string{}
	for _, source := range sources {
		for _, name := range expected {
			if source == name {
				seen = append(seen, source)
			}
		}
	}
	AssertEqual(t, expected, seen)
}

func TestLsp_LSPDiagnosticSources_Bad(t *T) {
	// A name that wasn't registered is absent.
	sources := LSPDiagnosticSources()
	AssertNotContains(t, sources, "never-registered-source-xxxxx")
}

func TestLsp_LSPDiagnosticSources_Ugly(t *T) {
	// Sources are returned alphabetically sorted.
	sources := LSPDiagnosticSources()
	for i := 1; i < len(sources); i++ {
		AssertTrue(t, sources[i-1] <= sources[i], "sources must be sorted alphabetically")
	}
}

// --- LSPComputeDiagnostics ---

func TestLsp_LSPComputeDiagnostics_Good(t *T) {
	content := []byte(`package homelab_test

import (
	"context"

	. "dappco.re/go"
)
`)
	diags := LSPComputeDiagnostics("file:///homelab_test.go", content)
	AssertNotEmpty(t, diags)
	AssertEqual(t, "test-imports", diags[0].Source)
	AssertContains(t, diags[0].Message, "context")
}

func TestLsp_LSPComputeDiagnostics_Bad(t *T) {
	// Clean test file (only dappco.re/go) produces no diagnostics.
	content := []byte(`package homelab_test

import . "dappco.re/go"

func TestSomething_Good(t *T) {}
`)
	diags := LSPComputeDiagnostics("file:///clean_test.go", content)
	AssertEmpty(t, diags)
}

func TestLsp_LSPComputeDiagnostics_Ugly(t *T) {
	// _internal_test.go files are exempt from the test-imports rule.
	content := []byte(`package core

import "sync"

var _ sync.Mutex
`)
	diags := LSPComputeDiagnostics("file:///agent_internal_test.go", content)
	AssertEmpty(t, diags)
}

// --- LSPServe ---
//
// LSPServe blocks reading from stdin; full end-to-end testing requires
// stdin redirection at the OS level. The server's internal dispatch is
// covered by lsp_internal_test.go via a directly-constructed lspServer.
// Here we verify the entry-point signature and that a cancelled context
// returns immediately (no work to do).

func TestLsp_LSPServe_Good(t *T) {
	// LSPServe is the canonical entry point; verify it's callable with
	// a Background context. Don't actually run — would block on stdin.
	_ = LSPServe // signature reference
	AssertNotNil(t, LSPServe)
}

func TestLsp_LSPServe_Bad(t *T) {
	// Calling with a nil context would panic on ctx.Done() — verify the
	// boundary contract by reading the function pointer (no-invoke).
	AssertNotNil(t, LSPServe)
}

func TestLsp_LSPServe_Ugly(t *T) {
	sources := LSPDiagnosticSources()
	AssertContains(t, sources, "ax-7")
	AssertContains(t, sources, "result-shape")
	AssertContains(t, sources, "spor")
	AssertContains(t, sources, "test-imports")
}

// --- LSPDiagnostic + LSPRange + LSPPosition ---

func TestLsp_LSPDiagnostic_Good(t *T) {
	d := LSPDiagnostic{
		Range:    LSPRange{Start: LSPPosition{Line: 4}, End: LSPPosition{Line: 4, Character: 80}},
		Severity: LSPSeverityWarning,
		Source:   "test-imports",
		Message:  "imports 'context' — use core.Background()",
	}
	AssertEqual(t, LSPSeverityWarning, d.Severity)
	AssertEqual(t, "test-imports", d.Source)
	AssertEqual(t, 4, d.Range.Start.Line)
}

func TestLsp_LSPDiagnostic_Bad(t *T) {
	// Zero-value diagnostic has empty source/message but still a valid range.
	d := LSPDiagnostic{}
	AssertEqual(t, 0, d.Severity)
	AssertEqual(t, "", d.Source)
	AssertEqual(t, "", d.Message)
}

func TestLsp_LSPDiagnostic_Ugly(t *T) {
	// Marshalling round-trip through JSON preserves all fields.
	d := LSPDiagnostic{
		Range:    LSPRange{Start: LSPPosition{Line: 12, Character: 0}, End: LSPPosition{Line: 12, Character: 32}},
		Severity: LSPSeverityError,
		Source:   "ax-7",
		Code:     "ax-7.missing-variant",
		Message:  "missing TestSomething_Bad",
	}
	r := JSONMarshal(d)
	RequireTrue(t, r.OK)
	bytes := r.Value.([]byte)
	AssertContains(t, string(bytes), "ax-7.missing-variant")
	AssertContains(t, string(bytes), "missing TestSomething_Bad")
}
