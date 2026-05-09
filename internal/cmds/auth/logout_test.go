package auth

import (
	"bytes"
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

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

	if err := config.Save(&config.PAT{
		Token:     "pat_secret",
		PATID:     "pat-1",
		UserEmail: "alice@example.com",
	}); err != nil {
		t.Fatalf("Save: %v", err)
	}

	var stdout bytes.Buffer
	if err := runLogout(context.Background(), &stdout); err != nil {
		t.Fatalf("runLogout: %v", err)
	}

	if gotMethod != http.MethodDelete {
		t.Errorf("method = %q, want DELETE", gotMethod)
	}
	if gotPath != "/v1/pats/pat-1" {
		t.Errorf("path = %q", gotPath)
	}
	if gotAuth != "Bearer pat_secret" {
		t.Errorf("auth = %q", gotAuth)
	}
	if _, err := config.Load(); !errors.Is(err, config.ErrNotLoggedIn) {
		t.Errorf("after logout, Load: %v, want ErrNotLoggedIn", err)
	}
	if !strings.Contains(stdout.String(), "Logged out") {
		t.Errorf("stdout = %q", stdout.String())
	}
}

func TestRunLogout_AlreadyLoggedOut(t *testing.T) {
	t.Setenv("HOME", t.TempDir())

	var stdout bytes.Buffer
	if err := runLogout(context.Background(), &stdout); err != nil {
		t.Fatalf("runLogout: %v", err)
	}
	if !strings.Contains(stdout.String(), "Already logged out") {
		t.Errorf("stdout = %q", stdout.String())
	}
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

	if err := config.Save(&config.PAT{
		Token: "pat_stale", PATID: "pat-stale", UserEmail: "x@y",
	}); err != nil {
		t.Fatalf("Save: %v", err)
	}

	var stdout bytes.Buffer
	if err := runLogout(context.Background(), &stdout); err != nil {
		t.Fatalf("runLogout: %v", err)
	}
	if _, err := config.Load(); !errors.Is(err, config.ErrNotLoggedIn) {
		t.Errorf("local session not cleared: %v", err)
	}
	if !strings.Contains(stdout.String(), "Logged out") {
		t.Errorf("stdout = %q", stdout.String())
	}
}
