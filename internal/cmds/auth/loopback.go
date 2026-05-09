package auth

import (
	"context"
	"crypto/subtle"
	"errors"
	"fmt"
	"net"
	"net/http"
	"time"
)

// loopbackResult is what the UI's auto-submitting form delivers back
// to the CLI's local HTTP server. Field names match the form input
// names rendered by the cli-auth-success.gohtml template in the
// monorepo UI.
type loopbackResult struct {
	Token     string
	PATID     string
	UserID    string
	UserEmail string
	ExpiresAt time.Time // zero if the user disabled expiry
}

// loopbackServer is a single-shot HTTP server bound to a random
// localhost port. It accepts one POST to /callback, validates the
// state nonce, and delivers either a parsed result or an error to
// callers waiting on Wait.
//
// "Single-shot" means: the first valid POST wins; subsequent POSTs
// (e.g. a reload of the success page, or a stray attacker probe) are
// ignored on the channel side, though they may still receive an HTTP
// response. This is fine because Wait returns immediately on the
// first delivery and the caller is expected to Close the server.
type loopbackServer struct {
	state    string
	listener net.Listener
	server   *http.Server
	result   chan loopbackOutcome
}

type loopbackOutcome struct {
	res *loopbackResult
	err error
}

func newLoopbackServer(state string) (*loopbackServer, error) {
	if state == "" {
		return nil, errors.New("loopback state must not be empty")
	}
	l, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		return nil, fmt.Errorf("bind loopback: %w", err)
	}
	s := &loopbackServer{
		state:    state,
		listener: l,
		result:   make(chan loopbackOutcome, 1),
	}
	mux := http.NewServeMux()
	mux.HandleFunc("/callback", s.handleCallback)
	s.server = &http.Server{
		Handler:           mux,
		ReadHeaderTimeout: 10 * time.Second,
	}
	go func() { _ = s.server.Serve(l) }()
	return s, nil
}

// Port returns the kernel-assigned port the server is listening on.
func (s *loopbackServer) Port() int {
	return s.listener.Addr().(*net.TCPAddr).Port
}

// Wait blocks until a callback is delivered or ctx is cancelled.
// Returns the parsed result on success, or the underlying error on
// failure (state mismatch, missing token, malformed payload, or
// ctx.Err() if the wait timed out).
func (s *loopbackServer) Wait(ctx context.Context) (*loopbackResult, error) {
	select {
	case out := <-s.result:
		return out.res, out.err
	case <-ctx.Done():
		return nil, ctx.Err()
	}
}

// Close shuts the HTTP server down. Safe to call multiple times.
func (s *loopbackServer) Close() error {
	return s.server.Close()
}

func (s *loopbackServer) handleCallback(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "POST required", http.StatusMethodNotAllowed)
		return
	}
	if err := r.ParseForm(); err != nil {
		http.Error(w, "invalid form data", http.StatusBadRequest)
		s.deliver(loopbackOutcome{err: fmt.Errorf("parse form: %w", err)})
		return
	}

	// Constant-time comparison out of habit; the state is single-use
	// and never reused, so timing attacks here are largely theoretical.
	got := r.FormValue("state")
	if subtle.ConstantTimeCompare([]byte(got), []byte(s.state)) != 1 {
		http.Error(w, "state mismatch", http.StatusBadRequest)
		s.deliver(loopbackOutcome{err: errors.New("state mismatch — refusing to accept callback")})
		return
	}

	token := r.FormValue("token")
	if token == "" {
		http.Error(w, "missing token", http.StatusBadRequest)
		s.deliver(loopbackOutcome{err: errors.New("callback delivered no token")})
		return
	}

	expiresAt, err := parseOptionalTime(r.FormValue("expires_at"))
	if err != nil {
		http.Error(w, "invalid expires_at", http.StatusBadRequest)
		s.deliver(loopbackOutcome{err: fmt.Errorf("parse expires_at: %w", err)})
		return
	}

	res := &loopbackResult{
		Token:     token,
		PATID:     r.FormValue("pat_id"),
		UserID:    r.FormValue("user_id"),
		UserEmail: r.FormValue("user_email"),
		ExpiresAt: expiresAt,
	}

	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	_, _ = w.Write([]byte(loopbackDoneHTML))

	s.deliver(loopbackOutcome{res: res})
}

// deliver sends out on the result channel, or drops it if the channel
// already holds a value. Keeps a duplicate or attacker-probe POST from
// blocking the handler goroutine.
func (s *loopbackServer) deliver(out loopbackOutcome) {
	select {
	case s.result <- out:
	default:
	}
}

// parseOptionalTime parses an RFC3339 timestamp or returns the zero
// time if the input is empty (the user opted out of expiry).
func parseOptionalTime(s string) (time.Time, error) {
	if s == "" {
		return time.Time{}, nil
	}
	return time.Parse(time.RFC3339, s)
}

const loopbackDoneHTML = `<!DOCTYPE html>
<html lang="en">
<head>
    <meta charset="UTF-8">
    <title>Everscribe CLI</title>
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
    <style>
        body { font-family: -apple-system, BlinkMacSystemFont, system-ui, sans-serif; display: flex; align-items: center; justify-content: center; min-height: 100vh; margin: 0; background: #f5f5f7; color: #1d1d1f; }
        .card { background: #fff; padding: 2rem 3rem; border-radius: 12px; box-shadow: 0 4px 14px rgba(0,0,0,0.08); text-align: center; max-width: 400px; }
        h1 { margin: 0 0 0.5rem 0; font-size: 1.5rem; }
        p { margin: 0; color: #6e6e73; line-height: 1.5; }
    </style>
</head>
<body>
    <div class="card">
        <h1>Login complete</h1>
        <p>You can close this tab and return to your terminal.</p>
    </div>
</body>
</html>
`
