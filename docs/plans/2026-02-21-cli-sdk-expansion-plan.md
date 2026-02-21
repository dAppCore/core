# CLI SDK Expansion (Phase 0) Implementation Plan

> **For Claude:** REQUIRED SUB-SKILL: Use superpowers:executing-plans to implement this plan task-by-task.

**Goal:** Extend `go/pkg/cli` with charmbracelet TUI primitives (Spinner, ProgressBar, List, TextInput, Viewport) so domain repos never import anything but `forge.lthn.ai/core/go/pkg/cli` for CLI concerns.

**Architecture:** Each TUI primitive gets its own file in `pkg/cli/`. Charmbracelet libraries (bubbletea, bubbles, lipgloss) are imported only inside `pkg/cli/` — the public API uses our own types. Stubs for future features (Form, FilePicker, Tabs) define the interface but fall back to simple bufio implementations until charm backends are wired in later.

**Tech Stack:** `charmbracelet/bubbletea` (app loop), `charmbracelet/bubbles` (spinner, progress, list, textinput, viewport), `charmbracelet/lipgloss` (styling — replaces our ANSI builder long-term)

---

## Context

`go/pkg/cli` currently provides:
- Cobra wrappers: `Command`, `NewCommand()`, `NewGroup()`, flag helpers
- Output: `Success()`, `Error()`, `Table`, `Section()`, `Label()`
- Prompts: `Confirm()`, `Question()`, `Choose()`, `ChooseMulti()` (all bufio-based)
- Styles: `AnsiStyle` builder with 17 pre-built styles, 47 Tailwind colour constants
- Glyphs: `:check:`, `:cross:` etc. with theme switching
- Layout: HLCRF composite renderer

Zero charmbracelet dependencies exist today. All styling is pure ANSI escape codes.

The 34 files in `cli/cmd/*` that import `github.com/spf13/cobra` directly need `cli.*` equivalents. This plan does NOT migrate those files — it builds the SDK surface they'll need. Migration happens in Phase 1+.

## Critical Files

All changes are in `/Users/snider/Code/host-uk/core/pkg/cli/`:

- `spinner.go` + `spinner_test.go` — Async spinner
- `progress.go` + `progress_test.go` — Progress bar
- `list.go` + `list_test.go` — Interactive scrollable list
- `textinput.go` + `textinput_test.go` — Styled text input
- `viewport.go` + `viewport_test.go` — Scrollable content pane
- `tui.go` + `tui_test.go` — RunTUI escape hatch + Model interface
- `stubs.go` + `stubs_test.go` — Form, FilePicker, Tabs interfaces (simple fallback)

---

### Task 1: Add charmbracelet dependencies

**Files:**
- Modify: `/Users/snider/Code/host-uk/core/go.mod`

**Step 1: Add bubbletea, bubbles, and lipgloss**

Run:
```bash
cd /Users/snider/Code/host-uk/core && go get github.com/charmbracelet/bubbletea/v2@latest github.com/charmbracelet/bubbles/v2@latest github.com/charmbracelet/lipgloss/v2@latest
```

**Step 2: Verify module resolves**

Run: `cd /Users/snider/Code/host-uk/core && go mod tidy`
Expected: Clean, no errors.

**Step 3: Verify existing tests still pass**

Run: `cd /Users/snider/Code/host-uk/core && go test ./pkg/cli/...`
Expected: All existing tests pass (no behaviour changed).

**Step 4: Commit**

```bash
cd /Users/snider/Code/host-uk/core && git add go.mod go.sum && git commit -m "chore(cli): add charmbracelet dependencies (bubbletea, bubbles, lipgloss)

Co-Authored-By: Virgil <virgil@lethean.io>"
```

---

### Task 2: Spinner

**Files:**
- Create: `/Users/snider/Code/host-uk/core/pkg/cli/spinner.go`
- Create: `/Users/snider/Code/host-uk/core/pkg/cli/spinner_test.go`

A non-blocking spinner that runs in a goroutine. The caller gets a handle to update the message, mark it done, or mark it failed. Uses `bubbles/spinner` internally.

