// SPDX-License-Identifier: EUPL-1.2

package core_test

import (
	. "dappco.re/go/core"
)

func TestRegexp_Regex_Good_Compile(t *T) {
	r := Regex(`\d+`)
	AssertTrue(t, r.OK)
	AssertNotNil(t, r.Value)
}

func TestRegexp_Regex_Bad_InvalidPattern(t *T) {
	r := Regex(`[`)
	AssertFalse(t, r.OK)
	AssertError(t, r.Value.(error))
}

func TestRegexp_MatchString_Good(t *T) {
	rx := Regex(`\d+`).Value.(*Regexp)
	AssertTrue(t, rx.MatchString("foo123bar"))
	AssertFalse(t, rx.MatchString("no digits"))
}

func TestRegexp_FindString_Good(t *T) {
	rx := Regex(`\d+`).Value.(*Regexp)
	AssertEqual(t, "42", rx.FindString("hello 42 world"))
	AssertEqual(t, "", rx.FindString("nothing here"))
}

func TestRegexp_FindAllString_Good(t *T) {
	rx := Regex(`\d+`).Value.(*Regexp)
	AssertEqual(t, []string{"1", "2", "3"}, rx.FindAllString("a1 b2 c3", -1))
	AssertEqual(t, []string{"1", "2"}, rx.FindAllString("a1 b2 c3", 2))
}

func TestRegexp_FindStringSubmatch_Good(t *T) {
	rx := Regex(`(\w+)=(\d+)`).Value.(*Regexp)
	got := rx.FindStringSubmatch("count=42")
	AssertEqual(t, []string{"count=42", "count", "42"}, got)
}

func TestRegexp_ReplaceAllString_Good(t *T) {
	rx := Regex(`\d`).Value.(*Regexp)
	AssertEqual(t, "aXbX", rx.ReplaceAllString("a1b2", "X"))
}

func TestRegexp_Split_Good(t *T) {
	rx := Regex(`,+`).Value.(*Regexp)
	AssertEqual(t, []string{"a", "b", "c"}, rx.Split("a,b,,c", -1))
}

func TestRegexp_String_Good(t *T) {
	rx := Regex(`abc`).Value.(*Regexp)
	AssertEqual(t, "abc", rx.String())
}
