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
