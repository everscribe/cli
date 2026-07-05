package output

import (
	"fmt"
	"io"
	"time"
)

// spinnerFrames are the braille dots used by most modern CLI spinners.
// Ten frames at ~80ms per frame yields a full revolution every 800ms.
var spinnerFrames = []rune{'⠋', '⠙', '⠹', '⠸', '⠼', '⠴', '⠦', '⠧', '⠇', '⠏'}

const spinnerInterval = 80 * time.Millisecond

// Spinner animates a single status line on w until Stop is called.
// When w is not a TTY (a pipe, a file, an io.Discard), Spinner is a
// no-op - CI logs and piped stderr stay free of carriage returns and
// escape codes.
//
// Spinner is single-shot: create a new one per operation. Stop is
// safe to call exactly once; calling it twice will block.
type Spinner struct {
	w       io.Writer
	enabled bool
	stop    chan struct{}
	done    chan struct{}
}

// StartSpinner begins animating message on w and returns a handle the
// caller must Stop before printing anything else to w.
func StartSpinner(w io.Writer, message string) *Spinner {
	s := &Spinner{
		w:       w,
		enabled: IsTTY(w),
		stop:    make(chan struct{}),
		done:    make(chan struct{}),
	}
	if !s.enabled {
		close(s.done)
		return s
	}
	go s.run(message)
	return s
}

func (s *Spinner) run(message string) {
	defer close(s.done)
	t := time.NewTicker(spinnerInterval)
	defer t.Stop()

	frame := 0
	// Initial paint so something appears before the first tick.
	fmt.Fprintf(s.w, "\r%c %s", spinnerFrames[frame], message)
	for {
		select {
		case <-s.stop:
			// \r returns the cursor to col 0; \x1b[K clears to end of line.
			fmt.Fprint(s.w, "\r\x1b[K")
			return
		case <-t.C:
			frame = (frame + 1) % len(spinnerFrames)
			fmt.Fprintf(s.w, "\r%c %s", spinnerFrames[frame], message)
		}
	}
}

// Stop halts the animation and clears the line. Safe to call when the
// spinner is the no-op (non-TTY) variant.
func (s *Spinner) Stop() {
	if !s.enabled {
		return
	}
	close(s.stop)
	<-s.done
}
