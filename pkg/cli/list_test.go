package cli

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestListModel_Good_Create(t *testing.T) {
	items := []string{"alpha", "beta", "gamma"}
	m := newListModel(items, "Pick one:")
	assert.Equal(t, 3, len(m.items))
	assert.Equal(t, 0, m.cursor)
	assert.Equal(t, "Pick one:", m.title)
}

func TestListModel_Good_MoveDown(t *testing.T) {
	m := newListModel([]string{"a", "b", "c"}, "")
	m.moveDown()
	assert.Equal(t, 1, m.cursor)
	m.moveDown()
	assert.Equal(t, 2, m.cursor)
}

func TestListModel_Good_MoveUp(t *testing.T) {
	m := newListModel([]string{"a", "b", "c"}, "")
	m.moveDown()
	m.moveDown()
	m.moveUp()
	assert.Equal(t, 1, m.cursor)
}

func TestListModel_Good_WrapAround(t *testing.T) {
	m := newListModel([]string{"a", "b", "c"}, "")
	m.moveUp() // Should wrap to bottom
	assert.Equal(t, 2, m.cursor)
}

func TestListModel_Good_View(t *testing.T) {
	m := newListModel([]string{"alpha", "beta"}, "Choose:")
	view := m.View()
	assert.Contains(t, view, "Choose:")
	assert.Contains(t, view, "alpha")
	assert.Contains(t, view, "beta")
}

func TestListModel_Good_Selected(t *testing.T) {
	m := newListModel([]string{"a", "b", "c"}, "")
	m.moveDown()
	m.selected = true
	assert.Equal(t, "b", m.items[m.cursor])
}
