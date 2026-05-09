package auth

import (
	"bytes"
	"errors"
	"strings"
	"testing"
	"time"

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
	if err := config.Save(pat); err != nil {
		t.Fatalf("Save: %v", err)
	}
	return pat
}

func TestRunWhoami_Table(t *testing.T) {
	savePAT(t)
	var buf bytes.Buffer
	if err := runWhoami(&buf, "table"); err != nil {
		t.Fatalf("runWhoami: %v", err)
	}
	out := buf.String()
	for _, want := range []string{"EMAIL", "USER ID", "PAT ID", "EXPIRES", "alice@example.com", "u_1", "pat-uuid", "2026-08-07"} {
		if !strings.Contains(out, want) {
			t.Errorf("table output missing %q:\n%s", want, out)
		}
	}
	// Token must NEVER appear in whoami output.
	if strings.Contains(out, "pat_secret") {
		t.Errorf("whoami leaked token in table output: %q", out)
	}
}

func TestRunWhoami_JSON(t *testing.T) {
	savePAT(t)
	var buf bytes.Buffer
	if err := runWhoami(&buf, "json"); err != nil {
		t.Fatalf("runWhoami: %v", err)
	}
	out := buf.String()
	for _, want := range []string{`"email": "alice@example.com"`, `"user_id": "u_1"`, `"pat_id": "pat-uuid"`} {
		if !strings.Contains(out, want) {
			t.Errorf("JSON output missing %q:\n%s", want, out)
		}
	}
	if strings.Contains(out, "pat_secret") {
		t.Errorf("whoami leaked token in JSON output: %q", out)
	}
}

func TestRunWhoami_YAML(t *testing.T) {
	savePAT(t)
	var buf bytes.Buffer
	if err := runWhoami(&buf, "yaml"); err != nil {
		t.Fatalf("runWhoami: %v", err)
	}
	out := buf.String()
	for _, want := range []string{"email: alice@example.com", "user_id: u_1", "pat_id: pat-uuid"} {
		if !strings.Contains(out, want) {
			t.Errorf("YAML output missing %q:\n%s", want, out)
		}
	}
}

func TestRunWhoami_NotLoggedIn(t *testing.T) {
	t.Setenv("HOME", t.TempDir())
	err := runWhoami(&bytes.Buffer{}, "table")
	if !errors.Is(err, config.ErrNotLoggedIn) {
		t.Errorf("err = %v, want ErrNotLoggedIn", err)
	}
}

func TestRunWhoami_RejectsBadFormat(t *testing.T) {
	savePAT(t)
	err := runWhoami(&bytes.Buffer{}, "xml")
	if err == nil || !strings.Contains(err.Error(), "invalid") {
		t.Errorf("err = %v, want format error", err)
	}
}
