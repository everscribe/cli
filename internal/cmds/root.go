// Package cmds assembles the cobra command tree for the es CLI.
package cmds

import (
	"github.com/spf13/cobra"

	"github.com/everscribe/cli/internal/cmds/auth"
	"github.com/everscribe/cli/internal/cmds/events"
	"github.com/everscribe/cli/internal/cmds/keys"
	"github.com/everscribe/cli/internal/cmds/projects"
	"github.com/everscribe/cli/internal/cmds/skills"
)

// NewRoot returns the top-level `es` command with all subcommands attached.
func NewRoot(version string) *cobra.Command {
	root := &cobra.Command{
		Use:           "es",
		Short:         "Everscribe command-line interface",
		Long:          "es is the Everscribe CLI for managing projects, API keys, and observing events.",
		Version:       version,
		SilenceUsage:  true,
		SilenceErrors: true,
	}

	root.AddCommand(
		auth.NewCmd(),
		projects.NewCmd(),
		keys.NewCmd(),
		events.NewCmd(),
		skills.NewCmd(),
	)
	return root
}
