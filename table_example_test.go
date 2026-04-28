package core_test

import . "dappco.re/go"

func ExampleNewTable() {
	buf := NewBuffer()
	table := NewTable(buf)
	table.Row("Name", "Status").Row("api", "ok")
	table.Flush()

	Println(Contains(buf.String(), "api"))
	// Output: true
}

func ExampleTable_Row() {
	buf := NewBuffer()
	NewTable(buf).Row("Service", "Port").Row("api", "8080").Flush()
	Println(Contains(buf.String(), "8080"))
	// Output: true
}

func ExampleTable_Flush() {
	buf := NewBuffer()
	table := NewTable(buf)
	table.Row("Name", "Status")
	Println(table.Flush() == nil)
	// Output: true
}
