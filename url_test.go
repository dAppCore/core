package core_test

import (
	. "dappco.re/go"
)

func TestUrl_URLDecode_Good(t *T) {
	r := URLDecode("agent+dispatch%2Fready")

	AssertTrue(t, r.OK)
	AssertEqual(t, "agent dispatch/ready", r.Value)
}

func TestUrl_URLDecode_Bad(t *T) {
	r := URLDecode("agent%zz")

	AssertFalse(t, r.OK)
	AssertError(t, r.Value.(error))
}

func TestUrl_URLDecode_Ugly(t *T) {
	r := URLDecode("")

	AssertTrue(t, r.OK)
	AssertEqual(t, "", r.Value)
}

func TestUrl_URLEncode_Good(t *T) {
	AssertEqual(t, "agent+dispatch%2Fready", URLEncode("agent dispatch/ready"))
}

func TestUrl_URLEncode_Bad(t *T) {
	AssertEqual(t, "", URLEncode(""))
}

func TestUrl_URLEncode_Ugly(t *T) {
	AssertEqual(t, "%25+%26%3D%3F%23", URLEncode("% &=?#"))
}

func TestUrl_URLNormalize_Good(t *T) {
	AssertEqual(t, "https://example.com/agent%20dispatch", URLNormalize("https://example.com/agent dispatch"))
}

func TestUrl_URLNormalize_Bad(t *T) {
	AssertEqual(t, "", URLNormalize("http://[::1"))
}

func TestUrl_URLNormalize_Ugly(t *T) {
	AssertEqual(t, "example.com/agent%20dispatch", URLNormalize("example.com/agent dispatch"))
}

func TestUrl_URLParse_Good(t *T) {
	r := URLParse("https://example.com/path?q=agent#ready")

	AssertTrue(t, r.OK)
	u := r.Value.(*URL)
	AssertEqual(t, "https", u.Scheme)
	AssertEqual(t, "example.com", u.Host)
	AssertEqual(t, "/path", u.Path)
	AssertEqual(t, "q=agent", u.RawQuery)
	AssertEqual(t, "ready", u.Fragment)
}

func TestUrl_URLParse_Bad(t *T) {
	r := URLParse("http://[::1")

	AssertFalse(t, r.OK)
	AssertError(t, r.Value.(error))
}

func TestUrl_URLParse_Ugly(t *T) {
	r := URLParse("mailto:agent@example.com")

	AssertTrue(t, r.OK)
	u := r.Value.(*URL)
	AssertEqual(t, "mailto", u.Scheme)
	AssertEqual(t, "agent@example.com", u.Opaque)
}

func TestUrl_URLPathEscape_Good(t *T) {
	AssertEqual(t, "agent%20dispatch%2Fready", URLPathEscape("agent dispatch/ready"))
}

func TestUrl_URLPathEscape_Bad(t *T) {
	AssertEqual(t, "", URLPathEscape(""))
}

func TestUrl_URLPathEscape_Ugly(t *T) {
	AssertEqual(t, "%25%3F%23", URLPathEscape("%?#"))
}
