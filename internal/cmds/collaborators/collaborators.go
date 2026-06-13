// Package collaborators implements `es collaborators ...` — managing the people
// you can share projects with.
package collaborators

import "github.com/spf13/cobra"

// NewCmd returns the `es collaborators` command tree.
func NewCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:     "collaborators",
		Aliases: []string{"collab", "collaborator"},
		Short:   "Manage collaborators you can share projects with",
	}
	cmd.AddCommand(
		newAddCmd(),
		newListCmd(),
		newRemoveCmd(),
	)
	return cmd
}
