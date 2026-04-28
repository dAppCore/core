package core_test

import . "dappco.re/go"

// --- mock stream for testing ---

type mockStream struct {
	sent     []byte
	response []byte
	closed   bool
	sendErr  error
	recvErr  error
}

func (s *mockStream) Send(data []byte) error {
	if s.sendErr != nil {
		return s.sendErr
	}
	s.sent = data
	return nil
}

func (s *mockStream) Receive() ([]byte, error) {
	if s.recvErr != nil {
		return nil, s.recvErr
	}
	return s.response, nil
}

func (s *mockStream) Close() error {
	s.closed = true
	return nil
}

func mockFactory(response string) StreamFactory {
	return func(handle *DriveHandle) (Stream, error) {
		return &mockStream{response: []byte(response)}, nil
	}
}

func mockStreamFactory(stream *mockStream) StreamFactory {
	return func(handle *DriveHandle) (Stream, error) {
		return stream, nil
	}
}

func mockFailingFactory(err error) StreamFactory {
	return func(handle *DriveHandle) (Stream, error) {
		return nil, err
	}
}

type failingAPIWriter struct{}

func (failingAPIWriter) Write(_ []byte) (int, error) {
	return 0, NewError("agent upload sink refused write")
}

// --- API ---

func TestApi_API_Good_Accessor(t *T) {
	c := New()
	AssertNotNil(t, c.API())
}

// --- RegisterProtocol ---

func TestApi_RegisterProtocol_Good(t *T) {
	c := New()
	c.API().RegisterProtocol("http", mockFactory("ok"))
	AssertContains(t, c.API().Protocols(), "http")
}

// --- Stream ---

func TestApi_Stream_Good(t *T) {
	c := New()
	c.API().RegisterProtocol("http", mockFactory("pong"))
	c.Drive().New(NewOptions(
		Option{Key: "name", Value: "charon"},
		Option{Key: "transport", Value: "http://10.69.69.165:9101/mcp"},
	))

	r := c.API().Stream("charon")
	AssertTrue(t, r.OK)

	stream := r.Value.(Stream)
	stream.Send([]byte("ping"))
	resp, err := stream.Receive()
	AssertNoError(t, err)
	AssertEqual(t, "pong", string(resp))
	stream.Close()
}

func TestApi_Stream_Bad_EndpointNotFound(t *T) {
	c := New()
	r := c.API().Stream("nonexistent")
	AssertFalse(t, r.OK)
}

func TestApi_Stream_Bad_NoProtocolHandler(t *T) {
	c := New()
	c.Drive().New(NewOptions(
		Option{Key: "name", Value: "unknown"},
		Option{Key: "transport", Value: "grpc://host:port"},
	))

	r := c.API().Stream("unknown")
	AssertFalse(t, r.OK)
}

// --- Call ---

func TestApi_Call_Good(t *T) {
	c := New()
	c.API().RegisterProtocol("http", mockFactory(`{"status":"ok"}`))
	c.Drive().New(NewOptions(
		Option{Key: "name", Value: "charon"},
		Option{Key: "transport", Value: "http://10.69.69.165:9101"},
	))

	r := c.API().Call("charon", "agentic.status", NewOptions())
	AssertTrue(t, r.OK)
	AssertContains(t, r.Value.(string), "ok")
}

func TestApi_Call_Bad_EndpointNotFound(t *T) {
	c := New()
	r := c.API().Call("missing", "action", NewOptions())
	AssertFalse(t, r.OK)
}

// --- RemoteAction ---

func TestApi_RemoteAction_Good_Local(t *T) {
	c := New()
	c.Action("local.action", func(_ Context, _ Options) Result {
		return Result{Value: "local", OK: true}
	})

	r := c.RemoteAction("local.action", Background(), NewOptions())
	AssertTrue(t, r.OK)
	AssertEqual(t, "local", r.Value)
}

