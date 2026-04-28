package core_test

import (
	"net/url"

	. "dappco.re/go"
)

// --- URLParse ---

func TestURL_Parse_Good(t *T) {
	r := URLParse("https://example.com/path?q=hello#frag")
	AssertTrue(t, r.OK)

	u, ok := r.Value.(*url.URL)
	AssertTrue(t, ok)
	AssertEqual(t, "https", u.Scheme)
	AssertEqual(t, "example.com", u.Host)
	AssertEqual(t, "/path", u.Path)
	AssertEqual(t, "q=hello", u.RawQuery)
	AssertEqual(t, "frag", u.Fragment)
}

func TestURL_Parse_Bad(t *T) {
	r := URLParse("http://[::1")
	AssertFalse(t, r.OK)
	_, ok := r.Value.(error)
	AssertTrue(t, ok)
}

func TestURL_Parse_Ugly(t *T) {
	r := URLParse("mailto:user@example.com")
	AssertTrue(t, r.OK)

	u, ok := r.Value.(*url.URL)
	AssertTrue(t, ok)
	AssertEqual(t, "mailto", u.Scheme)
	AssertEqual(t, "user@example.com", u.Opaque)
}

// --- URLEncode ---

func TestURL_Encode_Good(t *T) {
	AssertEqual(t, "hello+world%2Bcore", URLEncode("hello world+core"))
}

func TestURL_Encode_Bad(t *T) {
	AssertEqual(t, "", URLEncode(""))
}

func TestURL_Encode_Ugly(t *T) {
	AssertEqual(t, "%25+%26%3D%3F%23", URLEncode("% &=?#"))
}

// --- URLDecode ---

func TestURL_Decode_Good(t *T) {
	r := URLDecode("hello+world%2Bcore")
	AssertTrue(t, r.OK)
	AssertEqual(t, "hello world+core", r.Value)
}

func TestURL_Decode_Bad(t *T) {
	r := URLDecode("bad%zz")
	AssertFalse(t, r.OK)
	_, ok := r.Value.(error)
	AssertTrue(t, ok)
}

func TestURL_Decode_Ugly(t *T) {
	r := URLDecode("a%2Fb%3Fc%23d")
	AssertTrue(t, r.OK)
	AssertEqual(t, "a/b?c#d", r.Value)
}

// --- URLPathEscape ---

func TestURL_PathEscape_Good(t *T) {
	AssertEqual(t, "a%20b%2Fc", URLPathEscape("a b/c"))
}

func TestURL_PathEscape_Bad(t *T) {
	AssertEqual(t, "", URLPathEscape(""))
}

func TestURL_PathEscape_Ugly(t *T) {
	AssertEqual(t, "%25%3F%23", URLPathEscape("%?#"))
}

// --- URLNormalize ---

func TestURL_Normalize_Good(t *T) {
	AssertEqual(t, "https://example.com/a%20b", URLNormalize("https://example.com/a b"))
}

func TestURL_Normalize_Bad(t *T) {
	AssertEqual(t, "", URLNormalize("http://[::1"))
}

func TestURL_Normalize_Ugly(t *T) {
	AssertEqual(t, "example.com/a%20b", URLNormalize("example.com/a b"))
}
