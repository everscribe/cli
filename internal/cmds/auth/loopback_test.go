package auth

import (
	"context"
	"net/http"
	"net/url"
	"strconv"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// callbackURL builds the loopback URL for a given port.
func callbackURL(port int) string {
	return "http://127.0.0.1:" + strconv.Itoa(port) + "/callback"
}

// postForm POSTs form data and returns (status, err). Returning the
// error lets goroutine callers route failures through assert (the
// non-fatal counterpart) — require.X / t.Fatalf are unsafe outside
// the main test goroutine.
func postForm(port int, form url.Values) (int, error) {
	resp, err := http.PostForm(callbackURL(port), form)
	if err != nil {
		return 0, err
	}
	defer resp.Body.Close()
	return resp.StatusCode, nil
}

func TestLoopback_HappyPath(t *testing.T) {
	srv, err := newLoopbackServer("test-state-abc123")
	require.NoError(t, err)
	t.Cleanup(func() { srv.Close() })

	exp := time.Date(2026, 8, 7, 12, 0, 0, 0, time.UTC)
	form := url.Values{
		"state":      {"test-state-abc123"},
		"token":      {"pat_aaaa1111bbbb2222cccc3333dddd4444"},
		"pat_id":     {"pat-uuid-1"},
		"user_id":    {"u_1"},
		"user_email": {"alice@example.com"},
		"expires_at": {exp.Format(time.RFC3339)},
	}

	go func() {
		status, err := postForm(srv.Port(), form)
		assert.NoError(t, err)
		assert.Equal(t, http.StatusOK, status)
	}()

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	res, err := srv.Wait(ctx)
	require.NoError(t, err)
	require.Equal(t, "pat_aaaa1111bbbb2222cccc3333dddd4444", res.Token)
	require.Equal(t, "alice@example.com", res.UserEmail)
	require.True(t, res.ExpiresAt.Equal(exp))
}

func TestLoopback_StateMismatchRejected(t *testing.T) {
	srv, err := newLoopbackServer("right-state")
	require.NoError(t, err)
	t.Cleanup(func() { srv.Close() })

	go func() {
		status, err := postForm(srv.Port(), url.Values{
			"state": {"wrong-state"},
			"token": {"pat_x"},
		})
		assert.NoError(t, err)
		assert.Equal(t, http.StatusBadRequest, status)
	}()

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	_, err = srv.Wait(ctx)
	require.Error(t, err)
	require.Contains(t, err.Error(), "state mismatch")
}

func TestLoopback_MissingTokenRejected(t *testing.T) {
	srv, err := newLoopbackServer("st")
	require.NoError(t, err)
	t.Cleanup(func() { srv.Close() })

	go func() {
		status, err := postForm(srv.Port(), url.Values{"state": {"st"}})
		assert.NoError(t, err)
		assert.Equal(t, http.StatusBadRequest, status)
	}()

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	_, err = srv.Wait(ctx)
	require.Error(t, err)
	require.Contains(t, err.Error(), "no token")
}

func TestLoopback_NoExpiryParsesAsZero(t *testing.T) {
	srv, err := newLoopbackServer("st")
	require.NoError(t, err)
	t.Cleanup(func() { srv.Close() })

	go func() {
		_, err := postForm(srv.Port(), url.Values{
			"state": {"st"},
			"token": {"pat_x"},
			// expires_at intentionally omitted
		})
		assert.NoError(t, err)
	}()

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	res, err := srv.Wait(ctx)
	require.NoError(t, err)
	require.True(t, res.ExpiresAt.IsZero())
}

func TestLoopback_BadExpiresAtRejected(t *testing.T) {
	srv, err := newLoopbackServer("st")
	require.NoError(t, err)
	t.Cleanup(func() { srv.Close() })

	go func() {
		_, err := postForm(srv.Port(), url.Values{
			"state":      {"st"},
			"token":      {"pat_x"},
			"expires_at": {"not-a-time"},
		})
		assert.NoError(t, err)
	}()

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	_, err = srv.Wait(ctx)
	require.Error(t, err)
	require.Contains(t, err.Error(), "expires_at")
}

func TestLoopback_GetMethodRejected(t *testing.T) {
	srv, err := newLoopbackServer("st")
	require.NoError(t, err)
	t.Cleanup(func() { srv.Close() })

	resp, err := http.Get(callbackURL(srv.Port()))
	require.NoError(t, err)
	t.Cleanup(func() { resp.Body.Close() })
	require.Equal(t, http.StatusMethodNotAllowed, resp.StatusCode)
}

func TestLoopback_TimeoutPropagates(t *testing.T) {
	srv, err := newLoopbackServer("st")
	require.NoError(t, err)
	t.Cleanup(func() { srv.Close() })

	ctx, cancel := context.WithTimeout(context.Background(), 50*time.Millisecond)
	defer cancel()
	_, err = srv.Wait(ctx)
	require.ErrorIs(t, err, context.DeadlineExceeded)
}
