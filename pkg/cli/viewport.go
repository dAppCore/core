package cli

import (
	"fmt"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"golang.org/x/term"
)

// viewportModel is the internal bubbletea model for scrollable content.
type viewportModel struct {
	title   string
	lines   []string
	offset  int
	height  int
	quitted bool
}

func newViewportModel(content, title string, height int) *viewportModel {
	lines := strings.Split(content, "\n")
	return &viewportModel{
		title:  title,
		lines:  lines,
		height: height,
	}
}

func (m *viewportModel) scrollDown() {
	maxOffset := len(m.lines) - m.height
	if maxOffset < 0 {
		maxOffset = 0
	}
	if m.offset < maxOffset {
		m.offset++
	}
}

func (m *viewportModel) scrollUp() {
	if m.offset > 0 {
		m.offset--
	}
}

func (m *viewportModel) Init() tea.Cmd {
	return nil
}

func (m *viewportModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.Type {
		case tea.KeyUp:
			m.scrollUp()
		case tea.KeyDown:
			m.scrollDown()
		case tea.KeyPgUp:
			for i := 0; i < m.height; i++ {
				m.scrollUp()
			}
		case tea.KeyPgDown:
			for i := 0; i < m.height; i++ {
				m.scrollDown()
			}
		case tea.KeyHome:
			m.offset = 0
		case tea.KeyEnd:
			maxOffset := len(m.lines) - m.height
			if maxOffset < 0 {
				maxOffset = 0
			}
			m.offset = maxOffset
		case tea.KeyEscape, tea.KeyCtrlC:
			m.quitted = true
			return m, tea.Quit
		case tea.KeyRunes:
			switch string(msg.Runes) {
			case "q":
				m.quitted = true
				return m, tea.Quit
			case "j":
				m.scrollDown()
			case "k":
				m.scrollUp()
			case "g":
				m.offset = 0
			case "G":
				maxOffset := len(m.lines) - m.height
				if maxOffset < 0 {
					maxOffset = 0
				}
				m.offset = maxOffset
			}
		}
	}
	return m, nil
}

func (m *viewportModel) View() string {
	var sb strings.Builder

	if m.title != "" {
		sb.WriteString(BoldStyle.Render(m.title) + "\n")
		sb.WriteString(DimStyle.Render(strings.Repeat("\u2500", len(m.title))) + "\n")
	}

	// Visible window
	end := m.offset + m.height
	if end > len(m.lines) {
		end = len(m.lines)
	}
	for _, line := range m.lines[m.offset:end] {
		sb.WriteString(line + "\n")
	}

	// Scroll indicator
	total := len(m.lines)
	if total > m.height {
		pct := (m.offset * 100) / (total - m.height)
		sb.WriteString(DimStyle.Render(fmt.Sprintf("\n%d%% (%d/%d lines)", pct, m.offset+m.height, total)))
	}

	sb.WriteString("\n" + DimStyle.Render("\u2191/\u2193 scroll \u2022 PgUp/PgDn page \u2022 q quit"))

	return sb.String()
}

// ViewportOption configures Viewport behaviour.
type ViewportOption func(*viewportConfig)

type viewportConfig struct {
	title  string
	height int
}

// WithViewportTitle sets the title shown above the viewport.
func WithViewportTitle(title string) ViewportOption {
	return func(c *viewportConfig) {
		c.title = title
	}
}

// WithViewportHeight sets the visible height in lines.
func WithViewportHeight(n int) ViewportOption {
	return func(c *viewportConfig) {
		c.height = n
	}
}

// Viewport displays scrollable content in the terminal.
// Falls back to printing the full content when stdin is not a terminal.
//
//	cli.Viewport(longContent, WithViewportTitle("Build Log"), WithViewportHeight(20))
func Viewport(content string, opts ...ViewportOption) error {
	cfg := &viewportConfig{
		height: 20,
	}
	for _, opt := range opts {
		opt(cfg)
	}

	// Fall back to plain output if not a terminal
	if !term.IsTerminal(0) {
		if cfg.title != "" {
			fmt.Println(BoldStyle.Render(cfg.title))
			fmt.Println(DimStyle.Render(strings.Repeat("\u2500", len(cfg.title))))
		}
		fmt.Println(content)
		return nil
	}

	m := newViewportModel(content, cfg.title, cfg.height)
	p := tea.NewProgram(m)
	_, err := p.Run()
	return err
}
