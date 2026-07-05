// Package skills implements `es skills ...`, the distribution helper
// for Everscribe's Claude Code skills (e.g. `setup`). The actual
// instrumentation logic lives in the published SKILL.md content
// served from everscribe.io/manifests/; this command group just
// fetches the catalog and writes the markdown to the user's local
// Claude Code skills directory.
//
// The boundary is deliberate: this CLI fetches manifests at runtime;
// it doesn't depend on any server code at build time.
package skills

import "github.com/spf13/cobra"

// NewCmd returns the `es skills` command tree.
func NewCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:     "skills",
		Aliases: []string{"skill"},
		Short:   "Browse and install Everscribe Claude Code skills",
	}
	cmd.AddCommand(
		newListCmd(),
		newDescribeCmd(),
		newDownloadCmd(),
	)
	return cmd
}
