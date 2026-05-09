package auth

import (
	"errors"

	"github.com/spf13/cobra"
)

func newLoginCmd() *cobra.Command {
	var (
		expiresInDays int
		noBrowser     bool
	)
	cmd := &cobra.Command{
		Use:   "login",
		Short: "Authenticate by minting a personal access token via the browser",
		RunE: func(cmd *cobra.Command, args []string) error {
			_ = expiresInDays
			_ = noBrowser
			return errors.New("not implemented (step 3)")
		},
	}
	cmd.Flags().IntVar(&expiresInDays, "expires-in-days", 90, "PAT lifetime in days (max 1825; 0 = no expiry)")
	cmd.Flags().BoolVar(&noBrowser, "no-browser", false, "print the login URL instead of opening it in a browser")
	return cmd
}
