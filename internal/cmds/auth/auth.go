// Package auth implements `es auth ...`.
package auth

import "github.com/spf13/cobra"

// NewCmd returns the `es auth` command tree.
func NewCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "auth",
		Short: "Authenticate the CLI to Everscribe",
	}
	cmd.AddCommand(newLoginCmd(), newLogoutCmd(), newWhoamiCmd())
	return cmd
}
