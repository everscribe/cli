package auth

import (
	"context"
	"errors"
	"fmt"
	"io"
	"os/exec"
	"runtime"
	"time"

	"github.com/spf13/cobra"

	"github.com/everscribe/cli/internal/client"
	"github.com/everscribe/cli/internal/config"
)

// browserOpener is the indirection that makes login_test.go possible:
// tests swap in a fake that captures the URL instead of spawning a
// real browser.
var browserOpener = openBrowser

func newLoginCmd() *cobra.Command {
	var (
		noBrowser bool
	)
	cmd := &cobra.Command{
		Use:   "login",
		Short: "Authenticate by entering a one-time code in the browser",
		RunE: func(cmd *cobra.Command, args []string) error {
			return runLogin(cmd.Context(), cmd.OutOrStdout(), noBrowser)
		},
	}
	cmd.Flags().BoolVar(&noBrowser, "no-browser", false, "print the verification URL instead of opening it")
	return cmd
}

// runLogin runs the RFC 8628 device authorization grant:
//
//  1. POST /v1/cli/device-codes - receive (user_code, device_code) pair
//  2. show user_code + verification URL to the user (and open it in
//     a browser unless --no-browser)
//  3. poll POST /v1/cli/device-tokens at the server-suggested interval
//     until the user approves in the UI, the code expires, or ctx
//     is canceled
//  4. persist the resulting PAT to ~/.config/everscribe/pat.json
//
// PAT lifetime is server-controlled (90 days by default; the API can
// override). Users who want a different lifetime can mint via the
// /settings/developer page in the UI.
func runLogin(ctx context.Context, stdout io.Writer, noBrowser bool) error {
	c := client.New("")

	codes, err := c.IssueDeviceCode(ctx)
	if err != nil {
		return fmt.Errorf("requesting device code: %w", err)
	}

	fmt.Fprintln(stdout, "First copy your one-time code:")
	fmt.Fprintln(stdout)
	fmt.Fprintln(stdout, "    "+codes.UserCode)
	fmt.Fprintln(stdout)
	if noBrowser {
		fmt.Fprintln(stdout, "Then open this URL in your browser:")
		fmt.Fprintln(stdout)
		fmt.Fprintln(stdout, "    "+codes.VerificationURI)
	} else {
		fmt.Fprintln(stdout, "Then open this URL in your browser (we'll try to open it for you):")
		fmt.Fprintln(stdout)
		fmt.Fprintln(stdout, "    "+codes.VerificationURI)
		fmt.Fprintln(stdout)
		if err := browserOpener(codes.VerificationURIComplete); err != nil {
			fmt.Fprintf(stdout, "(Couldn't open browser: %v. Open the URL above manually.)\n", err)
		}
	}
	fmt.Fprintln(stdout)
	fmt.Fprintln(stdout, "Waiting for authorization... (press Ctrl+C to cancel)")

	interval := time.Duration(codes.Interval) * time.Second
	if interval <= 0 {
		interval = 5 * time.Second
	}
	deadline := time.Now().Add(time.Duration(codes.ExpiresIn) * time.Second)

	for {
		// Wait `interval` (or until context cancels) before each poll.
		// Polling immediately on entry would just waste a request - the
		// user hasn't even seen the code yet.
		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-time.After(interval):
		}

		if time.Now().After(deadline) {
			return errors.New("device code expired before authorization")
		}

		resp, err := c.ExchangeDeviceCode(ctx, codes.DeviceCode)
		if err == nil {
			pat := &config.PAT{
				Token:     resp.Plaintext,
				PATID:     resp.PATID,
				UserID:    resp.UserID,
				UserEmail: resp.UserEmail,
				ExpiresAt: resp.PATExpiresAt,
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

		switch {
		case client.IsAuthorizationPending(err):
			// Keep polling.
			continue
		case client.IsExpiredToken(err):
			return errors.New("device code expired before authorization")
		case client.IsAccessDenied(err):
			return errors.New("authorization denied")
		default:
			return fmt.Errorf("polling for authorization: %w", err)
		}
	}
}

// openBrowser launches the OS default browser at targetURL. Detached
// from the parent - Start (not Run) so we don't block on the browser
// process.
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
