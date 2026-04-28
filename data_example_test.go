package core_test

import . "dappco.re/go"

// ExampleData_New mounts a prompts directory as named embedded Lethean data. Mounted data
// can be read, listed, and extracted through Result-returning helpers.
func ExampleData_New() {
	fs := (&Fs{}).New("/")
	dir := fs.TempDir("core-data-example")
	defer fs.DeleteAll(dir)
	fs.Write(Path(dir, "prompts", "hello.txt"), "hello")

	c := New()
	r := c.Data().New(NewOptions(
		Option{Key: "name", Value: "agent"},
		Option{Key: "source", Value: DirFS(dir)},
		Option{Key: "path", Value: "prompts"},
	))

	Println(r.OK)
	Println(c.Data().Mounts())
	// Output:
	// true
	// [agent]
}

// ExampleData_ReadFile reads a named file through `Data.ReadFile` for embedded Lethean
// data. Mounted data can be read, listed, and extracted through Result-returning helpers.
func ExampleData_ReadFile() {
	fs := (&Fs{}).New("/")
	dir := fs.TempDir("core-data-example")
	defer fs.DeleteAll(dir)
	fs.Write(Path(dir, "prompts", "hello.txt"), "hello")

	c := New()
	c.Data().New(NewOptions(
		Option{Key: "name", Value: "agent"},
		Option{Key: "source", Value: DirFS(dir)},
		Option{Key: "path", Value: "prompts"},
	))

	r := c.Data().ReadFile("agent/hello.txt")
	Println(string(r.Value.([]byte)))
	// Output: hello
}

// ExampleData_ReadString reads text content through `Data.ReadString` for embedded Lethean
// data. Mounted data can be read, listed, and extracted through Result-returning helpers.
func ExampleData_ReadString() {
	fs := (&Fs{}).New("/")
	dir := fs.TempDir("core-data-example")
	defer fs.DeleteAll(dir)
	fs.Write(Path(dir, "prompts", "hello.txt"), "hello")

	c := New()
	c.Data().New(NewOptions(
		Option{Key: "name", Value: "agent"},
		Option{Key: "source", Value: DirFS(dir)},
		Option{Key: "path", Value: "prompts"},
	))

	r := c.Data().ReadString("agent/hello.txt")
	Println(r.Value)
	// Output: hello
}

// ExampleData_List lists entries through `Data.List` for embedded Lethean data. Mounted
// data can be read, listed, and extracted through Result-returning helpers.
func ExampleData_List() {
	fs := (&Fs{}).New("/")
	dir := fs.TempDir("core-data-example")
	defer fs.DeleteAll(dir)
	fs.Write(Path(dir, "prompts", "hello.txt"), "hello")

	c := New()
	c.Data().New(NewOptions(
		Option{Key: "name", Value: "agent"},
		Option{Key: "source", Value: DirFS(dir)},
		Option{Key: "path", Value: "prompts"},
	))

	r := c.Data().List("agent/.")
	Println(r.OK)
	// Output: true
}

// ExampleData_ListNames lists entry names through `Data.ListNames` for embedded Lethean
// data. Mounted data can be read, listed, and extracted through Result-returning helpers.
func ExampleData_ListNames() {
	fs := (&Fs{}).New("/")
	dir := fs.TempDir("core-data-example")
	defer fs.DeleteAll(dir)
	fs.Write(Path(dir, "prompts", "hello.txt"), "hello")

	c := New()
	c.Data().New(NewOptions(
		Option{Key: "name", Value: "agent"},
		Option{Key: "source", Value: DirFS(dir)},
		Option{Key: "path", Value: "prompts"},
	))

	r := c.Data().ListNames("agent/.")
	Println(r.Value)
	// Output: [hello]
}

// ExampleData_Extract extracts embedded files through `Data.Extract` for embedded Lethean
// data. Mounted data can be read, listed, and extracted through Result-returning helpers.
func ExampleData_Extract() {
	fs := (&Fs{}).New("/")
	source := fs.TempDir("core-data-source")
	target := fs.TempDir("core-data-target")
	defer fs.DeleteAll(source)
	defer fs.DeleteAll(target)

	fs.Write(Path(source, "templates", "README.md.tmpl"), "hello {{.Name}}")

	c := New()
	c.Data().New(NewOptions(
		Option{Key: "name", Value: "starter"},
		Option{Key: "source", Value: DirFS(source)},
		Option{Key: "path", Value: "templates"},
	))

	r := c.Data().Extract("starter/.", target, map[string]string{"Name": "codex"})
	Println(r.OK)
	Println(fs.Read(Path(target, "README.md")).Value)
	// Output:
	// true
	// hello codex
}

// ExampleData_Mounts lists mounted data sources through `Data.Mounts` for embedded Lethean
// data. Mounted data can be read, listed, and extracted through Result-returning helpers.
func ExampleData_Mounts() {
	c := New()
	Println(c.Data().Mounts())
	// Output: []
}
