package core_test

import (
	"context"

	. "dappco.re/go"
)

type exampleStream struct {
	response []byte
	sent     []byte
}

func (s *exampleStream) Send(data []byte) error {
	s.sent = data
	return nil
}

func (s *exampleStream) Receive() ([]byte, error) {
	return s.response, nil
}

func (s *exampleStream) Close() error {
	return nil
}

// ExampleAPI_RegisterProtocol registers a transport protocol through
// `API.RegisterProtocol` for a Lethean drive integration. Transport details stay behind
// the API wrapper while callers exchange drives, streams, and Results.
func ExampleAPI_RegisterProtocol() {
	c := New()
	c.API().RegisterProtocol("http", func(h *DriveHandle) (Stream, error) {
		return &exampleStream{response: []byte("pong")}, nil
	})
	Println(c.API().Protocols())
	// Output: [http]
}

// ExampleAPI_Stream opens a stream through `API.Stream` for a Lethean drive integration.
// Transport details stay behind the API wrapper while callers exchange drives, streams,
// and Results.
func ExampleAPI_Stream() {
	c := New()
	c.API().RegisterProtocol("http", func(h *DriveHandle) (Stream, error) {
		return &exampleStream{response: []byte(Concat("connected to ", h.Name))}, nil
	})
	c.Drive().New(NewOptions(
		Option{Key: "name", Value: "charon"},
		Option{Key: "transport", Value: "http://10.69.69.165:9101"},
	))

	r := c.API().Stream("charon")
	if r.OK {
		stream := r.Value.(Stream)
		resp, _ := stream.Receive()
		Println(string(resp))
		stream.Close()
	}
	// Output: connected to charon
}

// ExampleAPI_Call calls a remote method through `API.Call` for a Lethean drive
// integration. Transport details stay behind the API wrapper while callers exchange
// drives, streams, and Results.
func ExampleAPI_Call() {
	c := New()
	c.API().RegisterProtocol("http", func(_ *DriveHandle) (Stream, error) {
		return &exampleStream{response: []byte(`{"ok":true}`)}, nil
	})
	c.Drive().New(NewOptions(
		Option{Key: "name", Value: "charon"},
		Option{Key: "transport", Value: "http://10.69.69.165:9101"},
	))

	r := c.API().Call("charon", "agent.status", NewOptions())
	Println(r.Value)
	// Output: {"ok":true}
}

// ExampleAPI_Protocols lists transport protocols through `API.Protocols` for a Lethean
// drive integration. Transport details stay behind the API wrapper while callers exchange
// drives, streams, and Results.
func ExampleAPI_Protocols() {
	c := New()
	c.API().RegisterProtocol("http", func(_ *DriveHandle) (Stream, error) {
		return &exampleStream{}, nil
	})
	Println(c.API().Protocols())
	// Output: [http]
}

// ExampleCore_RemoteAction resolves an action name locally before a remote drive prefix is
// needed. Transport details stay behind the API wrapper while callers exchange drives,
// streams, and Results.
func ExampleCore_RemoteAction() {
	c := New()
	// Local action
	c.Action("status", func(_ context.Context, _ Options) Result {
		return Result{Value: "running", OK: true}
	})

	// No colon — resolves locally
	r := c.RemoteAction("status", context.Background(), NewOptions())
	Println(r.Value)
	// Output: running
}

// ExampleHTTPGet fetches a local health endpoint through the core HTTP client wrapper for
// a Lethean drive integration. Transport details stay behind the API wrapper while callers
// exchange drives, streams, and Results.
func ExampleHTTPGet() {
	srv := NewHTTPTestServer(HandlerFunc(func(w ResponseWriter, _ *Request) {
		WriteString(w, "ok")
	}))
	defer srv.Close()

	r := HTTPGet(srv.URL)
	defer r.Value.(*Response).Body.Close()
	body := ReadAll(r.Value.(*Response).Body)
	Println(body.Value)
	// Output: ok
}

// ExampleHTTPPost sends a reader-backed payload through the core HTTP client wrapper for a
// Lethean drive integration. Transport details stay behind the API wrapper while callers
// exchange drives, streams, and Results.
func ExampleHTTPPost() {
	srv := NewHTTPTestServer(HandlerFunc(func(w ResponseWriter, _ *Request) {
		WriteString(w, "created")
	}))
	defer srv.Close()

	r := HTTPPost(srv.URL, "text/plain", NewReader("payload"))
	defer r.Value.(*Response).Body.Close()
	body := ReadAll(r.Value.(*Response).Body)
	Println(body.Value)
	// Output: created
}

