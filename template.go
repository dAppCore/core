// SPDX-License-Identifier: EUPL-1.2

// Text templating primitive for the Core framework.
//
// Re-exports stdlib text/template types and provides Result-shape
// constructors for Compile / Parse / Execute. Consumer packages
// declare Template / FuncMap parameters via core without importing
// text/template directly.
//
// Usage:
//
//	r := core.NewTemplate("greeting").Parse("Hello {{.Name}}!")
//	if !r.OK { return r }
//	tmpl := r.Value.(*Template)
//
//	var buf bytes.Buffer
//	er := core.ExecuteTemplate(tmpl, &buf, map[string]string{"Name": "Snider"})
//	if !er.OK { return er }
package core

import (
	"text/template"
)

// Template is a parsed text template, ready to execute against data.
//
//	tmpl := core.NewTemplate("status")
//	r, err := tmpl.Parse("agent {{.Name}} ready")
//	if err == nil { _ = r }
type Template = template.Template

// FuncMap is a map of named functions exposed to templates.
//
//	funcs := core.FuncMap{"upper": core.Upper}
//	tmpl := core.NewTemplate("status").Funcs(funcs)
//	_, _ = tmpl.Parse("{{upper .Name}}")
type FuncMap = template.FuncMap

// NewTemplate creates an empty named template.
//
//	tmpl := core.NewTemplate("status")
func NewTemplate(name string) *Template {
	return template.New(name)
}

// ParseTemplate creates and parses a template from a string in one
// call. Returns Result wrapping *Template on success.
//
//	r := core.ParseTemplate("status", "Hello {{.Name}}!")
//	if !r.OK { return r }
//	tmpl := r.Value.(*Template)
func ParseTemplate(name, text string) Result {
	t, err := template.New(name).Parse(text)
	if err != nil {
		return Result{err, false}
	}
	return Result{t, true}
}

// ParseTemplateFiles parses one or more named template files into a
// single Template. The first file's basename becomes the template's
// name.
//
//	r := core.ParseTemplateFiles("layout.tmpl", "body.tmpl")
func ParseTemplateFiles(filenames ...string) Result {
	t, err := template.ParseFiles(filenames...)
	if err != nil {
		return Result{err, false}
	}
	return Result{t, true}
}

// ExecuteTemplate runs the template with the given data, writing to w.
//
//	r := core.ExecuteTemplate(tmpl, &buf, data)
//	if !r.OK { return r }
func ExecuteTemplate(t *Template, w Writer, data any) Result {
	if err := t.Execute(w, data); err != nil {
		return Result{err, false}
	}
	return Result{OK: true}
}
