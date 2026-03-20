package core_test

import (
	"testing"

	. "dappco.re/go/core/pkg/core"
	"github.com/stretchr/testify/assert"
)

// --- Option / Options ---

func TestOptions_Get_Good(t *testing.T) {
	opts := Options{
		{Key: "name", Value: "brain"},
		{Key: "port", Value: 8080},
	}
	r := opts.Get("name")
	assert.True(t, r.OK)
	assert.Equal(t, "brain", r.Value)
}

func TestOptions_Get_Bad(t *testing.T) {
	opts := Options{{Key: "name", Value: "brain"}}
	r := opts.Get("missing")
	assert.False(t, r.OK)
	assert.Nil(t, r.Value)
}

func TestOptions_Has_Good(t *testing.T) {
	opts := Options{{Key: "debug", Value: true}}
	assert.True(t, opts.Has("debug"))
	assert.False(t, opts.Has("missing"))
}

func TestOptions_String_Good(t *testing.T) {
	opts := Options{{Key: "name", Value: "brain"}}
	assert.Equal(t, "brain", opts.String("name"))
}

func TestOptions_String_Bad(t *testing.T) {
	opts := Options{{Key: "port", Value: 8080}}
	// Wrong type — returns empty string
	assert.Equal(t, "", opts.String("port"))
	// Missing key — returns empty string
	assert.Equal(t, "", opts.String("missing"))
}

func TestOptions_Int_Good(t *testing.T) {
	opts := Options{{Key: "port", Value: 8080}}
	assert.Equal(t, 8080, opts.Int("port"))
}

func TestOptions_Int_Bad(t *testing.T) {
	opts := Options{{Key: "name", Value: "brain"}}
	assert.Equal(t, 0, opts.Int("name"))
	assert.Equal(t, 0, opts.Int("missing"))
}

func TestOptions_Bool_Good(t *testing.T) {
	opts := Options{{Key: "debug", Value: true}}
	assert.True(t, opts.Bool("debug"))
}

func TestOptions_Bool_Bad(t *testing.T) {
	opts := Options{{Key: "name", Value: "brain"}}
	assert.False(t, opts.Bool("name"))
	assert.False(t, opts.Bool("missing"))
}

func TestOptions_TypedStruct_Good(t *testing.T) {
	// Packages plug typed structs into Option.Value
	type BrainConfig struct {
		Name       string
		OllamaURL  string
		Collection string
	}
	cfg := BrainConfig{Name: "brain", OllamaURL: "http://localhost:11434", Collection: "openbrain"}
	opts := Options{{Key: "config", Value: cfg}}

	r := opts.Get("config")
	assert.True(t, r.OK)
	bc, ok := r.Value.(BrainConfig)
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
