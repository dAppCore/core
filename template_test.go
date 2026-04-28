// SPDX-License-Identifier: EUPL-1.2

package core_test

import (
	"bytes"
	"testing"

	. "dappco.re/go/core"
)

func TestTemplate_NewTemplate_Good(t *testing.T) {
	tmpl := NewTemplate("test")
	AssertNotNil(t, tmpl)
	AssertEqual(t, "test", tmpl.Name())
}

func TestTemplate_ParseTemplate_Good(t *testing.T) {
	r := ParseTemplate("greeting", "Hello {{.Name}}!")
	AssertTrue(t, r.OK)
	AssertNotNil(t, r.Value)
}

func TestTemplate_ParseTemplate_Bad_Syntax(t *testing.T) {
	r := ParseTemplate("broken", "{{.Name")
	AssertFalse(t, r.OK)
}

func TestTemplate_ExecuteTemplate_Good(t *testing.T) {
	r := ParseTemplate("greeting", "Hello {{.Name}}!")
	AssertTrue(t, r.OK)
	tmpl := r.Value.(*Template)

	var buf bytes.Buffer
	er := ExecuteTemplate(tmpl, &buf, map[string]string{"Name": "Snider"})
	AssertTrue(t, er.OK)
	AssertEqual(t, "Hello Snider!", buf.String())
}

func TestTemplate_ExecuteTemplate_Bad_MissingField(t *testing.T) {
	// strict missing-key behaviour depends on Option settings; default
	// just renders <no value>. Test the err path with a misuse.
	r := ParseTemplate("typed", "{{.Number.Method}}")
	AssertTrue(t, r.OK)
	tmpl := r.Value.(*Template)

	var buf bytes.Buffer
	er := ExecuteTemplate(tmpl, &buf, map[string]int{"Number": 42})
	AssertFalse(t, er.OK)
}
