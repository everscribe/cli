package config

import (
	"errors"
	"os"
	"path/filepath"
	"testing"
	"time"
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
	if err := Save(want); err != nil {
		t.Fatalf("Save: %v", err)
	}

	got, err := Load()
	if err != nil {
		t.Fatalf("Load: %v", err)
	}
	if *got != *want {
		t.Errorf("round-trip mismatch:\n got: %+v\nwant: %+v", *got, *want)
	}
}

func TestSavePermissions(t *testing.T) {
	t.Setenv("HOME", t.TempDir())

	if err := Save(&PAT{Token: "pat_x"}); err != nil {
		t.Fatalf("Save: %v", err)
	}

	p, _ := Path()
	info, err := os.Stat(p)
	if err != nil {
		t.Fatalf("Stat file: %v", err)
	}
	if got := info.Mode().Perm(); got != 0o600 {
		t.Errorf("file perm = %o, want 0600", got)
	}

	dirInfo, err := os.Stat(filepath.Dir(p))
	if err != nil {
		t.Fatalf("Stat dir: %v", err)
	}
	if got := dirInfo.Mode().Perm(); got != 0o700 {
		t.Errorf("dir perm = %o, want 0700", got)
	}
}

func TestLoadMissing(t *testing.T) {
	t.Setenv("HOME", t.TempDir())

	_, err := Load()
	if !errors.Is(err, ErrNotLoggedIn) {
		t.Errorf("Load on missing file: err = %v, want ErrNotLoggedIn", err)
	}
}

func TestLoadEmptyTokenIsNotLoggedIn(t *testing.T) {
	t.Setenv("HOME", t.TempDir())

	if err := Save(&PAT{}); err != nil {
		t.Fatalf("Save: %v", err)
	}
	_, err := Load()
	if !errors.Is(err, ErrNotLoggedIn) {
		t.Errorf("Load with empty token: err = %v, want ErrNotLoggedIn", err)
	}
}

func TestDeleteIdempotent(t *testing.T) {
	t.Setenv("HOME", t.TempDir())

	if err := Delete(); err != nil {
		t.Errorf("Delete on missing file: %v", err)
	}
	if err := Save(&PAT{Token: "pat_x"}); err != nil {
		t.Fatalf("Save: %v", err)
	}
	if err := Delete(); err != nil {
		t.Errorf("Delete on existing file: %v", err)
	}
	if _, err := Load(); !errors.Is(err, ErrNotLoggedIn) {
		t.Errorf("after Delete, Load: err = %v, want ErrNotLoggedIn", err)
	}
}
