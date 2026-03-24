package core_test

import (
	"testing"

	. "dappco.re/go/core"
	"github.com/stretchr/testify/assert"
)

// --- App ---

func TestApp_Good(t *testing.T) {
	c := New(WithOptions(NewOptions(Option{Key: "name", Value: "myapp"}))).Value.(*Core)
	assert.Equal(t, "myapp", c.App().Name)
}

func TestApp_Empty_Good(t *testing.T) {
	c := New().Value.(*Core)
	assert.NotNil(t, c.App())
	assert.Equal(t, "", c.App().Name)
}

func TestApp_Runtime_Good(t *testing.T) {
	c := New().Value.(*Core)
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
