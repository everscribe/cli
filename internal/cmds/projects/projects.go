// Package projects implements `es projects ...`.
package projects

import "github.com/spf13/cobra"

// NewCmd returns the `es projects` command tree.
func NewCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:     "projects",
		Aliases: []string{"project"},
		Short:   "Manage Everscribe projects",
	}
	cmd.AddCommand(
		newListCmd(),
		newGetCmd(),
		newCreateCmd(),
		newUpdateCmd(),
		newDeleteCmd(),
	)
	return cmd
}
