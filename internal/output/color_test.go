package output

import (
	"bytes"
	"strings"
	"testing"
)

func TestStylistDisabled(t *testing.T) {
	s := PlainStylist()
	if got := s.Green("ok"); got != "ok" {
		t.Errorf("plain Green(ok) = %q, want unchanged", got)
	}
	if got := s.Red("err"); got != "err" {
		t.Errorf("plain Red(err) = %q, want unchanged", got)
	}
}

func TestStylistEnabledWraps(t *testing.T) {
	s := Stylist{enabled: true}
	got := s.Green("ok")
	if !strings.Contains(got, ansiGreen) || !strings.Contains(got, ansiReset) {
		t.Errorf("enabled Green(ok) = %q, want ANSI-wrapped", got)
	}
}

func TestColorEnabledNoTTY(t *testing.T) {
	if ColorEnabled(&bytes.Buffer{}) {
		t.Errorf("ColorEnabled on bytes.Buffer = true, want false")
	}
}

func TestColorEnabledRespectsNoColor(t *testing.T) {
	t.Setenv("NO_COLOR", "1")
	if ColorEnabled(&bytes.Buffer{}) {
		t.Errorf("ColorEnabled with NO_COLOR set = true, want false")
	}
}
