// SPDX-License-Identifier: EUPL-1.2

// Remote communication primitive for the Core framework.
// API manages named streams to remote endpoints. The transport protocol
// (HTTP, WebSocket, SSE, MCP, TCP) is handled by protocol handlers
// registered by consumer packages.
//
// Drive is the phone book (WHERE to connect).
// API is the phone (HOW to connect).
//
// Usage:
//
//	// Configure endpoint
//	c.Drive().New(core.NewOptions(
//	    core.Option{Key: "name", Value: "charon"},
//	    core.Option{Key: "transport", Value: "http://10.69.69.165:9101/mcp"},
//	))
//
//	// Open stream
//	s := c.API().Stream("charon")
//	if s.OK { stream := s.Value.(Stream) }
//
//	// Remote Action dispatch
//	r := c.API().Call("charon", "agentic.status", opts)
package core

import (
	"mime/multipart"
	"net/http"
	"net/http/httptest"
)

// Stream is a bidirectional connection to a remote endpoint.
// Consumers implement this for each transport protocol.
//
//	type httpStream struct { ... }
//	func (s *httpStream) Send(data []byte) error { ... }
//	func (s *httpStream) Receive() ([]byte, error) { ... }
//	func (s *httpStream) Close() error { ... }
type Stream interface {
	Send(data []byte) error
	Receive() ([]byte, error)
	Close() error
}

// StreamFactory creates a Stream from a DriveHandle's transport config.
// Registered per-protocol by consumer packages.
//
//	factory := func(handle *core.DriveHandle) (core.Stream, error) {
//	    return nil, core.NewError(core.Concat("protocol unavailable: ", handle.Transport))
//	}
//	core.New().API().RegisterProtocol("mcp", factory)
type StreamFactory func(handle *DriveHandle) (Stream, error)

// API manages remote streams and protocol handlers.
//
//	c := core.New()
//	api := c.API()
//	_ = api.Protocols()
type API struct {
	core      *Core
	protocols *Registry[StreamFactory]
}

// API returns the remote communication primitive.
//
//	c.API().Stream("charon")
func (c *Core) API() *API {
	return c.api
}

// RegisterProtocol registers a stream factory for a URL scheme.
// Consumer packages call this during OnStartup.
//
//	c.API().RegisterProtocol("http", httpStreamFactory)
//	c.API().RegisterProtocol("https", httpStreamFactory)
//	c.API().RegisterProtocol("mcp", mcpStreamFactory)
func (a *API) RegisterProtocol(scheme string, factory StreamFactory) {
	a.protocols.Set(scheme, factory)
}

// Stream opens a connection to a named endpoint.
// Looks up the endpoint in Drive, extracts the protocol from the transport URL,
// and delegates to the registered protocol handler.
//
//	r := c.API().Stream("charon")
//	if r.OK { stream := r.Value.(Stream) }
func (a *API) Stream(name string) Result {
	r := a.core.Drive().Get(name)
	if !r.OK {
		return Result{E("api.Stream", Concat("endpoint not found in Drive: ", name), nil), false}
	}

	handle := r.Value.(*DriveHandle)
	scheme := extractScheme(handle.Transport)

	fr := a.protocols.Get(scheme)
	if !fr.OK {
		return Result{E("api.Stream", Concat("no protocol handler for scheme: ", scheme), nil), false}
	}

	factory := fr.Value.(StreamFactory)
	stream, err := factory(handle)
	if err != nil {
		return Result{err, false}
	}
	return Result{stream, true}
}

// Call invokes a named Action on a remote endpoint.
// This is the remote equivalent of c.Action("name").Run(ctx, opts).
//
//	r := c.API().Call("charon", "agentic.status", opts)
func (a *API) Call(endpoint string, action string, opts Options) Result {
	r := a.Stream(endpoint)
	if !r.OK {
		return r
	}

	stream := r.Value.(Stream)
	defer stream.Close()

	// Encode the action call as JSON-RPC (MCP compatible)
	payload := Concat(`{"action":"`, action, `","options":`, JSONMarshalString(opts), `}`)

	if err := stream.Send([]byte(payload)); err != nil {
		return Result{err, false}
	}

	response, err := stream.Receive()
	if err != nil {
		return Result{err, false}
	}

	return Result{string(response), true}
}

