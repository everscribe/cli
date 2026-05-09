package auth

import (
	"context"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"testing"
	"time"
)

// callbackURL builds the loopback URL for a given port.
func callbackURL(port int) string {
	return "http://127.0.0.1:" + strconv.Itoa(port) + "/callback"
}

// postForm POSTs form data and returns (status, err). Returning the
// error lets callers in goroutines route failures through t.Errorf —
// t.Fatalf is unsafe outside the main test goroutine.
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
	if err != nil {
		t.Fatalf("newLoopbackServer: %v", err)
	}
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
		if err != nil {
			t.Errorf("POST: %v", err)
			return
		}
		if status != http.StatusOK {
			t.Errorf("POST status = %d, want 200", status)
		}
	}()

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	res, err := srv.Wait(ctx)
	if err != nil {
		t.Fatalf("Wait: %v", err)
	}
	if res.Token != "pat_aaaa1111bbbb2222cccc3333dddd4444" {
		t.Errorf("Token = %q", res.Token)
	}
	if res.UserEmail != "alice@example.com" {
		t.Errorf("UserEmail = %q", res.UserEmail)
	}
	if !res.ExpiresAt.Equal(exp) {
		t.Errorf("ExpiresAt = %v, want %v", res.ExpiresAt, exp)
	}
}

func TestLoopback_StateMismatchRejected(t *testing.T) {
	srv, err := newLoopbackServer("right-state")
	if err != nil {
		t.Fatalf("newLoopbackServer: %v", err)
	}
	t.Cleanup(func() { srv.Close() })

	form := url.Values{
		"state": {"wrong-state"},
		"token": {"pat_x"},
	}
	go func() {
		status, err := postForm(srv.Port(), form)
		if err != nil {
			t.Errorf("POST: %v", err)
			return
		}
		if status != http.StatusBadRequest {
			t.Errorf("POST status = %d, want 400", status)
		}
	}()

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	_, err = srv.Wait(ctx)
	if err == nil || !strings.Contains(err.Error(), "state mismatch") {
		t.Errorf("err = %v, want state mismatch", err)
	}
}

func TestLoopback_MissingTokenRejected(t *testing.T) {
	srv, err := newLoopbackServer("st")
	if err != nil {
		t.Fatalf("newLoopbackServer: %v", err)
	}
	t.Cleanup(func() { srv.Close() })

	go func() {
		status, err := postForm(srv.Port(), url.Values{"state": {"st"}})
		if err != nil {
			t.Errorf("POST: %v", err)
			return
		}
		if status != http.StatusBadRequest {
			t.Errorf("POST status = %d, want 400", status)
		}
	}()
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	_, err = srv.Wait(ctx)
	if err == nil || !strings.Contains(err.Error(), "no token") {
		t.Errorf("err = %v, want no token", err)
	}
}

func TestLoopback_NoExpiryParsesAsZero(t *testing.T) {
	srv, err := newLoopbackServer("st")
	if err != nil {
		t.Fatalf("newLoopbackServer: %v", err)
	}
	t.Cleanup(func() { srv.Close() })

	go func() {
		if _, err := postForm(srv.Port(), url.Values{
			"state": {"st"},
			"token": {"pat_x"},
			// expires_at intentionally omitted
		}); err != nil {
			t.Errorf("POST: %v", err)
		}
	}()

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	res, err := srv.Wait(ctx)
	if err != nil {
		t.Fatalf("Wait: %v", err)
	}
	if !res.ExpiresAt.IsZero() {
		t.Errorf("ExpiresAt = %v, want zero", res.ExpiresAt)
	}
}

func TestLoopback_BadExpiresAtRejected(t *testing.T) {
	srv, err := newLoopbackServer("st")
	if err != nil {
		t.Fatalf("newLoopbackServer: %v", err)
	}
	t.Cleanup(func() { srv.Close() })

	go func() {
		if _, err := postForm(srv.Port(), url.Values{
			"state":      {"st"},
			"token":      {"pat_x"},
			"expires_at": {"not-a-time"},
		}); err != nil {
			t.Errorf("POST: %v", err)
		}
	}()

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	_, err = srv.Wait(ctx)
	if err == nil || !strings.Contains(err.Error(), "expires_at") {
		t.Errorf("err = %v, want parse error", err)
	}
}

func TestLoopback_GetMethodRejected(t *testing.T) {
	srv, err := newLoopbackServer("st")
	if err != nil {
		t.Fatalf("newLoopbackServer: %v", err)
	}
	t.Cleanup(func() { srv.Close() })

	resp, err := http.Get(callbackURL(srv.Port()))
	if err != nil {
		t.Fatalf("GET: %v", err)
	}
	t.Cleanup(func() { resp.Body.Close() })
	if resp.StatusCode != http.StatusMethodNotAllowed {
		t.Errorf("status = %d, want 405", resp.StatusCode)
	}
}

func TestLoopback_TimeoutPropagates(t *testing.T) {
	srv, err := newLoopbackServer("st")
	if err != nil {
		t.Fatalf("newLoopbackServer: %v", err)
	}
	t.Cleanup(func() { srv.Close() })

	ctx, cancel := context.WithTimeout(context.Background(), 50*time.Millisecond)
	defer cancel()
	_, err = srv.Wait(ctx)
	if err != context.DeadlineExceeded {
		t.Errorf("err = %v, want DeadlineExceeded", err)
	}
}
