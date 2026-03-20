package core_test

import (
	"testing"

	. "forge.lthn.ai/core/go/pkg/core"
	"github.com/stretchr/testify/assert"
)

// --- App ---

func TestApp_Good(t *testing.T) {
	c := New(Options{{Key: "name", Value: "myapp"}})
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
	r := Find("go", "go")
	assert.True(t, r.OK)
	app := r.Value.(*App)
	assert.NotEmpty(t, app.Path)
}

func TestApp_Find_Bad(t *testing.T) {
	r := Find("nonexistent-binary-xyz", "test")
	assert.False(t, r.OK)
}
