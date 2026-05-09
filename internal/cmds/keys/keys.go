// Package keys implements `es keys ...` for project-scoped ingest API keys.
package keys

import "github.com/spf13/cobra"

// NewCmd returns the `es keys` command tree.
func NewCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:     "keys",
		Aliases: []string{"key"},
		Short:   "Manage project-scoped ingest API keys",
	}
	cmd.AddCommand(newListCmd(), newCreateCmd(), newRevokeCmd())
	return cmd
}
