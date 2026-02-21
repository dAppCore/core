package cli

import (
	tea "github.com/charmbracelet/bubbletea"
)

// Model is the interface for interactive TUI applications.
// It mirrors bubbletea's Model but uses our own types so domain
// packages never import bubbletea directly.
type Model interface {
	// Init returns an initial command to run.
	Init() Cmd

	// Update handles a message and returns the updated model and command.
	Update(msg Msg) (Model, Cmd)

	// View returns the string representation of the UI.
	View() string
}

// Msg is a message passed to Update. Can be any type.
type Msg = tea.Msg

// Cmd is a function that returns a message. Nil means no command.
type Cmd = tea.Cmd

// Quit is a command that tells the TUI to exit.
var Quit = tea.Quit

// KeyMsg represents a key press event.
type KeyMsg = tea.KeyMsg

// KeyType represents the type of key pressed.
type KeyType = tea.KeyType

// Key type constants.
const (
	KeyEnter     KeyType = tea.KeyEnter
	KeyEsc       KeyType = tea.KeyEscape
	KeyCtrlC     KeyType = tea.KeyCtrlC
	KeyUp        KeyType = tea.KeyUp
	KeyDown      KeyType = tea.KeyDown
	KeyLeft      KeyType = tea.KeyLeft
	KeyRight     KeyType = tea.KeyRight
	KeyTab       KeyType = tea.KeyTab
	KeyBackspace KeyType = tea.KeyBackspace
	KeySpace     KeyType = tea.KeySpace
	KeyHome      KeyType = tea.KeyHome
	KeyEnd       KeyType = tea.KeyEnd
	KeyPgUp      KeyType = tea.KeyPgUp
	KeyPgDown    KeyType = tea.KeyPgDown
	KeyDelete    KeyType = tea.KeyDelete
	KeyShiftTab  KeyType = tea.KeyShiftTab
	KeyRunes     KeyType = tea.KeyRunes
)

// adapter wraps our Model interface into a bubbletea.Model.
type adapter struct {
	inner Model
}

func (a adapter) Init() tea.Cmd {
	return a.inner.Init()
}

func (a adapter) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	m, cmd := a.inner.Update(msg)
	return adapter{inner: m}, cmd
}

func (a adapter) View() string {
	return a.inner.View()
}

// RunTUI runs an interactive TUI application using the provided Model.
// This is the escape hatch for complex interactive UIs that need the
// full bubbletea event loop. For simple spinners, progress bars, and
// lists, use the dedicated helpers instead.
//
//	err := cli.RunTUI(&myModel{items: items})
func RunTUI(m Model) error {
	p := tea.NewProgram(adapter{inner: m})
	_, err := p.Run()
	return err
}
