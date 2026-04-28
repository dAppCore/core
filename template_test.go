// SPDX-License-Identifier: EUPL-1.2

package core_test

import (
	. "dappco.re/go"
)

func TestTemplate_ExecuteTemplate_Good(t *T) {
	r := ParseTemplate("status", "agent {{.Name}} ready")
	RequireTrue(t, r.OK)
	tmpl := r.Value.(*Template)
	buf := NewBuffer()

	er := ExecuteTemplate(tmpl, buf, map[string]string{"Name": "codex"})

	AssertTrue(t, er.OK)
	AssertEqual(t, "agent codex ready", buf.String())
}

func TestTemplate_ExecuteTemplate_Bad(t *T) {
	r := ParseTemplate("status", "{{.Agent.Ready}}")
	RequireTrue(t, r.OK)
	buf := NewBuffer()

	er := ExecuteTemplate(r.Value.(*Template), buf, map[string]int{"Agent": 42})

	AssertFalse(t, er.OK)
	AssertError(t, er.Value.(error))
}

func TestTemplate_ExecuteTemplate_Ugly(t *T) {
	r := ParseTemplate("empty", "")
	RequireTrue(t, r.OK)
	buf := NewBuffer()

	er := ExecuteTemplate(r.Value.(*Template), buf, nil)

	AssertTrue(t, er.OK)
	AssertEqual(t, "", buf.String())
}

func TestTemplate_NewTemplate_Good(t *T) {
	tmpl := NewTemplate("status")

	AssertNotNil(t, tmpl)
	AssertEqual(t, "status", tmpl.Name())
}

func TestTemplate_NewTemplate_Bad(t *T) {
	tmpl := NewTemplate("")

	AssertNotNil(t, tmpl)
	AssertEqual(t, "", tmpl.Name())
}

func TestTemplate_NewTemplate_Ugly(t *T) {
	tmpl := NewTemplate("agent/status").Funcs(FuncMap{"upper": Upper})
	parsed, err := tmpl.Parse("{{upper .}}")
	RequireNoError(t, err)
	buf := NewBuffer()

	er := ExecuteTemplate(parsed, buf, "codex")

	AssertTrue(t, er.OK)
	AssertEqual(t, "CODEX", buf.String())
}

func TestTemplate_ParseTemplate_Good(t *T) {
	r := ParseTemplate("status", "agent {{.Name}} ready")

	AssertTrue(t, r.OK)
	AssertEqual(t, "status", r.Value.(*Template).Name())
}

func TestTemplate_ParseTemplate_Bad(t *T) {
	r := ParseTemplate("broken", "{{.Name")

	AssertFalse(t, r.OK)
	AssertError(t, r.Value.(error))
}

func TestTemplate_ParseTemplate_Ugly(t *T) {
	r := ParseTemplate("", "")

	AssertTrue(t, r.OK)
	AssertEqual(t, "", r.Value.(*Template).Name())
}

func TestTemplate_ParseTemplateFiles_Good(t *T) {
	dir := t.TempDir()
	path := PathJoin(dir, "status.tmpl")
	RequireTrue(t, WriteFile(path, []byte("agent {{.}}"), 0o644).OK)

	r := ParseTemplateFiles(path)

	AssertTrue(t, r.OK)
	AssertEqual(t, "status.tmpl", r.Value.(*Template).Name())
}

func TestTemplate_ParseTemplateFiles_Bad(t *T) {
	r := ParseTemplateFiles(PathJoin(t.TempDir(), "missing.tmpl"))

	AssertFalse(t, r.OK)
	AssertError(t, r.Value.(error))
}

func TestTemplate_ParseTemplateFiles_Ugly(t *T) {
	dir := t.TempDir()
	path := PathJoin(dir, "empty.tmpl")
	RequireTrue(t, WriteFile(path, []byte(""), 0o644).OK)

	r := ParseTemplateFiles(path)

	AssertTrue(t, r.OK)
	AssertEqual(t, "empty.tmpl", r.Value.(*Template).Name())
}

func TestTemplate_ParseTemplateFS_Good(t *T) {
	dir := t.TempDir()
	RequireTrue(t, WriteFile(PathJoin(dir, "status.tmpl"), []byte("agent {{.}}"), 0o644).OK)

	r := ParseTemplateFS(DirFS(dir), "status.tmpl")

	AssertTrue(t, r.OK)
	AssertNotNil(t, r.Value.(*Template).Lookup("status.tmpl"))
}

func TestTemplate_ParseTemplateFS_Bad(t *T) {
	r := ParseTemplateFS(DirFS(t.TempDir()), "missing.tmpl")

	AssertFalse(t, r.OK)
	AssertError(t, r.Value.(error))
}

func TestTemplate_ParseTemplateFS_Ugly(t *T) {
	dir := t.TempDir()
	RequireTrue(t, WriteFile(PathJoin(dir, "a.tmpl"), []byte("a"), 0o644).OK)
	RequireTrue(t, WriteFile(PathJoin(dir, "b.tmpl"), []byte("b"), 0o644).OK)

	r := ParseTemplateFS(DirFS(dir), "*.tmpl")

	AssertTrue(t, r.OK)
	AssertNotNil(t, r.Value.(*Template).Lookup("a.tmpl"))
	AssertNotNil(t, r.Value.(*Template).Lookup("b.tmpl"))
}