**Step 1: Write the tests**

```go
// spinner_test.go
package cli

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestSpinner_Good_CreateAndStop(t *testing.T) {
	s := NewSpinner("Loading...")
	require.NotNil(t, s)
	assert.Equal(t, "Loading...", s.Message())
	s.Stop()
}

func TestSpinner_Good_UpdateMessage(t *testing.T) {
	s := NewSpinner("Step 1")
	s.Update("Step 2")
	assert.Equal(t, "Step 2", s.Message())
	s.Stop()
}

func TestSpinner_Good_Done(t *testing.T) {
	s := NewSpinner("Building")
	s.Done("Build complete")
	// After Done, spinner is stopped — calling Stop again is safe
	s.Stop()
}

func TestSpinner_Good_Fail(t *testing.T) {
	s := NewSpinner("Checking")
	s.Fail("Check failed")
	s.Stop()
}

func TestSpinner_Good_DoubleStop(t *testing.T) {
	s := NewSpinner("Loading")
	s.Stop()
	s.Stop() // Should not panic
}
```

**Step 2: Run tests to verify they fail**

Run: `cd /Users/snider/Code/host-uk/core && go test -run TestSpinner ./pkg/cli/...`
Expected: FAIL — `NewSpinner` undefined.

**Step 3: Write the implementation**

```go
// spinner.go
package cli

import (
	"fmt"
	"sync"
	"time"
)

// SpinnerHandle controls a running spinner.
type SpinnerHandle struct {
	mu      sync.Mutex
	message string
	done    bool
	ticker  *time.Ticker
	stopCh  chan struct{}
}

// NewSpinner starts an async spinner with the given message.
// Call Stop(), Done(), or Fail() to stop it.
func NewSpinner(message string) *SpinnerHandle {
	s := &SpinnerHandle{
		message: message,
		ticker:  time.NewTicker(100 * time.Millisecond),
		stopCh:  make(chan struct{}),
	}

	frames := []string{"⠋", "⠙", "⠹", "⠸", "⠼", "⠴", "⠦", "⠧", "⠇", "⠏"}
	if !ColorEnabled() {
		frames = []string{"|", "/", "-", "\\"}
	}

	go func() {
		i := 0
		for {
			select {
			case <-s.stopCh:
				return
			case <-s.ticker.C:
				s.mu.Lock()
				if !s.done {
					fmt.Printf("\033[2K\r%s %s", DimStyle.Render(frames[i%len(frames)]), s.message)
				}
				s.mu.Unlock()
				i++
			}
		}
	}()

	return s
}

// Message returns the current spinner message.
func (s *SpinnerHandle) Message() string {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.message
}

// Update changes the spinner message.
func (s *SpinnerHandle) Update(message string) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.message = message
}

// Stop stops the spinner silently (clears the line).
func (s *SpinnerHandle) Stop() {
	s.mu.Lock()
	defer s.mu.Unlock()
	if s.done {
		return
	}
	s.done = true
	s.ticker.Stop()
	close(s.stopCh)
	fmt.Print("\033[2K\r")
}

// Done stops the spinner with a success message.
func (s *SpinnerHandle) Done(message string) {
	s.mu.Lock()
	alreadyDone := s.done
	s.done = true
	s.mu.Unlock()

	if alreadyDone {
		return
	}
	s.ticker.Stop()
	close(s.stopCh)
	fmt.Printf("\033[2K\r%s\n", SuccessStyle.Render(Glyph(":check:")+" "+message))
}

// Fail stops the spinner with an error message.
func (s *SpinnerHandle) Fail(message string) {
	s.mu.Lock()
	alreadyDone := s.done
	s.done = true
	s.mu.Unlock()

	if alreadyDone {
		return
	}
	s.ticker.Stop()
	close(s.stopCh)
	fmt.Printf("\033[2K\r%s\n", ErrorStyle.Render(Glyph(":cross:")+" "+message))
}
```

