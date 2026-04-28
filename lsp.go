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
// test-imports, result-shape, SPOR, and AX-7 drift sources.
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
var lspNamingCache = map[string]struct {
	Stamp    string
	Suffixes map[string]bool
}{}
var lspNamingCacheMu Mutex

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
	LSPRegisterDiagnostic("ax-7", lspNamingDiagnostic)
	LSPRegisterDiagnostic("result-shape", lspResultShapeDiagnostic)
	LSPRegisterDiagnostic("spor", lspSporDiagnostic)
	LSPRegisterDiagnostic("test-imports", lspTestImportsDiagnostic)
}

// lspSporDiagnostic flags production files that import a SPOR-protected
// stdlib package outside that package's single owner file.
func lspSporDiagnostic(uri string, content []byte) []LSPDiagnostic {
	if !HasSuffix(uri, ".go") {
		return nil
	}
	if HasSuffix(uri, "_test.go") || HasSuffix(uri, "_example_test.go") || HasSuffix(uri, "_fuzz_test.go") || HasSuffix(uri, "_internal_test.go") {
		return nil
	}

	owners := map[string]string{
		"bufio":             "scanner.go",
		"bytes":             "io.go",
		"cmp":               "math.go",
		"compress/gzip":     "embed.go",
		"context":           "context.go",
		"crypto/hkdf":       "hash.go",
		"crypto/hmac":       "hash.go",
		"crypto/rand":       "random.go",
		"crypto/sha256":     "hash.go",
		"crypto/sha3":       "sha3.go",
		"crypto/sha512":     "hash.go",
		"database/sql":      "sql.go",
		"embed":             "embed.go",
		"encoding/base64":   "encode.go",
		"encoding/binary":   "encode.go",
		"encoding/hex":      "encode.go",
		"encoding/json":     "json.go",
		"errors":            "error.go",
		"fmt":               "format.go",
		"go/ast":            "embed.go",
		"go/parser":         "embed.go",
		"go/token":          "embed.go",
		"hash":              "hash.go",
		"html":              "string.go",
		"html/template":     "template.go",
		"io":                "io.go",
		"io/fs":             "fs.go",
		"iter":              "iter.go",
		"maps":              "map.go",
		"math":              "math.go",
		"math/big":          "math.go",
		"math/bits":         "sha3.go",
		"math/rand/v2":      "random.go",
		"mime/multipart":    "api.go",
		"net":               "net.go",
		"net/http":          "api.go",
		"net/http/httptest": "api.go",
		"net/url":           "api.go",
		"os":                "os.go",
		"os/exec":           "process.go",
		"os/user":           "user.go",
		"path/filepath":     "path.go",
		"reflect":           "reflect.go",
		"regexp":            "regexp.go",
		"runtime":           "info.go",
		"runtime/debug":     "info.go",
		"slices":            "slice.go",
		"sort":              "slice.go",
		"strconv":           "int.go",
		"strings":           "string.go",
		"sync":              "sync.go",
		"sync/atomic":       "atomic.go",
		"testing":           "test.go",
		"text/tabwriter":    "table.go",
		"text/template":     "template.go",
		"time":              "time.go",
		"unicode":           "unicode.go",
		"unicode/utf8":      "string.go",
	}

	fileName := PathBase(TrimPrefix(uri, "file://"))
	var diags []LSPDiagnostic
	lines := Split(string(content), "\n")
	inImportBlock := false
	for i, line := range lines {
		trimmed := Trim(line)

		checkImport := func(importPath string) {
			owner, ok := owners[importPath]
			if !ok || fileName == owner {
				return
			}
			diags = append(diags, LSPDiagnostic{
				Range: LSPRange{
					Start: LSPPosition{Line: i, Character: 0},
					End:   LSPPosition{Line: i, Character: len(line)},
				},
				Severity: LSPSeverityWarning,
				Source:   "spor",
				Code:     "spor.violation",
				Message:  Sprintf("imports '%s' but %s is owned by %s; route through the core wrapper", importPath, importPath, owner),
			})
		}

		if !inImportBlock {
			if HasPrefix(trimmed, "import (") {
				inImportBlock = true
				continue
			}
			if HasPrefix(trimmed, "import ") {
				if path, ok := lspExtractImportPath(trimmed); ok {
					checkImport(path)
				}
			}
			continue
		}

		if trimmed == ")" {
			inImportBlock = false
			continue
		}
		if trimmed == "" || HasPrefix(trimmed, "//") {
			continue
		}
		if path, ok := lspExtractImportPath(trimmed); ok {
			checkImport(path)
		}
	}
	return diags
}