// ExampleHTTPPostForm submits form data through the core HTTP client wrapper for a Lethean
// drive integration. Transport details stay behind the API wrapper while callers exchange
// drives, streams, and Results.
func ExampleHTTPPostForm() {
	srv := NewHTTPTestServer(HandlerFunc(func(w ResponseWriter, _ *Request) {
		WriteString(w, "submitted")
	}))
	defer srv.Close()

	r := HTTPPostForm(srv.URL, nil)
	defer r.Value.(*Response).Body.Close()
	body := ReadAll(r.Value.(*Response).Body)
	Println(body.Value)
	// Output: submitted
}

// ExampleNewHTTPRequest builds a POST request with a payload for a deployment endpoint.
// Transport details stay behind the API wrapper while callers exchange drives, streams,
// and Results.
func ExampleNewHTTPRequest() {
	r := NewHTTPRequest("POST", "https://example.com/deploy", NewReader("payload"))
	req := r.Value.(*Request)
	Println(req.Method)
	Println(req.URL.Path)
	// Output:
	// POST
	// /deploy
}

// ExampleNewHTTPRequestContext builds a request bound to the active Core context for a
// status endpoint. Transport details stay behind the API wrapper while callers exchange
// drives, streams, and Results.
func ExampleNewHTTPRequestContext() {
	ctx := New().Context()
	r := NewHTTPRequestContext(ctx, "GET", "https://example.com/status", nil)
	req := r.Value.(*Request)
	Println(req.Context() == ctx)
	Println(req.Method)
	// Output:
	// true
	// GET
}

// ExampleHTTPStatusText reads status reason text through `HTTPStatusText` for a Lethean
// drive integration. Transport details stay behind the API wrapper while callers exchange
// drives, streams, and Results.
func ExampleHTTPStatusText() {
	Println(HTTPStatusText(201))
	// Output: Created
}

// ExampleNewMultipartWriter creates multipart form data through `NewMultipartWriter` for a
// Lethean drive integration. Transport details stay behind the API wrapper while callers
// exchange drives, streams, and Results.
func ExampleNewMultipartWriter() {
	buf := NewBuffer()
	writer := NewMultipartWriter(buf)
	writer.WriteField("name", "codex")
	boundary := writer.Boundary()
	writer.Close()

	Println(boundary != "")
	Println(Contains(buf.String(), "codex"))
	// Output:
	// true
	// true
}

// ExampleNewMultipartReader reads multipart form data through `NewMultipartReader` for a
// Lethean drive integration. Transport details stay behind the API wrapper while callers
// exchange drives, streams, and Results.
func ExampleNewMultipartReader() {
	buf := NewBuffer()
	writer := NewMultipartWriter(buf)
	writer.WriteField("name", "codex")
	boundary := writer.Boundary()
	writer.Close()

	reader := NewMultipartReader(buf, boundary)
	part, _ := reader.NextPart()
	data := ReadAll(part)
	Println(part.FormName())
	Println(data.Value)
	// Output:
	// name
	// codex
}

// ExampleNewHTTPTestServer creates an HTTP fixture server for handler validation.
// Transport details stay behind the API wrapper while callers exchange drives, streams,
// and Results.
func ExampleNewHTTPTestServer() {
	srv := NewHTTPTestServer(HandlerFunc(func(w ResponseWriter, _ *Request) {
		WriteString(w, "ok")
	}))
	defer srv.Close()

	Println(HasPrefix(srv.URL, "http://"))
	// Output: true
}

// ExampleNewHTTPTestTLSServer creates a TLS fixture server for HTTPS handler validation.
// Transport details stay behind the API wrapper while callers exchange drives, streams,
// and Results.
func ExampleNewHTTPTestTLSServer() {
	srv := NewHTTPTestTLSServer(HandlerFunc(func(w ResponseWriter, _ *Request) {
		WriteString(w, "ok")
	}))
	defer srv.Close()

	Println(HasPrefix(srv.URL, "https://"))
	// Output: true
}

// ExampleNewHTTPTestRecorder records response status for handler validation. Transport
// details stay behind the API wrapper while callers exchange drives, streams, and Results.
func ExampleNewHTTPTestRecorder() {
	rec := NewHTTPTestRecorder()
	rec.WriteHeader(202)
	Println(rec.Code)
	// Output: 202
}

// ExampleNewHTTPTestRequest builds a request fixture for a status route. Transport details
// stay behind the API wrapper while callers exchange drives, streams, and Results.
func ExampleNewHTTPTestRequest() {
	req := NewHTTPTestRequest("GET", "/status", nil)
	Println(req.Method)
	Println(req.URL.Path)
	// Output:
	// GET
	// /status
}
