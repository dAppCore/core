package core_test

import (
	"testing"

	. "dappco.re/go/core"
	"github.com/stretchr/testify/assert"
)

type testJSON struct {
	Name string `json:"name"`
	Port int    `json:"port"`
}

// --- JSONMarshal ---

func TestJson_JSONMarshal_Good(t *testing.T) {
	r := JSONMarshal(testJSON{Name: "brain", Port: 8080})
	assert.True(t, r.OK)
	assert.Contains(t, string(r.Value.([]byte)), `"name":"brain"`)
}

func TestJson_JSONMarshal_Bad_Unmarshalable(t *testing.T) {
	r := JSONMarshal(make(chan int))
	assert.False(t, r.OK)
}

// --- JSONMarshalString ---

func TestJson_JSONMarshalString_Good(t *testing.T) {
	s := JSONMarshalString(testJSON{Name: "x", Port: 1})
	assert.Contains(t, s, `"name":"x"`)
}

func TestJson_JSONMarshalString_Ugly_Fallback(t *testing.T) {
	s := JSONMarshalString(make(chan int))
	assert.Equal(t, "{}", s)
}

// --- JSONUnmarshal ---

func TestJson_JSONUnmarshal_Good(t *testing.T) {
	var target testJSON
	r := JSONUnmarshal([]byte(`{"name":"brain","port":8080}`), &target)
	assert.True(t, r.OK)
	assert.Equal(t, "brain", target.Name)
	assert.Equal(t, 8080, target.Port)
}

func TestJson_JSONUnmarshal_Bad_Invalid(t *testing.T) {
	var target testJSON
	r := JSONUnmarshal([]byte(`not json`), &target)
	assert.False(t, r.OK)
}

// --- JSONUnmarshalString ---

func TestJson_JSONUnmarshalString_Good(t *testing.T) {
	var target testJSON
	r := JSONUnmarshalString(`{"name":"x","port":1}`, &target)
	assert.True(t, r.OK)
	assert.Equal(t, "x", target.Name)
}
