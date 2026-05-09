package output

import (
	"bytes"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestStylistDisabled(t *testing.T) {
	s := PlainStylist()
	require.Equal(t, "ok", s.Green("ok"))
	require.Equal(t, "err", s.Red("err"))
}

func TestStylistEnabledWraps(t *testing.T) {
	s := Stylist{enabled: true}
	got := s.Green("ok")
	require.Contains(t, got, ansiGreen)
	require.Contains(t, got, ansiReset)
}

func TestColorEnabledNoTTY(t *testing.T) {
	require.False(t, ColorEnabled(&bytes.Buffer{}), "non-TTY writer should disable color")
}

func TestColorEnabledRespectsNoColor(t *testing.T) {
	t.Setenv("NO_COLOR", "1")
	require.False(t, ColorEnabled(&bytes.Buffer{}))
}
