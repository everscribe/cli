package auth

import (
	"bytes"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/everscribe/cli/internal/config"
)

func TestRunLogout_DeletesLocalSession(t *testing.T) {
	t.Setenv("HOME", t.TempDir())
	require.NoError(t, config.Save(&config.PAT{
		Token:     "pat_x",
		PATID:     "pat-1",
		UserEmail: "alice@example.com",
	}))

	var buf bytes.Buffer
	require.NoError(t, runLogout(&buf))

	_, err := config.Load()
	require.ErrorIs(t, err, config.ErrNotLoggedIn)

	out := buf.String()
	require.Contains(t, out, "Logged out")
	require.NotContains(t, out, "Warning", "no server-side revoke is attempted, so no warnings")
	require.NotContains(t, out, "403", "the patSelfGuard 403 must never reach the user")
}

func TestRunLogout_AlreadyLoggedOut(t *testing.T) {
	t.Setenv("HOME", t.TempDir())

	var buf bytes.Buffer
	require.NoError(t, runLogout(&buf))
	require.Contains(t, buf.String(), "Already logged out")
}
