package core_test

import (
	"testing"

	. "dappco.re/go/core"
	"github.com/stretchr/testify/assert"
)

func TestMap_MapKeys_Good(t *testing.T) {
	keys := MapKeys(map[string]int{"a": 1, "b": 2})

	assert.ElementsMatch(t, []string{"a", "b"}, keys)
}

func TestMap_MapKeys_Bad(t *testing.T) {
	keys := MapKeys[string, int](nil)

	assert.Empty(t, keys)
}

func TestMap_MapKeys_Ugly(t *testing.T) {
	keys := MapKeys(map[int]string{2: "b", 1: "a", 3: "c"})

	SliceSort(keys)
	assert.Equal(t, []int{1, 2, 3}, keys)
}

func TestMap_MapValues_Good(t *testing.T) {
	values := MapValues(map[string]int{"a": 1, "b": 2})

	assert.ElementsMatch(t, []int{1, 2}, values)
}

func TestMap_MapValues_Bad(t *testing.T) {
	values := MapValues[string, int](nil)

	assert.Empty(t, values)
}

func TestMap_MapValues_Ugly(t *testing.T) {
	values := MapValues(map[int]string{2: "b", 1: "a", 3: "c"})

	SliceSort(values)
	assert.Equal(t, []string{"a", "b", "c"}, values)
}

func TestMap_MapClone_Good(t *testing.T) {
	clone := MapClone(map[string]int{"a": 1})

	assert.Equal(t, map[string]int{"a": 1}, clone)
}

func TestMap_MapClone_Bad(t *testing.T) {
	clone := MapClone[string, int](nil)

	assert.Nil(t, clone)
}

func TestMap_MapClone_Ugly(t *testing.T) {
	original := map[string]int{"a": 1}
	clone := MapClone(original)
	clone["a"] = 2

	assert.Equal(t, 1, original["a"])
	assert.Equal(t, 2, clone["a"])
}
