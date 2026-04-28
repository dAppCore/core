package core_test

import . "dappco.re/go"

type failingWriter struct{}

func (failingWriter) Write(_ []byte) (int, error) {
	return 0, E("test.failingWriter", "write failed", nil)
}

func TestTable_NewTable_Good(t *T) {
	out := NewBuffer()

	AssertNotNil(t, NewTable(out))
}

func TestTable_NewTable_Bad(t *T) {
	table := NewTable(nil)

	AssertError(t, table.Flush())
}

func TestTable_NewTable_Ugly(t *T) {
	out := NewBuffer()
	table := NewTable(out)

	AssertNoError(t, table.Row("Name", "Status").Flush())
	AssertContains(t, out.String(), "Name")
}

func TestTable_Row_Good(t *T) {
	out := NewBuffer()
	table := NewTable(out)

	AssertSame(t, table, table.Row("Name", "Status"))
	AssertNoError(t, table.Flush())
	AssertContains(t, out.String(), "Name")
	AssertContains(t, out.String(), "Status")
}

func TestTable_Row_Bad(t *T) {
	out := NewBuffer()
	table := NewTable(out)

	AssertSame(t, table, table.Row())
	AssertNoError(t, table.Flush())
	AssertEqual(t, "\n", out.String())
}

func TestTable_Row_Ugly(t *T) {
	out := NewBuffer()
	table := NewTable(out)

	table.Row("Name", "Status").Row("api", "ok")
	AssertNoError(t, table.Flush())
	AssertContains(t, out.String(), "api")
	AssertContains(t, out.String(), "ok")
}

func TestTable_Flush_Good(t *T) {
	out := NewBuffer()

	err := NewTable(out).Row("A", "B").Flush()

	AssertNoError(t, err)
	AssertContains(t, out.String(), "A")
}

func TestTable_Flush_Bad(t *T) {
	table := NewTable(failingWriter{})

	AssertError(t, table.Row("A").Flush())
}

func TestTable_Flush_Ugly(t *T) {
	out := NewBuffer()
	table := NewTable(out).Row("A")

	AssertNoError(t, table.Flush())
	AssertNoError(t, table.Flush())
}

func TestTable_Table_Row_Good(t *T) {
	out := NewBuffer()
	table := NewTable(out)

	AssertSame(t, table, table.Row("Name", "Status"))
	AssertNoError(t, table.Flush())
	AssertContains(t, out.String(), "Name")
	AssertContains(t, out.String(), "Status")
}

func TestTable_Table_Row_Bad(t *T) {
	var table *Table

	AssertNil(t, table.Row("Name"))
}

func TestTable_Table_Row_Ugly(t *T) {
	out := NewBuffer()
	table := NewTable(out)

	table.Row("Name", "Status", "Updated").Row("agent.dispatch", "ok", "now")

	AssertNoError(t, table.Flush())
	AssertContains(t, out.String(), "agent.dispatch")
	AssertContains(t, out.String(), "Updated")
}

func TestTable_Table_Flush_Good(t *T) {
	out := NewBuffer()

	err := NewTable(out).Row("Name", "Status").Flush()

	AssertNoError(t, err)
	AssertContains(t, out.String(), "Status")
}

func TestTable_Table_Flush_Bad(t *T) {
	var table *Table

	err := table.Flush()

	AssertError(t, err, "table is nil")
}

func TestTable_Table_Flush_Ugly(t *T) {
	out := NewBuffer()
	table := NewTable(out)

	AssertNoError(t, table.Flush())
	AssertEqual(t, "", out.String())
}
