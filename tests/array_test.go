package core_test

import (
	"testing"

	. "forge.lthn.ai/core/go/pkg/core"
	"github.com/stretchr/testify/assert"
)

// --- Array[T] ---

func TestArray_New_Good(t *testing.T) {
	a := NewArray("a", "b", "c")
	assert.Equal(t, 3, a.Len())
}

func TestArray_Add_Good(t *testing.T) {
	a := NewArray[string]()
	a.Add("x", "y")
	assert.Equal(t, 2, a.Len())
	assert.True(t, a.Contains("x"))
	assert.True(t, a.Contains("y"))
}

func TestArray_AddUnique_Good(t *testing.T) {
	a := NewArray("a", "b")
	a.AddUnique("b", "c")
	assert.Equal(t, 3, a.Len())
}

func TestArray_Contains_Good(t *testing.T) {
	a := NewArray(1, 2, 3)
	assert.True(t, a.Contains(2))
	assert.False(t, a.Contains(99))
}

func TestArray_Filter_Good(t *testing.T) {
	a := NewArray(1, 2, 3, 4, 5)
	r := a.Filter(func(n int) bool { return n%2 == 0 })
	assert.True(t, r.OK)
	evens := r.Value.(*Array[int])
	assert.Equal(t, 2, evens.Len())
	assert.True(t, evens.Contains(2))
	assert.True(t, evens.Contains(4))
}

func TestArray_Each_Good(t *testing.T) {
	a := NewArray("a", "b", "c")
	var collected []string
	a.Each(func(s string) { collected = append(collected, s) })
	assert.Equal(t, []string{"a", "b", "c"}, collected)
}

func TestArray_Remove_Good(t *testing.T) {
	a := NewArray("a", "b", "c")
	a.Remove("b")
	assert.Equal(t, 2, a.Len())
	assert.False(t, a.Contains("b"))
}

func TestArray_Remove_Bad(t *testing.T) {
	a := NewArray("a", "b")
	a.Remove("missing")
	assert.Equal(t, 2, a.Len())
}

func TestArray_Deduplicate_Good(t *testing.T) {
	a := NewArray("a", "b", "a", "c", "b")
	a.Deduplicate()
	assert.Equal(t, 3, a.Len())
}

func TestArray_Clear_Good(t *testing.T) {
	a := NewArray(1, 2, 3)
	a.Clear()
	assert.Equal(t, 0, a.Len())
}

func TestArray_AsSlice_Good(t *testing.T) {
	a := NewArray("x", "y")
	s := a.AsSlice()
	assert.Equal(t, []string{"x", "y"}, s)
}

func TestArray_Empty_Good(t *testing.T) {
	a := NewArray[int]()
	assert.Equal(t, 0, a.Len())
	assert.False(t, a.Contains(0))
	assert.Equal(t, []int(nil), a.AsSlice())
}
