package auth

import (
	"context"
	"crypto/rand"
	"encoding/base64"
	"errors"
	"fmt"
	"io"
	"net/url"
	"os/exec"
	"runtime"
	"strconv"
	"time"

	"github.com/spf13/cobra"

	"github.com/everscribe/cli/internal/client"
	"github.com/everscribe/cli/internal/config"
)

const (
	cliAuthPath  = "/cli/auth"
	loginTimeout = 2 * time.Minute
	stateBytes   = 32
)

// browserOpener is the indirection that makes login_test.go possible:
// tests swap in a fake that captures the URL instead of spawning a
// real browser.
var browserOpener = openBrowser

func newLoginCmd() *cobra.Command {
	var (
		expiresInDays int
		noBrowser     bool
	)
	cmd := &cobra.Command{
		Use:   "login",
		Short: "Authenticate by minting a personal access token via the browser",
		RunE: func(cmd *cobra.Command, args []string) error {
			return runLogin(cmd.Context(), cmd.OutOrStdout(), expiresInDays, noBrowser)
		},
	}
	cmd.Flags().IntVar(&expiresInDays, "expires-in-days", 90, "PAT lifetime in days (max 1825; 0 = no expiry)")
	cmd.Flags().BoolVar(&noBrowser, "no-browser", false, "print the login URL instead of opening it in a browser")
	return cmd
}

// runLogin runs the loopback browser-callback flow:
//
//  1. generate a 32-byte random state
//  2. bind a localhost listener and start the loopback HTTP server
//  3. open the user's browser at <UI>/cli/auth?callback_port=&state=
//     (or print the URL with --no-browser)
//  4. wait up to loginTimeout for the UI's auto-submitting form to
//     POST /callback with the freshly minted PAT
//  5. persist the result to ~/.config/everscribe/pat.json (mode 0600)
func runLogin(ctx context.Context, stdout io.Writer, expiresInDays int, noBrowser bool) error {
	state, err := generateState()
	if err != nil {
		return fmt.Errorf("generate state: %w", err)
	}

	srv, err := newLoopbackServer(state)
	if err != nil {
		return err
	}
	defer srv.Close()

	loginURL := buildLoginURL(srv.Port(), state, expiresInDays)

	if noBrowser {
		fmt.Fprintf(stdout, "Open this URL in your browser to log in:\n\n  %s\n\n", loginURL)
	} else {
		fmt.Fprintln(stdout, "Opening browser to authorize the CLI…")
		if err := browserOpener(loginURL); err != nil {
			fmt.Fprintf(stdout, "Couldn't open browser (%v).\nOpen this URL manually:\n\n  %s\n\n", err, loginURL)
		}
	}
	fmt.Fprintf(stdout, "Waiting for callback (timeout %s)…\n", loginTimeout)

	waitCtx, cancel := context.WithTimeout(ctx, loginTimeout)
	defer cancel()

	res, err := srv.Wait(waitCtx)
	if err != nil {
		if errors.Is(err, context.DeadlineExceeded) {
			return fmt.Errorf("timed out after %s waiting for browser callback", loginTimeout)
		}
		return err
	}

	pat := &config.PAT{
		Token:     res.Token,
		PATID:     res.PATID,
		UserID:    res.UserID,
		UserEmail: res.UserEmail,
		ExpiresAt: res.ExpiresAt,
	}
	if err := config.Save(pat); err != nil {
		return fmt.Errorf("save token: %w", err)
	}

	identifier := pat.UserEmail
	if identifier == "" {
		identifier = pat.UserID
	}
	fmt.Fprintf(stdout, "Logged in as %s.\n", identifier)
	return nil
}

func buildLoginURL(port int, state string, expiresInDays int) string {
	q := url.Values{
		"callback_port": {strconv.Itoa(port)},
		"state":         {state},
	}
	if expiresInDays > 0 {
		q.Set("expires_in_days", strconv.Itoa(expiresInDays))
	}
	return client.UIBaseURL() + cliAuthPath + "?" + q.Encode()
}

func generateState() (string, error) {
	buf := make([]byte, stateBytes)
	if _, err := rand.Read(buf); err != nil {
		return "", err
	}
	return base64.RawURLEncoding.EncodeToString(buf), nil
}

// openBrowser launches the OS default browser at targetURL. Detached
// from the parent process: Start (not Run) so we don't block waiting
// for the browser to exit.
func openBrowser(targetURL string) error {
	var cmd string
	var args []string
	switch runtime.GOOS {
	case "darwin":
		cmd = "open"
		args = []string{targetURL}
	case "windows":
		cmd = "rundll32"
		args = []string{"url.dll,FileProtocolHandler", targetURL}
	default:
		cmd = "xdg-open"
		args = []string{targetURL}
	}
	return exec.Command(cmd, args...).Start()
}
