// SPDX-License-Identifier: EUPL-1.2

package core_test

import (
	. "dappco.re/go"
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

func TestRegexp_Regex_Good(t *T) {
	r := Regex(`agent-[0-9]+`)

	AssertTrue(t, r.OK)
	AssertEqual(t, `agent-[0-9]+`, r.Value.(*Regexp).String())
}

func TestRegexp_Regex_Bad(t *T) {
	r := Regex(`[agent`)

	AssertFalse(t, r.OK)
	AssertError(t, r.Value.(error))
}

func TestRegexp_Regex_Ugly(t *T) {
	r := Regex("")

	AssertTrue(t, r.OK)
	AssertTrue(t, r.Value.(*Regexp).MatchString("agent"))
}

func TestRegexp_Regexp_FindAllString_Good(t *T) {
	rx := Regex(`agent-[0-9]+`).Value.(*Regexp)

	AssertEqual(t, []string{"agent-1", "agent-2"}, rx.FindAllString("agent-1 agent-2", -1))
}

func TestRegexp_Regexp_FindAllString_Bad(t *T) {
	rx := Regex(`agent-[0-9]+`).Value.(*Regexp)

	AssertNil(t, rx.FindAllString("no agents", -1))
}

func TestRegexp_Regexp_FindAllString_Ugly(t *T) {
	rx := Regex(`agent-[0-9]+`).Value.(*Regexp)

	AssertNil(t, rx.FindAllString("agent-1 agent-2", 0))
}

func TestRegexp_Regexp_FindString_Good(t *T) {
	rx := Regex(`agent-[0-9]+`).Value.(*Regexp)

	AssertEqual(t, "agent-7", rx.FindString("dispatch agent-7 ready"))
}

func TestRegexp_Regexp_FindString_Bad(t *T) {
	rx := Regex(`agent-[0-9]+`).Value.(*Regexp)

	AssertEqual(t, "", rx.FindString("dispatch ready"))
}

func TestRegexp_Regexp_FindString_Ugly(t *T) {
	rx := Regex(``).Value.(*Regexp)

	AssertEqual(t, "", rx.FindString("dispatch"))
}

func TestRegexp_Regexp_FindStringSubmatch_Good(t *T) {
	rx := Regex(`agent=([a-z]+)`).Value.(*Regexp)

	AssertEqual(t, []string{"agent=codex", "codex"}, rx.FindStringSubmatch("agent=codex"))
}

func TestRegexp_Regexp_FindStringSubmatch_Bad(t *T) {
	rx := Regex(`agent=([a-z]+)`).Value.(*Regexp)

	AssertNil(t, rx.FindStringSubmatch("agent=42"))
}

func TestRegexp_Regexp_FindStringSubmatch_Ugly(t *T) {
	rx := Regex(`(agent)?`).Value.(*Regexp)

	AssertEqual(t, []string{"", ""}, rx.FindStringSubmatch(""))
}

func TestRegexp_Regexp_MatchString_Good(t *T) {
	rx := Regex(`^agent\.`).Value.(*Regexp)

	AssertTrue(t, rx.MatchString("agent.dispatch"))
}

func TestRegexp_Regexp_MatchString_Bad(t *T) {
	rx := Regex(`^agent\.`).Value.(*Regexp)

	AssertFalse(t, rx.MatchString("task.dispatch"))
}

func TestRegexp_Regexp_MatchString_Ugly(t *T) {
	rx := Regex(`^$`).Value.(*Regexp)

	AssertTrue(t, rx.MatchString(""))
}

func TestRegexp_Regexp_ReplaceAllString_Good(t *T) {
	rx := Regex(`/+`).Value.(*Regexp)

	AssertEqual(t, "agent.dispatch.ready", rx.ReplaceAllString("agent/dispatch//ready", "."))
}

func TestRegexp_Regexp_ReplaceAllString_Bad(t *T) {
	rx := Regex(`missing`).Value.(*Regexp)

	AssertEqual(t, "agent.dispatch", rx.ReplaceAllString("agent.dispatch", "task"))
}

func TestRegexp_Regexp_ReplaceAllString_Ugly(t *T) {
	rx := Regex(`agent-([0-9]+)`).Value.(*Regexp)

	AssertEqual(t, "id=42", rx.ReplaceAllString("agent-42", "id=$1"))
}

func TestRegexp_Regexp_Split_Good(t *T) {
	rx := Regex(`\s+`).Value.(*Regexp)

	AssertEqual(t, []string{"agent", "dispatch", "ready"}, rx.Split("agent dispatch ready", -1))
}

func TestRegexp_Regexp_Split_Bad(t *T) {
	rx := Regex(`,`).Value.(*Regexp)

	AssertNil(t, rx.Split("agent,dispatch", 0))
}

func TestRegexp_Regexp_Split_Ugly(t *T) {
	rx := Regex(`,`).Value.(*Regexp)

	AssertEqual(t, []string{"", "agent", ""}, rx.Split(",agent,", -1))
}

func TestRegexp_Regexp_String_Good(t *T) {
	rx := Regex(`agent-[0-9]+`).Value.(*Regexp)

	AssertEqual(t, `agent-[0-9]+`, rx.String())
}

func TestRegexp_Regexp_String_Bad(t *T) {
	rx := Regex(`\.`).Value.(*Regexp)

	AssertEqual(t, `\.`, rx.String())
}

func TestRegexp_Regexp_String_Ugly(t *T) {
	rx := Regex("").Value.(*Regexp)

	AssertEqual(t, "", rx.String())
}
