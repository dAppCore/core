package core_test

import (
	"testing"

	. "forge.lthn.ai/core/go/pkg/core"
	"github.com/stretchr/testify/assert"
)

// --- Option / Options ---

func TestOptions_Get_Good(t *testing.T) {
	opts := Options{
		{K: "name", V: "brain"},
		{K: "port", V: 8080},
	}
	val, ok := opts.Get("name")
	assert.True(t, ok)
	assert.Equal(t, "brain", val)
}

func TestOptions_Get_Bad(t *testing.T) {
	opts := Options{{K: "name", V: "brain"}}
	val, ok := opts.Get("missing")
	assert.False(t, ok)
	assert.Nil(t, val)
}

func TestOptions_Has_Good(t *testing.T) {
	opts := Options{{K: "debug", V: true}}
	assert.True(t, opts.Has("debug"))
	assert.False(t, opts.Has("missing"))
}

func TestOptions_String_Good(t *testing.T) {
	opts := Options{{K: "name", V: "brain"}}
	assert.Equal(t, "brain", opts.String("name"))
}

func TestOptions_String_Bad(t *testing.T) {
	opts := Options{{K: "port", V: 8080}}
	// Wrong type — returns empty string
	assert.Equal(t, "", opts.String("port"))
	// Missing key — returns empty string
	assert.Equal(t, "", opts.String("missing"))
}

func TestOptions_Int_Good(t *testing.T) {
	opts := Options{{K: "port", V: 8080}}
	assert.Equal(t, 8080, opts.Int("port"))
}

func TestOptions_Int_Bad(t *testing.T) {
	opts := Options{{K: "name", V: "brain"}}
	assert.Equal(t, 0, opts.Int("name"))
	assert.Equal(t, 0, opts.Int("missing"))
}

func TestOptions_Bool_Good(t *testing.T) {
	opts := Options{{K: "debug", V: true}}
	assert.True(t, opts.Bool("debug"))
}

func TestOptions_Bool_Bad(t *testing.T) {
	opts := Options{{K: "name", V: "brain"}}
	assert.False(t, opts.Bool("name"))
	assert.False(t, opts.Bool("missing"))
}

func TestOptions_TypedStruct_Good(t *testing.T) {
	// Packages plug typed structs into Option.V
	type BrainConfig struct {
		Name       string
		OllamaURL  string
		Collection string
	}
	cfg := BrainConfig{Name: "brain", OllamaURL: "http://localhost:11434", Collection: "openbrain"}
	opts := Options{{K: "config", V: cfg}}

	val, ok := opts.Get("config")
	assert.True(t, ok)
	bc, ok := val.(BrainConfig)
	assert.True(t, ok)
	assert.Equal(t, "brain", bc.Name)
	assert.Equal(t, "http://localhost:11434", bc.OllamaURL)
}

func TestOptions_Empty_Good(t *testing.T) {
	opts := Options{}
	assert.False(t, opts.Has("anything"))
	assert.Equal(t, "", opts.String("anything"))
	assert.Equal(t, 0, opts.Int("anything"))
	assert.False(t, opts.Bool("anything"))
}
