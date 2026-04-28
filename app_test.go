package core_test

import (
	. "dappco.re/go"
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

// --- AX-7 canonical triplets ---

func TestApp_App_New_Good(t *T) {
	app := App{}.New(NewOptions(
		Option{Key: "name", Value: "Lethean Agent"},
		Option{Key: "version", Value: "0.9.0"},
		Option{Key: "description", Value: "agent dispatch runtime"},
		Option{Key: "filename", Value: "lethean-agent"},
	))
	AssertEqual(t, "Lethean Agent", app.Name)
	AssertEqual(t, "0.9.0", app.Version)
	AssertEqual(t, "agent dispatch runtime", app.Description)
	AssertEqual(t, "lethean-agent", app.Filename)
}

func TestApp_App_New_Bad(t *T) {
	app := App{}.New(NewOptions())
	AssertEqual(t, "", app.Name)
	AssertEqual(t, "", app.Version)
	AssertEqual(t, "", app.Description)
}

func TestApp_App_New_Ugly(t *T) {
	base := App{Name: "Existing", Version: "0.8.0", Description: "old", Filename: "old-agent"}
	app := base.New(NewOptions(Option{Key: "name", Value: "Updated"}))
	AssertEqual(t, "Updated", app.Name)
	AssertEqual(t, "0.8.0", app.Version)
	AssertEqual(t, "old", app.Description)
	AssertEqual(t, "old-agent", app.Filename)
}

func TestApp_App_Find_Good(t *T) {
	dir := t.TempDir()
	bin := Path(dir, "agent-dispatch")
	AssertTrue(t, WriteFile(bin, []byte("#!/bin/sh\n"), 0755).OK)

	r := App{}.Find(bin, "Agent Dispatch")
	AssertTrue(t, r.OK)
	app := r.Value.(*App)
	AssertEqual(t, "Agent Dispatch", app.Name)
	AssertEqual(t, bin, app.Filename)
	AssertEqual(t, bin, app.Path)
}

func TestApp_App_Find_Bad(t *T) {
	r := App{}.Find("definitely-missing-agent-binary", "Missing Agent")
	AssertFalse(t, r.OK)
}

func TestApp_App_Find_Ugly(t *T) {
	oldPath, hadPath := LookupEnv("PATH")
	RequireTrue(t, Setenv("PATH", "").OK)
	defer func() {
		if hadPath {
			Setenv("PATH", oldPath)
		} else {
			Unsetenv("PATH")
		}
	}()

	r := App{}.Find("agent-dispatch", "Agent Dispatch")
	AssertFalse(t, r.OK)
	AssertContains(t, r.Error(), "PATH is empty")
}