Note: This initial implementation uses a goroutine + ticker rather than bubbletea, keeping it simple and non-blocking. The bubbletea spinner can replace the internals later without changing the public API.

**Step 4: Run tests to verify they pass**

Run: `cd /Users/snider/Code/host-uk/core && go test -run TestSpinner ./pkg/cli/... -v`
Expected: All 5 tests PASS.

**Step 5: Commit**

```bash
cd /Users/snider/Code/host-uk/core && git add pkg/cli/spinner.go pkg/cli/spinner_test.go && git commit -m "feat(cli): add Spinner with async handle (Update, Done, Fail)

Co-Authored-By: Virgil <virgil@lethean.io>"
```

---

### Task 3: ProgressBar

**Files:**
- Create: `/Users/snider/Code/host-uk/core/pkg/cli/progressbar.go`
- Create: `/Users/snider/Code/host-uk/core/pkg/cli/progressbar_test.go`

A progress bar that renders inline. Shows percentage, bar, and optional message.

**Step 1: Write the tests**

```go
// progressbar_test.go
package cli

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestProgressBar_Good_Create(t *testing.T) {
	pb := NewProgressBar(100)
	require.NotNil(t, pb)
	assert.Equal(t, 0, pb.Current())
	assert.Equal(t, 100, pb.Total())
}

func TestProgressBar_Good_Increment(t *testing.T) {
	pb := NewProgressBar(10)
	pb.Increment()
	assert.Equal(t, 1, pb.Current())
	pb.Increment()
	assert.Equal(t, 2, pb.Current())
}

func TestProgressBar_Good_SetMessage(t *testing.T) {
	pb := NewProgressBar(10)
	pb.SetMessage("Processing file.go")
	assert.Equal(t, "Processing file.go", pb.message)
}

func TestProgressBar_Good_Set(t *testing.T) {
	pb := NewProgressBar(100)
	pb.Set(50)
	assert.Equal(t, 50, pb.Current())
}

func TestProgressBar_Good_Done(t *testing.T) {
	pb := NewProgressBar(5)
	for i := 0; i < 5; i++ {
		pb.Increment()
	}
	pb.Done()
	// After Done, Current == Total
	assert.Equal(t, 5, pb.Current())
}

func TestProgressBar_Bad_ExceedsTotal(t *testing.T) {
	pb := NewProgressBar(2)
	pb.Increment()
	pb.Increment()
	pb.Increment() // Should clamp to total
	assert.Equal(t, 2, pb.Current())
}

func TestProgressBar_Good_Render(t *testing.T) {
	pb := NewProgressBar(10)
	pb.Set(5)
	rendered := pb.String()
	assert.Contains(t, rendered, "50%")
}
```

**Step 2: Run tests to verify they fail**

Run: `cd /Users/snider/Code/host-uk/core && go test -run TestProgressBar ./pkg/cli/... -v`
Expected: FAIL — `NewProgressBar` undefined.

**Step 3: Write the implementation**

