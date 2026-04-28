package core_test

import . "dappco.re/go"

// ExampleNewTemplate creates a template through `NewTemplate` for operator-facing
// templates. Parsing and execution use core template wrappers for operator-facing text.
func ExampleNewTemplate() {
	tmpl, _ := NewTemplate("greeting").Parse("hello {{.Name}}")
	buf := NewBuffer()
	ExecuteTemplate(tmpl, buf, map[string]string{"Name": "codex"})
	Println(buf.String())
	// Output: hello codex
}

// ExampleParseTemplate parses template text through `ParseTemplate` for operator-facing
// templates. Parsing and execution use core template wrappers for operator-facing text.
func ExampleParseTemplate() {
	r := ParseTemplate("greeting", "hello {{.Name}}")
	buf := NewBuffer()
	ExecuteTemplate(r.Value.(*Template), buf, map[string]string{"Name": "codex"})
	Println(buf.String())
	// Output: hello codex
}

// ExampleParseTemplateFiles parses templates from files through `ParseTemplateFiles` for
// operator-facing templates. Parsing and execution use core template wrappers for
// operator-facing text.
func ExampleParseTemplateFiles() {
	fs := (&Fs{}).New("/")
	dir := fs.TempDir("core-template-example")
	defer fs.DeleteAll(dir)

	path := Path(dir, "greeting.tmpl")
	fs.Write(path, "hello {{.Name}}")

	r := ParseTemplateFiles(path)
	buf := NewBuffer()
	ExecuteTemplate(r.Value.(*Template), buf, map[string]string{"Name": "codex"})
	Println(buf.String())
	// Output: hello codex
}

// ExampleExecuteTemplate executes a template through `ExecuteTemplate` for operator-facing
// templates. Parsing and execution use core template wrappers for operator-facing text.
func ExampleExecuteTemplate() {
	tmpl := ParseTemplate("greeting", "hello {{.Name}}").Value.(*Template)
	buf := NewBuffer()
	r := ExecuteTemplate(tmpl, buf, map[string]string{"Name": "codex"})
	Println(r.OK)
	Println(buf.String())
	// Output:
	// true
	// hello codex
}
