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

func ExampleAPI_RegisterProtocol() {
	c := New()
	c.API().RegisterProtocol("http", func(h *DriveHandle) (Stream, error) {
		return &exampleStream{response: []byte("pong")}, nil
	})
	Println(c.API().Protocols())
	// Output: [http]
}

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

func ExampleAPI_Protocols() {
	c := New()
	c.API().RegisterProtocol("http", func(_ *DriveHandle) (Stream, error) {
		return &exampleStream{}, nil
	})
	Println(c.API().Protocols())
	// Output: [http]
}

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

func ExampleNewHTTPRequest() {
	r := NewHTTPRequest("POST", "https://example.com/deploy", NewReader("payload"))
	req := r.Value.(*Request)
	Println(req.Method)
	Println(req.URL.Path)
	// Output:
	// POST
	// /deploy
}

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

func ExampleHTTPStatusText() {
	Println(HTTPStatusText(201))
	// Output: Created
}

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

func ExampleNewHTTPTestServer() {
	srv := NewHTTPTestServer(HandlerFunc(func(w ResponseWriter, _ *Request) {
		WriteString(w, "ok")
	}))
	defer srv.Close()

	Println(HasPrefix(srv.URL, "http://"))
	// Output: true
}

func ExampleNewHTTPTestTLSServer() {
	srv := NewHTTPTestTLSServer(HandlerFunc(func(w ResponseWriter, _ *Request) {
		WriteString(w, "ok")
	}))
	defer srv.Close()

	Println(HasPrefix(srv.URL, "https://"))
	// Output: true
}

func ExampleNewHTTPTestRecorder() {
	rec := NewHTTPTestRecorder()
	rec.WriteHeader(202)
	Println(rec.Code)
	// Output: 202
}

func ExampleNewHTTPTestRequest() {
	req := NewHTTPTestRequest("GET", "/status", nil)
	Println(req.Method)
	Println(req.URL.Path)
	// Output:
	// GET
	// /status
}
