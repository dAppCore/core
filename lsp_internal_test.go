// SPDX-License-Identifier: EUPL-1.2

package core

import "bytes"

// --- lspExtractImportPath ---

func TestLsp_lspExtractImportPath_Good(t *T) {
	path, ok := lspExtractImportPath(`	"dappco.re/go"`)
	AssertTrue(t, ok)
	AssertEqual(t, "dappco.re/go", path)
}

func TestLsp_lspExtractImportPath_Bad(t *T) {
	_, ok := lspExtractImportPath(`	// no import here`)
	AssertFalse(t, ok)
}

func TestLsp_lspExtractImportPath_Ugly(t *T) {
	// Aliased import — `name "path"` form
	path, ok := lspExtractImportPath(`	cryptorand "crypto/rand"`)
	AssertTrue(t, ok)
	AssertEqual(t, "crypto/rand", path)
}

// --- lspMakeDiagnostic ---

func TestLsp_lspMakeDiagnostic_Good(t *T) {
	d := lspMakeDiagnostic(4, `	"context"`, "context")
	AssertEqual(t, 4, d.Range.Start.Line)
	AssertEqual(t, LSPSeverityWarning, d.Severity)
	AssertEqual(t, "test-imports", d.Source)
	AssertContains(t, d.Message, "context")
}

func TestLsp_lspMakeDiagnostic_Bad(t *T) {
	d := lspMakeDiagnostic(0, "", "")
	AssertEqual(t, 0, d.Range.Start.Line)
	AssertEqual(t, "test-imports", d.Source)
}

func TestLsp_lspMakeDiagnostic_Ugly(t *T) {
	// Long line text — End.Character matches len(line)
	line := `	"github.com/some/very/long/package/path"`
	d := lspMakeDiagnostic(99, line, "github.com/some/very/long/package/path")
	AssertEqual(t, 99, d.Range.Start.Line)
	AssertEqual(t, len(line), d.Range.End.Character)
}

// --- lspExtractDocument ---

func TestLsp_lspExtractDocument_Good(t *T) {
	uri, content := lspExtractDocument(map[string]any{
		"textDocument": map[string]any{
			"uri":  "file:///agent_test.go",
			"text": "package agent_test",
		},
	})
	AssertEqual(t, "file:///agent_test.go", uri)
	AssertEqual(t, "package agent_test", string(content))
}

func TestLsp_lspExtractDocument_Bad(t *T) {
	uri, content := lspExtractDocument(map[string]any{})
	AssertEqual(t, "", uri)
	AssertNil(t, content)
}

func TestLsp_lspExtractDocument_Ugly(t *T) {
	// Wrong shape — params is not a map at all
	uri, content := lspExtractDocument("not-a-map")
	AssertEqual(t, "", uri)
	AssertNil(t, content)
}

// --- lspExtractDocumentChange ---

func TestLsp_lspExtractDocumentChange_Good(t *T) {
	uri, content := lspExtractDocumentChange(map[string]any{
		"textDocument": map[string]any{"uri": "file:///agent_test.go"},
		"contentChanges": []any{
			map[string]any{"text": "package agent_test\n"},
		},
	})
	AssertEqual(t, "file:///agent_test.go", uri)
	AssertEqual(t, "package agent_test\n", string(content))
}

func TestLsp_lspExtractDocumentChange_Bad(t *T) {
	// Missing contentChanges
	uri, content := lspExtractDocumentChange(map[string]any{
		"textDocument": map[string]any{"uri": "file:///x"},
	})
	AssertEqual(t, "file:///x", uri)
	AssertNil(t, content)
}

func TestLsp_lspExtractDocumentChange_Ugly(t *T) {
	// Empty contentChanges array
	uri, content := lspExtractDocumentChange(map[string]any{
		"textDocument":   map[string]any{"uri": "file:///x"},
		"contentChanges": []any{},
	})
	AssertEqual(t, "file:///x", uri)
	AssertNil(t, content)
}

// --- lspTestImportsDiagnostic ---

