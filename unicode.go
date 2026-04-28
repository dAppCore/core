// SPDX-License-Identifier: EUPL-1.2

// Unicode rune helpers for the Core framework.
// Wraps unicode so consumers don't import it directly.
package core

import "unicode"

// IsLetter reports whether r is a letter.
//
//	core.IsLetter('A')  // true
func IsLetter(r rune) bool {
	return unicode.IsLetter(r)
}

// IsDigit reports whether r is a decimal digit.
//
//	core.IsDigit('9')  // true
func IsDigit(r rune) bool {
	return unicode.IsDigit(r)
}

// IsSpace reports whether r is a space character.
//
//	core.IsSpace('\n')  // true
func IsSpace(r rune) bool {
	return unicode.IsSpace(r)
}

// IsUpper reports whether r is an uppercase letter.
//
//	core.IsUpper('A')  // true
func IsUpper(r rune) bool {
	return unicode.IsUpper(r)
}

// IsLower reports whether r is a lowercase letter.
//
//	core.IsLower('a')  // true
func IsLower(r rune) bool {
	return unicode.IsLower(r)
}

// ToUpper maps r to uppercase.
//
//	core.ToUpper('a')  // 'A'
func ToUpper(r rune) rune {
	return unicode.ToUpper(r)
}

// ToLower maps r to lowercase.
//
//	core.ToLower('A')  // 'a'
func ToLower(r rune) rune {
	return unicode.ToLower(r)
}
