package auth

import (
	"errors"

	"github.com/spf13/cobra"
)

func newLogoutCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "logout",
		Short: "Revoke the saved PAT and remove the local session file",
		RunE: func(cmd *cobra.Command, args []string) error {
			return errors.New("not implemented (step 3)")
		},
	}
}
