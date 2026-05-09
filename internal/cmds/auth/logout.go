package auth

import (
	"context"
	"errors"
	"fmt"
	"io"
	"net/http"

	"github.com/spf13/cobra"

	"github.com/everscribe/cli/internal/client"
	"github.com/everscribe/cli/internal/config"
)

func newLogoutCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "logout",
		Short: "Revoke the saved PAT and remove the local session file",
		RunE: func(cmd *cobra.Command, args []string) error {
			return runLogout(cmd.Context(), cmd.OutOrStdout())
		},
	}
}

func runLogout(ctx context.Context, stdout io.Writer) error {
	pat, err := config.Load()
	if errors.Is(err, config.ErrNotLoggedIn) {
		fmt.Fprintln(stdout, "Already logged out.")
		return nil
	}
	if err != nil {
		return err
	}

	// Best-effort server-side revoke. If the token is already invalid,
	// gone, or unreachable, we still proceed to clear the local file —
	// otherwise a stale or misconfigured token would wedge the user
	// out of `es auth login`.
	if pat.PATID != "" {
		c := client.New(pat.Token)
		if err := c.Do(ctx, http.MethodDelete, "/v1/pats/"+pat.PATID, nil, nil, nil); err != nil {
			if !client.IsUnauthorized(err) && !client.IsNotFound(err) {
				fmt.Fprintf(stdout, "Warning: server-side revoke failed (%v). Removing local session anyway.\n", err)
			}
		}
	}

	if err := config.Delete(); err != nil {
		return fmt.Errorf("remove local session: %w", err)
	}
	fmt.Fprintln(stdout, "Logged out.")
	return nil
}