func TestApi_RemoteAction_Good_Remote(t *T) {
	c := New()
	c.API().RegisterProtocol("http", mockFactory(`{"value":"remote"}`))
	c.Drive().New(NewOptions(
		Option{Key: "name", Value: "charon"},
		Option{Key: "transport", Value: "http://10.69.69.165:9101"},
	))

	r := c.RemoteAction("charon:agentic.status", Background(), NewOptions())
	AssertTrue(t, r.OK)
	AssertContains(t, r.Value.(string), "remote")
}

func TestApi_RemoteAction_Ugly_NoColon(t *T) {
	c := New()
	// No colon — falls through to local action (which doesn't exist)
	r := c.RemoteAction("nonexistent", Background(), NewOptions())
	AssertFalse(t, r.OK, "non-existent local action should fail")
}

// --- extractScheme ---

func TestApi_Ugly_SchemeExtraction(t *T) {
	c := New()
	// Verify scheme parsing works by registering different protocols
	c.API().RegisterProtocol("http", mockFactory("http"))
	c.API().RegisterProtocol("mcp", mockFactory("mcp"))
	c.API().RegisterProtocol("ws", mockFactory("ws"))

	AssertEqual(t, 3, len(c.API().Protocols()))
}

// --- AX-7 canonical triplets ---

func TestApi_Core_API_Good(t *T) {
	c := New()
	AssertNotNil(t, c.API())
	AssertEmpty(t, c.API().Protocols())
}

func TestApi_Core_API_Bad(t *T) {
	c := New()
	AssertSame(t, c.API(), c.API(), "API accessor is infallible and stable")
}

func TestApi_Core_API_Ugly(t *T) {
	c := New()
	c.API().RegisterProtocol("agent", mockFactory("ready"))
	AssertContains(t, c.API().Protocols(), "agent")
}

func TestApi_API_RegisterProtocol_Good(t *T) {
	c := New()
	c.API().RegisterProtocol("mcp", mockFactory("registered"))
	AssertContains(t, c.API().Protocols(), "mcp")
}

func TestApi_API_RegisterProtocol_Bad(t *T) {
	c := New()
	c.API().RegisterProtocol("", mockFactory("empty scheme"))
	AssertContains(t, c.API().Protocols(), "")
}

func TestApi_API_RegisterProtocol_Ugly(t *T) {
	c := New()
	c.API().RegisterProtocol("http", mockFactory("old"))
	c.API().RegisterProtocol("http", mockFactory("new"))
	c.Drive().New(NewOptions(
		Option{Key: "name", Value: "forge"},
		Option{Key: "transport", Value: "http://forge.local/rpc"},
	))

	r := c.API().Call("forge", "agent.dispatch", NewOptions())
	AssertTrue(t, r.OK)
	AssertEqual(t, "new", r.Value.(string))
}

func TestApi_API_Protocols_Good(t *T) {
	c := New()
	c.API().RegisterProtocol("http", mockFactory("http"))
	c.API().RegisterProtocol("mcp", mockFactory("mcp"))
	AssertEqual(t, []string{"http", "mcp"}, c.API().Protocols())
}

func TestApi_API_Protocols_Bad(t *T) {
	c := New()
	AssertEmpty(t, c.API().Protocols())
}

func TestApi_API_Protocols_Ugly(t *T) {
	c := New()
	c.API().RegisterProtocol("ssh", mockFactory("ssh"))
	protocols := c.API().Protocols()
	protocols[0] = "mutated"
	AssertEqual(t, []string{"ssh"}, c.API().Protocols())
}

func TestApi_API_Stream_Good(t *T) {
	c := New()
	c.API().RegisterProtocol("http", mockFactory("pong"))
	c.Drive().New(NewOptions(
		Option{Key: "name", Value: "homelab"},
		Option{Key: "transport", Value: "http://10.69.69.165:9101/mcp"},
	))

	r := c.API().Stream("homelab")
	AssertTrue(t, r.OK)
	AssertNotNil(t, r.Value.(Stream))
}

