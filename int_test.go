package core_test

import (
	"testing"

	. "dappco.re/go/core"
	"github.com/stretchr/testify/assert"
)

// --- Atoi ---

func TestInt_Atoi_Good(t *testing.T) {
	r := Atoi("42")
	assert.True(t, r.OK)
	assert.Equal(t, 42, r.Value)
}

func TestInt_Atoi_Bad(t *testing.T) {
	r := Atoi("not-an-int")
	assert.False(t, r.OK)
	_, ok := r.Value.(error)
	assert.True(t, ok)
}

func TestInt_Atoi_Ugly(t *testing.T) {
	r := Atoi("999999999999999999999999999999")
	assert.False(t, r.OK)
	_, ok := r.Value.(error)
	assert.True(t, ok)
}

// --- Itoa ---

func TestInt_Itoa_Good(t *testing.T) {
	assert.Equal(t, "42", Itoa(42))
}

func TestInt_Itoa_Bad(t *testing.T) {
	assert.Equal(t, "-42", Itoa(-42))
}

func TestInt_Itoa_Ugly(t *testing.T) {
	assert.Equal(t, "0", Itoa(0))
}

// --- FormatInt ---

func TestInt_FormatInt_Good(t *testing.T) {
	assert.Equal(t, "ff", FormatInt(255, 16))
}

func TestInt_FormatInt_Bad(t *testing.T) {
	assert.Equal(t, "-ff", FormatInt(-255, 16))
}

func TestInt_FormatInt_Ugly(t *testing.T) {
	assert.Equal(t, "z", FormatInt(35, 36))
	assert.Equal(t, "0", FormatInt(0, 2))
}

// --- ParseInt ---

func TestInt_ParseInt_Good(t *testing.T) {
	r := ParseInt("ff", 16, 64)
	assert.True(t, r.OK)
	assert.Equal(t, int64(255), r.Value)
}

func TestInt_ParseInt_Bad(t *testing.T) {
	r := ParseInt("not-an-int", 10, 64)
	assert.False(t, r.OK)
	_, ok := r.Value.(error)
	assert.True(t, ok)
}

func TestInt_ParseInt_Ugly(t *testing.T) {
	r := ParseInt("255", 10, 8)
	assert.False(t, r.OK)
	_, ok := r.Value.(error)
	assert.True(t, ok)
}
