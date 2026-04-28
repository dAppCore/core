package core_test

import . "dappco.re/go"

func ExampleParseIP() {
	ip := ParseIP("192.0.2.10")
	Println(ip.String())
	// Output: 192.0.2.10
}

func ExampleParseCIDR() {
	r := ParseCIDR("192.0.2.0/24")
	parts := r.Value.([]any)
	Println(parts[0])
	Println(parts[1])
	// Output:
	// 192.0.2.0
	// 192.0.2.0/24
}

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

func ExampleNetListen() {
	r := NetListen("tcp", "127.0.0.1:0")
	Println(r.OK)
	if r.OK {
		r.Value.(Listener).Close()
	}
	// Output: true
}

func ExampleNetListenPacket() {
	r := NetListenPacket("udp", "127.0.0.1:0")
	Println(r.OK)
	if r.OK {
		r.Value.(PacketConn).Close()
	}
	// Output: true
}

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
