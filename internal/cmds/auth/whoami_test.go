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

func TestRunWhoami_Table(t *testing.T) {
	savePAT(t)
	var buf bytes.Buffer
	require.NoError(t, runWhoami(&buf, "table"))
	out := buf.String()
	for _, want := range []string{"EMAIL", "USER ID", "PAT ID", "EXPIRES", "alice@example.com", "u_1", "pat-uuid", "2026-08-07"} {
		require.Containsf(t, out, want, "table output missing %q", want)
	}
	require.NotContains(t, out, "pat_secret", "whoami must not leak token")
}

func TestRunWhoami_JSON(t *testing.T) {
	savePAT(t)
	var buf bytes.Buffer
	require.NoError(t, runWhoami(&buf, "json"))
	out := buf.String()
	for _, want := range []string{`"email": "alice@example.com"`, `"user_id": "u_1"`, `"pat_id": "pat-uuid"`} {
		require.Containsf(t, out, want, "JSON output missing %q", want)
	}
	require.NotContains(t, out, "pat_secret", "whoami must not leak token")
}

func TestRunWhoami_YAML(t *testing.T) {
	savePAT(t)
	var buf bytes.Buffer
	require.NoError(t, runWhoami(&buf, "yaml"))
	out := buf.String()
	for _, want := range []string{"email: alice@example.com", "user_id: u_1", "pat_id: pat-uuid"} {
		require.Containsf(t, out, want, "YAML output missing %q", want)
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
	require.Contains(t, err.Error(), "invalid")
}