```go
// progressbar.go
package cli

import (
	"fmt"
	"strings"
	"sync"
)

// ProgressHandle controls a progress bar.
type ProgressHandle struct {
	mu      sync.Mutex
	current int
	total   int
	message string
	width   int
}

// NewProgressBar creates a new progress bar with the given total.
func NewProgressBar(total int) *ProgressHandle {
	return &ProgressHandle{
		total: total,
		width: 30,
	}
}

// Current returns the current progress value.
func (p *ProgressHandle) Current() int {
	p.mu.Lock()
	defer p.mu.Unlock()
	return p.current
}

// Total returns the total value.
func (p *ProgressHandle) Total() int {
	return p.total
}

// Increment advances the progress by 1.
func (p *ProgressHandle) Increment() {
	p.mu.Lock()
	defer p.mu.Unlock()
	if p.current < p.total {
		p.current++
	}
	p.render()
}

// Set sets the progress to a specific value.
func (p *ProgressHandle) Set(n int) {
	p.mu.Lock()
	defer p.mu.Unlock()
	if n > p.total {
		n = p.total
	}
	if n < 0 {
		n = 0
	}
	p.current = n
	p.render()
}

// SetMessage sets the message displayed alongside the bar.
func (p *ProgressHandle) SetMessage(msg string) {
	p.mu.Lock()
	defer p.mu.Unlock()
	p.message = msg
	p.render()
}

// Done completes the progress bar and moves to a new line.
func (p *ProgressHandle) Done() {
	p.mu.Lock()
	defer p.mu.Unlock()
	p.current = p.total
	p.render()
	fmt.Println()
}

// String returns the rendered progress bar without ANSI cursor control.
func (p *ProgressHandle) String() string {
	pct := 0
	if p.total > 0 {
		pct = (p.current * 100) / p.total
	}

	filled := (p.width * p.current) / p.total
	if filled > p.width {
		filled = p.width
	}
	empty := p.width - filled

	bar := "[" + strings.Repeat("█", filled) + strings.Repeat("░", empty) + "]"

	if p.message != "" {
		return fmt.Sprintf("%s %3d%% %s", bar, pct, p.message)
	}
	return fmt.Sprintf("%s %3d%%", bar, pct)
}

// render outputs the progress bar, overwriting the current line.
func (p *ProgressHandle) render() {
	fmt.Printf("\033[2K\r%s", p.String())
}
```

**Step 4: Run tests to verify they pass**

Run: `cd /Users/snider/Code/host-uk/core && go test -run TestProgressBar ./pkg/cli/... -v`
Expected: All 7 tests PASS.

**Step 5: Commit**

```bash
cd /Users/snider/Code/host-uk/core && git add pkg/cli/progressbar.go pkg/cli/progressbar_test.go && git commit -m "feat(cli): add ProgressBar with Increment, Set, SetMessage, Done

Co-Authored-By: Virgil <virgil@lethean.io>"
```

---

### Task 4: TUI runner (RunTUI + Model interface)

**Files:**
- Create: `/Users/snider/Code/host-uk/core/pkg/cli/tui.go`
- Create: `/Users/snider/Code/host-uk/core/pkg/cli/tui_test.go`

The escape hatch for complex interactive UIs. Wraps `bubbletea.Program` behind our own `Model` interface so domain packages never import bubbletea directly.

**Step 1: Write the tests**

```go
// tui_test.go
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

func TestKeyMsg_Good_String(t *testing.T) {
	k := KeyMsg{Type: KeyEnter}
	assert.Equal(t, KeyEnter, k.Type)
}

func TestKeyTypes_Good_Constants(t *testing.T) {
	// Verify key type constants exist
	assert.NotEmpty(t, string(KeyEnter))
	assert.NotEmpty(t, string(KeyEsc))
	assert.NotEmpty(t, string(KeyCtrlC))
	assert.NotEmpty(t, string(KeyUp))
	assert.NotEmpty(t, string(KeyDown))
	assert.NotEmpty(t, string(KeyTab))
	assert.NotEmpty(t, string(KeyBackspace))
}
```

**Step 2: Run tests to verify they fail**

Run: `cd /Users/snider/Code/host-uk/core && go test -run TestModel ./pkg/cli/... -v && go test -run TestQuit ./pkg/cli/... -v && go test -run TestKey ./pkg/cli/... -v`
Expected: FAIL — types undefined.

**Step 3: Write the implementation**

```go
// tui.go
package cli

import (
	tea "github.com/charmbracelet/bubbletea/v2"
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
```

**Step 4: Run tests to verify they pass**

Run: `cd /Users/snider/Code/host-uk/core && go test -run "TestModel|TestQuit|TestKey" ./pkg/cli/... -v`
Expected: All 4 tests PASS.

**Step 5: Commit**

```bash
cd /Users/snider/Code/host-uk/core && git add pkg/cli/tui.go pkg/cli/tui_test.go && git commit -m "feat(cli): add RunTUI escape hatch with Model/Msg/Cmd/KeyMsg types

Wraps bubbletea behind our own interface so domain packages
never import charmbracelet directly.

Co-Authored-By: Virgil <virgil@lethean.io>"
```

