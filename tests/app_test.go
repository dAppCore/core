package core_test

import (
	"testing"

	. "forge.lthn.ai/core/go/pkg/core"
	"github.com/stretchr/testify/assert"
)

// --- App ---

func TestApp_Good(t *testing.T) {
	c := New(Options{{K: "name", V: "myapp"}})
	assert.Equal(t, "myapp", c.App().Name)
}

func TestApp_Empty_Good(t *testing.T) {
	c := New()
	assert.NotNil(t, c.App())
	assert.Equal(t, "", c.App().Name)
}

func TestApp_Runtime_Good(t *testing.T) {
	c := New()
	c.App().Runtime = &struct{ Name string }{Name: "wails"}
	assert.NotNil(t, c.App().Runtime)
}

func TestApp_Find_Good(t *testing.T) {
	app := Find("go", "go")
	// Find looks for a binary — go should be in PATH
	if app != nil {
		assert.NotEmpty(t, app.Path)
	}
}

func TestApp_Find_Bad(t *testing.T) {
	app := Find("nonexistent-binary-xyz", "test")
	assert.Nil(t, app)
}