func TestLsp_lspTestImportsDiagnostic_Good(t *T) {
	content := []byte(`package agent_test

import (
	"context"

	. "dappco.re/go"
)
`)
	diags := lspTestImportsDiagnostic("file:///agent_test.go", content)
	AssertLen(t, diags, 1)
	AssertContains(t, diags[0].Message, "context")
}

func TestLsp_lspTestImportsDiagnostic_Bad(t *T) {
	// Non-test file path returns nil (rule only applies to *_test.go).
	content := []byte(`package agent

import "context"
`)
	diags := lspTestImportsDiagnostic("file:///agent.go", content)
	AssertEmpty(t, diags)
}

func TestLsp_lspTestImportsDiagnostic_Ugly(t *T) {
	// _internal_test.go files are exempt
	content := []byte(`package core

import "sync"

var _ sync.Mutex
`)
	diags := lspTestImportsDiagnostic("file:///agent_internal_test.go", content)
	AssertEmpty(t, diags)
}

// --- lspServer methods ---
//
// Construct an lspServer with bytes.Buffer-backed in/out so the
// dispatch + frame-handling logic can be exercised without OS stdio.

func newTestLSPServer() (*lspServer, *bytes.Buffer, *bytes.Buffer) {
	in := &bytes.Buffer{}
	out := &bytes.Buffer{}
	srv := &lspServer{
		in:        NewBufReader(in),
		out:       out,
		documents: map[string][]byte{},
	}
	return srv, in, out
}

func TestLsp_lspServer_run_Good(t *T) {
	srv, _, _ := newTestLSPServer()
	ctx, cancel := WithCancel(Background())
	cancel()
	r := srv.run(ctx)
	AssertFalse(t, r.OK)
}

func TestLsp_lspServer_run_Bad(t *T) {
	// Empty stdin: readMessage returns EOF, run returns OK.
	srv, _, _ := newTestLSPServer()
	r := srv.run(Background())
	AssertTrue(t, r.OK)
}

func TestLsp_lspServer_run_Ugly(t *T) {
	// Malformed frame: missing Content-Length triggers an error.
	srv, in, _ := newTestLSPServer()
	in.WriteString("garbage\r\n\r\n{}")
	srv.in = NewBufReader(in)
	r := srv.run(Background())
	AssertFalse(t, r.OK)
}

func TestLsp_lspServer_readMessage_Good(t *T) {
	srv, in, _ := newTestLSPServer()
	body := `{"jsonrpc":"2.0","method":"initialize","id":1}`
	in.WriteString("Content-Length: ")
	in.WriteString(Itoa(len(body)))
	in.WriteString("\r\n\r\n")
	in.WriteString(body)
	srv.in = NewBufReader(in)

	got, err := srv.readMessage()
	AssertNoError(t, err)
	AssertEqual(t, body, string(got))
}

func TestLsp_lspServer_readMessage_Bad(t *T) {
	srv, _, _ := newTestLSPServer()
	_, err := srv.readMessage()
	AssertError(t, err)
}

func TestLsp_lspServer_readMessage_Ugly(t *T) {
	// Missing Content-Length but valid blank line: should error
	srv, in, _ := newTestLSPServer()
	in.WriteString("X-Header: noise\r\n\r\n")
	srv.in = NewBufReader(in)
	_, err := srv.readMessage()
	AssertError(t, err)
}

func TestLsp_lspServer_writeMessage_Good(t *T) {
	srv, _, out := newTestLSPServer()
	err := srv.writeMessage(map[string]any{"jsonrpc": "2.0", "id": 1, "result": "ok"})
	AssertNoError(t, err)
	AssertContains(t, out.String(), "Content-Length:")
	AssertContains(t, out.String(), `"result":"ok"`)
}

func TestLsp_lspServer_writeMessage_Bad(t *T) {
	srv, _, out := newTestLSPServer()
	// Channel can't be JSON-marshalled — JSONMarshal returns OK=false.
	ch := make(chan int)
	err := srv.writeMessage(ch)
	AssertError(t, err)
	AssertEqual(t, 0, out.Len())
}