// lspNamingDiagnostic flags production symbols that do not have the
// Test*_{Symbol}_{Good,Bad,Ugly} triplet in the same directory's tests.
func lspNamingDiagnostic(uri string, content []byte) []LSPDiagnostic {
	if !HasSuffix(uri, ".go") || HasSuffix(uri, "_test.go") {
		return nil
	}

	topResult := Regex(`^func ([A-Za-z][A-Za-z0-9_]*)\s*[\[(]`)
	methodResult := Regex(`^func \([^)]*?\*?([A-Za-z][A-Za-z0-9_]*)(?:\[[^\]]+\])?\) ([A-Za-z][A-Za-z0-9_]*)\s*[\[(]`)
	testResult := Regex(`^func (Test[A-Za-z0-9_]+)\s*\(`)
	if !topResult.OK || !methodResult.OK || !testResult.OK {
		return nil
	}
	top := topResult.Value.(*Regexp)
	method := methodResult.Value.(*Regexp)
	test := testResult.Value.(*Regexp)

	path := TrimPrefix(uri, "file://")
	dir := PathDir(path)
	read := ReadDir(DirFS(dir), ".")
	testFiles := []string{}
	stampParts := []string{}
	if read.OK {
		for _, entry := range read.Value.([]FsDirEntry) {
			name := entry.Name()
			if !HasSuffix(name, "_test.go") && !HasSuffix(name, "_internal_test.go") {
				continue
			}
			testFiles = append(testFiles, name)
			modTime := int64(0)
			size := int64(0)
			stat := Stat(PathJoin(dir, name))
			if stat.OK {
				info := stat.Value.(FsFileInfo)
				modTime = info.ModTime().UnixNano()
				size = info.Size()
			}
			stampParts = append(stampParts, Sprintf("%s:%d:%d", name, modTime, size))
		}
	}
	SliceSort(testFiles)
	SliceSort(stampParts)
	stamp := Join("|", stampParts...)

	lspNamingCacheMu.Lock()
	cached, ok := lspNamingCache[dir]
	if ok && cached.Stamp == stamp {
		lspNamingCacheMu.Unlock()
		return lspNamingDiagnosticsFromSuffixes(content, top, method, cached.Suffixes)
	}
	lspNamingCacheMu.Unlock()

	suffixes := map[string]bool{}
	for _, name := range testFiles {
		file := ReadFile(PathJoin(dir, name))
		if !file.OK {
			continue
		}
		for _, line := range Split(string(file.Value.([]byte)), "\n") {
			matches := test.FindStringSubmatch(line)
			if len(matches) < 2 {
				continue
			}
			testName := matches[1]
			underscore := -1
			for i := 0; i < len(testName); i++ {
				if testName[i] == '_' {
					underscore = i
					break
				}
			}
			if underscore >= 0 && underscore+1 < len(testName) {
				suffixes[testName[underscore+1:]] = true
			}
		}
	}

	lspNamingCacheMu.Lock()
	lspNamingCache[dir] = struct {
		Stamp    string
		Suffixes map[string]bool
	}{Stamp: stamp, Suffixes: suffixes}
	lspNamingCacheMu.Unlock()

	return lspNamingDiagnosticsFromSuffixes(content, top, method, suffixes)
}

func lspNamingDiagnosticsFromSuffixes(content []byte, top, method *Regexp, suffixes map[string]bool) []LSPDiagnostic {
	var diags []LSPDiagnostic
	lines := Split(string(content), "\n")
	variants := []string{"Good", "Bad", "Ugly"}

	hasVariant := func(symbol, variant string) bool {
		wanted := Concat(symbol, "_", variant)
		target := Concat("_", wanted)
		for suffix := range suffixes {
			if suffix == wanted || HasSuffix(suffix, target) {
				return true
			}
		}
		return false
	}

	for i, line := range lines {
		symbol := ""
		if matches := top.FindStringSubmatch(line); len(matches) >= 2 {
			symbol = matches[1]
		}
		if matches := method.FindStringSubmatch(line); len(matches) >= 3 {
			symbol = Concat(matches[1], "_", matches[2])
		}
		if symbol == "" {
			continue
		}
		for _, variant := range variants {
			if hasVariant(symbol, variant) {
				continue
			}
			diags = append(diags, LSPDiagnostic{
				Range: LSPRange{
					Start: LSPPosition{Line: i, Character: 0},
					End:   LSPPosition{Line: i, Character: len(line)},
				},
				Severity: LSPSeverityHint,
				Source:   "ax-7",
				Code:     "ax-7.missing-variant",
				Message:  Sprintf("missing Test*_%s_%s (Good|Bad|Ugly) — write it to satisfy the AX-7 triplet", symbol, variant),
			})
		}
	}
	return diags
}

