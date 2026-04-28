package core_test

import . "dappco.re/go"

// ExampleParseIP parses an IP address through `ParseIP` for network health checks. Network
// primitives are reached through core wrappers and Result-shaped calls.
func ExampleParseIP() {
	ip := ParseIP("192.0.2.10")
	Println(ip.String())
	// Output: 192.0.2.10
}

// ExampleParseCIDR parses a CIDR range through `ParseCIDR` for network health checks.
// Network primitives are reached through core wrappers and Result-shaped calls.
func ExampleParseCIDR() {
	r := ParseCIDR("192.0.2.0/24")
	parts := r.Value.([]any)
	Println(parts[0])
	Println(parts[1])
	// Output:
	// 192.0.2.0
	// 192.0.2.0/24
}

// ExampleNetDial dials a network address through `NetDial` for network health checks.
// Network primitives are reached through core wrappers and Result-shaped calls.
func ExampleNetDial() {
	ln := NetListen("tcp", "127.0.0.1:0").Value.(Listener)
	defer ln.Close()

	r := NetDial("tcp", ln.Addr().String())
	Println(r.OK)
	if r.OK {
		r.Value.(Conn).Close()
	}
	// Output: true
}

// ExampleNetDialTimeout dials a network address with a timeout through `NetDialTimeout`
// for network health checks. Network primitives are reached through core wrappers and
// Result-shaped calls.
func ExampleNetDialTimeout() {
	ln := NetListen("tcp", "127.0.0.1:0").Value.(Listener)
	defer ln.Close()

	r := NetDialTimeout("tcp", ln.Addr().String(), 0)
	Println(r.OK)
	if r.OK {
		r.Value.(Conn).Close()
	}
	// Output: true
}

// ExampleNetListen listens on a network address through `NetListen` for network health
// checks. Network primitives are reached through core wrappers and Result-shaped calls.
func ExampleNetListen() {
	r := NetListen("tcp", "127.0.0.1:0")
	Println(r.OK)
	if r.OK {
		r.Value.(Listener).Close()
	}
	// Output: true
}

// ExampleNetListenPacket listens for packets through `NetListenPacket` for network health
// checks. Network primitives are reached through core wrappers and Result-shaped calls.
func ExampleNetListenPacket() {
	r := NetListenPacket("udp", "127.0.0.1:0")
	Println(r.OK)
	if r.OK {
		r.Value.(PacketConn).Close()
	}
	// Output: true
}

// ExampleNetPipe creates an in-memory connection pair through `NetPipe` for network health
// checks. Network primitives are reached through core wrappers and Result-shaped calls.
func ExampleNetPipe() {
	left, right := NetPipe()
	defer right.Close()

	go func() {
		WriteString(left, "ping")
		left.Close()
	}()

	data := ReadAll(right)
	Println(data.Value)
	// Output: ping
}