func TestApi_API_Stream_Bad(t *T) {
	c := New()
	r := c.API().Stream("missing-endpoint")
	AssertFalse(t, r.OK)
}

func TestApi_API_Stream_Ugly(t *T) {
	c := New()
	c.API().RegisterProtocol("mcp", mockFailingFactory(NewError("mcp handshake refused")))
	c.Drive().New(NewOptions(
		Option{Key: "name", Value: "agent"},
		Option{Key: "transport", Value: "mcp://agent.local"},
	))

	r := c.API().Stream("agent")
	AssertFalse(t, r.OK)
	AssertContains(t, r.Error(), "handshake")
}

func TestApi_API_Call_Good(t *T) {
	c := New()
	c.API().RegisterProtocol("http", mockFactory(`{"status":"ready"}`))
	c.Drive().New(NewOptions(
		Option{Key: "name", Value: "dispatch"},
		Option{Key: "transport", Value: "http://dispatch.local/rpc"},
	))

	r := c.API().Call("dispatch", "agent.status", NewOptions(Option{Key: "agent", Value: "codex"}))
	AssertTrue(t, r.OK)
	AssertContains(t, r.Value.(string), "ready")
}

func TestApi_API_Call_Bad(t *T) {
	stream := &mockStream{sendErr: NewError("session token rejected")}
	c := New()
	c.API().RegisterProtocol("http", mockStreamFactory(stream))
	c.Drive().New(NewOptions(
		Option{Key: "name", Value: "dispatch"},
		Option{Key: "transport", Value: "http://dispatch.local/rpc"},
	))

	r := c.API().Call("dispatch", "agent.status", NewOptions())
	AssertFalse(t, r.OK)
	AssertTrue(t, stream.closed)
}

func TestApi_API_Call_Ugly(t *T) {
	stream := &mockStream{response: []byte("{}")}
	c := New()
	c.API().RegisterProtocol("", mockStreamFactory(stream))
	c.Drive().New(NewOptions(
		Option{Key: "name", Value: "loopback"},
		Option{Key: "transport", Value: ""},
	))

	r := c.API().Call("loopback", "", NewOptions())
	AssertTrue(t, r.OK)
	AssertContains(t, string(stream.sent), `"action":""`)
	AssertTrue(t, stream.closed)
}

func TestApi_Core_RemoteAction_Good(t *T) {
	c := New()
	c.API().RegisterProtocol("http", mockFactory(`{"value":"remote"}`))
	c.Drive().New(NewOptions(
		Option{Key: "name", Value: "charon"},
		Option{Key: "transport", Value: "http://charon.local/rpc"},
	))

	r := c.RemoteAction("charon:agent.health", Background(), NewOptions())
	AssertTrue(t, r.OK)
	AssertContains(t, r.Value.(string), "remote")
}

func TestApi_Core_RemoteAction_Bad(t *T) {
	c := New()
	r := c.RemoteAction("missing:agent.health", Background(), NewOptions())
	AssertFalse(t, r.OK)
}

func TestApi_Core_RemoteAction_Ugly(t *T) {
	c := New()
	c.Action("agent.local", func(_ Context, _ Options) Result {
		return Result{Value: "local dispatch", OK: true}
	})

	r := c.RemoteAction("agent.local", Background(), NewOptions())
	AssertTrue(t, r.OK)
	AssertEqual(t, "local dispatch", r.Value)
}

func TestApi_HTTPGet_Good(t *T) {
	srv := NewHTTPTestServer(HandlerFunc(func(w ResponseWriter, r *Request) {
		WriteString(w, "homelab ready")
	}))
	defer srv.Close()

	r := HTTPGet(srv.URL)
	AssertTrue(t, r.OK)
	resp := r.Value.(*Response)
	AssertEqual(t, 200, resp.StatusCode)
	body := ReadAll(resp.Body)
	AssertTrue(t, body.OK)
	AssertEqual(t, "homelab ready", body.Value.(string))
}

