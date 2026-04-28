// SPDX-License-Identifier: EUPL-1.2

package core_test

import (
	. "dappco.re/go/core"
)

func TestNet_ParseIP_Good_IPv4(t *T) {
	ip := ParseIP("192.0.2.1")
	AssertNotNil(t, ip)
	AssertEqual(t, "192.0.2.1", ip.String())
}

func TestNet_ParseIP_Good_IPv6(t *T) {
	ip := ParseIP("2001:db8::1")
	AssertNotNil(t, ip)
}

func TestNet_ParseIP_Bad_Garbage(t *T) {
	AssertNil(t, ParseIP("not-an-ip"))
}

func TestNet_ParseCIDR_Good(t *T) {
	r := ParseCIDR("10.0.0.0/24")
	AssertTrue(t, r.OK)
	parts := r.Value.([]any)
	AssertLen(t, parts, 2)
	AssertNotNil(t, parts[0]) // IP
	AssertNotNil(t, parts[1]) // *IPNet
}

func TestNet_ParseCIDR_Bad(t *T) {
	r := ParseCIDR("not/a/cidr")
	AssertFalse(t, r.OK)
}

func TestNet_NetPipe_Good_BidirectionalIO(t *T) {
	a, b := NetPipe()
	defer a.Close()
	defer b.Close()

	go func() { _, _ = a.Write([]byte("ping")) }()
	buf := make([]byte, 4)
	n, err := b.Read(buf)
	AssertNoError(t, err)
	AssertEqual(t, 4, n)
	AssertEqual(t, "ping", string(buf))
}

func TestNet_NetListen_Good_Tcp(t *T) {
	r := NetListen("tcp", "127.0.0.1:0")
	AssertTrue(t, r.OK)
	ln := r.Value.(Listener)
	defer ln.Close()
	AssertNotNil(t, ln.Addr())
}

func TestNet_NetListen_Bad_InvalidAddress(t *T) {
	r := NetListen("tcp", "not:a:valid:address")
	AssertFalse(t, r.OK)
}

func TestNet_NetDial_Good_RoundTrip(t *T) {
	listener := NetListen("tcp", "127.0.0.1:0")
	AssertTrue(t, listener.OK)
	ln := listener.Value.(Listener)
	defer ln.Close()

	r := NetDial("tcp", ln.Addr().String())
	AssertTrue(t, r.OK)
	conn := r.Value.(Conn)
	conn.Close()
}
