package cli

import (
	"fmt"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"golang.org/x/term"
)

// listModel is the internal bubbletea model for interactive list selection.
type listModel struct {
	items    []string
	cursor   int
	title    string
	selected bool
	quitted  bool
}

func newListModel(items []string, title string) *listModel {
	return &listModel{
		items: items,
		title: title,
	}
}

func (m *listModel) moveDown() {
	m.cursor++
	if m.cursor >= len(m.items) {
		m.cursor = 0
	}
}

func (m *listModel) moveUp() {
	m.cursor--
	if m.cursor < 0 {
		m.cursor = len(m.items) - 1
	}
}

func (m *listModel) Init() tea.Cmd {
	return nil
}

func (m *listModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.Type {
		case tea.KeyUp, tea.KeyShiftTab:
			m.moveUp()
		case tea.KeyDown, tea.KeyTab:
			m.moveDown()
		case tea.KeyEnter:
			m.selected = true
			return m, tea.Quit
		case tea.KeyEscape, tea.KeyCtrlC:
			m.quitted = true
			return m, tea.Quit
		case tea.KeyRunes:
			switch string(msg.Runes) {
			case "j":
				m.moveDown()
			case "k":
				m.moveUp()
			}
		}
	}
	return m, nil
}

func (m *listModel) View() string {
	var sb strings.Builder

	if m.title != "" {
		sb.WriteString(BoldStyle.Render(m.title) + "\n\n")
	}

	for i, item := range m.items {
		cursor := "  "
		style := DimStyle
		if i == m.cursor {
			cursor = AccentStyle.Render(Glyph(":pointer:")) + " "
			style = BoldStyle
		}
		sb.WriteString(fmt.Sprintf("%s%s\n", cursor, style.Render(item)))
	}

	sb.WriteString("\n" + DimStyle.Render("↑/↓ navigate • enter select • esc cancel"))

	return sb.String()
}

// ListOption configures InteractiveList behaviour.
type ListOption func(*listConfig)

type listConfig struct {
	height int
}

// WithListHeight sets the visible height of the list (number of items shown).
func WithListHeight(n int) ListOption {
	return func(c *listConfig) {
		c.height = n
	}
}

// InteractiveList presents an interactive scrollable list and returns the
// selected item's index and value. Returns -1 and empty string if cancelled.
//
// Falls back to numbered Select() when stdin is not a terminal (e.g. piped input).
//
//	idx, value := cli.InteractiveList("Pick a repo:", repos)
func InteractiveList(title string, items []string, opts ...ListOption) (int, string) {
	if len(items) == 0 {
		return -1, ""
	}

	// Fall back to simple Select if not a terminal
	if !term.IsTerminal(0) {
		result, err := Select(title, items)
		if err != nil {
			return -1, ""
		}
		for i, item := range items {
			if item == result {
				return i, result
			}
		}
		return -1, ""
	}

	m := newListModel(items, title)
	p := tea.NewProgram(m)
	finalModel, err := p.Run()
	if err != nil {
		return -1, ""
	}

	final := finalModel.(*listModel)
	if final.quitted || !final.selected {
		return -1, ""
	}
	return final.cursor, final.items[final.cursor]
}
