// SPDX-License-Identifier: EUPL-1.2

// Language Server Protocol implementation for the Core framework.
//
// LSPServe starts a JSON-RPC-over-stdio Language Server that surfaces
// the same drift signals as tests/cli/imports, tests/cli/naming, and
// tests/cli/test_imports — but as inline editor diagnostics. Editors
// (VS Code, Zed) and AI agents (Claude Code, Codex) connect over
// stdio and receive textDocument/publishDiagnostics on every save,
// closing the feedback loop from "fail at CI" to "fail at keystroke".
//
// # Architecture
//
// The server core handles LSP plumbing — message framing, method
// dispatch, document state. Diagnostic-producing logic lives in
// pluggable DiagnosticSource functions registered via
// LSPRegisterDiagnostic. The default registration includes
// test-imports drift; add SPOR and AX-7 sources as the LSP matures.
//
// # Wire protocol
//
// LSP messages frame on stdin/stdout as:
//
//	Content-Length: N\r\n
//	\r\n
//	{ JSON-RPC body }
//
// The body uses jsonrpc 2.0. core/go ships zero stdlib LSP deps;
// JSON marshalling routes through core.JSONMarshal/Unmarshal.
//
// Usage
//
//	c := core.New()
//	core.LSPServe(core.Background())  // blocks until stdin closes
//
// Editor integration: point the LSP client at the binary and use
// stdio transport. No port, no socket — straight pipes.
package core

// LSPServerName identifies this language server in client capability
// negotiation.
//
//	core.Println(core.LSPServerName)  // "core-go-lsp"
const LSPServerName = "core-go-lsp"

// LSPServerVersion follows core/go's release version. Bumped alongside
// core/go releases so editor capability negotiation stays accurate.
const LSPServerVersion = "0.9.0"

// LSPDiagnosticSeverity values match LSP spec — Error=1, Warning=2,
// Information=3, Hint=4. Drift signals default to Warning so they
// surface inline without blocking edit-time iteration.
const (
	LSPSeverityError       = 1
	LSPSeverityWarning     = 2
	LSPSeverityInformation = 3
	LSPSeverityHint        = 4
)

// LSPPosition is a 0-based line/character location inside a document.
//
//	pos := core.LSPPosition{Line: 12, Character: 0}
type LSPPosition struct {
	Line      int `json:"line"`
	Character int `json:"character"`
}

// LSPRange spans Start..End within a document.
//
//	r := core.LSPRange{Start: core.LSPPosition{Line: 12}, End: core.LSPPosition{Line: 12, Character: 80}}
type LSPRange struct {
	Start LSPPosition `json:"start"`
	End   LSPPosition `json:"end"`
}

// LSPDiagnostic is one drift finding at a specific document range.
// Severity uses the LSPSeverity* constants. Source identifies the
// producing rule (e.g. "test-imports", "spor", "ax-7").
//
//	d := core.LSPDiagnostic{
//	    Range:    core.LSPRange{Start: core.LSPPosition{Line: 4}, End: core.LSPPosition{Line: 4, Character: 80}},
//	    Severity: core.LSPSeverityWarning,
//	    Source:   "test-imports",
//	    Message:  "imports 'context' — use core.Background()/core.Context",
//	}
type LSPDiagnostic struct {
	Range    LSPRange `json:"range"`
	Severity int      `json:"severity"`
	Source   string   `json:"source"`
	Message  string   `json:"message"`
	Code     string   `json:"code,omitempty"`
}

// LSPDiagnosticSource produces diagnostics for one document. uri is
// the LSP document URI ("file:///abs/path"); content is the current
// file body. Return an empty slice when nothing applies.
//
//	core.LSPRegisterDiagnostic("test-imports", func(uri string, content []byte) []core.LSPDiagnostic {
//	    if !core.HasSuffix(uri, "_test.go") { return nil }
//	    // ... parse imports, return diagnostics ...
//	    return nil
//	})
type LSPDiagnosticSource func(uri string, content []byte) []LSPDiagnostic

var lspSources = map[string]LSPDiagnosticSource{}
var lspSourcesMu RWMutex

// LSPRegisterDiagnostic adds a named diagnostic source. The source runs
// on every textDocument/didOpen and didSave event. Replacing a source
// with the same name overwrites the previous registration.
//
//	core.LSPRegisterDiagnostic("test-imports", testImportsDiagnostic)
//	core.LSPRegisterDiagnostic("spor", sporDiagnostic)
func LSPRegisterDiagnostic(name string, fn LSPDiagnosticSource) {
	lspSourcesMu.Lock()
	defer lspSourcesMu.Unlock()
	lspSources[name] = fn
}

// LSPDiagnosticSources returns the registered source names sorted
// alphabetically — useful for capability negotiation and debug pages.
//
//	for _, name := range core.LSPDiagnosticSources() {
//	    core.Println("LSP source:", name)
//	}
func LSPDiagnosticSources() []string {
	lspSourcesMu.RLock()
	defer lspSourcesMu.RUnlock()
	names := make([]string, 0, len(lspSources))
	for k := range lspSources {
		names = append(names, k)
	}
	SliceSort(names)
	return names
}

