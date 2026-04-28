package core_test

import . "dappco.re/go"

func ExampleNewTemplate() {
	tmpl, _ := NewTemplate("greeting").Parse("hello {{.Name}}")
	buf := NewBuffer()
	ExecuteTemplate(tmpl, buf, map[string]string{"Name": "codex"})
	Println(buf.String())
	// Output: hello codex
}

func ExampleParseTemplate() {
	r := ParseTemplate("greeting", "hello {{.Name}}")
	buf := NewBuffer()
	ExecuteTemplate(r.Value.(*Template), buf, map[string]string{"Name": "codex"})
	Println(buf.String())
	// Output: hello codex
}

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
