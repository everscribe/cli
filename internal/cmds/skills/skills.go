// Package skills implements `es skills ...` — the distribution helper
// for Everscribe's Claude Code skills (e.g. `setup`). The actual
// instrumentation logic lives in the published SKILL.md content
// served from everscribe.io/manifests/; this command group just
// fetches the catalog and writes the markdown to the user's local
// Claude Code skills directory.
//
// See docs/auto-instrument-spec.md → "Code organization" for the
// boundary: this CLI fetches manifests at runtime; it doesn't depend
// on any monorepo code at build time.
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
		newDownloadCmd(),
	)
	return cmd
}
