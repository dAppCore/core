// SPDX-License-Identifier: EUPL-1.2

// Network primitives for the Core framework.
//
// Re-exports stdlib net types and provides Result-shape constructors
// for the common Dial / Listen / ParseIP patterns. Consumer packages
// reach IP/Conn/Listener types via core without importing net directly.
//
// Higher-level HTTP machinery lives alongside Stream in api.go (see
// HTTPGet / HTTPServer / HTTPClient). What's here is the lower layer
// — TCP/UDP/Unix sockets, IP parsing, and the connection interfaces
// HTTP and other protocols build on top of.
//
// Usage:
//
//	addr := core.ParseIP("192.0.2.1")
//	if addr == nil { return core.E("bad", "invalid IP", nil) }
//
//	conn := core.NetDial("tcp", "10.0.0.1:8080")
//	if !conn.OK { return conn }
//	defer conn.Value.(core.Conn).Close()
//
//	a, b := core.NetPipe()  // in-memory test connection pair
package core

import (
	"net"
	"time"
)

// Conn is the canonical network connection interface.
type Conn = net.Conn

// Listener accepts inbound connections.
type Listener = net.Listener

// PacketConn is a connectionless packet-oriented connection (UDP, etc.).
type PacketConn = net.PacketConn

// Addr is a network endpoint address.
type Addr = net.Addr

// TCPConn is a connected TCP socket.
type TCPConn = net.TCPConn

// TCPListener accepts inbound TCP connections.
type TCPListener = net.TCPListener

// TCPAddr is a TCP endpoint address.
type TCPAddr = net.TCPAddr

// UDPConn is a UDP packet connection.
type UDPConn = net.UDPConn

// UDPAddr is a UDP endpoint address.
type UDPAddr = net.UDPAddr

// UnixConn is a Unix-domain socket connection.
type UnixConn = net.UnixConn

// UnixListener accepts inbound Unix-domain connections.
type UnixListener = net.UnixListener

// UnixAddr is a Unix-domain socket address.
type UnixAddr = net.UnixAddr

// IP is a single IP address.
type IP = net.IP

// IPNet is an IP network (address + mask).
type IPNet = net.IPNet

// IPMask is the mask portion of an IPNet.
type IPMask = net.IPMask

// Dialer holds connection-establishment options.
type Dialer = net.Dialer

// Resolver looks up host names.
type Resolver = net.Resolver

// ParseIP parses an IPv4 or IPv6 textual address. Returns nil on
// malformed input — same shape as net.ParseIP.
//
//	ip := core.ParseIP("2001:db8::1")
func ParseIP(s string) IP {
	return net.ParseIP(s)
}

// ParseCIDR parses a CIDR notation IP/mask. Returns Result wrapping
// (IP, *IPNet) on success.
//
//	r := core.ParseCIDR("10.0.0.0/24")
func ParseCIDR(s string) Result {
	ip, ipnet, err := net.ParseCIDR(s)
	if err != nil {
		return Result{err, false}
	}
	return Result{[]any{ip, ipnet}, true}
}

// NetDial opens a connection to the given network/address.
//
//	r := core.NetDial("tcp", "127.0.0.1:8080")
//	if !r.OK { return r }
//	conn := r.Value.(Conn)
func NetDial(network, address string) Result {
	c, err := net.Dial(network, address)
	if err != nil {
		return Result{err, false}
	}
	return Result{c, true}
}

// NetDialTimeout opens a connection with a deadline.
//
//	r := core.NetDialTimeout("tcp", "host:port", 5*time.Second)
func NetDialTimeout(network, address string, timeout time.Duration) Result {
	c, err := net.DialTimeout(network, address, timeout)
	if err != nil {
		return Result{err, false}
	}
	return Result{c, true}
}

// NetListen binds a listener on the given network/address.
//
//	r := core.NetListen("tcp", ":0")
//	if !r.OK { return r }
//	ln := r.Value.(Listener)
func NetListen(network, address string) Result {
	ln, err := net.Listen(network, address)
	if err != nil {
		return Result{err, false}
	}
	return Result{ln, true}
}

// NetListenPacket binds a packet-oriented listener (UDP etc.).
func NetListenPacket(network, address string) Result {
	pc, err := net.ListenPacket(network, address)
	if err != nil {
		return Result{err, false}
	}
	return Result{pc, true}
}

// NetPipe returns a connected, in-memory, full-duplex Conn pair. Useful
// for testing protocols against a real Conn without a real socket.
//
//	a, b := core.NetPipe()
//	defer a.Close(); defer b.Close()
func NetPipe() (Conn, Conn) {
	return net.Pipe()
}
