package core_test

import (
	"context"

	. "dappco.re/go/core"
)

func ExampleAPI_RegisterProtocol() {
	c := New()
	c.API().RegisterProtocol("http", func(h *DriveHandle) (Stream, error) {
		return &mockStream{response: []byte("pong")}, nil
	})
	Println(c.API().Protocols())
	// Output: [http]
}

func ExampleAPI_Stream() {
	c := New()
	c.API().RegisterProtocol("http", func(h *DriveHandle) (Stream, error) {
		return &mockStream{response: []byte(Concat("connected to ", h.Name))}, nil
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
