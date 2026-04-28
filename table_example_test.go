package core_test

import . "dappco.re/go"

// ExampleNewTable creates a table through `NewTable` for CLI table output. Tabular CLI
// output is built and flushed through the table wrapper.
func ExampleNewTable() {
	buf := NewBuffer()
	table := NewTable(buf)
	table.Row("Name", "Status").Row("api", "ok")
	table.Flush()

	Println(Contains(buf.String(), "api"))
	// Output: true
}

// ExampleTable_Row adds a row through `Table.Row` for CLI table output. Tabular CLI output
// is built and flushed through the table wrapper.
func ExampleTable_Row() {
	buf := NewBuffer()
	NewTable(buf).Row("Service", "Port").Row("api", "8080").Flush()
	Println(Contains(buf.String(), "8080"))
	// Output: true
}

// ExampleTable_Flush flushes a table through `Table.Flush` for CLI table output. Tabular
// CLI output is built and flushed through the table wrapper.
func ExampleTable_Flush() {
	buf := NewBuffer()
	table := NewTable(buf)
	table.Row("Name", "Status")
	Println(table.Flush() == nil)
	// Output: true
}