---

### Task 5: Interactive List

**Files:**
- Create: `/Users/snider/Code/host-uk/core/pkg/cli/list.go`
- Create: `/Users/snider/Code/host-uk/core/pkg/cli/list_test.go`

An interactive scrollable list for terminal selection. Uses our `RunTUI` internally. Falls back to numbered `Select()` when stdin is not a terminal.

**Step 1: Write the tests**

```go
// list_test.go
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
```

**Step 2: Run tests to verify they fail**

Run: `cd /Users/snider/Code/host-uk/core && go test -run TestListModel ./pkg/cli/... -v`
Expected: FAIL — `newListModel` undefined.

**Step 3: Write the implementation**

```go
// list.go
package cli

import (
	"fmt"
	"strings"

	tea "github.com/charmbracelet/bubbletea/v2"
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
		default:
			// Handle j/k vim-style navigation
			if msg.String() == "j" {
				m.moveDown()
			} else if msg.String() == "k" {
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

// ListOption configures List behaviour.
type ListOption func(*listConfig)

type listConfig struct {
	height int
}

// WithHeight sets the visible height of the list (number of items shown).
func WithHeight(n int) ListOption {
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
	if !term.IsTerminal(int(StdinFd())) {
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

// StdinFd returns the file descriptor for stdin.
// Extracted for testing.
func StdinFd() uintptr {
	return uintptr(0) // stdin
}
```

**Step 4: Run tests to verify they pass**

Run: `cd /Users/snider/Code/host-uk/core && go test -run TestListModel ./pkg/cli/... -v`
Expected: All 6 tests PASS.

**Step 5: Commit**

```bash
cd /Users/snider/Code/host-uk/core && git add pkg/cli/list.go pkg/cli/list_test.go && git commit -m "feat(cli): add InteractiveList with keyboard navigation and terminal fallback

Co-Authored-By: Virgil <virgil@lethean.io>"
```

---

### Task 6: TextInput

**Files:**
- Create: `/Users/snider/Code/host-uk/core/pkg/cli/textinput.go`
- Create: `/Users/snider/Code/host-uk/core/pkg/cli/textinput_test.go`

A styled single-line text input with placeholder, validation, and optional masking (for passwords). Falls back to `Question()` when stdin is not a terminal.

**Step 1: Write the tests**

```go
// textinput_test.go
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
```

**Step 2: Run tests to verify they fail**

Run: `cd /Users/snider/Code/host-uk/core && go test -run TestTextInputModel ./pkg/cli/... -v`
Expected: FAIL — `newTextInputModel` undefined.

**Step 3: Write the implementation**

```go
// textinput.go
package cli

import (
	"fmt"
	"strings"

	tea "github.com/charmbracelet/bubbletea/v2"
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
		default:
			if msg.Text != "" {
				for _, ch := range msg.Text {
					m.insertChar(ch)
				}
				m.err = nil
			}
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
	sb.WriteString(AccentStyle.Render("█")) // Cursor

	if m.err != nil {
		sb.WriteString("\n" + ErrorStyle.Render(fmt.Sprintf("  %s", m.err)))
	}

	sb.WriteString("\n\n" + DimStyle.Render("enter submit • esc cancel"))

	return sb.String()
}

// TextInputOption configures TextInput behaviour.
type TextInputOption func(*textInputConfig)

type textInputConfig struct {
	placeholder string
	masked      bool
	validator   func(string) error
}

// WithPlaceholder sets placeholder text shown when input is empty.
func WithPlaceholder(text string) TextInputOption {
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
//	name, err := cli.TextInput("Enter your name:", WithPlaceholder("Anonymous"))
//	pass, err := cli.TextInput("Password:", WithMask())
func TextInput(title string, opts ...TextInputOption) (string, error) {
	cfg := &textInputConfig{}
	for _, opt := range opts {
		opt(cfg)
	}

	// Fall back to simple Question if not a terminal
	if !term.IsTerminal(int(StdinFd())) {
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
```

**Step 4: Run tests to verify they pass**

