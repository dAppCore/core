package core_test

import (
	"testing"

	. "dappco.re/go/core"
)

type testJSON struct {
	Name string `json:"name"`
	Port int    `json:"port"`
}

// --- JSONMarshal ---

func TestJson_JSONMarshal_Good(t *testing.T) {
	r := JSONMarshal(testJSON{Name: "brain", Port: 8080})
	AssertTrue(t, r.OK)
	AssertContains(t, string(r.Value.([]byte)), `"name":"brain"`)
}

func TestJson_JSONMarshal_Bad_Unmarshalable(t *testing.T) {
	r := JSONMarshal(make(chan int))
	AssertFalse(t, r.OK)
}

// --- JSONMarshalString ---

func TestJson_JSONMarshalString_Good(t *testing.T) {
	s := JSONMarshalString(testJSON{Name: "x", Port: 1})
	AssertContains(t, s, `"name":"x"`)
}

func TestJson_JSONMarshalString_Ugly_Fallback(t *testing.T) {
	s := JSONMarshalString(make(chan int))
	AssertEqual(t, "{}", s)
}

// --- JSONUnmarshal ---

func TestJson_JSONUnmarshal_Good(t *testing.T) {
	var target testJSON
	r := JSONUnmarshal([]byte(`{"name":"brain","port":8080}`), &target)
	AssertTrue(t, r.OK)
	AssertEqual(t, "brain", target.Name)
	AssertEqual(t, 8080, target.Port)
}

func TestJson_JSONUnmarshal_Bad_Invalid(t *testing.T) {
	var target testJSON
	r := JSONUnmarshal([]byte(`not json`), &target)
	AssertFalse(t, r.OK)
}

// --- JSONUnmarshalString ---

func TestJson_JSONUnmarshalString_Good(t *testing.T) {
	var target testJSON
	r := JSONUnmarshalString(`{"name":"x","port":1}`, &target)
	AssertTrue(t, r.OK)
	AssertEqual(t, "x", target.Name)
}
