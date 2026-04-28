package core_test

import (
	"embed"
	"testing"

	. "dappco.re/go/core"
)

//go:embed testdata
var testFS embed.FS

// --- Data (Embedded Content Mounts) ---

func mountTestData(t *testing.T, c *Core, name string) {
	t.Helper()

	r := c.Data().New(NewOptions(
		Option{Key: "name", Value: name},
		Option{Key: "source", Value: testFS},
		Option{Key: "path", Value: "testdata"},
	))
	AssertTrue(t, r.OK)
}

func TestData_New_Good(t *testing.T) {
	c := New()
	r := c.Data().New(NewOptions(
		Option{Key: "name", Value: "test"},
		Option{Key: "source", Value: testFS},
		Option{Key: "path", Value: "testdata"},
	))
	AssertTrue(t, r.OK)
	AssertNotNil(t, r.Value)
}

func TestData_New_Bad(t *testing.T) {
	c := New()

	r := c.Data().New(NewOptions(Option{Key: "source", Value: testFS}))
	AssertFalse(t, r.OK)

	r = c.Data().New(NewOptions(Option{Key: "name", Value: "test"}))
	AssertFalse(t, r.OK)

	r = c.Data().New(NewOptions(Option{Key: "name", Value: "test"}, Option{Key: "source", Value: "not-an-fs"}))
	AssertFalse(t, r.OK)
}

func TestData_ReadString_Good(t *testing.T) {
	c := New()
	mountTestData(t, c, "app")
	r := c.Data().ReadString("app/test.txt")
	AssertTrue(t, r.OK)
	AssertEqual(t, "hello from testdata\n", r.Value.(string))
}

func TestData_ReadString_Bad(t *testing.T) {
	c := New()
	r := c.Data().ReadString("nonexistent/file.txt")
	AssertFalse(t, r.OK)
}

func TestData_ReadFile_Good(t *testing.T) {
	c := New()
	mountTestData(t, c, "app")
	r := c.Data().ReadFile("app/test.txt")
	AssertTrue(t, r.OK)
	AssertEqual(t, "hello from testdata\n", string(r.Value.([]byte)))
}

func TestData_Get_Good(t *testing.T) {
	c := New()
	mountTestData(t, c, "brain")
	gr := c.Data().Get("brain")
	AssertTrue(t, gr.OK)
	emb := gr.Value.(*Embed)

	r := emb.Open("test.txt")
	AssertTrue(t, r.OK)
	cr := ReadAll(r.Value)
	AssertTrue(t, cr.OK)
	AssertEqual(t, "hello from testdata\n", cr.Value)
}

func TestData_Get_Bad(t *testing.T) {
	c := New()
	r := c.Data().Get("nonexistent")
	AssertFalse(t, r.OK)
}

func TestData_Mounts_Good(t *testing.T) {
	c := New()
	mountTestData(t, c, "a")
	mountTestData(t, c, "b")
	mounts := c.Data().Mounts()
	AssertLen(t, mounts, 2)
}

func TestData_List_Good(t *testing.T) {
	c := New()
	mountTestData(t, c, "app")
	r := c.Data().List("app/.")
	AssertTrue(t, r.OK)
}

func TestData_List_Bad(t *testing.T) {
	c := New()
	r := c.Data().List("nonexistent/path")
	AssertFalse(t, r.OK)
}

func TestData_ListNames_Good(t *testing.T) {
	c := New()
	mountTestData(t, c, "app")
	r := c.Data().ListNames("app/.")
	AssertTrue(t, r.OK)
	AssertContains(t, r.Value.([]string), "test")
}

func TestData_Extract_Good(t *testing.T) {
	c := New()
	mountTestData(t, c, "app")
	r := c.Data().Extract("app/.", t.TempDir(), nil)
	AssertTrue(t, r.OK)
}

func TestData_Extract_Bad(t *testing.T) {
	c := New()
	r := c.Data().Extract("nonexistent/path", t.TempDir(), nil)
	AssertFalse(t, r.OK)
}
