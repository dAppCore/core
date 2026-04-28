package core_test

import (
	"context"

	. "dappco.re/go/core"
)

// --- mock stream for testing ---

type mockStream struct {
	sent     []byte
	response []byte
	closed   bool
}

func (s *mockStream) Send(data []byte) error {
	s.sent = data
	return nil
}

func (s *mockStream) Receive() ([]byte, error) {
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
	c.Action("local.action", func(_ context.Context, _ Options) Result {
		return Result{Value: "local", OK: true}
	})

	r := c.RemoteAction("local.action", context.Background(), NewOptions())
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

	r := c.RemoteAction("charon:agentic.status", context.Background(), NewOptions())
	AssertTrue(t, r.OK)
	AssertContains(t, r.Value.(string), "remote")
}

func TestApi_RemoteAction_Ugly_NoColon(t *T) {
	c := New()
	// No colon — falls through to local action (which doesn't exist)
	r := c.RemoteAction("nonexistent", context.Background(), NewOptions())
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
