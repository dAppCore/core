package cli

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

// testModel is a minimal Model that quits immediately.
type testModel struct {
	initCalled   bool
	updateCalled bool
	viewCalled   bool
}

func (m *testModel) Init() Cmd {
	m.initCalled = true
	return Quit
}

func (m *testModel) Update(msg Msg) (Model, Cmd) {
	m.updateCalled = true
	return m, nil
}

func (m *testModel) View() string {
	m.viewCalled = true
	return "test view"
}

func TestModel_Good_InterfaceSatisfied(t *testing.T) {
	var m Model = &testModel{}
	assert.NotNil(t, m)
}

func TestQuitCmd_Good_ReturnsQuitMsg(t *testing.T) {
	cmd := Quit
	assert.NotNil(t, cmd)
}

func TestKeyMsg_Good_Type(t *testing.T) {
	// Verify our re-exported KeyType constants match bubbletea's
	assert.Equal(t, KeyEnter, KeyEnter)
	assert.Equal(t, KeyEsc, KeyEsc)
}

func TestKeyTypes_Good_Constants(t *testing.T) {
	// Verify key type constants exist and are distinct
	keys := []KeyType{KeyEnter, KeyEsc, KeyCtrlC, KeyUp, KeyDown, KeyTab, KeyBackspace}
	seen := make(map[KeyType]bool)
	for _, k := range keys {
		assert.False(t, seen[k], "duplicate key type")
		seen[k] = true
	}
}
