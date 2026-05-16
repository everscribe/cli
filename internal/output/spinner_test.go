package output

import (
	"bytes"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

// TestSpinner_NonTTYIsSilent: a *bytes.Buffer isn't a TTY, so the
// spinner must write nothing and Stop must not block. This is the
// contract CI / piped-stderr usage relies on.
func TestSpinner_NonTTYIsSilent(t *testing.T) {
	var buf bytes.Buffer
	sp := StartSpinner(&buf, "loading...")
	time.Sleep(150 * time.Millisecond) // long enough for a TTY spinner to tick
	sp.Stop()
	require.Empty(t, buf.String(), "non-TTY writer must receive no spinner output")
}

func TestSpinner_StopReturnsPromptly(t *testing.T) {
	var buf bytes.Buffer
	sp := StartSpinner(&buf, "x")
	done := make(chan struct{})
	go func() {
		sp.Stop()
		close(done)
	}()
	select {
	case <-done:
	case <-time.After(time.Second):
		t.Fatal("Stop did not return — spinner goroutine likely leaked")
	}
}
