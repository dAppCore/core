package cli

import (
	"fmt"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"golang.org/x/term"
)

// textInputModel is the internal bubbletea model for text input.
type textInputModel struct {
	title       string
	placeholder string
	value       string
	masked      bool
	submitted   bool
	cancelled   bool
	cursorPos   int
	validator   func(string) error
	err         error
}

func newTextInputModel(title, placeholder string) *textInputModel {
	return &textInputModel{
		title:       title,
		placeholder: placeholder,
	}
}

func (m *textInputModel) insertChar(ch rune) {
	m.value = m.value[:m.cursorPos] + string(ch) + m.value[m.cursorPos:]
	m.cursorPos++
}

func (m *textInputModel) backspace() {
	if m.cursorPos > 0 {
		m.value = m.value[:m.cursorPos-1] + m.value[m.cursorPos:]
		m.cursorPos--
	}
}

func (m *textInputModel) Init() tea.Cmd {
	return nil
}

func (m *textInputModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.Type {
		case tea.KeyEnter:
			if m.validator != nil {
				if err := m.validator(m.value); err != nil {
					m.err = err
					return m, nil
				}
			}
			if m.value == "" && m.placeholder != "" {
				m.value = m.placeholder
			}
			m.submitted = true
			return m, tea.Quit
		case tea.KeyEscape, tea.KeyCtrlC:
			m.cancelled = true
			return m, tea.Quit
		case tea.KeyBackspace:
			m.backspace()
			m.err = nil
		case tea.KeyLeft:
			if m.cursorPos > 0 {
				m.cursorPos--
			}
		case tea.KeyRight:
			if m.cursorPos < len(m.value) {
				m.cursorPos++
			}
		case tea.KeyRunes:
			for _, ch := range msg.Runes {
				m.insertChar(ch)
			}
			m.err = nil
		}
	}
	return m, nil
}

func (m *textInputModel) View() string {
	var sb strings.Builder

	sb.WriteString(BoldStyle.Render(m.title) + "\n\n")

	display := m.value
	if m.masked {
		display = strings.Repeat("*", len(m.value))
	}

	if display == "" && m.placeholder != "" {
		sb.WriteString(DimStyle.Render(m.placeholder))
	} else {
		sb.WriteString(display)
	}
	sb.WriteString(AccentStyle.Render("\u2588")) // Cursor block

	if m.err != nil {
		sb.WriteString("\n" + ErrorStyle.Render(fmt.Sprintf("  %s", m.err)))
	}

	sb.WriteString("\n\n" + DimStyle.Render("enter submit \u2022 esc cancel"))

	return sb.String()
}

// TextInputOption configures TextInput behaviour.
type TextInputOption func(*textInputConfig)

type textInputConfig struct {
	placeholder string
	masked      bool
	validator   func(string) error
}

// WithTextPlaceholder sets placeholder text shown when input is empty.
func WithTextPlaceholder(text string) TextInputOption {
	return func(c *textInputConfig) {
		c.placeholder = text
	}
}

// WithMask hides input characters (for passwords).
func WithMask() TextInputOption {
	return func(c *textInputConfig) {
		c.masked = true
	}
}

// WithInputValidator adds a validation function for the input.
func WithInputValidator(fn func(string) error) TextInputOption {
	return func(c *textInputConfig) {
		c.validator = fn
	}
}

// TextInput presents a styled text input prompt and returns the entered value.
// Returns empty string if cancelled.
//
// Falls back to Question() when stdin is not a terminal.
//
//	name, err := cli.TextInput("Enter your name:", cli.WithTextPlaceholder("Anonymous"))
//	pass, err := cli.TextInput("Password:", cli.WithMask())
func TextInput(title string, opts ...TextInputOption) (string, error) {
	cfg := &textInputConfig{}
	for _, opt := range opts {
		opt(cfg)
	}

	// Fall back to simple Question if not a terminal
	if !term.IsTerminal(0) {
		var qopts []QuestionOption
		if cfg.placeholder != "" {
			qopts = append(qopts, WithDefault(cfg.placeholder))
		}
		if cfg.validator != nil {
			qopts = append(qopts, WithValidator(cfg.validator))
		}
		return Question(title, qopts...), nil
	}

	m := newTextInputModel(title, cfg.placeholder)
	m.masked = cfg.masked
	m.validator = cfg.validator

	p := tea.NewProgram(m)
	finalModel, err := p.Run()
	if err != nil {
		return "", err
	}

	final := finalModel.(*textInputModel)
	if final.cancelled {
		return "", nil
	}
	return final.value, nil
}