func TestLsp_lspServer_writeMessage_Ugly(t *T) {
	// Nil payload still serialises (as "null").
	srv, _, out := newTestLSPServer()
	err := srv.writeMessage(nil)
	AssertNoError(t, err)
	AssertContains(t, out.String(), "Content-Length:")
}

func TestLsp_lspServer_dispatch_Good(t *T) {
	srv, _, out := newTestLSPServer()
	id := 1
	srv.dispatch([]byte(`{"jsonrpc":"2.0","method":"initialize","id":1}`))
	_ = id
	AssertContains(t, out.String(), "core-go-lsp")
}

func TestLsp_lspServer_dispatch_Bad(t *T) {
	// Malformed JSON — dispatch silently ignores.
	srv, _, out := newTestLSPServer()
	srv.dispatch([]byte(`not json`))
	AssertEqual(t, 0, out.Len())
}

func TestLsp_lspServer_dispatch_Ugly(t *T) {
	// Unknown method — silently ignored, no response.
	srv, _, out := newTestLSPServer()
	srv.dispatch([]byte(`{"jsonrpc":"2.0","method":"unknown/method","id":42}`))
	AssertEqual(t, 0, out.Len())
}

func TestLsp_lspServer_handleInitialize_Good(t *T) {
	srv, _, out := newTestLSPServer()
	id := 5
	srv.handleInitialize(lspMessage{ID: &id, Method: "initialize"})
	AssertContains(t, out.String(), "core-go-lsp")
	AssertContains(t, out.String(), "0.9.0")
}

func TestLsp_lspServer_handleInitialize_Bad(t *T) {
	srv, _, out := newTestLSPServer()
	srv.handleInitialize(lspMessage{Method: "initialize"})
	AssertContains(t, out.String(), "core-go-lsp")
}

func TestLsp_lspServer_handleInitialize_Ugly(t *T) {
	// Initialize response advertises capabilities including diagnosticProvider.
	srv, _, out := newTestLSPServer()
	id := 1
	srv.handleInitialize(lspMessage{ID: &id, Method: "initialize"})
	AssertContains(t, out.String(), "diagnosticProvider")
	AssertContains(t, out.String(), "textDocumentSync")
}

func TestLsp_lspServer_handleDocumentSync_Good(t *T) {
	srv, _, out := newTestLSPServer()
	srv.handleDocumentSync(lspMessage{
		Method: "textDocument/didOpen",
		Params: map[string]any{
			"textDocument": map[string]any{
				"uri":  "file:///agent_test.go",
				"text": "package agent_test\n\nimport \"context\"\n",
			},
		},
	})
	AssertContains(t, out.String(), "publishDiagnostics")
	AssertContains(t, out.String(), "context")
}

func TestLsp_lspServer_handleDocumentSync_Bad(t *T) {
	srv, _, out := newTestLSPServer()
	srv.handleDocumentSync(lspMessage{Method: "textDocument/didOpen", Params: nil})
	AssertEqual(t, 0, out.Len())
}

func TestLsp_lspServer_handleDocumentSync_Ugly(t *T) {
	// Clean test file produces empty diagnostics array (still a valid notification).
	srv, _, out := newTestLSPServer()
	srv.handleDocumentSync(lspMessage{
		Method: "textDocument/didOpen",
		Params: map[string]any{
			"textDocument": map[string]any{
				"uri":  "file:///agent_test.go",
				"text": "package agent_test\n\nimport . \"dappco.re/go\"\n",
			},
		},
	})
	AssertContains(t, out.String(), "publishDiagnostics")
	AssertContains(t, out.String(), `"diagnostics":[]`)
}

func TestLsp_lspServer_handleDocumentChange_Good(t *T) {
	srv, _, out := newTestLSPServer()
	srv.handleDocumentChange(lspMessage{
		Method: "textDocument/didChange",
		Params: map[string]any{
			"textDocument": map[string]any{"uri": "file:///agent_test.go"},
			"contentChanges": []any{
				map[string]any{"text": "package agent_test\n\nimport \"sync\"\n"},
			},
		},
	})
	AssertContains(t, out.String(), "sync")
}