Run: `cd /Users/snider/Code/host-uk/core && go test -run TestTextInputModel ./pkg/cli/... -v`
Expected: All 7 tests PASS.

**Step 5: Commit**

```bash
cd /Users/snider/Code/host-uk/core && git add pkg/cli/textinput.go pkg/cli/textinput_test.go && git commit -m "feat(cli): add TextInput with placeholder, masking, validation

Co-Authored-By: Virgil <virgil@lethean.io>"
```

---

### Task 7: Viewport (scrollable content)

**Files:**
- Create: `/Users/snider/Code/host-uk/core/pkg/cli/viewport.go`
- Create: `/Users/snider/Code/host-uk/core/pkg/cli/viewport_test.go`

A scrollable content pane for displaying long output (logs, diffs, docs). Uses bubbletea internally.

**Step 1: Write the tests**

```go
// viewport_test.go
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
	// Should clamp — can't scroll past content
	assert.GreaterOrEqual(t, m.offset, 0)
}

func TestViewportModel_Good_View(t *testing.T) {
	m := newViewportModel("line 1\nline 2", "My Title", 10)
	view := m.View()
	assert.Contains(t, view, "My Title")
	assert.Contains(t, view, "line 1")
	assert.Contains(t, view, "line 2")
}
```

**Step 2: Run tests to verify they fail**

Run: `cd /Users/snider/Code/host-uk/core && go test -run TestViewportModel ./pkg/cli/... -v`
Expected: FAIL — `newViewportModel` undefined.

**Step 3: Write the implementation**

```go
// viewport.go
package cli

import (
	"fmt"
	"strings"

	tea "github.com/charmbracelet/bubbletea/v2"
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
		default:
			switch msg.String() {
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
		sb.WriteString(DimStyle.Render(strings.Repeat("─", len(m.title))) + "\n")
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

	sb.WriteString("\n" + DimStyle.Render("↑/↓ scroll • PgUp/PgDn page • q quit"))

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
	if !term.IsTerminal(int(StdinFd())) {
		if cfg.title != "" {
			fmt.Println(BoldStyle.Render(cfg.title))
			fmt.Println(DimStyle.Render(strings.Repeat("─", len(cfg.title))))
		}
		fmt.Println(content)
		return nil
	}

	m := newViewportModel(content, cfg.title, cfg.height)
	p := tea.NewProgram(m)
	_, err := p.Run()
	return err
}
```

**Step 4: Run tests to verify they pass**

Run: `cd /Users/snider/Code/host-uk/core && go test -run TestViewportModel ./pkg/cli/... -v`
Expected: All 6 tests PASS.

**Step 5: Commit**

```bash
cd /Users/snider/Code/host-uk/core && git add pkg/cli/viewport.go pkg/cli/viewport_test.go && git commit -m "feat(cli): add Viewport for scrollable content (logs, diffs, docs)

Co-Authored-By: Virgil <virgil@lethean.io>"
```

---

### Task 8: Future stubs (Form, FilePicker, Tabs)

**Files:**
- Create: `/Users/snider/Code/host-uk/core/pkg/cli/stubs.go`
- Create: `/Users/snider/Code/host-uk/core/pkg/cli/stubs_test.go`

Interface definitions for features we'll build later. Simple fallback implementations so the API is usable today.

**Step 1: Write the tests**

```go
// stubs_test.go
package cli

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestFormField_Good_Types(t *testing.T) {
	fields := []FormField{
		{Label: "Name", Key: "name", Type: FieldText},
		{Label: "Password", Key: "pass", Type: FieldPassword},
		{Label: "Accept", Key: "ok", Type: FieldConfirm},
	}
	assert.Equal(t, 3, len(fields))
	assert.Equal(t, FieldText, fields[0].Type)
	assert.Equal(t, FieldPassword, fields[1].Type)
	assert.Equal(t, FieldConfirm, fields[2].Type)
}

func TestFieldType_Good_Constants(t *testing.T) {
	assert.Equal(t, FieldType("text"), FieldText)
	assert.Equal(t, FieldType("password"), FieldPassword)
	assert.Equal(t, FieldType("confirm"), FieldConfirm)
	assert.Equal(t, FieldType("select"), FieldSelect)
}

func TestTabItem_Good_Structure(t *testing.T) {
	tabs := []TabItem{
		{Title: "Overview", Content: "overview content"},
		{Title: "Details", Content: "detail content"},
	}
	assert.Equal(t, 2, len(tabs))
	assert.Equal(t, "Overview", tabs[0].Title)
}
```

