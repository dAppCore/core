package lthn

import (
	"sync"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestHash(t *testing.T) {
	hash := Hash("hello")
	assert.NotEmpty(t, hash)
}

func TestVerify(t *testing.T) {
	hash := Hash("hello")
	assert.True(t, Verify("hello", hash))
	assert.False(t, Verify("world", hash))
}

func TestCreateSalt_Good(t *testing.T) {
	// "hello" reversed: "olleh" -> "0113h"
	expected := "0113h"
	actual := createSalt("hello")
	assert.Equal(t, expected, actual, "Salt should be correctly created for 'hello'")
}

func TestCreateSalt_Bad(t *testing.T) {
	// Test with an empty string
	expected := ""
	actual := createSalt("")
	assert.Equal(t, expected, actual, "Salt for an empty string should be empty")
}

func TestCreateSalt_Ugly(t *testing.T) {
	// Test with characters not in the keyMap
	input := "world123"
	// "world123" reversed: "321dlrow" -> "e2ld1r0w"
	expected := "e2ld1r0w"
	actual := createSalt(input)
	assert.Equal(t, expected, actual, "Salt should handle characters not in the keyMap")

	// Test with only characters in the keyMap
	input = "oleta"
	// "oleta" reversed: "atelo" -> "47310"
	expected = "47310"
	actual = createSalt(input)
	assert.Equal(t, expected, actual, "Salt should correctly handle strings with only keyMap characters")
}

var testKeyMapMu sync.Mutex

func TestSetKeyMap(t *testing.T) {
	testKeyMapMu.Lock()
	originalKeyMap := GetKeyMap()
	t.Cleanup(func() {
		SetKeyMap(originalKeyMap)
		testKeyMapMu.Unlock()
	})

	newKeyMap := map[rune]rune{
		'a': 'b',
	}
	SetKeyMap(newKeyMap)
	assert.Equal(t, newKeyMap, GetKeyMap())
}
