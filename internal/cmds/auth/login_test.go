package auth

import (
	"bytes"
	"context"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"testing"
	"time"

	"github.com/everscribe/cli/internal/config"
)

// stubBrowser swaps browserOpener for a function that captures the
// login URL and (optionally) drives the loopback callback. Restore is
// registered with t.Cleanup. Returns a pointer that, after runLogin
// returns, will hold the URL the CLI tried to open.
func stubBrowser(t *testing.T, simulateCallback func(target string)) *string {
	t.Helper()
	captured := new(string)
	prev := browserOpener
	t.Cleanup(func() { browserOpener = prev })
	browserOpener = func(target string) error {
		*captured = target
		if simulateCallback != nil {
			simulateCallback(target)
		}
		return nil
	}
	return captured
}

// simulateUICallback parses the login URL the CLI built, then mimics
// the UI's auto-submitting form by POSTing a successful payload back
// to the loopback. Mirrors what cli-auth-success.gohtml does in prod.
func simulateUICallback(t *testing.T, target string, payload url.Values) {
	t.Helper()
	u, err := url.Parse(target)
	if err != nil {
		t.Errorf("parse target URL: %v", err)
		return
	}
	port, err := strconv.Atoi(u.Query().Get("callback_port"))
	if err != nil {
		t.Errorf("parse callback_port: %v", err)
		return
	}
	state := u.Query().Get("state")
	if state == "" {
		t.Errorf("state missing from URL")
		return
	}
	// Use the state from the captured URL so a future state-handling
	// regression in buildLoginURL still trips the test.
	if payload.Get("state") == "" {
		payload.Set("state", state)
	}
	go func() {
		resp, err := http.PostForm("http://127.0.0.1:"+strconv.Itoa(port)+"/callback", payload)
		if err != nil {
			t.Errorf("loopback POST: %v", err)
			return
		}
		resp.Body.Close()
	}()
}

func TestRunLogin_HappyPath(t *testing.T) {
	t.Setenv("HOME", t.TempDir())
	t.Setenv("EVERSCRIBE_UI_URL_OVERRIDE", "https://test.example")

	exp := time.Date(2026, 8, 7, 12, 0, 0, 0, time.UTC)
	payload := url.Values{
		"token":      {"pat_aaaa1111bbbb2222cccc3333dddd4444"},
		"pat_id":     {"pat-uuid-1"},
		"user_id":    {"u_1"},
		"user_email": {"alice@example.com"},
		"expires_at": {exp.Format(time.RFC3339)},
	}
	captured := stubBrowser(t, func(target string) {
		simulateUICallback(t, target, payload)
	})

	var stdout bytes.Buffer
	if err := runLogin(context.Background(), &stdout, 90, false); err != nil {
		t.Fatalf("runLogin: %v", err)
	}

	if !strings.HasPrefix(*captured, "https://test.example/cli/auth?") {
		t.Errorf("captured URL prefix wrong: %q", *captured)
	}
	if !strings.Contains(*captured, "expires_in_days=90") {
		t.Errorf("expires_in_days missing from URL: %q", *captured)
	}

	pat, err := config.Load()
	if err != nil {
		t.Fatalf("config.Load: %v", err)
	}
	if pat.Token != "pat_aaaa1111bbbb2222cccc3333dddd4444" {
		t.Errorf("Token = %q", pat.Token)
	}
	if pat.UserEmail != "alice@example.com" {
		t.Errorf("UserEmail = %q", pat.UserEmail)
	}
	if !pat.ExpiresAt.Equal(exp) {
		t.Errorf("ExpiresAt = %v, want %v", pat.ExpiresAt, exp)
	}

	out := stdout.String()
	if !strings.Contains(out, "alice@example.com") {
		t.Errorf("stdout missing email: %q", out)
	}
}

func TestRunLogin_NoBrowserPrintsURL(t *testing.T) {
	t.Setenv("HOME", t.TempDir())
	t.Setenv("EVERSCRIBE_UI_URL_OVERRIDE", "https://test.example")

	// Browser opener should NOT be called when --no-browser is set.
	browserCalled := false
	prev := browserOpener
	t.Cleanup(func() { browserOpener = prev })
	browserOpener = func(target string) error {
		browserCalled = true
		return nil
	}

	// Drive the callback ourselves by reading the URL from stdout.
	type result struct {
		err error
		out string
	}
	done := make(chan result, 1)
	stdout := &bytes.Buffer{}
	go func() {
		err := runLogin(context.Background(), stdout, 0, true)
		done <- result{err: err, out: stdout.String()}
	}()

	// Poll stdout for the URL the CLI printed.
	var loginURL string
	deadline := time.Now().Add(2 * time.Second)
	for time.Now().Before(deadline) {
		s := stdout.String()
		if i := strings.Index(s, "https://test.example/cli/auth?"); i >= 0 {
			rest := s[i:]
			end := strings.IndexAny(rest, " \n")
			if end < 0 {
				end = len(rest)
			}
			loginURL = rest[:end]
			break
		}
		time.Sleep(10 * time.Millisecond)
	}
	if loginURL == "" {
		t.Fatalf("login URL never printed; stdout: %q", stdout.String())
	}

	simulateUICallback(t, loginURL, url.Values{
		"token":      {"pat_x"},
		"user_email": {"bob@example.com"},
	})

	res := <-done
	if res.err != nil {
		t.Fatalf("runLogin: %v", res.err)
	}
	if browserCalled {
		t.Errorf("browser opener called even with --no-browser")
	}
	// expires_in_days=0 means the flag was zero — buildLoginURL should
	// omit it from the query string.
	if strings.Contains(loginURL, "expires_in_days") {
		t.Errorf("expires_in_days should be absent when 0: %q", loginURL)
	}
}

func TestRunLogin_TimeoutErrors(t *testing.T) {
	t.Setenv("HOME", t.TempDir())
	t.Setenv("EVERSCRIBE_UI_URL_OVERRIDE", "https://test.example")

	// Stub browser to do nothing — no callback will ever arrive.
	prev := browserOpener
	t.Cleanup(func() { browserOpener = prev })
	browserOpener = func(string) error { return nil }

	// Cancel the parent context before runLogin's 2-min timer would
	// fire — keeps the test fast. runLogin's internal context inherits
	// the cancellation.
	ctx, cancel := context.WithCancel(context.Background())
	go func() {
		time.Sleep(50 * time.Millisecond)
		cancel()
	}()

	err := runLogin(ctx, &bytes.Buffer{}, 30, true)
	if err == nil {
		t.Fatal("runLogin returned nil after cancel; want context error")
	}
}

func TestBuildLoginURL_OmitsZeroExpiry(t *testing.T) {
	t.Setenv("EVERSCRIBE_UI_URL_OVERRIDE", "https://test.example")

	got := buildLoginURL(45123, "abc-123", 0)
	if strings.Contains(got, "expires_in_days") {
		t.Errorf("URL should omit expires_in_days when 0: %q", got)
	}
	if !strings.Contains(got, "callback_port=45123") || !strings.Contains(got, "state=abc-123") {
		t.Errorf("URL missing required params: %q", got)
	}
}

func TestGenerateState_Unique(t *testing.T) {
	a, err := generateState()
	if err != nil {
		t.Fatal(err)
	}
	b, err := generateState()
	if err != nil {
		t.Fatal(err)
	}
	if a == b {
		t.Errorf("two states collided: %q", a)
	}
	// 32 bytes base64url-encoded with no padding ≈ 43 chars.
	if len(a) < 40 {
		t.Errorf("state suspiciously short: %q", a)
	}
}