**Step 2: Run tests to verify they fail**

Run: `cd /Users/snider/Code/host-uk/core && go test -run "TestFormField|TestFieldType|TestTabItem" ./pkg/cli/... -v`
Expected: FAIL — types undefined.

**Step 3: Write the implementation**

```go
// stubs.go
package cli

// ──────────────────────────────────────────────────────────────────────────────
// Form (stubbed — simple fallback, will use charmbracelet/huh later)
// ──────────────────────────────────────────────────────────────────────────────

// FieldType defines the type of a form field.
type FieldType string

const (
	FieldText     FieldType = "text"
	FieldPassword FieldType = "password"
	FieldConfirm  FieldType = "confirm"
	FieldSelect   FieldType = "select"
)

// FormField describes a single field in a form.
type FormField struct {
	Label       string
	Key         string
	Type        FieldType
	Default     string
	Placeholder string
	Options     []string // For FieldSelect
	Required    bool
	Validator   func(string) error
}

// Form presents a multi-field form and returns the values keyed by FormField.Key.
// Currently falls back to sequential Question()/Confirm()/Select() calls.
// Will be replaced with charmbracelet/huh interactive form later.
//
//	results, err := cli.Form([]cli.FormField{
//	    {Label: "Name", Key: "name", Type: cli.FieldText, Required: true},
//	    {Label: "Password", Key: "pass", Type: cli.FieldPassword},
//	    {Label: "Accept terms?", Key: "terms", Type: cli.FieldConfirm},
//	})
func Form(fields []FormField) (map[string]string, error) {
	results := make(map[string]string, len(fields))

	for _, f := range fields {
		switch f.Type {
		case FieldPassword:
			val := Question(f.Label+":", WithDefault(f.Default))
			results[f.Key] = val
		case FieldConfirm:
			if Confirm(f.Label) {
				results[f.Key] = "true"
			} else {
				results[f.Key] = "false"
			}
		case FieldSelect:
			val, err := Select(f.Label, f.Options)
			if err != nil {
				return nil, err
			}
			results[f.Key] = val
		default: // FieldText
			var opts []QuestionOption
			if f.Default != "" {
				opts = append(opts, WithDefault(f.Default))
			}
			if f.Required {
				opts = append(opts, RequiredInput())
			}
			if f.Validator != nil {
				opts = append(opts, WithValidator(f.Validator))
			}
			results[f.Key] = Question(f.Label+":", opts...)
		}
	}

	return results, nil
}

// ──────────────────────────────────────────────────────────────────────────────
// FilePicker (stubbed — will use charmbracelet/filepicker later)
// ──────────────────────────────────────────────────────────────────────────────

// FilePickerOption configures FilePicker behaviour.
type FilePickerOption func(*filePickerConfig)

type filePickerConfig struct {
	dir        string
	extensions []string
}

// InDirectory sets the starting directory for the file picker.
func InDirectory(dir string) FilePickerOption {
	return func(c *filePickerConfig) {
		c.dir = dir
	}
}

// WithExtensions filters to specific file extensions (e.g. ".go", ".yaml").
func WithExtensions(exts ...string) FilePickerOption {
	return func(c *filePickerConfig) {
		c.extensions = exts
	}
}

// FilePicker presents a file browser and returns the selected path.
// Currently falls back to a text prompt. Will be replaced with an
// interactive file browser later.
//
//	path, err := cli.FilePicker(cli.InDirectory("."), cli.WithExtensions(".go"))
func FilePicker(opts ...FilePickerOption) (string, error) {
	cfg := &filePickerConfig{dir: "."}
	for _, opt := range opts {
		opt(cfg)
	}

	hint := "File path"
	if cfg.dir != "." {
		hint += " (from " + cfg.dir + ")"
	}
	return Question(hint + ":"), nil
}

// ──────────────────────────────────────────────────────────────────────────────
// Tabs (stubbed — will use bubbletea model later)
// ──────────────────────────────────────────────────────────────────────────────

// TabItem describes a tab with a title and content.
type TabItem struct {
	Title   string
	Content string
}

// Tabs displays tabbed content. Currently prints all tabs sequentially.
// Will be replaced with an interactive tab switcher later.
//
//	cli.Tabs([]cli.TabItem{
//	    {Title: "Overview", Content: summaryText},
//	    {Title: "Details", Content: detailText},
//	})
func Tabs(items []TabItem) error {
	for i, tab := range items {
		if i > 0 {
			Blank()
		}
		Section(tab.Title)
		Println("%s", tab.Content)
	}
	return nil
}
```

