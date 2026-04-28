package core_test

import . "dappco.re/go"

func ExampleApp_New() {
	app := App{}.New(NewOptions(
		Option{Key: "name", Value: "forge"},
		Option{Key: "version", Value: "1.2.3"},
		Option{Key: "description", Value: "deployment tool"},
		Option{Key: "filename", Value: "forge"},
	))

	Println(app.Name)
	Println(app.Version)
	Println(app.Description)
	Println(app.Filename)
	// Output:
	// forge
	// 1.2.3
	// deployment tool
	// forge
}

func ExampleApp_Find() {
	fs := (&Fs{}).New("/")
	dir := fs.TempDir("core-app-example")
	defer fs.DeleteAll(dir)

	bin := Path(dir, "forge")
	fs.WriteMode(bin, "#!/bin/sh\n", 0755)

	r := App{}.Find(bin, "Forge")
	app := r.Value.(*App)
	Println(r.OK)
	Println(app.Name)
	Println(PathBase(app.Filename))
	Println(PathBase(app.Path))
	// Output:
	// true
	// Forge
	// forge
	// forge
}
