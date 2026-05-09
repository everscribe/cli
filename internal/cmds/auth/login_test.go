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

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

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
	require.NoError(t, err, "parse target URL")

	port, err := strconv.Atoi(u.Query().Get("callback_port"))
	require.NoError(t, err, "parse callback_port")

	state := u.Query().Get("state")
	require.NotEmpty(t, state, "state missing from URL")

	// Echo the captured state through unless the caller already set
	// one — keeps a future state-handling regression in buildLoginURL
	// from silently passing.
	if payload.Get("state") == "" {
		payload.Set("state", state)
	}
	go func() {
		resp, err := http.PostForm("http://127.0.0.1:"+strconv.Itoa(port)+"/callback", payload)
		if !assert.NoError(t, err, "loopback POST") {
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
	require.NoError(t, runLogin(context.Background(), &stdout, 90, false))

	require.True(t, strings.HasPrefix(*captured, "https://test.example/cli/auth?"),
		"captured URL = %q", *captured)
	require.Contains(t, *captured, "expires_in_days=90")

	pat, err := config.Load()
	require.NoError(t, err)
	require.Equal(t, "pat_aaaa1111bbbb2222cccc3333dddd4444", pat.Token)
	require.Equal(t, "alice@example.com", pat.UserEmail)
	require.True(t, pat.ExpiresAt.Equal(exp))
	require.Contains(t, stdout.String(), "alice@example.com")
}

func TestRunLogin_NoBrowserPrintsURL(t *testing.T) {
	t.Setenv("HOME", t.TempDir())
	t.Setenv("EVERSCRIBE_UI_URL_OVERRIDE", "https://test.example")

	browserCalled := false
	prev := browserOpener
	t.Cleanup(func() { browserOpener = prev })
	browserOpener = func(target string) error {
		browserCalled = true
		return nil
	}

	type result struct {
		err error
	}
	done := make(chan result, 1)
	stdout := &bytes.Buffer{}
	go func() {
		done <- result{err: runLogin(context.Background(), stdout, 0, true)}
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
	require.NotEmpty(t, loginURL, "login URL never printed; stdout: %q", stdout.String())

	simulateUICallback(t, loginURL, url.Values{
		"token":      {"pat_x"},
		"user_email": {"bob@example.com"},
	})

	res := <-done
	require.NoError(t, res.err)
	require.False(t, browserCalled, "browser opener called even with --no-browser")
	require.NotContains(t, loginURL, "expires_in_days",
		"expires_in_days should be absent when 0: %q", loginURL)
}

func TestRunLogin_TimeoutErrors(t *testing.T) {
	t.Setenv("HOME", t.TempDir())
	t.Setenv("EVERSCRIBE_UI_URL_OVERRIDE", "https://test.example")

	prev := browserOpener
	t.Cleanup(func() { browserOpener = prev })
	browserOpener = func(string) error { return nil }

	// Cancel parent context before runLogin's 2-min timer fires —
	// keeps the test fast. runLogin's internal context inherits the
	// cancellation.
	ctx, cancel := context.WithCancel(context.Background())
	go func() {
		time.Sleep(50 * time.Millisecond)
		cancel()
	}()

	err := runLogin(ctx, &bytes.Buffer{}, 30, true)
	require.Error(t, err, "expected context error after cancel")
}

func TestBuildLoginURL_OmitsZeroExpiry(t *testing.T) {
	t.Setenv("EVERSCRIBE_UI_URL_OVERRIDE", "https://test.example")

	got := buildLoginURL(45123, "abc-123", 0)
	require.NotContains(t, got, "expires_in_days")
	require.Contains(t, got, "callback_port=45123")
	require.Contains(t, got, "state=abc-123")
}

func TestGenerateState_Unique(t *testing.T) {
	a, err := generateState()
	require.NoError(t, err)
	b, err := generateState()
	require.NoError(t, err)
	require.NotEqual(t, a, b, "two states collided")
	// 32 bytes base64url-encoded with no padding ≈ 43 chars.
	require.GreaterOrEqual(t, len(a), 40)
}
