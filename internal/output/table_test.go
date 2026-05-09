package output

import (
	"bytes"
	"strings"
	"testing"
)

func TestTableAlignment(t *testing.T) {
	var buf bytes.Buffer
	tbl := NewTable(&buf)
	tbl.Header("ID", "NAME", "STATUS")
	tbl.Row("a1", "short", "ok")
	tbl.Row("a22", "much-longer-name", "denied")
	if err := tbl.Flush(); err != nil {
		t.Fatal(err)
	}

	lines := strings.Split(strings.TrimRight(buf.String(), "\n"), "\n")
	if len(lines) != 3 {
		t.Fatalf("got %d lines, want 3:\n%s", len(lines), buf.String())
	}

	idIdx := strings.Index(lines[0], "NAME")
	for _, l := range lines[1:] {
		if strings.Index(l, strings.Fields(l)[1]) != idIdx {
			t.Errorf("column NAME not aligned at index %d in row %q", idIdx, l)
		}
	}
}

func TestTableEmptyCellRendersDash(t *testing.T) {
	var buf bytes.Buffer
	tbl := NewTable(&buf)
	tbl.Row("a", "", "c")
	if err := tbl.Flush(); err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(buf.String(), "-") {
		t.Errorf("empty cell not rendered as dash: %q", buf.String())
	}
}
