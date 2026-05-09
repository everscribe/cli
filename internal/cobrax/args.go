// Package cobrax provides small extensions over spf13/cobra. Lives at
// the package root rather than inside internal/cmds so subcommand
// packages can import it without creating a cycle through cmds/root.go.
package cobrax

import (
	"fmt"

	"github.com/spf13/cobra"
)

// RequireArgs returns a cobra.PositionalArgs validator that requires
// exactly the named positional arguments. Unlike cobra.ExactArgs, the
// error names the specific missing argument so the user sees
//
//	Error: missing required argument: <event-id>
//
// instead of cobra's default
//
//	Error: accepts 1 arg(s), received 0
//
// Names should be the bare argument label (e.g. "event-id", "id"); the
// helper wraps each in <…> when rendering the error.
func RequireArgs(names ...string) cobra.PositionalArgs {
	return func(cmd *cobra.Command, args []string) error {
		switch {
		case len(args) < len(names):
			return fmt.Errorf("missing required argument: <%s>", names[len(args)])
		case len(args) > len(names):
			extras := args[len(names):]
			return fmt.Errorf("unexpected extra argument(s): %v", extras)
		}
		return nil
	}
}
