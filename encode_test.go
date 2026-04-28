package core_test

import (
	. "dappco.re/go/core"
)

// --- HexEncode ---

func TestEncode_HexEncode_Good(t *T) {
	AssertEqual(t, "68656c6c6f", HexEncode([]byte("hello")))
}

func TestEncode_HexEncode_Bad(t *T) {
	AssertEqual(t, "", HexEncode(nil))
	AssertEqual(t, HexEncode(nil), HexEncode([]byte{}))
}

func TestEncode_HexEncode_Ugly(t *T) {
	src := []byte{0x00, 0x0f, 0x10, 0xff}
	encoded := HexEncode(src)

	src[0] = 0xff

	AssertEqual(t, "000f10ff", encoded)
	AssertNotEqual(t, encoded, HexEncode(src))
}

// --- HexDecode ---

func TestEncode_HexDecode_Good(t *T) {
	r := HexDecode("68656c6c6f")
	AssertTrue(t, r.OK)
	AssertEqual(t, []byte("hello"), r.Value)
}

func TestEncode_HexDecode_Bad(t *T) {
	r := HexDecode("not-hex")
	AssertFalse(t, r.OK)
	_, ok := r.Value.(error)
	AssertTrue(t, ok)
}

func TestEncode_HexDecode_Ugly(t *T) {
	r := HexDecode("abc")
	AssertFalse(t, r.OK)
	_, ok := r.Value.(error)
	AssertTrue(t, ok)
}

// --- Base64Encode ---

func TestEncode_Base64Encode_Good(t *T) {
	AssertEqual(t, "aGVsbG8=", Base64Encode([]byte("hello")))
}

func TestEncode_Base64Encode_Bad(t *T) {
	AssertEqual(t, "", Base64Encode(nil))
	AssertEqual(t, Base64Encode(nil), Base64Encode([]byte{}))
}

func TestEncode_Base64Encode_Ugly(t *T) {
	src := []byte{0xfb, 0xff, 0xff}
	encoded := Base64Encode(src)

	src[0] = 0x00

	AssertEqual(t, "+///", encoded)
	AssertNotEqual(t, encoded, Base64Encode(src))
}

// --- Base64Decode ---

func TestEncode_Base64Decode_Good(t *T) {
	r := Base64Decode("aGVsbG8=")
	AssertTrue(t, r.OK)
	AssertEqual(t, []byte("hello"), r.Value)
}

func TestEncode_Base64Decode_Bad(t *T) {
	r := Base64Decode("not-base64")
	AssertFalse(t, r.OK)
	_, ok := r.Value.(error)
	AssertTrue(t, ok)
}

func TestEncode_Base64Decode_Ugly(t *T) {
	r := Base64Decode("aGVsbG8===")
	AssertFalse(t, r.OK)
	_, ok := r.Value.(error)
	AssertTrue(t, ok)
}

// --- Base64URLEncode ---

func TestEncode_Base64URLEncode_Good(t *T) {
	AssertEqual(t, "aGVsbG8=", Base64URLEncode([]byte("hello")))
}

func TestEncode_Base64URLEncode_Bad(t *T) {
	AssertEqual(t, "", Base64URLEncode(nil))
	AssertEqual(t, Base64URLEncode(nil), Base64URLEncode([]byte{}))
}

func TestEncode_Base64URLEncode_Ugly(t *T) {
	src := []byte{0xfb, 0xff, 0xff}
	encoded := Base64URLEncode(src)

	src[0] = 0x00

	AssertEqual(t, "-___", encoded)
	AssertNotEqual(t, encoded, Base64URLEncode(src))
}

// --- Base64URLDecode ---

func TestEncode_Base64URLDecode_Good(t *T) {
	r := Base64URLDecode("aGVsbG8=")
	AssertTrue(t, r.OK)
	AssertEqual(t, []byte("hello"), r.Value)
}

func TestEncode_Base64URLDecode_Bad(t *T) {
	r := Base64URLDecode("not+url")
	AssertFalse(t, r.OK)
	_, ok := r.Value.(error)
	AssertTrue(t, ok)
}

func TestEncode_Base64URLDecode_Ugly(t *T) {
	r := Base64URLDecode("aGVsbG8===")
	AssertFalse(t, r.OK)
	_, ok := r.Value.(error)
	AssertTrue(t, ok)
}
