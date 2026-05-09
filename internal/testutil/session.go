// Package testutil provides shared helpers for CLI command tests.
//
// Most CLI commands need an authenticated session (~/.config/everscribe/pat.json)
// and an API server to talk to. SetupSession wires both up against a
// per-test temp HOME and an httptest.Server, plus the env override
// that points the client at it.
package testutil

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/everscribe/cli/internal/config"
)

// SetupSession boots an httptest.Server with the given handler, points
// the client at it via EVERSCRIBE_API_URL_OVERRIDE, sets HOME to a
// fresh temp dir, and saves a fake pat.json. Returns the server (so
// tests can inspect its URL) and the saved PAT (so tests can assert
// against the bearer token they expect to see).
//
// All cleanup (server close, env restore via t.Setenv) is registered
// with the testing.T.
func SetupSession(t *testing.T, handler http.HandlerFunc) (*httptest.Server, *config.PAT) {
	t.Helper()

	srv := httptest.NewServer(handler)
	t.Cleanup(srv.Close)

	t.Setenv("HOME", t.TempDir())
	t.Setenv("EVERSCRIBE_API_URL_OVERRIDE", srv.URL)

	pat := &config.PAT{
		Token:     "pat_test_token",
		PATID:     "pat-test-id",
		UserID:    "u_test",
		UserEmail: "test@example.com",
	}
	require.NoError(t, config.Save(pat))
	return srv, pat
}
