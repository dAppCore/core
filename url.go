// SPDX-License-Identifier: EUPL-1.2

// URL helpers for the Core framework.
// Wraps net/url so consumers can use core primitives for URL parsing
// and escaping.
package core

import "net/url"

// URL is the canonical parsed URL type.
//
//	r := core.URLParse("https://example.com/path")
//	if r.OK { u := r.Value.(*core.URL); _ = u }
type URL = url.URL

// URLValues is the canonical URL form/query values map.
//
//	values := core.URLValues{"key": {"value"}}
type URLValues = url.Values

// URLParse parses a raw URL string.
//
//	r := core.URLParse("https://example.com/path")
//	if r.OK { u := r.Value.(*core.URL) }
func URLParse(rawURL string) Result {
	u, err := url.Parse(rawURL)
	if err != nil {
		return Result{err, false}
	}
	return Result{u, true}
}

// URLEncode escapes a string for use in URL query components.
//
//	s := core.URLEncode("hello world")
func URLEncode(s string) string {
	return url.QueryEscape(s)
}

// URLDecode unescapes a URL query component string.
//
//	r := core.URLDecode("hello+world")
//	if r.OK { s := r.Value.(string) }
func URLDecode(s string) Result {
	decoded, err := url.QueryUnescape(s)
	if err != nil {
		return Result{err, false}
	}
	return Result{decoded, true}
}

// URLPathEscape escapes a string for use in URL path components.
//
//	s := core.URLPathEscape("a/b")
func URLPathEscape(s string) string {
	return url.PathEscape(s)
}

// URLNormalize parses and re-encodes a URL into net/url's canonical string form.
//
//	s := core.URLNormalize("https://example.com/a b")
func URLNormalize(rawURL string) string {
	u, err := url.Parse(rawURL)
	if err != nil {
		return ""
	}
	return u.String()
}
