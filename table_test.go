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

	AssertFalse(t, table.Flush().OK)
}

func TestTable_NewTable_Ugly(t *T) {
	out := NewBuffer()
	table := NewTable(out)

	AssertTrue(t, table.Row("Name", "Status").Flush().OK)
	AssertContains(t, out.String(), "Name")
}

func TestTable_Row_Good(t *T) {
	out := NewBuffer()
	table := NewTable(out)

	AssertSame(t, table, table.Row("Name", "Status"))
	AssertTrue(t, table.Flush().OK)
	AssertContains(t, out.String(), "Name")
	AssertContains(t, out.String(), "Status")
}

func TestTable_Row_Bad(t *T) {
	out := NewBuffer()
	table := NewTable(out)

	AssertSame(t, table, table.Row())
	AssertTrue(t, table.Flush().OK)
	AssertEqual(t, "\n", out.String())
}

func TestTable_Row_Ugly(t *T) {
	out := NewBuffer()
	table := NewTable(out)

	table.Row("Name", "Status").Row("api", "ok")
	AssertTrue(t, table.Flush().OK)
	AssertContains(t, out.String(), "api")
	AssertContains(t, out.String(), "ok")
}

func TestTable_Flush_Good(t *T) {
	out := NewBuffer()

	r := NewTable(out).Row("A", "B").Flush()

	AssertTrue(t, r.OK)
	AssertContains(t, out.String(), "A")
}

func TestTable_Flush_Bad(t *T) {
	table := NewTable(failingWriter{})

	AssertFalse(t, table.Row("A").Flush().OK)
}

func TestTable_Flush_Ugly(t *T) {
	out := NewBuffer()
	table := NewTable(out).Row("A")

	AssertTrue(t, table.Flush().OK)
	AssertTrue(t, table.Flush().OK)
}

func TestTable_Table_Row_Good(t *T) {
	out := NewBuffer()
	table := NewTable(out)

	AssertSame(t, table, table.Row("Name", "Status"))
	AssertTrue(t, table.Flush().OK)
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

	AssertTrue(t, table.Flush().OK)
	AssertContains(t, out.String(), "agent.dispatch")
	AssertContains(t, out.String(), "Updated")
}

func TestTable_Table_Flush_Good(t *T) {
	out := NewBuffer()

	r := NewTable(out).Row("Name", "Status").Flush()

	AssertTrue(t, r.OK)
	AssertContains(t, out.String(), "Status")
}

func TestTable_Table_Flush_Bad(t *T) {
	var table *Table

	r := table.Flush()

	AssertFalse(t, r.OK)
	AssertContains(t, r.Error(), "table is nil")
}

func TestTable_Table_Flush_Ugly(t *T) {
	out := NewBuffer()
	table := NewTable(out)

	AssertTrue(t, table.Flush().OK)
	AssertEqual(t, "", out.String())
}