func TestLsp_lspServer_handleDocumentChange_Bad(t *T) {
	srv, _, out := newTestLSPServer()
	srv.handleDocumentChange(lspMessage{Params: nil})
	AssertEqual(t, 0, out.Len())
}

func TestLsp_lspServer_handleDocumentChange_Ugly(t *T) {
	srv, _, _ := newTestLSPServer()
	srv.handleDocumentChange(lspMessage{
		Params: map[string]any{
			"textDocument":   map[string]any{"uri": "file:///x_test.go"},
			"contentChanges": []any{},
		},
	})
	// Empty contentChanges still tracks the doc with nil content.
	AssertEqual(t, 1, len(srv.documents))
	AssertNil(t, srv.documents["file:///x_test.go"])
}

func TestLsp_lspServer_handleDocumentClose_Good(t *T) {
	srv, _, _ := newTestLSPServer()
	srv.documents["file:///agent_test.go"] = []byte("content")
	srv.handleDocumentClose(lspMessage{
		Method: "textDocument/didClose",
		Params: map[string]any{
			"textDocument": map[string]any{"uri": "file:///agent_test.go"},
		},
	})
	AssertEqual(t, 0, len(srv.documents))
}

func TestLsp_lspServer_handleDocumentClose_Bad(t *T) {
	srv, _, _ := newTestLSPServer()
	srv.handleDocumentClose(lspMessage{Params: nil})
	AssertEqual(t, 0, len(srv.documents))
}

func TestLsp_lspServer_handleDocumentClose_Ugly(t *T) {
	// Closing an unopened doc is a no-op.
	srv, _, _ := newTestLSPServer()
	srv.handleDocumentClose(lspMessage{
		Params: map[string]any{
			"textDocument": map[string]any{"uri": "file:///never-opened.go"},
		},
	})
	AssertEqual(t, 0, len(srv.documents))
}

func TestLsp_lspServer_publishDiagnostics_Good(t *T) {
	srv, _, out := newTestLSPServer()
	srv.publishDiagnostics("file:///agent_test.go", []byte("package agent_test\n\nimport \"sync\"\n"))
	AssertContains(t, out.String(), "publishDiagnostics")
	AssertContains(t, out.String(), "sync")
}

func TestLsp_lspServer_publishDiagnostics_Bad(t *T) {
	// Empty content — no diagnostics, but notification still sent with [].
	srv, _, out := newTestLSPServer()
	srv.publishDiagnostics("file:///agent_test.go", []byte{})
	AssertContains(t, out.String(), `"diagnostics":[]`)
}

func TestLsp_lspServer_publishDiagnostics_Ugly(t *T) {
	// Non-test path: no test-imports diagnostic produced.
	srv, _, out := newTestLSPServer()
	srv.publishDiagnostics("file:///agent.go", []byte("package agent\n\nimport \"sync\"\n"))
	AssertContains(t, out.String(), `"diagnostics":[]`)
}

func TestLsp_lspServer_respond_Good(t *T) {
	srv, _, out := newTestLSPServer()
	id := 7
	srv.respond(&id, "result-payload", nil)
	AssertContains(t, out.String(), `"result":"result-payload"`)
}

func TestLsp_lspServer_respond_Bad(t *T) {
	srv, _, out := newTestLSPServer()
	id := 8
	srv.respond(&id, nil, &lspError{Code: -32601, Message: "method not found"})
	AssertContains(t, out.String(), "method not found")
}

func TestLsp_lspServer_respond_Ugly(t *T) {
	// Notification-style response with nil id.
	srv, _, out := newTestLSPServer()
	srv.respond(nil, "ok", nil)
	AssertContains(t, out.String(), "ok")
}

