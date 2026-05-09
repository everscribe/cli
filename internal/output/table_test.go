package output

import (
	"bytes"
	"strings"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestTableAlignment(t *testing.T) {
	var buf bytes.Buffer
	tbl := NewTable(&buf)
	tbl.Header("ID", "NAME", "STATUS")
	tbl.Row("a1", "short", "ok")
	tbl.Row("a22", "much-longer-name", "denied")
	require.NoError(t, tbl.Flush())

	lines := strings.Split(strings.TrimRight(buf.String(), "\n"), "\n")
	require.Len(t, lines, 3)

	nameIdx := strings.Index(lines[0], "NAME")
	for _, l := range lines[1:] {
		require.Equalf(t, nameIdx, strings.Index(l, strings.Fields(l)[1]),
			"NAME column not aligned in row %q", l)
	}
}

func TestTableEmptyCellRendersDash(t *testing.T) {
	var buf bytes.Buffer
	tbl := NewTable(&buf)
	tbl.Row("a", "", "c")
	require.NoError(t, tbl.Flush())
	require.Contains(t, buf.String(), "-", "empty cells should render as dash")
}
