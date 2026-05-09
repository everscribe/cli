package output

import (
	"bytes"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestParseFormat(t *testing.T) {
	cases := []struct {
		name       string
		in         string
		allowTable bool
		want       Format
		wantErr    string
	}{
		{name: "table allowed", in: "table", allowTable: true, want: FormatTable},
		{name: "table rejected", in: "table", allowTable: false, wantErr: "not supported"},
		{name: "json", in: "json", allowTable: true, want: FormatJSON},
		{name: "json no-table", in: "json", allowTable: false, want: FormatJSON},
		{name: "yaml", in: "yaml", allowTable: true, want: FormatYAML},
		{name: "invalid", in: "xml", allowTable: true, wantErr: "invalid"},
		{name: "empty", in: "", allowTable: true, wantErr: "invalid"},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			got, err := ParseFormat(tc.in, tc.allowTable)
			if tc.wantErr != "" {
				require.Error(t, err)
				require.Contains(t, err.Error(), tc.wantErr)
				return
			}
			require.NoError(t, err)
			require.Equal(t, tc.want, got)
		})
	}
}

func TestJSON(t *testing.T) {
	var buf bytes.Buffer
	require.NoError(t, JSON(&buf, map[string]string{"k": "v"}))
	got := buf.String()
	require.Contains(t, got, `"k": "v"`, "should pretty-print")
	require.True(t, len(got) > 0 && got[len(got)-1] == '\n', "should end with newline")
}

func TestYAML(t *testing.T) {
	var buf bytes.Buffer
	require.NoError(t, YAML(&buf, map[string]string{"k": "v"}))
	require.Contains(t, buf.String(), "k: v")
}
