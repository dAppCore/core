// SPDX-License-Identifier: EUPL-1.2

// Tabular text output for the Core framework.

package core

import (
	"io"
	"text/tabwriter"
)

// Table writes tab-aligned rows to an underlying writer.
//
//	table := core.NewTable(out)
//	table.Row("Name", "Status").Row("api", "ok")
//	_ = table.Flush()
type Table struct {
	writer *tabwriter.Writer
	err    error
}

// NewTable creates a Table that writes tab-aligned rows to w.
//
//	table := core.NewTable(out)
func NewTable(w io.Writer) *Table {
	if w == nil {
		return &Table{
			writer: tabwriter.NewWriter(io.Discard, 0, 0, 2, ' ', 0),
			err:    E("core.NewTable", "writer is nil", nil),
		}
	}
	return &Table{writer: tabwriter.NewWriter(w, 0, 0, 2, ' ', 0)}
}

// Row writes one row and returns the Table for chaining.
//
//	table.Row("Name", "Status").Row("api", "ok")
func (t *Table) Row(cells ...string) *Table {
	if t == nil {
		return t
	}
	for i, cell := range cells {
		if i > 0 {
			t.write("\t")
		}
		t.write(cell)
	}
	t.write("\n")
	return t
}

// Flush flushes buffered table output to the underlying writer.
//
//	if err := table.Flush(); err != nil { return err }
func (t *Table) Flush() error {
	if t == nil {
		return E("core.Table.Flush", "table is nil", nil)
	}
	if t.err != nil {
		return t.err
	}
	return t.writer.Flush()
}

func (t *Table) write(s string) {
	if t.err != nil {
		return
	}
	_, err := t.writer.Write([]byte(s))
	if err != nil {
		t.err = err
	}
}