// LSPComputeDiagnostics runs every registered source against (uri, content)
// and concatenates the results. Used internally by LSPServe and exposed
// here so tests/agents can compute diagnostics without spinning up the
// full server.
//
//	diags := core.LSPComputeDiagnostics("file:///path/to/foo_test.go", content)
//	for _, d := range diags { core.Println(d.Message) }
func LSPComputeDiagnostics(uri string, content []byte) []LSPDiagnostic {
	lspSourcesMu.RLock()
	srcs := make([]LSPDiagnosticSource, 0, len(lspSources))
	for _, fn := range lspSources {
		srcs = append(srcs, fn)
	}
	lspSourcesMu.RUnlock()

	var out []LSPDiagnostic
	for _, fn := range srcs {
		out = append(out, fn(uri, content)...)
	}
	return out
}

// LSPServe starts the language server. Reads JSON-RPC messages from
// stdin and writes responses + notifications to stdout. Blocks until
// stdin closes (EOF) or ctx cancels.
//
//	core.LSPServe(core.Background())
//
// Output is interleaved with diagnostic notifications; stderr remains
// available for log output. Editor clients connect via the standard
// stdio transport.
func LSPServe(ctx Context) Result {
	srv := &lspServer{
		in:        NewBufReader(Stdin()),
		out:       Stdout(),
		documents: map[string][]byte{},
	}
	return srv.run(ctx)
}

// --- internal LSP server state and message dispatch ---

type lspServer struct {
	in        *BufReader
	out       Writer
	documents map[string][]byte
	docsMu    Mutex
}

type lspMessage struct {
	JSONRPC string    `json:"jsonrpc"`
	ID      *int      `json:"id,omitempty"`
	Method  string    `json:"method,omitempty"`
	Params  any       `json:"params,omitempty"`
	Result  any       `json:"result,omitempty"`
	Error   *lspError `json:"error,omitempty"`
}

type lspError struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
}

func (s *lspServer) run(ctx Context) Result {
	for {
		select {
		case <-ctx.Done():
			return Result{Value: ctx.Err(), OK: false}
		default:
		}

		body, err := s.readMessage()
		if err != nil {
			if err == EOF {
				return Result{OK: true}
			}
			return Result{Value: err, OK: false}
		}

		s.dispatch(body)
	}
}

// readMessage reads one LSP frame: Content-Length header + blank line + JSON body.
func (s *lspServer) readMessage() ([]byte, error) {
	var contentLength int
	for {
		line, err := s.in.ReadString('\n')
		if err != nil {
			return nil, err
		}
		line = Trim(line)
		if line == "" {
			break
		}
		if HasPrefix(line, "Content-Length:") {
			n := Trim(TrimPrefix(line, "Content-Length:"))
			r := Atoi(n)
			if r.OK {
				contentLength = r.Value.(int)
			}
		}
	}
	if contentLength <= 0 {
		return nil, E("lsp.read", "missing Content-Length header", nil)
	}
	buf := make([]byte, contentLength)
	if _, err := s.in.Read(buf); err != nil {
		return nil, err
	}
	return buf, nil
}

// writeMessage sends one LSP frame. Marshals payload to JSON, prepends
// the Content-Length header, writes to stdout.
func (s *lspServer) writeMessage(payload any) error {
	r := JSONMarshal(payload)
	if !r.OK {
		return r.Value.(error)
	}
	body := r.Value.([]byte)
	header := Sprintf("Content-Length: %d\r\n\r\n", len(body))
	if rh := WriteString(s.out, header); !rh.OK {
		return rh.Value.(error)
	}
	if _, err := s.out.Write(body); err != nil {
		return err
	}
	return nil
}

func (s *lspServer) dispatch(raw []byte) {
	var msg lspMessage
	if r := JSONUnmarshal(raw, &msg); !r.OK {
		return
	}

	switch msg.Method {
	case "initialize":
		s.handleInitialize(msg)
	case "initialized":
		// notification — no response
	case "shutdown":
		s.respond(msg.ID, nil, nil)
	case "exit":
		// run loop exits on EOF; nothing to do here
	case "textDocument/didOpen", "textDocument/didSave":
		s.handleDocumentSync(msg)
	case "textDocument/didChange":
		s.handleDocumentChange(msg)
	case "textDocument/didClose":
		s.handleDocumentClose(msg)
	}
}

func (s *lspServer) respond(id *int, result any, lerr *lspError) {
	resp := lspMessage{JSONRPC: "2.0", ID: id, Result: result, Error: lerr}
	_ = s.writeMessage(resp)
}

func (s *lspServer) notify(method string, params any) {
	note := lspMessage{JSONRPC: "2.0", Method: method, Params: params}
	_ = s.writeMessage(note)
}

func (s *lspServer) handleInitialize(msg lspMessage) {
	caps := map[string]any{
		"capabilities": map[string]any{
			"textDocumentSync": 1, // full document sync
			"diagnosticProvider": map[string]any{
				"interFileDependencies": false,
				"workspaceDiagnostics":  false,
			},
		},
		"serverInfo": map[string]any{
			"name":    LSPServerName,
			"version": LSPServerVersion,
		},
	}
	s.respond(msg.ID, caps, nil)
}

