package core_test

import (
	"embed"
	"io"
	"testing"

	. "forge.lthn.ai/core/go/pkg/core"
	"github.com/stretchr/testify/assert"
)

//go:embed testdata
var testFS embed.FS

// --- Data (Embedded Content Mounts) ---

func TestData_New_Good(t *testing.T) {
	c := New()
	r := c.Data().New(Options{
		{K: "name", V: "test"},
		{K: "source", V: testFS},
		{K: "path", V: "testdata"},
	})
	assert.True(t, r.OK)
	assert.NotNil(t, r.Value)
}

func TestData_New_Bad(t *testing.T) {
	c := New()

	// Missing name
	r := c.Data().New(Options{
		{K: "source", V: testFS},
	})
	assert.False(t, r.OK)

	// Missing source
	r = c.Data().New(Options{
		{K: "name", V: "test"},
	})
	assert.False(t, r.OK)

	// Wrong source type
	r = c.Data().New(Options{
		{K: "name", V: "test"},
		{K: "source", V: "not-an-fs"},
	})
	assert.False(t, r.OK)
}

func TestData_ReadString_Good(t *testing.T) {
	c := New()
	c.Data().New(Options{
		{K: "name", V: "app"},
		{K: "source", V: testFS},
		{K: "path", V: "testdata"},
	})
	content, err := c.Data().ReadString("app/test.txt")
	assert.NoError(t, err)
	assert.Equal(t, "hello from testdata\n", content)
}

func TestData_ReadString_Bad(t *testing.T) {
	c := New()
	_, err := c.Data().ReadString("nonexistent/file.txt")
	assert.Error(t, err)
}

func TestData_ReadFile_Good(t *testing.T) {
	c := New()
	c.Data().New(Options{
		{K: "name", V: "app"},
		{K: "source", V: testFS},
		{K: "path", V: "testdata"},
	})
	data, err := c.Data().ReadFile("app/test.txt")
	assert.NoError(t, err)
	assert.Equal(t, "hello from testdata\n", string(data))
}

func TestData_Get_Good(t *testing.T) {
	c := New()
	c.Data().New(Options{
		{K: "name", V: "brain"},
		{K: "source", V: testFS},
		{K: "path", V: "testdata"},
	})
	emb := c.Data().Get("brain")
	assert.NotNil(t, emb)

	// Read via the Embed directly
	file, err := emb.Open("test.txt")
	assert.NoError(t, err)
	defer file.Close()
	content, _ := io.ReadAll(file)
	assert.Equal(t, "hello from testdata\n", string(content))
}

func TestData_Get_Bad(t *testing.T) {
	c := New()
	emb := c.Data().Get("nonexistent")
	assert.Nil(t, emb)
}

func TestData_Mounts_Good(t *testing.T) {
	c := New()
	c.Data().New(Options{{K: "name", V: "a"}, {K: "source", V: testFS}, {K: "path", V: "testdata"}})
	c.Data().New(Options{{K: "name", V: "b"}, {K: "source", V: testFS}, {K: "path", V: "testdata"}})
	mounts := c.Data().Mounts()
	assert.Len(t, mounts, 2)
	assert.Contains(t, mounts, "a")
	assert.Contains(t, mounts, "b")
}

// --- Legacy Embed() accessor ---

func TestEmbed_Legacy_Good(t *testing.T) {
	c := New()
	c.Data().New(Options{
		{K: "name", V: "app"},
		{K: "source", V: testFS},
		{K: "path", V: "testdata"},
	})
	// Legacy accessor reads from Data mount "app"
	emb := c.Embed()
	assert.NotNil(t, emb)
}
