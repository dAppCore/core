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
