package auth

import (
	"errors"
	"fmt"
	"io"

	"github.com/spf13/cobra"

	"github.com/everscribe/cli/internal/config"
)

func newLogoutCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "logout",
		Short: "Remove the local session",
		Long: `Remove the saved PAT from ~/.config/everscribe/pat.json.

The PAT itself remains valid server-side until it expires (90 days by
default). The Everscribe API explicitly rejects PAT-authenticated
callers attempting to revoke their own token — by design, so a leaked
PAT can't quietly self-revoke out of the audit trail. To invalidate
the token across every device that has a copy, revoke it from
Developer Settings (https://everscribe.io/settings/developer).`,
		RunE: func(cmd *cobra.Command, args []string) error {
			return runLogout(cmd.OutOrStdout())
		},
	}
}

func runLogout(stdout io.Writer) error {
	if _, err := config.Load(); errors.Is(err, config.ErrNotLoggedIn) {
		fmt.Fprintln(stdout, "Already logged out.")
		return nil
	} else if err != nil {
		return err
	}
	if err := config.Delete(); err != nil {
		return fmt.Errorf("remove local session: %w", err)
	}
	fmt.Fprintln(stdout, "Logged out.")
	return nil
}