func TestApi_HTTPGet_Bad(t *T) {
	r := HTTPGet("://missing-scheme")
	AssertFalse(t, r.OK)
}

func TestApi_HTTPGet_Ugly(t *T) {
	srv := NewHTTPTestServer(HandlerFunc(func(w ResponseWriter, r *Request) {
		w.WriteHeader(204)
	}))
	defer srv.Close()

	r := HTTPGet(srv.URL)
	AssertTrue(t, r.OK)
	resp := r.Value.(*Response)
	AssertEqual(t, 204, resp.StatusCode)
	body := ReadAll(resp.Body)
	AssertTrue(t, body.OK)
	AssertEqual(t, "", body.Value.(string))
}

func TestApi_HTTPPost_Good(t *T) {
	srv := NewHTTPTestServer(HandlerFunc(func(w ResponseWriter, r *Request) {
		body := ReadAll(r.Body)
		AssertTrue(t, body.OK)
		AssertEqual(t, "POST", r.Method)
		AssertEqual(t, "application/json", r.Header.Get("Content-Type"))
		WriteString(w, body.Value.(string))
	}))
	defer srv.Close()

	r := HTTPPost(srv.URL, "application/json", NewReader(`{"agent":"codex"}`))
	AssertTrue(t, r.OK)
	body := ReadAll(r.Value.(*Response).Body)
	AssertTrue(t, body.OK)
	AssertEqual(t, `{"agent":"codex"}`, body.Value.(string))
}

func TestApi_HTTPPost_Bad(t *T) {
	r := HTTPPost("://missing-scheme", "application/json", NewReader("{}"))
	AssertFalse(t, r.OK)
}

func TestApi_HTTPPost_Ugly(t *T) {
	srv := NewHTTPTestServer(HandlerFunc(func(w ResponseWriter, r *Request) {
		body := ReadAll(r.Body)
		AssertTrue(t, body.OK)
		AssertEqual(t, "", body.Value.(string))
		w.WriteHeader(202)
	}))
	defer srv.Close()

	r := HTTPPost(srv.URL, "text/plain", nil)
	AssertTrue(t, r.OK)
	AssertEqual(t, 202, r.Value.(*Response).StatusCode)
	ReadAll(r.Value.(*Response).Body)
}

func TestApi_HTTPPostForm_Good(t *T) {
	srv := NewHTTPTestServer(HandlerFunc(func(w ResponseWriter, r *Request) {
		AssertNoError(t, r.ParseForm())
		AssertEqual(t, "session-token", r.Form.Get("token"))
		WriteString(w, "accepted")
	}))
	defer srv.Close()

	r := HTTPPostForm(srv.URL, URLValues{"token": []string{"session-token"}})
	AssertTrue(t, r.OK)
	body := ReadAll(r.Value.(*Response).Body)
	AssertTrue(t, body.OK)
	AssertEqual(t, "accepted", body.Value.(string))
}

func TestApi_HTTPPostForm_Bad(t *T) {
	r := HTTPPostForm("://missing-scheme", URLValues{"token": []string{"session-token"}})
	AssertFalse(t, r.OK)
}

func TestApi_HTTPPostForm_Ugly(t *T) {
	srv := NewHTTPTestServer(HandlerFunc(func(w ResponseWriter, r *Request) {
		AssertNoError(t, r.ParseForm())
		AssertEqual(t, "", r.Form.Get("token"))
		w.WriteHeader(204)
	}))
	defer srv.Close()

	r := HTTPPostForm(srv.URL, URLValues{})
	AssertTrue(t, r.OK)
	AssertEqual(t, 204, r.Value.(*Response).StatusCode)
	ReadAll(r.Value.(*Response).Body)
}

func TestApi_NewHTTPRequest_Good(t *T) {
	r := NewHTTPRequest("POST", "https://api.lethean.example/v1/tasks", NewReader("{}"))
	AssertTrue(t, r.OK)
	req := r.Value.(*Request)
	AssertEqual(t, "POST", req.Method)
	AssertEqual(t, "api.lethean.example", req.URL.Host)
}

