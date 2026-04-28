package core_test

import (
	"bytes"
	"testing"

	. "dappco.re/go/core"
	"github.com/stretchr/testify/assert"
)

type failingWriter struct{}

func (failingWriter) Write(_ []byte) (int, error) {
	return 0, E("test.failingWriter", "write failed", nil)
}

func TestTable_NewTable_Good(t *testing.T) {
	var out bytes.Buffer

	assert.NotNil(t, NewTable(&out))
}

func TestTable_NewTable_Bad(t *testing.T) {
	table := NewTable(nil)

	assert.Error(t, table.Flush())
}

func TestTable_NewTable_Ugly(t *testing.T) {
	var out bytes.Buffer
	table := NewTable(&out)

	assert.NoError(t, table.Row("Name", "Status").Flush())
	assert.Contains(t, out.String(), "Name")
}

func TestTable_Row_Good(t *testing.T) {
	var out bytes.Buffer
	table := NewTable(&out)

	assert.Same(t, table, table.Row("Name", "Status"))
	assert.NoError(t, table.Flush())
	assert.Contains(t, out.String(), "Name")
	assert.Contains(t, out.String(), "Status")
}

func TestTable_Row_Bad(t *testing.T) {
	var out bytes.Buffer
	table := NewTable(&out)

	assert.Same(t, table, table.Row())
	assert.NoError(t, table.Flush())
	assert.Equal(t, "\n", out.String())
}

func TestTable_Row_Ugly(t *testing.T) {
	var out bytes.Buffer
	table := NewTable(&out)

	table.Row("Name", "Status").Row("api", "ok")
	assert.NoError(t, table.Flush())
	assert.Contains(t, out.String(), "api")
	assert.Contains(t, out.String(), "ok")
}

func TestTable_Flush_Good(t *testing.T) {
	var out bytes.Buffer

	err := NewTable(&out).Row("A", "B").Flush()

	assert.NoError(t, err)
	assert.Contains(t, out.String(), "A")
}

func TestTable_Flush_Bad(t *testing.T) {
	table := NewTable(failingWriter{})

	assert.Error(t, table.Row("A").Flush())
}

func TestTable_Flush_Ugly(t *testing.T) {
	var out bytes.Buffer
	table := NewTable(&out).Row("A")

	assert.NoError(t, table.Flush())
	assert.NoError(t, table.Flush())
}
