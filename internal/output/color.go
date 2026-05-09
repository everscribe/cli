package output

import (
	"io"
	"os"

	"golang.org/x/term"
)

const (
	ansiReset  = "\x1b[0m"
	ansiRed    = "\x1b[31m"
	ansiGreen  = "\x1b[32m"
	ansiYellow = "\x1b[33m"
)

// IsTTY reports whether w is an *os.File backed by a terminal.
func IsTTY(w io.Writer) bool {
	f, ok := w.(*os.File)
	if !ok {
		return false
	}
	return term.IsTerminal(int(f.Fd()))
}

// ColorEnabled reports whether ANSI color is appropriate for w. False if
// stdout is not a TTY or if NO_COLOR is set (https://no-color.org).
func ColorEnabled(w io.Writer) bool {
	if os.Getenv("NO_COLOR") != "" {
		return false
	}
	return IsTTY(w)
}

// Stylist applies ANSI color to text when its target writer supports it.
type Stylist struct {
	enabled bool
}

// NewStylist captures the color decision for w once, so every subsequent
// call avoids re-checking the environment.
func NewStylist(w io.Writer) Stylist {
	return Stylist{enabled: ColorEnabled(w)}
}

// PlainStylist always returns text uncolored. Useful in tests.
func PlainStylist() Stylist {
	return Stylist{enabled: false}
}

func (s Stylist) Green(text string) string  { return s.wrap(ansiGreen, text) }
func (s Stylist) Red(text string) string    { return s.wrap(ansiRed, text) }
func (s Stylist) Yellow(text string) string { return s.wrap(ansiYellow, text) }

func (s Stylist) wrap(code, text string) string {
	if !s.enabled {
		return text
	}
	return code + text + ansiReset
}
