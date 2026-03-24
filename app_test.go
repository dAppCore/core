package core_test

import (
	"testing"

	. "dappco.re/go/core"
	"github.com/stretchr/testify/assert"
)

// --- App.New ---

func TestApp_New_Good(t *testing.T) {
	app := App{}.New(NewOptions(
		Option{Key: "name", Value: "myapp"},
		Option{Key: "version", Value: "1.0.0"},
		Option{Key: "description", Value: "test app"},
	))
	assert.Equal(t, "myapp", app.Name)
	assert.Equal(t, "1.0.0", app.Version)
	assert.Equal(t, "test app", app.Description)
}

func TestApp_New_Empty_Good(t *testing.T) {
	app := App{}.New(NewOptions())
	assert.Equal(t, "", app.Name)
	assert.Equal(t, "", app.Version)
}

func TestApp_New_Partial_Good(t *testing.T) {
	app := App{}.New(NewOptions(
		Option{Key: "name", Value: "myapp"},
	))
	assert.Equal(t, "myapp", app.Name)
	assert.Equal(t, "", app.Version)
}

// --- App via Core ---

func TestApp_Core_Good(t *testing.T) {
	c := New(WithOption("name", "myapp")).Value.(*Core)
	assert.Equal(t, "myapp", c.App().Name)
}

func TestApp_Core_Empty_Good(t *testing.T) {
	c := New().Value.(*Core)
	assert.NotNil(t, c.App())
	assert.Equal(t, "", c.App().Name)
}

func TestApp_Runtime_Good(t *testing.T) {
	c := New().Value.(*Core)
	c.App().Runtime = &struct{ Name string }{Name: "wails"}
	assert.NotNil(t, c.App().Runtime)
}

// --- App.Find ---

func TestApp_Find_Good(t *testing.T) {
	r := App{}.Find("go", "go")
	assert.True(t, r.OK)
	app := r.Value.(*App)
	assert.NotEmpty(t, app.Path)
}

func TestApp_Find_Bad(t *testing.T) {
	r := App{}.Find("nonexistent-binary-xyz", "test")
	assert.False(t, r.OK)
}
