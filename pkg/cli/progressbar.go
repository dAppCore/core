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

	filled := 0
	if p.total > 0 {
		filled = (p.width * p.current) / p.total
	}
	if filled > p.width {
		filled = p.width
	}
	empty := p.width - filled

	bar := "[" + strings.Repeat("\u2588", filled) + strings.Repeat("\u2591", empty) + "]"

	if p.message != "" {
		return fmt.Sprintf("%s %3d%% %s", bar, pct, p.message)
	}
	return fmt.Sprintf("%s %3d%%", bar, pct)
}

// render outputs the progress bar, overwriting the current line.
func (p *ProgressHandle) render() {
	fmt.Printf("\033[2K\r%s", p.String())
}
