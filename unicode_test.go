package core_test

import (
	"testing"

	. "dappco.re/go/core"
	"github.com/stretchr/testify/assert"
)

// --- Unicode Operations ---

func TestUnicode_IsLetter_Good(t *testing.T) {
	assert.True(t, IsLetter('A'))
	assert.True(t, IsLetter('é'))
	assert.True(t, IsLetter('界'))
}

func TestUnicode_IsLetter_Bad(t *testing.T) {
	assert.False(t, IsLetter('7'))
	assert.False(t, IsLetter('_'))
	assert.False(t, IsLetter(' '))
}

func TestUnicode_IsLetter_Ugly(t *testing.T) {
	assert.False(t, IsLetter(rune(0)))
	assert.False(t, IsLetter(-1))
}

func TestUnicode_IsDigit_Good(t *testing.T) {
	assert.True(t, IsDigit('0'))
	assert.True(t, IsDigit('9'))
	assert.True(t, IsDigit('٥'))
}

func TestUnicode_IsDigit_Bad(t *testing.T) {
	assert.False(t, IsDigit('A'))
	assert.False(t, IsDigit('.'))
	assert.False(t, IsDigit(' '))
}

func TestUnicode_IsDigit_Ugly(t *testing.T) {
	assert.False(t, IsDigit(rune(0)))
	assert.False(t, IsDigit(-1))
}

func TestUnicode_IsSpace_Good(t *testing.T) {
	assert.True(t, IsSpace(' '))
	assert.True(t, IsSpace('\n'))
	assert.True(t, IsSpace('\u00a0'))
}

func TestUnicode_IsSpace_Bad(t *testing.T) {
	assert.False(t, IsSpace('A'))
	assert.False(t, IsSpace('0'))
	assert.False(t, IsSpace('_'))
}

func TestUnicode_IsSpace_Ugly(t *testing.T) {
	assert.False(t, IsSpace(rune(0)))
	assert.False(t, IsSpace(-1))
}

func TestUnicode_IsUpper_Good(t *testing.T) {
	assert.True(t, IsUpper('A'))
	assert.True(t, IsUpper('É'))
	assert.True(t, IsUpper('Ω'))
}

func TestUnicode_IsUpper_Bad(t *testing.T) {
	assert.False(t, IsUpper('a'))
	assert.False(t, IsUpper('7'))
	assert.False(t, IsUpper(' '))
}

func TestUnicode_IsUpper_Ugly(t *testing.T) {
	assert.False(t, IsUpper(rune(0)))
	assert.False(t, IsUpper(-1))
}

func TestUnicode_IsLower_Good(t *testing.T) {
	assert.True(t, IsLower('a'))
	assert.True(t, IsLower('é'))
	assert.True(t, IsLower('ω'))
}

func TestUnicode_IsLower_Bad(t *testing.T) {
	assert.False(t, IsLower('A'))
	assert.False(t, IsLower('7'))
	assert.False(t, IsLower(' '))
}

func TestUnicode_IsLower_Ugly(t *testing.T) {
	assert.False(t, IsLower(rune(0)))
	assert.False(t, IsLower(-1))
}

func TestUnicode_ToUpper_Good(t *testing.T) {
	assert.Equal(t, 'A', ToUpper('a'))
	assert.Equal(t, 'É', ToUpper('é'))
	assert.Equal(t, 'Ω', ToUpper('ω'))
}

func TestUnicode_ToUpper_Bad(t *testing.T) {
	assert.Equal(t, 'A', ToUpper('A'))
	assert.Equal(t, '7', ToUpper('7'))
	assert.Equal(t, ' ', ToUpper(' '))
}

func TestUnicode_ToUpper_Ugly(t *testing.T) {
	assert.Equal(t, rune(0), ToUpper(rune(0)))
	assert.Equal(t, rune(-1), ToUpper(-1))
}

func TestUnicode_ToLower_Good(t *testing.T) {
	assert.Equal(t, 'a', ToLower('A'))
	assert.Equal(t, 'é', ToLower('É'))
	assert.Equal(t, 'ω', ToLower('Ω'))
}

func TestUnicode_ToLower_Bad(t *testing.T) {
	assert.Equal(t, 'a', ToLower('a'))
	assert.Equal(t, '7', ToLower('7'))
	assert.Equal(t, ' ', ToLower(' '))
}

func TestUnicode_ToLower_Ugly(t *testing.T) {
	assert.Equal(t, rune(0), ToLower(rune(0)))
	assert.Equal(t, rune(-1), ToLower(-1))
}
