package auth

import (
	"bytes"
	"testing"
	"time"

	"github.com/stretchr/testify/require"

	"github.com/everscribe/cli/internal/config"
)

func savePAT(t *testing.T) *config.PAT {
	t.Helper()
	t.Setenv("HOME", t.TempDir())
	pat := &config.PAT{
		Token:     "pat_secret",
		PATID:     "pat-uuid",
		UserID:    "u_1",
		UserEmail: "alice@example.com",
		ExpiresAt: time.Date(2026, 8, 7, 12, 0, 0, 0, time.UTC),
	}
	require.NoError(t, config.Save(pat))
	return pat
}

// TestRunWhoami_FormatOutputs covers the three render paths in one
// place. Each format has different column/key syntax, but all three
// must (a) include the user metadata and (b) never leak the PAT.
func TestRunWhoami_FormatOutputs(t *testing.T) {
	cases := []struct {
		name    string
		format  string
		wantSub []string
	}{
		{
			name:    "table",
			format:  "table",
			wantSub: []string{"EMAIL", "USER ID", "PAT ID", "EXPIRES", "alice@example.com", "u_1", "pat-uuid", "2026-08-07"},
		},
		{
			name:    "json",
			format:  "json",
			wantSub: []string{`"email": "alice@example.com"`, `"user_id": "u_1"`, `"pat_id": "pat-uuid"`},
		},
		{
			name:    "yaml",
			format:  "yaml",
			wantSub: []string{"email: alice@example.com", "user_id: u_1", "pat_id: pat-uuid"},
		},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			savePAT(t)
			var buf bytes.Buffer
			require.NoError(t, runWhoami(&buf, tc.format))
			out := buf.String()
			for _, want := range tc.wantSub {
				require.Containsf(t, out, want, "%s output missing %q", tc.format, want)
			}
			require.NotContains(t, out, "pat_secret", "whoami must not leak token")
		})
	}
}

func TestRunWhoami_NotLoggedIn(t *testing.T) {
	t.Setenv("HOME", t.TempDir())
	err := runWhoami(&bytes.Buffer{}, "table")
	require.ErrorIs(t, err, config.ErrNotLoggedIn)
}

func TestRunWhoami_RejectsBadFormat(t *testing.T) {
	savePAT(t)
	err := runWhoami(&bytes.Buffer{}, "xml")
	require.Error(t, err)
	require.ErrorContains(t, err, "invalid")
}
