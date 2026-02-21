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
