// SPDX-License-Identifier: EUPL-1.2

// Tabular text output for the Core framework.

package core

import "text/tabwriter"

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
func NewTable(w Writer) *Table {
	if w == nil {
		return &Table{
			writer: tabwriter.NewWriter(Discard, 0, 0, 2, ' ', 0),
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
// Returns Result with OK=false + Code "table.flush.failed" when the
// underlying writer rejects the flush.
//
//	r := table.Flush()
//	if !r.OK { return r }
func (t *Table) Flush() Result {
	if t == nil {
		return Result{Value: NewCode("table.nil", "table is nil"), OK: false}
	}
	if t.err != nil {
		return Result{Value: WrapCode(t.err, "table.flush.failed", "Table.Flush", "buffered write failed"), OK: false}
	}
	if err := t.writer.Flush(); err != nil {
		return Result{Value: WrapCode(err, "table.flush.failed", "Table.Flush", "writer rejected flush"), OK: false}
	}
	return Result{OK: true}
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