// Protocols returns all registered protocol scheme names.
//
//	c := core.New()
//	schemes := c.API().Protocols()
//	core.Println(core.Join(", ", schemes...))
func (a *API) Protocols() []string {
	return a.protocols.Names()
}

// extractScheme pulls the protocol from a transport URL.
// "http://host:port/path" → "http"
// "mcp://host:port" → "mcp"
func extractScheme(transport string) string {
	for i, c := range transport {
		if c == ':' {
			return transport[:i]
		}
	}
	return transport
}

// RemoteAction resolves "host:action.name" syntax for transparent remote dispatch.
// If the action name contains ":", the prefix is the endpoint and the suffix is the action.
//
//	c.Action("charon:agentic.status")  // → c.API().Call("charon", "agentic.status", opts)
func (c *Core) RemoteAction(name string, ctx Context, opts Options) Result {
	for i, ch := range name {
		if ch == ':' {
			endpoint := name[:i]
			action := name[i+1:]
			return c.API().Call(endpoint, action, opts)
		}
	}
	// No ":" — local action
	return c.Action(name).Run(ctx, opts)
}

// ---------------------------------------------------------------------
// HTTP — net/http aliases + Result-shape helpers
//
// Snider 2026-04-28: net/http folds into api.go alongside Stream so
// consumers reach HTTP types via core without importing net/http.
//
// Aliases keep type compatibility with stdlib (a *core.Request IS an
// *http.Request — pass it anywhere). Result-shape helpers wrap the
// common request/response patterns so consumers get core.E error
// formatting and Result.OK propagation without ad-hoc if-err checks.
// ---------------------------------------------------------------------

// Request is the canonical HTTP request, exported as core.Request.
//
//	r := core.NewHTTPRequest("GET", "https://api.lethean.example/health", nil)
//	if !r.OK {
//	    return r
//	}
//	req := r.Value.(*core.Request)
//	req.Header.Set("User-Agent", "dappcore-agent/1.0")
type Request = http.Request

// Response is the canonical HTTP response.
//
//	r := core.HTTPGet("https://api.lethean.example/health")
//	if !r.OK {
//	    return r
//	}
//	resp := r.Value.(*core.Response)
//	defer resp.Body.Close()
type Response = http.Response

// ResponseWriter is the canonical HTTP response-writer interface used
// by handlers.
//
//	handler := core.HandlerFunc(func(w core.ResponseWriter, r *core.Request) {
//	    core.WriteString(w, "ok\n")
//	})
//	_ = handler
type ResponseWriter = http.ResponseWriter

// Handler is the canonical HTTP handler interface.
//
//	var handler core.Handler = core.HandlerFunc(func(w core.ResponseWriter, r *core.Request) {
//	    core.WriteString(w, "ready\n")
//	})
//	_ = handler
type Handler = http.Handler

// HandlerFunc is the canonical HTTP handler-as-function adapter.
//
//	handler := core.HandlerFunc(func(w core.ResponseWriter, r *core.Request) {
//	    core.WriteString(w, "accepted\n")
//	})
//	core.NewHTTPTestServer(handler).Close()
type HandlerFunc = http.HandlerFunc

// ServeMux is the canonical HTTP request multiplexer.
//
//	mux := &core.ServeMux{}
//	mux.Handle("/health", core.HandlerFunc(func(w core.ResponseWriter, r *core.Request) {
//	    core.WriteString(w, "ok\n")
//	}))
type ServeMux = http.ServeMux

// HTTPServer is an HTTP server. Named HTTPServer (not Server) because
// core may grow other server types; this keeps the namespace clean.
//
//	srv := &core.HTTPServer{Addr: ":8080", Handler: &core.ServeMux{}}
//	go srv.ListenAndServe()
//	defer srv.Close()
type HTTPServer = http.Server

