package config

import (
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

func TestRoundTrip(t *testing.T) {
	t.Setenv("HOME", t.TempDir())

	want := &PAT{
		Token:     "pat_aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa",
		PATID:     "pat-uuid",
		UserID:    "u-1",
		UserEmail: "alice@example.com",
		ExpiresAt: time.Date(2026, 8, 1, 12, 0, 0, 0, time.UTC),
	}
	require.NoError(t, Save(want))

	got, err := Load()
	require.NoError(t, err)
	require.Equal(t, *want, *got)
}

func TestSavePermissions(t *testing.T) {
	t.Setenv("HOME", t.TempDir())

	require.NoError(t, Save(&PAT{Token: "pat_x"}))

	p, err := Path()
	require.NoError(t, err)

	info, err := os.Stat(p)
	require.NoError(t, err)
	require.Equal(t, os.FileMode(0o600), info.Mode().Perm(), "pat.json must be 0600")

	dirInfo, err := os.Stat(filepath.Dir(p))
	require.NoError(t, err)
	require.Equal(t, os.FileMode(0o700), dirInfo.Mode().Perm(), "config dir must be 0700")
}

func TestLoadMissing(t *testing.T) {
	t.Setenv("HOME", t.TempDir())

	_, err := Load()
	require.ErrorIs(t, err, ErrNotLoggedIn)
}

func TestLoadEmptyTokenIsNotLoggedIn(t *testing.T) {
	t.Setenv("HOME", t.TempDir())

	require.NoError(t, Save(&PAT{}))
	_, err := Load()
	require.ErrorIs(t, err, ErrNotLoggedIn)
}

func TestDeleteIdempotent(t *testing.T) {
	t.Setenv("HOME", t.TempDir())

	require.NoError(t, Delete(), "Delete on missing file should be no-op")
	require.NoError(t, Save(&PAT{Token: "pat_x"}))
	require.NoError(t, Delete(), "Delete on existing file")

	_, err := Load()
	require.ErrorIs(t, err, ErrNotLoggedIn)
}
