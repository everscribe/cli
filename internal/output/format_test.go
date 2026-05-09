package output

import (
	"bytes"
	"strings"
	"testing"
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
				if err == nil || !strings.Contains(err.Error(), tc.wantErr) {
					t.Fatalf("err = %v, want containing %q", err, tc.wantErr)
				}
				return
			}
			if err != nil {
				t.Fatalf("unexpected err: %v", err)
			}
			if got != tc.want {
				t.Errorf("got %q, want %q", got, tc.want)
			}
		})
	}
}

func TestJSON(t *testing.T) {
	var buf bytes.Buffer
	if err := JSON(&buf, map[string]string{"k": "v"}); err != nil {
		t.Fatal(err)
	}
	got := buf.String()
	if !strings.Contains(got, `"k": "v"`) {
		t.Errorf("JSON output missing pretty-printed pair: %q", got)
	}
	if !strings.HasSuffix(got, "\n") {
		t.Errorf("JSON output missing trailing newline: %q", got)
	}
}

func TestYAML(t *testing.T) {
	var buf bytes.Buffer
	if err := YAML(&buf, map[string]string{"k": "v"}); err != nil {
		t.Fatal(err)
	}
	got := buf.String()
	if !strings.Contains(got, "k: v") {
		t.Errorf("YAML output missing key: %q", got)
	}
}
