package core_test

import (
	"embed"
	"io"
	"testing"

	. "dappco.re/go/core"
	"github.com/stretchr/testify/assert"
)

//go:embed testdata
var testFS embed.FS

// --- Data (Embedded Content Mounts) ---

func TestData_New_Good(t *testing.T) {
	c := New()
	r := c.Data().New(Options{
		{Key: "name", Value: "test"},
		{Key: "source", Value: testFS},
		{Key: "path", Value: "testdata"},
	})
	assert.True(t, r.OK)
	assert.NotNil(t, r.Value)
}

func TestData_New_Bad(t *testing.T) {
	c := New()

	r := c.Data().New(Options{{Key: "source", Value: testFS}})
	assert.False(t, r.OK)

	r = c.Data().New(Options{{Key: "name", Value: "test"}})
	assert.False(t, r.OK)

	r = c.Data().New(Options{{Key: "name", Value: "test"}, {Key: "source", Value: "not-an-fs"}})
	assert.False(t, r.OK)
}

func TestData_ReadString_Good(t *testing.T) {
	c := New()
	c.Data().New(Options{{Key: "name", Value: "app"}, {Key: "source", Value: testFS}, {Key: "path", Value: "testdata"}})
	r := c.Data().ReadString("app/test.txt")
	assert.True(t, r.OK)
	assert.Equal(t, "hello from testdata\n", r.Value.(string))
}

func TestData_ReadString_Bad(t *testing.T) {
	c := New()
	r := c.Data().ReadString("nonexistent/file.txt")
	assert.False(t, r.OK)
}

func TestData_ReadFile_Good(t *testing.T) {
	c := New()
	c.Data().New(Options{{Key: "name", Value: "app"}, {Key: "source", Value: testFS}, {Key: "path", Value: "testdata"}})
	r := c.Data().ReadFile("app/test.txt")
	assert.True(t, r.OK)
	assert.Equal(t, "hello from testdata\n", string(r.Value.([]byte)))
}

func TestData_Get_Good(t *testing.T) {
	c := New()
	c.Data().New(Options{{Key: "name", Value: "brain"}, {Key: "source", Value: testFS}, {Key: "path", Value: "testdata"}})
	gr := c.Data().Get("brain")
	assert.True(t, gr.OK)
	emb := gr.Value.(*Embed)

	r := emb.Open("test.txt")
	assert.True(t, r.OK)
	file := r.Value.(io.ReadCloser)
	defer file.Close()
	content, _ := io.ReadAll(file)
	assert.Equal(t, "hello from testdata\n", string(content))
}

func TestData_Get_Bad(t *testing.T) {
	c := New()
	r := c.Data().Get("nonexistent")
	assert.False(t, r.OK)
}

func TestData_Mounts_Good(t *testing.T) {
	c := New()
	c.Data().New(Options{{Key: "name", Value: "a"}, {Key: "source", Value: testFS}, {Key: "path", Value: "testdata"}})
	c.Data().New(Options{{Key: "name", Value: "b"}, {Key: "source", Value: testFS}, {Key: "path", Value: "testdata"}})
	mounts := c.Data().Mounts()
	assert.Len(t, mounts, 2)
}

func TestEmbed_Legacy_Good(t *testing.T) {
	c := New()
	c.Data().New(Options{{Key: "name", Value: "app"}, {Key: "source", Value: testFS}, {Key: "path", Value: "testdata"}})
	assert.NotNil(t, c.Embed())
}

func TestData_List_Good(t *testing.T) {
	c := New()
	c.Data().New(Options{{Key: "name", Value: "app"}, {Key: "source", Value: testFS}, {Key: "path", Value: "."}})
	r := c.Data().List("app/testdata")
	assert.True(t, r.OK)
}

func TestData_List_Bad(t *testing.T) {
	c := New()
	r := c.Data().List("nonexistent/path")
	assert.False(t, r.OK)
}

func TestData_ListNames_Good(t *testing.T) {
	c := New()
	c.Data().New(Options{{Key: "name", Value: "app"}, {Key: "source", Value: testFS}, {Key: "path", Value: "."}})
	r := c.Data().ListNames("app/testdata")
	assert.True(t, r.OK)
	assert.Contains(t, r.Value.([]string), "test")
}

func TestData_Extract_Good(t *testing.T) {
	c := New()
	c.Data().New(Options{{Key: "name", Value: "app"}, {Key: "source", Value: testFS}, {Key: "path", Value: "."}})
	r := c.Data().Extract("app/testdata", t.TempDir(), nil)
	assert.True(t, r.OK)
}

func TestData_Extract_Bad(t *testing.T) {
	c := New()
	r := c.Data().Extract("nonexistent/path", t.TempDir(), nil)
	assert.False(t, r.OK)
}