func TestApi_NewHTTPRequest_Bad(t *T) {
	r := NewHTTPRequest("GET", "http://[::1", nil)
	AssertFalse(t, r.OK)
}

func TestApi_NewHTTPRequest_Ugly(t *T) {
	r := NewHTTPRequest("", "https://api.lethean.example/health", nil)
	AssertTrue(t, r.OK)
	AssertEqual(t, "GET", r.Value.(*Request).Method)
}

func TestApi_NewHTTPRequestContext_Good(t *T) {
	ctx := WithValue(Background(), "request-id", "agent-dispatch-42")
	r := NewHTTPRequestContext(ctx, "GET", "https://api.lethean.example/health", nil)
	AssertTrue(t, r.OK)
	AssertEqual(t, "agent-dispatch-42", r.Value.(*Request).Context().Value("request-id"))
}

func TestApi_NewHTTPRequestContext_Bad(t *T) {
	r := NewHTTPRequestContext(Background(), "GET", "http://[::1", nil)
	AssertFalse(t, r.OK)
}

func TestApi_NewHTTPRequestContext_Ugly(t *T) {
	ctx, cancel := WithCancel(Background())
	cancel()
	r := NewHTTPRequestContext(ctx, "GET", "https://api.lethean.example/health", nil)
	AssertTrue(t, r.OK)
	AssertError(t, r.Value.(*Request).Context().Err())
}

func TestApi_HTTPStatusText_Good(t *T) {
	AssertEqual(t, "Accepted", HTTPStatusText(202))
}

func TestApi_HTTPStatusText_Bad(t *T) {
	AssertEqual(t, "", HTTPStatusText(799))
}

func TestApi_HTTPStatusText_Ugly(t *T) {
	AssertEqual(t, "", HTTPStatusText(0))
}

func TestApi_NewHTTPTestServer_Good(t *T) {
	srv := NewHTTPTestServer(HandlerFunc(func(w ResponseWriter, r *Request) {
		WriteString(w, "ok")
	}))
	defer srv.Close()

	r := HTTPGet(srv.URL)
	AssertTrue(t, r.OK)
	body := ReadAll(r.Value.(*Response).Body)
	AssertTrue(t, body.OK)
	AssertEqual(t, "ok", body.Value.(string))
}

func TestApi_NewHTTPTestServer_Bad(t *T) {
	srv := NewHTTPTestServer(HandlerFunc(func(w ResponseWriter, r *Request) {
		w.WriteHeader(503)
	}))
	defer srv.Close()

	r := HTTPGet(srv.URL)
	AssertTrue(t, r.OK)
	AssertEqual(t, 503, r.Value.(*Response).StatusCode)
	ReadAll(r.Value.(*Response).Body)
}

func TestApi_NewHTTPTestServer_Ugly(t *T) {
	srv := NewHTTPTestServer(HandlerFunc(func(w ResponseWriter, r *Request) { /* no-op handler exercises empty response lifecycle */ }))
	srv.Close()
	AssertNotEmpty(t, srv.URL)
}

func TestApi_NewHTTPTestTLSServer_Good(t *T) {
	srv := NewHTTPTestTLSServer(HandlerFunc(func(w ResponseWriter, r *Request) {
		WriteString(w, "secure")
	}))
	defer srv.Close()

	resp, err := srv.Client().Get(srv.URL)
	AssertNoError(t, err)
	body := ReadAll(resp.Body)
	AssertTrue(t, body.OK)
	AssertEqual(t, "secure", body.Value.(string))
}

func TestApi_NewHTTPTestTLSServer_Bad(t *T) {
	srv := NewHTTPTestTLSServer(HandlerFunc(func(w ResponseWriter, r *Request) {
		WriteString(w, "secure")
	}))
	defer srv.Close()

	r := HTTPGet(srv.URL)
	AssertFalse(t, r.OK)
}

