package auth

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync/atomic"
	"testing"
	"time"

	"github.com/stretchr/testify/require"

	"github.com/everscribe/cli/internal/config"
	"github.com/everscribe/cli/internal/types"
)

// stubBrowser swaps browserOpener for a function that captures the
// verification URL. Restore happens via t.Cleanup.
func stubBrowser(t *testing.T) *string {
	t.Helper()
	captured := new(string)
	prev := browserOpener
	t.Cleanup(func() { browserOpener = prev })
	browserOpener = func(target string) error {
		*captured = target
		return nil
	}
	return captured
}

// deviceFlowServer returns an httptest.Server that:
//   - on /v1/cli/device-codes returns a fixed (user_code, device_code) pair
//   - on /v1/cli/device-tokens, returns 400 authorization_pending until
//     `approveAfter` polls have happened, then returns 200 with the
//     fixture PAT data
func deviceFlowServer(t *testing.T, approveAfter int32) *httptest.Server {
	t.Helper()
	var polls int32
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/v1/cli/device-codes":
			_ = json.NewEncoder(w).Encode(types.IssueDeviceCodeResponse{
				UserCode:                "ABCD-EFGH",
				DeviceCode:              "device-code-xyz",
				VerificationURI:         "https://everscribe.io/cli/verify",
				VerificationURIComplete: "https://everscribe.io/cli/verify?user_code=ABCD-EFGH",
				ExpiresIn:               60,
				Interval:                1,
			})
		case "/v1/cli/device-tokens":
			n := atomic.AddInt32(&polls, 1)
			if n < approveAfter {
				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(http.StatusBadRequest)
				_, _ = w.Write([]byte(`{"error":"authorization_pending"}`))
				return
			}
			_ = json.NewEncoder(w).Encode(types.ExchangeDeviceCodeResponse{
				Plaintext: "pat_secret",
				PATID:     "pat-1",
				UserID:    "u_1",
				UserEmail: "alice@example.com",
				PATExpiresAt: time.Date(2026, 8, 7, 12, 0, 0, 0, time.UTC),
			})
		default:
			http.NotFound(w, r)
		}
	}))
	t.Cleanup(srv.Close)
	return srv
}

func TestRunLogin_HappyPath(t *testing.T) {
	t.Setenv("HOME", t.TempDir())
	srv := deviceFlowServer(t, 1) // approve on first poll
	t.Setenv("EVERSCRIBE_API_URL_OVERRIDE", srv.URL)

	captured := stubBrowser(t)

	var stdout bytes.Buffer
	require.NoError(t, runLogin(t.Context(), &stdout, false))

	require.Equal(t, "https://everscribe.io/cli/verify?user_code=ABCD-EFGH", *captured)

	out := stdout.String()
	require.Contains(t, out, "ABCD-EFGH", "user_code should be displayed")
	require.Contains(t, out, "https://everscribe.io/cli/verify")
	require.Contains(t, out, "Logged in as alice@example.com")

	pat, err := config.Load()
	require.NoError(t, err)
	require.Equal(t, "pat_secret", pat.Token)
	require.Equal(t, "alice@example.com", pat.UserEmail)
	require.Equal(t, "pat-1", pat.PATID)
}

func TestRunLogin_NoBrowserSkipsOpener(t *testing.T) {
	t.Setenv("HOME", t.TempDir())
	srv := deviceFlowServer(t, 1)
	t.Setenv("EVERSCRIBE_API_URL_OVERRIDE", srv.URL)

	called := false
	prev := browserOpener
	t.Cleanup(func() { browserOpener = prev })
	browserOpener = func(string) error {
		called = true
		return nil
	}

	var stdout bytes.Buffer
	require.NoError(t, runLogin(t.Context(), &stdout, true))

	require.False(t, called, "browser opener must not run with --no-browser")
}

func TestRunLogin_PollsUntilApproved(t *testing.T) {
	t.Setenv("HOME", t.TempDir())
	srv := deviceFlowServer(t, 3) // pending twice, approve on third
	t.Setenv("EVERSCRIBE_API_URL_OVERRIDE", srv.URL)

	stubBrowser(t)
	require.NoError(t, runLogin(t.Context(), &bytes.Buffer{}, true))

	pat, err := config.Load()
	require.NoError(t, err)
	require.Equal(t, "pat_secret", pat.Token)
}

func TestRunLogin_ExpiredTokenErrors(t *testing.T) {
	t.Setenv("HOME", t.TempDir())
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/v1/cli/device-codes":
			_ = json.NewEncoder(w).Encode(types.IssueDeviceCodeResponse{
				UserCode:   "ABCD-EFGH",
				DeviceCode: "x",
				ExpiresIn:  60,
				Interval:   1,
			})
		case "/v1/cli/device-tokens":
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusBadRequest)
			_, _ = w.Write([]byte(`{"error":"expired_token"}`))
		}
	}))
	t.Cleanup(srv.Close)
	t.Setenv("EVERSCRIBE_API_URL_OVERRIDE", srv.URL)

	stubBrowser(t)
	err := runLogin(t.Context(), &bytes.Buffer{}, true)
	require.Error(t, err)
	require.ErrorContains(t, err, "expired")
}

func TestRunLogin_ContextCancelExits(t *testing.T) {
	t.Setenv("HOME", t.TempDir())
	srv := deviceFlowServer(t, 9999) // never approves
	t.Setenv("EVERSCRIBE_API_URL_OVERRIDE", srv.URL)

	stubBrowser(t)
	ctx, cancel := context.WithCancel(t.Context())
	go func() {
		time.Sleep(50 * time.Millisecond)
		cancel()
	}()
	err := runLogin(ctx, &bytes.Buffer{}, true)
	require.Error(t, err)
}

func TestRunLogin_IssueErrorPropagates(t *testing.T) {
	t.Setenv("HOME", t.TempDir())
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		http.Error(w, "boom", http.StatusInternalServerError)
	}))
	t.Cleanup(srv.Close)
	t.Setenv("EVERSCRIBE_API_URL_OVERRIDE", srv.URL)

	stubBrowser(t)
	err := runLogin(t.Context(), &bytes.Buffer{}, true)
	require.Error(t, err)
	require.True(t, strings.Contains(err.Error(), "device code"),
		"want context about issue step, got %v", err)
}
