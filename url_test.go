package core_test

import (
	"net/url"
	"testing"

	. "dappco.re/go/core"
	"github.com/stretchr/testify/assert"
)

// --- URLParse ---

func TestURL_Parse_Good(t *testing.T) {
	r := URLParse("https://example.com/path?q=hello#frag")
	assert.True(t, r.OK)

	u, ok := r.Value.(*url.URL)
	assert.True(t, ok)
	assert.Equal(t, "https", u.Scheme)
	assert.Equal(t, "example.com", u.Host)
	assert.Equal(t, "/path", u.Path)
	assert.Equal(t, "q=hello", u.RawQuery)
	assert.Equal(t, "frag", u.Fragment)
}

func TestURL_Parse_Bad(t *testing.T) {
	r := URLParse("http://[::1")
	assert.False(t, r.OK)
	_, ok := r.Value.(error)
	assert.True(t, ok)
}

func TestURL_Parse_Ugly(t *testing.T) {
	r := URLParse("mailto:user@example.com")
	assert.True(t, r.OK)

	u, ok := r.Value.(*url.URL)
	assert.True(t, ok)
	assert.Equal(t, "mailto", u.Scheme)
	assert.Equal(t, "user@example.com", u.Opaque)
}

// --- URLEncode ---

func TestURL_Encode_Good(t *testing.T) {
	assert.Equal(t, "hello+world%2Bcore", URLEncode("hello world+core"))
}

func TestURL_Encode_Bad(t *testing.T) {
	assert.Equal(t, "", URLEncode(""))
}

func TestURL_Encode_Ugly(t *testing.T) {
	assert.Equal(t, "%25+%26%3D%3F%23", URLEncode("% &=?#"))
}

// --- URLDecode ---

func TestURL_Decode_Good(t *testing.T) {
	r := URLDecode("hello+world%2Bcore")
	assert.True(t, r.OK)
	assert.Equal(t, "hello world+core", r.Value)
}

func TestURL_Decode_Bad(t *testing.T) {
	r := URLDecode("bad%zz")
	assert.False(t, r.OK)
	_, ok := r.Value.(error)
	assert.True(t, ok)
}

func TestURL_Decode_Ugly(t *testing.T) {
	r := URLDecode("a%2Fb%3Fc%23d")
	assert.True(t, r.OK)
	assert.Equal(t, "a/b?c#d", r.Value)
}

// --- URLPathEscape ---

func TestURL_PathEscape_Good(t *testing.T) {
	assert.Equal(t, "a%20b%2Fc", URLPathEscape("a b/c"))
}

func TestURL_PathEscape_Bad(t *testing.T) {
	assert.Equal(t, "", URLPathEscape(""))
}

func TestURL_PathEscape_Ugly(t *testing.T) {
	assert.Equal(t, "%25%3F%23", URLPathEscape("%?#"))
}

// --- URLNormalize ---

func TestURL_Normalize_Good(t *testing.T) {
	assert.Equal(t, "https://example.com/a%20b", URLNormalize("https://example.com/a b"))
}

func TestURL_Normalize_Bad(t *testing.T) {
	assert.Equal(t, "", URLNormalize("http://[::1"))
}

func TestURL_Normalize_Ugly(t *testing.T) {
	assert.Equal(t, "example.com/a%20b", URLNormalize("example.com/a b"))
}
