// SPDX-License-Identifier: EUPL-1.2

package core_test

import (
	. "dappco.re/go"
)

func TestNet_NetDial_Good(t *T) {
	listener := NetListen("tcp", "127.0.0.1:0")
	RequireTrue(t, listener.OK)
	ln := listener.Value.(Listener)
	defer ln.Close()
	go func() {
		conn, err := ln.Accept()
		if err == nil {
			conn.Close()
		}
	}()

	r := NetDial("tcp", ln.Addr().String())

	AssertTrue(t, r.OK)
	r.Value.(Conn).Close()
}

func TestNet_NetDial_Bad(t *T) {
	r := NetDial("bad-network", "127.0.0.1:1")

	AssertFalse(t, r.OK)
	AssertError(t, r.Value.(error))
}

func TestNet_NetDial_Ugly(t *T) {
	r := NetDial("tcp", "")

	AssertFalse(t, r.OK)
	AssertError(t, r.Value.(error))
}

func TestNet_NetDialTimeout_Good(t *T) {
	listener := NetListen("tcp", "127.0.0.1:0")
	RequireTrue(t, listener.OK)
	ln := listener.Value.(Listener)
	defer ln.Close()
	go func() {
		conn, err := ln.Accept()
		if err == nil {
			conn.Close()
		}
	}()

	r := NetDialTimeout("tcp", ln.Addr().String(), Second)

	AssertTrue(t, r.OK)
	r.Value.(Conn).Close()
}

func TestNet_NetDialTimeout_Bad(t *T) {
	r := NetDialTimeout("bad-network", "127.0.0.1:1", Millisecond)

	AssertFalse(t, r.OK)
	AssertError(t, r.Value.(error))
}

func TestNet_NetDialTimeout_Ugly(t *T) {
	r := NetDialTimeout("tcp", "", 0)

	AssertFalse(t, r.OK)
	AssertError(t, r.Value.(error))
}

func TestNet_NetListen_Good(t *T) {
	r := NetListen("tcp", "127.0.0.1:0")

	AssertTrue(t, r.OK)
	ln := r.Value.(Listener)
	defer ln.Close()
	AssertContains(t, ln.Addr().String(), "127.0.0.1:")
}

func TestNet_NetListen_Bad(t *T) {
	r := NetListen("bad-network", "127.0.0.1:0")

	AssertFalse(t, r.OK)
	AssertError(t, r.Value.(error))
}

func TestNet_NetListen_Ugly(t *T) {
	r := NetListen("tcp", "localhost:0")

	AssertTrue(t, r.OK)
	r.Value.(Listener).Close()
}

func TestNet_NetListenPacket_Good(t *T) {
	r := NetListenPacket("udp", "127.0.0.1:0")

	AssertTrue(t, r.OK)
	pc := r.Value.(PacketConn)
	defer pc.Close()
	AssertContains(t, pc.LocalAddr().String(), "127.0.0.1:")
}

func TestNet_NetListenPacket_Bad(t *T) {
	r := NetListenPacket("bad-network", "127.0.0.1:0")

	AssertFalse(t, r.OK)
	AssertError(t, r.Value.(error))
}

func TestNet_NetListenPacket_Ugly(t *T) {
	r := NetListenPacket("udp4", "127.0.0.1:0")

	AssertTrue(t, r.OK)
	r.Value.(PacketConn).Close()
}

func TestNet_NetPipe_Good(t *T) {
	a, b := NetPipe()
	go func() {
		_ = WriteString(a, "ping")
		a.Close()
	}()

	r := ReadAll(b)

	AssertTrue(t, r.OK)
	AssertEqual(t, "ping", r.Value)
}

func TestNet_NetPipe_Bad(t *T) {
	a, b := NetPipe()
	defer b.Close()
	a.Close()

	_, err := a.Write([]byte("ping"))

	AssertError(t, err)
}

func TestNet_NetPipe_Ugly(t *T) {
	a, b := NetPipe()
	go func() { a.Close() }()

	r := ReadAll(b)

	AssertTrue(t, r.OK)
	AssertEqual(t, "", r.Value)
}

func TestNet_ParseCIDR_Good(t *T) {
	r := ParseCIDR("10.0.0.0/24")

	AssertTrue(t, r.OK)
	parts := r.Value.([]any)
	AssertLen(t, parts, 2)
	AssertEqual(t, "10.0.0.0", parts[0].(IP).String())
	AssertEqual(t, "10.0.0.0/24", parts[1].(*IPNet).String())
}

func TestNet_ParseCIDR_Bad(t *T) {
	r := ParseCIDR("not/a/cidr")

	AssertFalse(t, r.OK)
	AssertError(t, r.Value.(error))
}

func TestNet_ParseCIDR_Ugly(t *T) {
	r := ParseCIDR("127.0.0.1/32")

	AssertTrue(t, r.OK)
	AssertEqual(t, "127.0.0.1/32", r.Value.([]any)[1].(*IPNet).String())
}

func TestNet_ParseIP_Good(t *T) {
	ip := ParseIP("192.0.2.1")

	AssertNotNil(t, ip)
	AssertEqual(t, "192.0.2.1", ip.String())
}

func TestNet_ParseIP_Bad(t *T) {
	AssertNil(t, ParseIP("not-an-ip"))
}

func TestNet_ParseIP_Ugly(t *T) {
	ip := ParseIP("2001:db8::1")

	AssertNotNil(t, ip)
	AssertEqual(t, "2001:db8::1", ip.String())
}