// lspResultShapeDiagnostic flags the classic Go `(value, error)` and
// `if err := f(); err != nil` patterns. core/go-shaped code returns
// core.Result, so any `, err :=` or `if err :=` is a tell that the
// callee hasn't been wrapped (or the caller hasn't switched to the
// Result-returning helper).
//
// Production .go files: every assignment-with-error is a smell.
// Test .go files: tests should call core helpers, not stdlib direct.
// *_internal_test.go and *_fuzz_test.go: NOT exempt — same rule.
//
// Examples flagged:
//
//	x, err := foo()                 // → use r := foo(); r.Value.(*X)
//	_, err := bar()                 // → r := bar(); if !r.OK { ... }
//	if x, err := f(); err != nil {  // → r := f(); if !r.OK { ... }
//	if err := g(); err != nil {     // → r := g(); if !r.OK { ... }
//
// String literals and comments are best-effort skipped — diagnostic
// is advisory anyway. Apply quick-fix: convert callee to Result return
// or wrap call site with `Result{}.New(value, err)`.
func lspResultShapeDiagnostic(uri string, content []byte) []LSPDiagnostic {
	if !HasSuffix(uri, ".go") {
		return nil
	}

	var diags []LSPDiagnostic
	lines := Split(string(content), "\n")
	inBlockComment := false
	for i, line := range lines {
		stripped := lspStripGoSyntax(line, &inBlockComment)
		if stripped == "" {
			continue
		}
		if msg, ok := lspMatchResultPattern(stripped); ok {
			diags = append(diags, LSPDiagnostic{
				Range: LSPRange{
					Start: LSPPosition{Line: i, Character: 0},
					End:   LSPPosition{Line: i, Character: len(line)},
				},
				Severity: LSPSeverityWarning,
				Source:   "result-shape",
				Code:     "result-shape.error-pair",
				Message:  msg,
			})
		}
	}
	return diags
}

// lspMatchResultPattern returns (suggestion, true) when line contains
// one of the (T, error) or single-error idioms. Patterns covered:
//
//	"x, err := f(...)"           — value + error declaration
//	"x, err = f(...)"            — value + error assignment
//	"_, err := f(...)"           — discarded value + error
//	"if x, err := f(...);"        — if-init with value + error
//	"if err := f(...);"           — if-init with single error
//
// The match runs against a syntax-stripped line (strings/comments
// removed) so fixture content inside raw strings doesn't false-trip.
func lspMatchResultPattern(stripped string) (string, bool) {
	// Trim leading whitespace
	s := stripped
	for len(s) > 0 && (s[0] == ' ' || s[0] == '\t') {
		s = s[1:]
	}

	// "if err := X(...); err != nil"
	if HasPrefix(s, "if err :=") || HasPrefix(s, "if err  :=") {
		return "result-shape: `if err := f(); err != nil` — switch to `r := f(); if !r.OK`", true
	}

	// "if x, err := f(...);"
	if HasPrefix(s, "if ") && Contains(s, ", err :=") && Contains(s, ";") {
		return "result-shape: `if x, err := f(); err != nil` — wrap callee in Result or use core's Result-returning equivalent", true
	}

	// Bare assignment forms: detect `, err :=` or `, err =` after an identifier
	if Contains(s, ", err :=") {
		return "result-shape: `x, err := f()` — wrap callee in Result or use `Result{}.New(value, err)`", true
	}
	if Contains(s, ", err =") {
		return "result-shape: `x, err = f()` — wrap callee in Result or use `Result{}.New(value, err)`", true
	}

	return "", false
}

// lspStripGoSyntax removes string literals and comments from line so
// pattern detectors don't false-trip on fixture content. Tracks
// inBlockComment across calls (Go /* ... */ block comments span lines).
func lspStripGoSyntax(line string, inBlockComment *bool) string {
	out := make([]byte, 0, len(line))
	i := 0
	n := len(line)
	for i < n {
		// Inside a block comment — skip until "*/"
		if *inBlockComment {
			if i+1 < n && line[i] == '*' && line[i+1] == '/' {
				*inBlockComment = false
				i += 2
				continue
			}
			i++
			continue
		}
		// Line comment "//"
		if i+1 < n && line[i] == '/' && line[i+1] == '/' {
			break
		}
		// Block comment open "/*"
		if i+1 < n && line[i] == '/' && line[i+1] == '*' {
			*inBlockComment = true
			i += 2
			continue
		}
		// Raw string `...`
		if line[i] == '`' {
			i++
			for i < n && line[i] != '`' {
				i++
			}
			if i < n {
				i++
			}
			continue
		}
		// Quoted string "..." or '...'
		if line[i] == '"' || line[i] == '\'' {
			delim := line[i]
			i++
			for i < n && line[i] != delim {
				if line[i] == '\\' && i+1 < n {
					i += 2
					continue
				}
				i++
			}
			if i < n {
				i++
			}
			continue
		}
		out = append(out, line[i])
		i++
	}
	return string(out)
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
