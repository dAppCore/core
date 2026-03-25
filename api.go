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

import "context"

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
type StreamFactory func(handle *DriveHandle) (Stream, error)

// API manages remote streams and protocol handlers.
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
	payload := Concat(`{"action":"`, action, `","options":`, optionsToJSON(opts), `}`)

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

// optionsToJSON is a minimal JSON serialiser for Options.
// core/go stays stdlib-only — no encoding/json import.
func optionsToJSON(opts Options) string {
	b := NewBuilder()
	b.WriteString("{")
	first := true
	for i := 0; ; i++ {
		r := opts.Get(Sprintf("_key_%d", i))
		if !r.OK {
			break
		}
		// This is a placeholder — real implementation needs proper iteration
		_ = first
		first = false
	}
	// Simple fallback: serialize known keys
	b.WriteString("}")
	return b.String()
}

// RemoteAction resolves "host:action.name" syntax for transparent remote dispatch.
// If the action name contains ":", the prefix is the endpoint and the suffix is the action.
//
//	c.Action("charon:agentic.status")  // → c.API().Call("charon", "agentic.status", opts)
func (c *Core) RemoteAction(name string, ctx context.Context, opts Options) Result {
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
