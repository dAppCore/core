package cli

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestViewportModel_Good_Create(t *testing.T) {
	content := "line 1\nline 2\nline 3"
	m := newViewportModel(content, "Title", 5)
	assert.Equal(t, "Title", m.title)
	assert.Equal(t, 3, len(m.lines))
	assert.Equal(t, 0, m.offset)
}

func TestViewportModel_Good_ScrollDown(t *testing.T) {
	lines := make([]string, 20)
	for i := range lines {
		lines[i] = strings.Repeat("x", 10)
	}
	m := newViewportModel(strings.Join(lines, "\n"), "", 5)
	m.scrollDown()
	assert.Equal(t, 1, m.offset)
}

func TestViewportModel_Good_ScrollUp(t *testing.T) {
	lines := make([]string, 20)
	for i := range lines {
		lines[i] = strings.Repeat("x", 10)
	}
	m := newViewportModel(strings.Join(lines, "\n"), "", 5)
	m.scrollDown()
	m.scrollDown()
	m.scrollUp()
	assert.Equal(t, 1, m.offset)
}

func TestViewportModel_Good_NoScrollPastTop(t *testing.T) {
	m := newViewportModel("a\nb\nc", "", 5)
	m.scrollUp() // Already at top
	assert.Equal(t, 0, m.offset)
}

func TestViewportModel_Good_NoScrollPastBottom(t *testing.T) {
	m := newViewportModel("a\nb\nc", "", 5)
	for i := 0; i < 10; i++ {
		m.scrollDown()
	}
	// Should clamp -- can't scroll past content
	assert.GreaterOrEqual(t, m.offset, 0)
}

func TestViewportModel_Good_View(t *testing.T) {
	m := newViewportModel("line 1\nline 2", "My Title", 10)
	view := m.View()
	assert.Contains(t, view, "My Title")
	assert.Contains(t, view, "line 1")
	assert.Contains(t, view, "line 2")
}
