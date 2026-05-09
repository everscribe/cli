package auth

import (
	"bytes"
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/everscribe/cli/internal/config"
)

func TestRunLogout_RevokesAndDeletes(t *testing.T) {
	t.Setenv("HOME", t.TempDir())

	var gotMethod, gotPath, gotAuth string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotMethod = r.Method
		gotPath = r.URL.Path
		gotAuth = r.Header.Get("Authorization")
		w.WriteHeader(http.StatusNoContent)
	}))
	t.Cleanup(srv.Close)
	t.Setenv("EVERSCRIBE_API_URL_OVERRIDE", srv.URL)

	require.NoError(t, config.Save(&config.PAT{
		Token:     "pat_secret",
		PATID:     "pat-1",
		UserEmail: "alice@example.com",
	}))

	var stdout bytes.Buffer
	require.NoError(t, runLogout(context.Background(), &stdout))

	require.Equal(t, http.MethodDelete, gotMethod)
	require.Equal(t, "/v1/pats/pat-1", gotPath)
	require.Equal(t, "Bearer pat_secret", gotAuth)

	_, err := config.Load()
	require.ErrorIs(t, err, config.ErrNotLoggedIn)
	require.Contains(t, stdout.String(), "Logged out")
}

func TestRunLogout_AlreadyLoggedOut(t *testing.T) {
	t.Setenv("HOME", t.TempDir())

	var stdout bytes.Buffer
	require.NoError(t, runLogout(context.Background(), &stdout))
	require.Contains(t, stdout.String(), "Already logged out")
}

// TestRunLogout_TolerantToServerErrors makes sure a stale or invalid
// token doesn't wedge the user — the local file gets cleared even
// when the server-side revoke fails.
func TestRunLogout_TolerantToServerErrors(t *testing.T) {
	t.Setenv("HOME", t.TempDir())

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		http.Error(w, "token not found", http.StatusNotFound)
	}))
	t.Cleanup(srv.Close)
	t.Setenv("EVERSCRIBE_API_URL_OVERRIDE", srv.URL)

	require.NoError(t, config.Save(&config.PAT{
		Token: "pat_stale", PATID: "pat-stale", UserEmail: "x@y",
	}))

	var stdout bytes.Buffer
	require.NoError(t, runLogout(context.Background(), &stdout))

	_, err := config.Load()
	require.ErrorIs(t, err, config.ErrNotLoggedIn, "local session should still be cleared")
	require.Contains(t, stdout.String(), "Logged out")
}
