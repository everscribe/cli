package cobrax

import (
	"testing"

	"github.com/spf13/cobra"
	"github.com/stretchr/testify/require"
)

func TestRequireArgs(t *testing.T) {
	cases := []struct {
		name           string
		argNames       []string
		input          []string
		wantErrSubstrs []string
	}{
		{
			name:     "exact match",
			argNames: []string{"id"},
			input:    []string{"abc"},
		},
		{
			name:     "no names and no args is ok",
			argNames: nil,
			input:    nil,
		},
		{
			name:           "missing names the only arg",
			argNames:       []string{"event-id"},
			input:          nil,
			wantErrSubstrs: []string{"missing required argument: <event-id>"},
		},
		{
			name:           "missing names the next arg, not the first",
			argNames:       []string{"project-id", "event-id"},
			input:          []string{"proj-1"},
			wantErrSubstrs: []string{"missing required argument: <event-id>"},
		},
		{
			name:           "too many args lists the extras",
			argNames:       []string{"id"},
			input:          []string{"a", "b", "c"},
			wantErrSubstrs: []string{"unexpected extra argument", "[b c]"},
		},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			err := RequireArgs(tc.argNames...)(&cobra.Command{}, tc.input)
			if len(tc.wantErrSubstrs) == 0 {
				require.NoError(t, err)
				return
			}
			require.Error(t, err)
			for _, want := range tc.wantErrSubstrs {
				require.ErrorContains(t, err, want)
			}
		})
	}
}
