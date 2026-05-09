package output

import (
	"fmt"
	"io"
	"strings"
	"text/tabwriter"
)

// Table renders aligned rows in the kubectl style.
type Table struct {
	w *tabwriter.Writer
}

// NewTable creates a Table writing to out. Flush must be called when done.
func NewTable(out io.Writer) *Table {
	return &Table{w: tabwriter.NewWriter(out, 0, 0, 3, ' ', 0)}
}

// Header writes a header row. Conventionally called once before any Row.
func (t *Table) Header(cols ...string) {
	t.Row(cols...)
}

// Row writes one data row. Empty cells are rendered as "-".
func (t *Table) Row(cells ...string) {
	out := make([]string, len(cells))
	for i, c := range cells {
		if c == "" {
			out[i] = "-"
		} else {
			out[i] = c
		}
	}
	fmt.Fprintln(t.w, strings.Join(out, "\t"))
}

// Flush emits any buffered output. Must be called before discarding the Table.
func (t *Table) Flush() error {
	return t.w.Flush()
}
