package core_test

import (
	"testing"

	. "dappco.re/go/core"
	"github.com/stretchr/testify/assert"
)

// --- NewOptions ---

func TestNewOptions_Good(t *testing.T) {
	opts := NewOptions(
		Option{Key: "name", Value: "brain"},
		Option{Key: "port", Value: 8080},
	)
	assert.Equal(t, 2, opts.Len())
}

func TestNewOptions_Empty_Good(t *testing.T) {
	opts := NewOptions()
	assert.Equal(t, 0, opts.Len())
	assert.False(t, opts.Has("anything"))
}

// --- Options.Set ---

func TestOptions_Set_Good(t *testing.T) {
	opts := NewOptions()
	opts.Set("name", "brain")
	assert.Equal(t, "brain", opts.String("name"))
}

func TestOptions_Set_Update_Good(t *testing.T) {
	opts := NewOptions(Option{Key: "name", Value: "old"})
	opts.Set("name", "new")
	assert.Equal(t, "new", opts.String("name"))
	assert.Equal(t, 1, opts.Len())
}

// --- Options.Get ---

func TestOptions_Get_Good(t *testing.T) {
	opts := NewOptions(
		Option{Key: "name", Value: "brain"},
		Option{Key: "port", Value: 8080},
	)
	r := opts.Get("name")
	assert.True(t, r.OK)
	assert.Equal(t, "brain", r.Value)
}

func TestOptions_Get_Bad(t *testing.T) {
	opts := NewOptions(Option{Key: "name", Value: "brain"})
	r := opts.Get("missing")
	assert.False(t, r.OK)
	assert.Nil(t, r.Value)
}

// --- Options.Has ---

func TestOptions_Has_Good(t *testing.T) {
	opts := NewOptions(Option{Key: "debug", Value: true})
	assert.True(t, opts.Has("debug"))
	assert.False(t, opts.Has("missing"))
}

// --- Options.String ---

func TestOptions_String_Good(t *testing.T) {
	opts := NewOptions(Option{Key: "name", Value: "brain"})
	assert.Equal(t, "brain", opts.String("name"))
}

func TestOptions_String_Bad(t *testing.T) {
	opts := NewOptions(Option{Key: "port", Value: 8080})
	assert.Equal(t, "", opts.String("port"))
	assert.Equal(t, "", opts.String("missing"))
}

// --- Options.Int ---

func TestOptions_Int_Good(t *testing.T) {
	opts := NewOptions(Option{Key: "port", Value: 8080})
	assert.Equal(t, 8080, opts.Int("port"))
}

func TestOptions_Int_Bad(t *testing.T) {
	opts := NewOptions(Option{Key: "name", Value: "brain"})
	assert.Equal(t, 0, opts.Int("name"))
	assert.Equal(t, 0, opts.Int("missing"))
}

// --- Options.Bool ---

func TestOptions_Bool_Good(t *testing.T) {
	opts := NewOptions(Option{Key: "debug", Value: true})
	assert.True(t, opts.Bool("debug"))
}

func TestOptions_Bool_Bad(t *testing.T) {
	opts := NewOptions(Option{Key: "name", Value: "brain"})
	assert.False(t, opts.Bool("name"))
	assert.False(t, opts.Bool("missing"))
}

// --- Options.Items ---

func TestOptions_Items_Good(t *testing.T) {
	opts := NewOptions(Option{Key: "a", Value: 1}, Option{Key: "b", Value: 2})
	items := opts.Items()
	assert.Len(t, items, 2)
}

// --- Options with typed struct ---

func TestOptions_TypedStruct_Good(t *testing.T) {
	type BrainConfig struct {
		Name       string
		OllamaURL  string
		Collection string
	}
	cfg := BrainConfig{Name: "brain", OllamaURL: "http://localhost:11434", Collection: "openbrain"}
	opts := NewOptions(Option{Key: "config", Value: cfg})

	r := opts.Get("config")
	assert.True(t, r.OK)
	bc, ok := r.Value.(BrainConfig)
	assert.True(t, ok)
	assert.Equal(t, "brain", bc.Name)
	assert.Equal(t, "http://localhost:11434", bc.OllamaURL)
}

// --- Result ---

func TestResult_New_Good(t *testing.T) {
	r := Result{}.New("value")
	assert.Equal(t, "value", r.Value)
}

func TestResult_New_Error_Bad(t *testing.T) {
	err := E("test", "failed", nil)
	r := Result{}.New(err)
	assert.False(t, r.OK)
	assert.Equal(t, err, r.Value)
}

func TestResult_Result_Good(t *testing.T) {
	r := Result{Value: "hello", OK: true}
	assert.Equal(t, r, r.Result())
}

func TestResult_Result_WithArgs_Good(t *testing.T) {
	r := Result{}.Result("value")
	assert.Equal(t, "value", r.Value)
}

func TestResult_Get_Good(t *testing.T) {
	r := Result{Value: "hello", OK: true}
	assert.True(t, r.Get().OK)
}

func TestResult_Get_Bad(t *testing.T) {
	r := Result{Value: "err", OK: false}
	assert.False(t, r.Get().OK)
}

// --- WithOption ---

func TestWithOption_Good(t *testing.T) {
	c := New(
		WithOption("name", "myapp"),
		WithOption("port", 8080),
	).Value.(*Core)
	assert.Equal(t, "myapp", c.App().Name)
	assert.Equal(t, 8080, c.Options().Int("port"))
}