func TestApi_NewHTTPTestTLSServer_Ugly(t *T) {
	srv := NewHTTPTestTLSServer(HandlerFunc(func(w ResponseWriter, r *Request) { /* no-op handler exercises empty response lifecycle */ }))
	srv.Close()
	AssertNotEmpty(t, srv.URL)
}

func TestApi_NewHTTPTestRecorder_Good(t *T) {
	rec := NewHTTPTestRecorder()
	req := NewHTTPTestRequest("GET", "/health", nil)
	HandlerFunc(func(w ResponseWriter, r *Request) {
		w.WriteHeader(202)
		WriteString(w, "queued")
	}).ServeHTTP(rec, req)

	AssertEqual(t, 202, rec.Code)
	AssertContains(t, rec.Body.String(), "queued")
}

func TestApi_NewHTTPTestRecorder_Bad(t *T) {
	rec := NewHTTPTestRecorder()
	HandlerFunc(func(w ResponseWriter, r *Request) {
		w.WriteHeader(500)
	}).ServeHTTP(rec, NewHTTPTestRequest("GET", "/health", nil))
	AssertEqual(t, 500, rec.Code)
}

func TestApi_NewHTTPTestRecorder_Ugly(t *T) {
	rec := NewHTTPTestRecorder()
	AssertEqual(t, 200, rec.Code)
	AssertEqual(t, "", rec.Body.String())
}

func TestApi_NewHTTPTestRequest_Good(t *T) {
	req := NewHTTPTestRequest("POST", "/agent/dispatch", NewReader("payload"))
	AssertEqual(t, "POST", req.Method)
	AssertEqual(t, "/agent/dispatch", req.URL.Path)
}

func TestApi_NewHTTPTestRequest_Bad(t *T) {
	AssertPanics(t, func() {
		NewHTTPTestRequest("GET", "http://[::1", nil)
	})
}

func TestApi_NewHTTPTestRequest_Ugly(t *T) {
	req := NewHTTPTestRequest("GET", "https://api.lethean.example/health?deep=1", nil)
	AssertEqual(t, "api.lethean.example", req.Host)
	AssertEqual(t, "1", req.URL.Query().Get("deep"))
}

func TestApi_NewMultipartReader_Good(t *T) {
	buf := NewBuffer()
	writer := NewMultipartWriter(buf)
	AssertNoError(t, writer.WriteField("agent", "codex"))
	AssertNoError(t, writer.Close())

	reader := NewMultipartReader(buf, writer.Boundary())
	part, err := reader.NextPart()
	AssertNoError(t, err)
	defer part.Close()
	body := ReadAll(part)
	AssertTrue(t, body.OK)
	AssertEqual(t, "agent", part.FormName())
	AssertEqual(t, "codex", body.Value.(string))
}

func TestApi_NewMultipartReader_Bad(t *T) {
	reader := NewMultipartReader(NewReader("not multipart"), "agent-boundary")
	_, err := reader.NextPart()
	AssertError(t, err)
}

func TestApi_NewMultipartReader_Ugly(t *T) {
	reader := NewMultipartReader(NewReader(""), "")
	_, err := reader.NextPart()
	AssertError(t, err)
}

func TestApi_NewMultipartWriter_Good(t *T) {
	buf := NewBuffer()
	writer := NewMultipartWriter(buf)
	AssertNoError(t, writer.WriteField("agent", "codex"))
	AssertNoError(t, writer.Close())
	AssertContains(t, buf.String(), "codex")
}

func TestApi_NewMultipartWriter_Bad(t *T) {
	writer := NewMultipartWriter(failingAPIWriter{})
	err := writer.WriteField("agent", "codex")
	AssertError(t, err)
}

func TestApi_NewMultipartWriter_Ugly(t *T) {
	buf := NewBuffer()
	writer := NewMultipartWriter(buf)
	AssertNoError(t, writer.Close())
	AssertContains(t, buf.String(), writer.Boundary())
}
