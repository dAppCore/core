package core_test

import (
	. "dappco.re/go/core"
)

// --- App.New ---

func TestApp_New_Good(t *T) {
	app := App{}.New(NewOptions(
		Option{Key: "name", Value: "myapp"},
		Option{Key: "version", Value: "1.0.0"},
		Option{Key: "description", Value: "test app"},
	))
	AssertEqual(t, "myapp", app.Name)
	AssertEqual(t, "1.0.0", app.Version)
	AssertEqual(t, "test app", app.Description)
}

func TestApp_New_Empty_Good(t *T) {
	app := App{}.New(NewOptions())
	AssertEqual(t, "", app.Name)
	AssertEqual(t, "", app.Version)
}

func TestApp_New_Partial_Good(t *T) {
	app := App{}.New(NewOptions(
		Option{Key: "name", Value: "myapp"},
	))
	AssertEqual(t, "myapp", app.Name)
	AssertEqual(t, "", app.Version)
}

// --- App via Core ---

func TestApp_Core_Good(t *T) {
	c := New(WithOption("name", "myapp"))
	AssertEqual(t, "myapp", c.App().Name)
}

func TestApp_Core_Empty_Good(t *T) {
	c := New()
	AssertNotNil(t, c.App())
	AssertEqual(t, "", c.App().Name)
}

func TestApp_Runtime_Good(t *T) {
	c := New()
	c.App().Runtime = &struct{ Name string }{Name: "wails"}
	AssertNotNil(t, c.App().Runtime)
}

// --- App.Find ---

func TestApp_Find_Good(t *T) {
	r := App{}.Find("go", "go")
	AssertTrue(t, r.OK)
	app := r.Value.(*App)
	AssertNotEmpty(t, app.Path)
}

func TestApp_Find_Bad(t *T) {
	r := App{}.Find("nonexistent-binary-xyz", "test")
	AssertFalse(t, r.OK)
}
