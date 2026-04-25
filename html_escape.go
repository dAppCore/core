// SPDX-License-Identifier: EUPL-1.2

// HTML escaping helpers for the Core framework.
// Wraps html so consumers can use core primitives for common HTML text
// escaping and unescaping operations.
package core

import "html"

// HTMLEscape returns s with special HTML characters escaped.
//
//	escaped := core.HTMLEscape(`<a href="/search?q=go&lang=en">Go</a>`)
func HTMLEscape(s string) string {
	return html.EscapeString(s)
}

// HTMLUnescape returns s with HTML character references unescaped.
//
//	unescaped := core.HTMLUnescape("&lt;strong&gt;Go&lt;/strong&gt;")
func HTMLUnescape(s string) string {
	return html.UnescapeString(s)
}