func (s *lspServer) handleDocumentSync(msg lspMessage) {
	uri, content := lspExtractDocument(msg.Params)
	if uri == "" {
		return
	}
	s.docsMu.Lock()
	s.documents[uri] = content
	s.docsMu.Unlock()
	s.publishDiagnostics(uri, content)
}

func (s *lspServer) handleDocumentChange(msg lspMessage) {
	uri, content := lspExtractDocumentChange(msg.Params)
	if uri == "" {
		return
	}
	s.docsMu.Lock()
	s.documents[uri] = content
	s.docsMu.Unlock()
	s.publishDiagnostics(uri, content)
}

func (s *lspServer) handleDocumentClose(msg lspMessage) {
	uri, _ := lspExtractDocument(msg.Params)
	if uri == "" {
		return
	}
	s.docsMu.Lock()
	delete(s.documents, uri)
	s.docsMu.Unlock()
}

func (s *lspServer) publishDiagnostics(uri string, content []byte) {
	diags := LSPComputeDiagnostics(uri, content)
	if diags == nil {
		diags = []LSPDiagnostic{}
	}
	s.notify("textDocument/publishDiagnostics", map[string]any{
		"uri":         uri,
		"diagnostics": diags,
	})
}

// lspExtractDocument pulls (uri, text) from a textDocument/didOpen or
// textDocument/didSave Params payload (best-effort, unmarshal-via-map).
func lspExtractDocument(params any) (string, []byte) {
	m, ok := params.(map[string]any)
	if !ok {
		return "", nil
	}
	td, ok := m["textDocument"].(map[string]any)
	if !ok {
		return "", nil
	}
	uri, _ := td["uri"].(string)
	text, _ := td["text"].(string)
	return uri, []byte(text)
}

// lspExtractDocumentChange pulls (uri, text) from textDocument/didChange
// — uses contentChanges[0].text under full-sync mode.
func lspExtractDocumentChange(params any) (string, []byte) {
	m, ok := params.(map[string]any)
	if !ok {
		return "", nil
	}
	td, ok := m["textDocument"].(map[string]any)
	if !ok {
		return "", nil
	}
	uri, _ := td["uri"].(string)
	changes, ok := m["contentChanges"].([]any)
	if !ok || len(changes) == 0 {
		return uri, nil
	}
	first, _ := changes[0].(map[string]any)
	text, _ := first["text"].(string)
	return uri, []byte(text)
}

// --- default diagnostic sources ---

func init() {
	LSPRegisterDiagnostic("test-imports", lspTestImportsDiagnostic)
}

// lspTestImportsDiagnostic flags any non-`dappco.re/go` import in
// *_test.go and *_example_test.go files. *_internal_test.go files are
// exempt (package core, may use any stdlib the production owner
// imports). One diagnostic per offending import line.
func lspTestImportsDiagnostic(uri string, content []byte) []LSPDiagnostic {
	if !HasSuffix(uri, "_test.go") {
		return nil
	}
	if HasSuffix(uri, "_internal_test.go") {
		return nil
	}

	var diags []LSPDiagnostic
	lines := Split(string(content), "\n")
	inImportBlock := false
	for i, line := range lines {
		trimmed := Trim(line)

		if !inImportBlock {
			if HasPrefix(trimmed, "import (") {
				inImportBlock = true
				continue
			}
			if HasPrefix(trimmed, "import ") {
				if path, ok := lspExtractImportPath(trimmed); ok && path != "dappco.re/go" {
					diags = append(diags, lspMakeDiagnostic(i, line, path))
				}
			}
			continue
		}

		// Inside import block.
		if trimmed == ")" {
			inImportBlock = false
			continue
		}
		if trimmed == "" || HasPrefix(trimmed, "//") {
			continue
		}
		if path, ok := lspExtractImportPath(trimmed); ok && path != "dappco.re/go" {
			diags = append(diags, lspMakeDiagnostic(i, line, path))
		}
	}
	return diags
}

func lspExtractImportPath(line string) (string, bool) {
	// `"path"` or `name "path"` or `. "path"` or `_ "path"`
	q1 := -1
	q2 := -1
	for i := 0; i < len(line); i++ {
		if line[i] == '"' {
			if q1 == -1 {
				q1 = i
			} else {
				q2 = i
				break
			}
		}
	}
	if q1 < 0 || q2 < 0 {
		return "", false
	}
	return line[q1+1 : q2], true
}

func lspMakeDiagnostic(lineNum int, lineText, importPath string) LSPDiagnostic {
	return LSPDiagnostic{
		Range: LSPRange{
			Start: LSPPosition{Line: lineNum, Character: 0},
			End:   LSPPosition{Line: lineNum, Character: len(lineText)},
		},
		Severity: LSPSeverityWarning,
		Source:   "test-imports",
		Code:     "test-imports.stdlib",
		Message:  Sprintf("imports %q — test files may import only \"dappco.re/go\"; use the core wrapper instead", importPath),
	}
}