func TestLsp_lspServer_notify_Good(t *T) {
	srv, _, out := newTestLSPServer()
	srv.notify("textDocument/publishDiagnostics", map[string]any{"uri": "file:///x", "diagnostics": []any{}})
	AssertContains(t, out.String(), "publishDiagnostics")
}

func TestLsp_lspServer_notify_Bad(t *T) {
	// Notification with nil params still sends a frame.
	srv, _, out := newTestLSPServer()
	srv.notify("custom/event", nil)
	AssertContains(t, out.String(), "custom/event")
}

func TestLsp_lspServer_notify_Ugly(t *T) {
	// Method with empty string — still serialises (frame produced).
	srv, _, out := newTestLSPServer()
	srv.notify("", map[string]any{"any": "thing"})
	AssertContains(t, out.String(), "Content-Length:")
}

// --- lspResultShapeDiagnostic ---

func TestLsp_lspResultShapeDiagnostic_Good(t *T) {
	content := []byte(`package agent

func dispatch() {
	x, err := callExternal()
	_ = x; _ = err
}
`)
	diags := lspResultShapeDiagnostic("file:///agent.go", content)
	AssertNotEmpty(t, diags)
	AssertEqual(t, "result-shape", diags[0].Source)
	AssertContains(t, diags[0].Message, "Result")
}

func TestLsp_lspResultShapeDiagnostic_Bad(t *T) {
	// Non-.go file — diagnostic doesn't apply.
	content := []byte("x, err := foo()")
	diags := lspResultShapeDiagnostic("file:///notes.md", content)
	AssertEmpty(t, diags)
}

func TestLsp_lspResultShapeDiagnostic_Ugly(t *T) {
	// (T, error) inside a raw-string fixture must NOT trip the detector
	// — the strip pass elides string literals first.
	content := []byte("package agent\n\nvar fixture = `x, err := callExternal()`\n")
	diags := lspResultShapeDiagnostic("file:///agent.go", content)
	AssertEmpty(t, diags)
}

// --- lspMatchResultPattern ---

func TestLsp_lspMatchResultPattern_Good(t *T) {
	cases := []string{
		"x, err := callExternal()",
		"if x, err := f(); err != nil {",
		"if err := g(); err != nil {",
		"_, err := openFile(path)",
	}
	for _, line := range cases {
		_, ok := lspMatchResultPattern(line)
		AssertTrue(t, ok, line)
	}
}

func TestLsp_lspMatchResultPattern_Bad(t *T) {
	// Lines that do NOT match the (T, error) idiom.
	clean := []string{
		"r := callExternal()",
		"if r := f(); !r.OK {",
		"return Result{Value: out, OK: true}",
		"package agent",
	}
	for _, line := range clean {
		_, ok := lspMatchResultPattern(line)
		AssertFalse(t, ok, line)
	}
}

func TestLsp_lspMatchResultPattern_Ugly(t *T) {
	// Edge cases — leading whitespace, indented forms.
	indented := "\t\tx, err := callExternal()"
	_, ok := lspMatchResultPattern(indented)
	AssertTrue(t, ok)
}

// --- lspStripGoSyntax ---

func TestLsp_lspStripGoSyntax_Good(t *T) {
	in := false
	out := lspStripGoSyntax(`agent := "snider" // a comment`, &in)
	AssertContains(t, out, "agent")
	AssertContains(t, out, ":=")
	AssertNotContains(t, out, "snider")
	AssertNotContains(t, out, "comment")
}

func TestLsp_lspStripGoSyntax_Bad(t *T) {
	in := false
	out := lspStripGoSyntax("", &in)
	AssertEqual(t, "", out)
}

func TestLsp_lspStripGoSyntax_Ugly(t *T) {
	// Block comment spans line — state must persist across calls.
	in := false
	first := lspStripGoSyntax("agent /* start", &in)
	AssertTrue(t, in)
	AssertContains(t, first, "agent")

	second := lspStripGoSyntax("inside */ end", &in)
	AssertFalse(t, in)
	AssertContains(t, second, "end")
	AssertNotContains(t, second, "inside")
}