// HTTPClient is an HTTP client. Named HTTPClient for symmetry with
// HTTPServer.
//
//	client := &core.HTTPClient{}
//	resp, err := client.Get("https://api.lethean.example/health")
//	if err == nil {
//	    defer resp.Body.Close()
//	}
type HTTPClient = http.Client

// Header is the canonical HTTP header map.
//
//	headers := core.Header{}
//	headers.Set("Authorization", "Bearer lethean-token")
//	headers.Set("X-Agent", "codex")
type Header = http.Header

// Cookie is the canonical HTTP cookie value type.
//
//	cookie := &core.Cookie{Name: "session", Value: "agent-session", Path: "/"}
//	_ = cookie.String()
type Cookie = http.Cookie

// HTTPGet performs an HTTP GET. Returns Result wrapping *Response on
// success or the error.
//
//	r := core.HTTPGet("https://api.example.com/health")
//	if !r.OK { return r }
//	defer r.Value.(*Response).Body.Close()
func HTTPGet(url string) Result {
	resp, err := http.Get(url)
	if err != nil {
		return Result{err, false}
	}
	return Result{resp, true}
}

// HTTPPost performs an HTTP POST with the given content type and body.
//
//	r := core.HTTPPost(url, "application/json", body)
func HTTPPost(url, contentType string, body Reader) Result {
	resp, err := http.Post(url, contentType, body)
	if err != nil {
		return Result{err, false}
	}
	return Result{resp, true}
}

// HTTPPostForm performs an HTTP POST with form-encoded values.
//
//	r := core.HTTPPostForm(url, core.URLValues{"key": {"value"}})
func HTTPPostForm(target string, data URLValues) Result {
	resp, err := http.PostForm(target, data)
	if err != nil {
		return Result{err, false}
	}
	return Result{resp, true}
}

// NewHTTPRequest constructs an *http.Request with the given method, URL,
// and body. Returns Result wrapping the request.
//
//	r := core.NewHTTPRequest("GET", url, nil)
//	if !r.OK { return r }
//	req := r.Value.(*Request)
func NewHTTPRequest(method, target string, body Reader) Result {
	req, err := http.NewRequest(method, target, body)
	if err != nil {
		return Result{err, false}
	}
	return Result{req, true}
}

// NewHTTPRequestContext is NewHTTPRequest with a Context attached.
//
//	ctx := Background()
//	body := core.NewBufferString(`{"agent":"codex"}`)
//	r := core.NewHTTPRequestContext(ctx, "POST", "https://api.lethean.example/v1/tasks", body)
//	if !r.OK {
//	    return r
//	}
//	req := r.Value.(*core.Request)
//	req.Header.Set("Content-Type", "application/json")
func NewHTTPRequestContext(ctx Context, method, target string, body Reader) Result {
	req, err := http.NewRequestWithContext(ctx, method, target, body)
	if err != nil {
		return Result{err, false}
	}
	return Result{req, true}
}

// HTTPStatusText returns the canonical text for an HTTP status code,
// e.g. 200 → "OK".
//
//	status := core.HTTPStatusText(202)
//	core.Println(status)
func HTTPStatusText(code int) string {
	return http.StatusText(code)
}

// ---------------------------------------------------------------------
// mime/multipart aliases (#1101)
//
// Multipart parsing/writing rides alongside HTTP for upload boundaries.
// ---------------------------------------------------------------------

// MultipartReader streams a multipart body.
//
//	body := core.NewBufferString("--agent-boundary\r\n\r\n--agent-boundary--\r\n")
//	reader := core.NewMultipartReader(body, "agent-boundary")
//	if part, err := reader.NextPart(); err == nil {
//	    defer part.Close()
//	}
type MultipartReader = multipart.Reader

// MultipartWriter assembles a multipart body.
//
//	buf := core.NewBuffer()
//	writer := core.NewMultipartWriter(buf)
//	field, _ := writer.CreateFormField("agent")
//	core.WriteString(field, "codex")
//	writer.Close()
type MultipartWriter = multipart.Writer

