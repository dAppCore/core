package core_test

import (
	. "dappco.re/go"
)

func TestFormat_Errorf_Good(t *T) {
	err := Errorf("dispatch %s: %w", "agent", AnError)

	AssertError(t, err, "dispatch agent")
	AssertTrue(t, Is(err, AnError))
}

func TestFormat_Errorf_Bad(t *T) {
	err := Errorf("agent %s", "codex")

	AssertError(t, err, "agent codex")
	AssertFalse(t, Is(err, AnError))
}

func TestFormat_Errorf_Ugly(t *T) {
	err := Errorf("")

	AssertError(t, err)
	AssertEqual(t, "", err.Error())
}

func TestFormat_Print_Good(t *T) {
	buf := NewBuffer()

	Print(buf, "agent %s", "ready")

	AssertEqual(t, "agent ready\n", buf.String())
}

func TestFormat_Print_Bad(t *T) {
	buf := NewBuffer()

	Print(buf, "")

	AssertEqual(t, "\n", buf.String())
}

func TestFormat_Print_Ugly(t *T) {
	buf := NewBuffer()

	Print(buf, "agent %s")

	AssertEqual(t, "agent %!s(MISSING)\n", buf.String())
}

func TestFormat_Println_Good(t *T) {
	AssertNotPanics(t, func() { Println("agent", "ready") })
}

func TestFormat_Println_Bad(t *T) {
	AssertNotPanics(t, func() { Println() })
}

func TestFormat_Println_Ugly(t *T) {
	AssertNotPanics(t, func() { Println(nil) })
}

func TestFormat_Sprint_Good(t *T) {
	AssertEqual(t, "agent42", Sprint("agent", 42))
}

func TestFormat_Sprint_Bad(t *T) {
	AssertEqual(t, "", Sprint())
}

func TestFormat_Sprint_Ugly(t *T) {
	AssertEqual(t, "<nil>", Sprint(nil))
}

func TestFormat_Sprintf_Good(t *T) {
	AssertEqual(t, "agent codex ready", Sprintf("agent %s ready", "codex"))
}

func TestFormat_Sprintf_Bad(t *T) {
	AssertEqual(t, `""`, Sprintf("%q", ""))
}

func TestFormat_Sprintf_Ugly(t *T) {
	AssertEqual(t, "100% ready", Sprintf("%d%% ready", 100))
}

func TestFormat_Sprintln_Good(t *T) {
	AssertEqual(t, "agent ready\n", Sprintln("agent", "ready"))
}

func TestFormat_Sprintln_Bad(t *T) {
	AssertEqual(t, "\n", Sprintln())
}

func TestFormat_Sprintln_Ugly(t *T) {
	AssertEqual(t, "<nil>\n", Sprintln(nil))
}
