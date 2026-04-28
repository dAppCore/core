package core_test

import (
	. "dappco.re/go"
)

type testJSON struct {
	Name string `json:"name"`
	Port int    `json:"port"`
}

// --- JSONMarshal ---

func TestJson_JSONMarshal_Good(t *T) {
	r := JSONMarshal(testJSON{Name: "brain", Port: 8080})
	AssertTrue(t, r.OK)
	AssertContains(t, string(r.Value.([]byte)), `"name":"brain"`)
}

func TestJson_JSONMarshal_Bad_Unmarshalable(t *T) {
	r := JSONMarshal(make(chan int))
	AssertFalse(t, r.OK)
}

// --- JSONMarshalString ---

func TestJson_JSONMarshalString_Good(t *T) {
	s := JSONMarshalString(testJSON{Name: "x", Port: 1})
	AssertContains(t, s, `"name":"x"`)
}

func TestJson_JSONMarshalString_Ugly_Fallback(t *T) {
	s := JSONMarshalString(make(chan int))
	AssertEqual(t, "{}", s)
}

// --- JSONUnmarshal ---

func TestJson_JSONUnmarshal_Good(t *T) {
	var target testJSON
	r := JSONUnmarshal([]byte(`{"name":"brain","port":8080}`), &target)
	AssertTrue(t, r.OK)
	AssertEqual(t, "brain", target.Name)
	AssertEqual(t, 8080, target.Port)
}

func TestJson_JSONUnmarshal_Bad_Invalid(t *T) {
	var target testJSON
	r := JSONUnmarshal([]byte(`not json`), &target)
	AssertFalse(t, r.OK)
}

// --- JSONUnmarshalString ---

func TestJson_JSONUnmarshalString_Good(t *T) {
	var target testJSON
	r := JSONUnmarshalString(`{"name":"x","port":1}`, &target)
	AssertTrue(t, r.OK)
	AssertEqual(t, "x", target.Name)
}

func TestJson_JSONMarshal_Bad(t *T) {
	r := JSONMarshal(make(chan int))

	AssertFalse(t, r.OK)
	AssertError(t, r.Value.(error))
}

func TestJson_JSONMarshal_Ugly(t *T) {
	r := JSONMarshal(nil)

	AssertTrue(t, r.OK)
	AssertEqual(t, "null", string(r.Value.([]byte)))
}

func TestJson_JSONMarshalIndent_Good(t *T) {
	r := JSONMarshalIndent(testJSON{Name: "codex", Port: 8080}, "", "  ")

	AssertTrue(t, r.OK)
	AssertContains(t, string(r.Value.([]byte)), "\n  \"name\": \"codex\"")
}

func TestJson_JSONMarshalIndent_Bad(t *T) {
	r := JSONMarshalIndent(make(chan int), "", "  ")

	AssertFalse(t, r.OK)
	AssertError(t, r.Value.(error))
}

func TestJson_JSONMarshalIndent_Ugly(t *T) {
	r := JSONMarshalIndent(testJSON{Name: "codex", Port: 8080}, ">>", "\t")

	AssertTrue(t, r.OK)
	AssertContains(t, string(r.Value.([]byte)), "\n>>\t\"name\": \"codex\"")
}

func TestJson_JSONMarshalString_Bad(t *T) {
	AssertEqual(t, "{}", JSONMarshalString(make(chan int)))
}

func TestJson_JSONMarshalString_Ugly(t *T) {
	AssertEqual(t, "null", JSONMarshalString(nil))
}

func TestJson_JSONUnmarshal_Bad(t *T) {
	var target testJSON
	r := JSONUnmarshal([]byte(`not json`), &target)

	AssertFalse(t, r.OK)
	AssertError(t, r.Value.(error))
}

func TestJson_JSONUnmarshal_Ugly(t *T) {
	var target testJSON
	r := JSONUnmarshal([]byte(`{"name":"codex","extra":true}`), &target)

	AssertTrue(t, r.OK)
	AssertEqual(t, "codex", target.Name)
	AssertEqual(t, 0, target.Port)
}

func TestJson_JSONUnmarshalString_Bad(t *T) {
	var target testJSON
	r := JSONUnmarshalString(`not json`, &target)

	AssertFalse(t, r.OK)
	AssertError(t, r.Value.(error))
}

func TestJson_JSONUnmarshalString_Ugly(t *T) {
	target := testJSON{Name: "codex", Port: 8080}
	r := JSONUnmarshalString(`{}`, &target)

	AssertTrue(t, r.OK)
	AssertEqual(t, "codex", target.Name)
	AssertEqual(t, 8080, target.Port)
}
