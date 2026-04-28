// SPDX-License-Identifier: EUPL-1.2

package core

func TestTable_Table_write_Good(t *T) {
	out := NewBuffer()
	table := NewTable(out)

	table.write("Name\tStatus\n")
	RequireNoError(t, table.Flush())

	AssertContains(t, out.String(), "Name")
	AssertContains(t, out.String(), "Status")
}
func TestTable_Table_write_Bad(t *T) {
	table := &Table{err: AnError}

	table.write("ignored")

	AssertError(t, table.err)
}
func TestTable_Table_write_Ugly(t *T) {
	table := NewTable(ax7FailingWriter{})

	table.write("agent\n")

	AssertError(t, table.Flush())
}