**Step 4: Run tests to verify they pass**

Run: `cd /Users/snider/Code/host-uk/core && go test -run "TestFormField|TestFieldType|TestTabItem" ./pkg/cli/... -v`
Expected: All 3 tests PASS.

**Step 5: Commit**

```bash
cd /Users/snider/Code/host-uk/core && git add pkg/cli/stubs.go pkg/cli/stubs_test.go && git commit -m "feat(cli): stub Form, FilePicker, Tabs with simple fallbacks

Interfaces defined for future charmbracelet/huh upgrade.
Current implementations use sequential prompts.

Co-Authored-By: Virgil <virgil@lethean.io>"
```

---

### Task 9: Run full test suite and verify

**Step 1: Run all cli package tests**

Run: `cd /Users/snider/Code/host-uk/core && go test ./pkg/cli/... -v -count=1`
Expected: All tests pass (existing + new).

**Step 2: Run full module tests**

Run: `cd /Users/snider/Code/host-uk/core && go test ./... 2>&1 | tail -30`
Expected: No regressions.

**Step 3: Verify no charmbracelet imports leaked outside pkg/cli**

Run: `cd /Users/snider/Code/host-uk/core && grep -r "charmbracelet" --include="*.go" . | grep -v pkg/cli/ | grep -v vendor/`
Expected: No output (charmbracelet only imported inside pkg/cli/).

---

## Verification

After all tasks:

1. `go test ./pkg/cli/... -v` — all pass (existing + ~34 new tests)
2. `go test ./...` — no regressions across the module
3. `grep -r "charmbracelet" --include="*.go" . | grep -v pkg/cli/` — empty (no leaks)
4. New public API surface:
   - `NewSpinner(msg)` → `*SpinnerHandle` (Update, Done, Fail, Stop)
   - `NewProgressBar(total)` → `*ProgressHandle` (Increment, Set, SetMessage, Done)
   - `InteractiveList(title, items)` → `(int, string)`
   - `TextInput(title, opts...)` → `(string, error)`
   - `Viewport(content, opts...)` → `error`
   - `RunTUI(model)` → `error` (escape hatch)
   - `Form(fields)` → `(map[string]string, error)` (stub)
   - `FilePicker(opts...)` → `(string, error)` (stub)
   - `Tabs(items)` → `error` (stub)
   - `Model`, `Msg`, `Cmd`, `KeyMsg`, `KeyType` + key constants

## Dependency Sequencing

```
Task 1 (add deps) ← Task 2 (Spinner)
Task 1 ← Task 3 (ProgressBar)
Task 1 ← Task 4 (TUI runner) ← Task 5 (List)
Task 4 ← Task 6 (TextInput)
Task 4 ← Task 7 (Viewport)
Task 1 ← Task 8 (Stubs)
Tasks 2-8 ← Task 9 (Verification)
```

Tasks 2, 3, and 8 are independent of each other (can run in parallel after Task 1). Tasks 5, 6, 7 depend on Task 4 (RunTUI) but are independent of each other.