// MultipartForm is a parsed multipart/form-data form.
//
//	form := &core.MultipartForm{Value: map[string][]string{"agent": {"codex"}}}
//	agents := form.Value["agent"]
//	core.Println(core.Join(", ", agents...))
type MultipartForm = multipart.Form

// MultipartFile is an open file from a multipart upload.
//
//	var file core.MultipartFile
//	if file != nil {
//	    defer file.Close()
//	}
type MultipartFile = multipart.File

// MultipartFileHeader is the header of a multipart upload file.
//
//	header := &core.MultipartFileHeader{Filename: "agent-report.json"}
//	core.Println(header.Filename)
type MultipartFileHeader = multipart.FileHeader

// NewMultipartReader returns a multipart reader for the given body and
// boundary.
//
//	r := core.NewMultipartReader(req.Body, boundary)
func NewMultipartReader(r Reader, boundary string) *MultipartReader {
	return multipart.NewReader(r, boundary)
}

// NewMultipartWriter returns a multipart writer that writes to w. The
// caller closes the writer to flush the trailing boundary.
//
//	w := core.NewMultipartWriter(buf)
//	w.WriteField("name", "value")
//	w.Close()
func NewMultipartWriter(w Writer) *MultipartWriter {
	return multipart.NewWriter(w)
}

// ---------------------------------------------------------------------
// net/http/httptest aliases (#1111)
//
// Test helpers fold here too — same domain, no good reason for a
// separate file.
// ---------------------------------------------------------------------

// HTTPTestServer is a test HTTP server, useful for exercising clients
// against a real listener without external dependencies.
//
//	server := core.NewHTTPTestServer(core.HandlerFunc(func(w core.ResponseWriter, r *core.Request) {
//	    core.WriteString(w, "ok\n")
//	}))
//	defer server.Close()
//	_ = core.HTTPGet(server.URL)
type HTTPTestServer = httptest.Server

// HTTPTestRecorder is an http.ResponseWriter that records the response
// for handler tests.
//
//	rec := core.NewHTTPTestRecorder()
//	req := core.NewHTTPTestRequest("GET", "/health", nil)
//	handler := core.HandlerFunc(func(w core.ResponseWriter, r *core.Request) { core.WriteString(w, "ok") })
//	handler.ServeHTTP(rec, req)
type HTTPTestRecorder = httptest.ResponseRecorder

// NewHTTPTestServer starts a new test server with the given handler.
// Caller must Close() the returned server.
//
//	srv := core.NewHTTPTestServer(handler)
//	defer srv.Close()
//	resp := core.HTTPGet(srv.URL)
func NewHTTPTestServer(handler Handler) *HTTPTestServer {
	return httptest.NewServer(handler)
}

// NewHTTPTestTLSServer starts a new test server using TLS.
//
//	server := core.NewHTTPTestTLSServer(core.HandlerFunc(func(w core.ResponseWriter, r *core.Request) {
//	    core.WriteString(w, "secure\n")
//	}))
//	defer server.Close()
func NewHTTPTestTLSServer(handler Handler) *HTTPTestServer {
	return httptest.NewTLSServer(handler)
}

// NewHTTPTestRecorder returns a fresh ResponseRecorder for handler tests.
//
//	rec := core.NewHTTPTestRecorder()
//	handler.ServeHTTP(rec, req)
//	if rec.Code != 200 { ... }
func NewHTTPTestRecorder() *HTTPTestRecorder {
	return httptest.NewRecorder()
}

// NewHTTPTestRequest constructs an *http.Request suitable for handler
// tests. Wraps httptest.NewRequest with Result-shape — though
// httptest.NewRequest itself never errors, the Result shape keeps the
// API uniform with NewHTTPRequest.
//
//	req := core.NewHTTPTestRequest("GET", "/path", nil)
func NewHTTPTestRequest(method, target string, body Reader) *Request {
	return httptest.NewRequest(method, target, body)
}
