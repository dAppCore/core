package core_test

import (
	. "dappco.re/go/core"
)

// --- Unicode Operations ---

func TestUnicode_IsLetter_Good(t *T) {
	AssertTrue(t, IsLetter('A'))
	AssertTrue(t, IsLetter('é'))
	AssertTrue(t, IsLetter('界'))
}

func TestUnicode_IsLetter_Bad(t *T) {
	AssertFalse(t, IsLetter('7'))
	AssertFalse(t, IsLetter('_'))
	AssertFalse(t, IsLetter(' '))
}

func TestUnicode_IsLetter_Ugly(t *T) {
	AssertFalse(t, IsLetter(rune(0)))
	AssertFalse(t, IsLetter(-1))
}

func TestUnicode_IsDigit_Good(t *T) {
	AssertTrue(t, IsDigit('0'))
	AssertTrue(t, IsDigit('9'))
	AssertTrue(t, IsDigit('٥'))
}

func TestUnicode_IsDigit_Bad(t *T) {
	AssertFalse(t, IsDigit('A'))
	AssertFalse(t, IsDigit('.'))
	AssertFalse(t, IsDigit(' '))
}

func TestUnicode_IsDigit_Ugly(t *T) {
	AssertFalse(t, IsDigit(rune(0)))
	AssertFalse(t, IsDigit(-1))
}

func TestUnicode_IsSpace_Good(t *T) {
	AssertTrue(t, IsSpace(' '))
	AssertTrue(t, IsSpace('\n'))
	AssertTrue(t, IsSpace('\u00a0'))
}

func TestUnicode_IsSpace_Bad(t *T) {
	AssertFalse(t, IsSpace('A'))
	AssertFalse(t, IsSpace('0'))
	AssertFalse(t, IsSpace('_'))
}

func TestUnicode_IsSpace_Ugly(t *T) {
	AssertFalse(t, IsSpace(rune(0)))
	AssertFalse(t, IsSpace(-1))
}

func TestUnicode_IsUpper_Good(t *T) {
	AssertTrue(t, IsUpper('A'))
	AssertTrue(t, IsUpper('É'))
	AssertTrue(t, IsUpper('Ω'))
}

func TestUnicode_IsUpper_Bad(t *T) {
	AssertFalse(t, IsUpper('a'))
	AssertFalse(t, IsUpper('7'))
	AssertFalse(t, IsUpper(' '))
}

func TestUnicode_IsUpper_Ugly(t *T) {
	AssertFalse(t, IsUpper(rune(0)))
	AssertFalse(t, IsUpper(-1))
}

func TestUnicode_IsLower_Good(t *T) {
	AssertTrue(t, IsLower('a'))
	AssertTrue(t, IsLower('é'))
	AssertTrue(t, IsLower('ω'))
}

func TestUnicode_IsLower_Bad(t *T) {
	AssertFalse(t, IsLower('A'))
	AssertFalse(t, IsLower('7'))
	AssertFalse(t, IsLower(' '))
}

func TestUnicode_IsLower_Ugly(t *T) {
	AssertFalse(t, IsLower(rune(0)))
	AssertFalse(t, IsLower(-1))
}

func TestUnicode_ToUpper_Good(t *T) {
	AssertEqual(t, 'A', ToUpper('a'))
	AssertEqual(t, 'É', ToUpper('é'))
	AssertEqual(t, 'Ω', ToUpper('ω'))
}

func TestUnicode_ToUpper_Bad(t *T) {
	AssertEqual(t, 'A', ToUpper('A'))
	AssertEqual(t, '7', ToUpper('7'))
	AssertEqual(t, ' ', ToUpper(' '))
}

func TestUnicode_ToUpper_Ugly(t *T) {
	AssertEqual(t, rune(0), ToUpper(rune(0)))
	AssertEqual(t, rune(-1), ToUpper(-1))
}

func TestUnicode_ToLower_Good(t *T) {
	AssertEqual(t, 'a', ToLower('A'))
	AssertEqual(t, 'é', ToLower('É'))
	AssertEqual(t, 'ω', ToLower('Ω'))
}

func TestUnicode_ToLower_Bad(t *T) {
	AssertEqual(t, 'a', ToLower('a'))
	AssertEqual(t, '7', ToLower('7'))
	AssertEqual(t, ' ', ToLower(' '))
}

func TestUnicode_ToLower_Ugly(t *T) {
	AssertEqual(t, rune(0), ToLower(rune(0)))
	AssertEqual(t, rune(-1), ToLower(-1))
}
