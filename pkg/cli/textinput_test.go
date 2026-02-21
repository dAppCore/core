package cli

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestTextInputModel_Good_Create(t *testing.T) {
	m := newTextInputModel("Enter name:", "")
	assert.Equal(t, "Enter name:", m.title)
	assert.Equal(t, "", m.value)
}

func TestTextInputModel_Good_WithPlaceholder(t *testing.T) {
	m := newTextInputModel("Name:", "John")
	assert.Equal(t, "John", m.placeholder)
}

func TestTextInputModel_Good_TypeCharacters(t *testing.T) {
	m := newTextInputModel("Name:", "")
	m.insertChar('H')
	m.insertChar('i')
	assert.Equal(t, "Hi", m.value)
}

func TestTextInputModel_Good_Backspace(t *testing.T) {
	m := newTextInputModel("Name:", "")
	m.insertChar('A')
	m.insertChar('B')
	m.backspace()
	assert.Equal(t, "A", m.value)
}

func TestTextInputModel_Good_BackspaceEmpty(t *testing.T) {
	m := newTextInputModel("Name:", "")
	m.backspace() // Should not panic
	assert.Equal(t, "", m.value)
}

func TestTextInputModel_Good_Masked(t *testing.T) {
	m := newTextInputModel("Password:", "")
	m.masked = true
	m.insertChar('s')
	m.insertChar('e')
	m.insertChar('c')
	assert.Equal(t, "sec", m.value) // Internal value is real
	view := m.View()
	assert.NotContains(t, view, "sec") // Display is masked
	assert.Contains(t, view, "***")
}

func TestTextInputModel_Good_View(t *testing.T) {
	m := newTextInputModel("Enter:", "")
	m.insertChar('X')
	view := m.View()
	assert.Contains(t, view, "Enter:")
	assert.Contains(t, view, "X")
}
