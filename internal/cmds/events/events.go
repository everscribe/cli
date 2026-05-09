// Package events implements `es events ...`.
package events

import "github.com/spf13/cobra"

// NewCmd returns the `es events` command tree.
func NewCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:     "events",
		Aliases: []string{"event"},
		Short:   "List, watch, describe, and diff project events",
	}
	cmd.AddCommand(
		newListCmd(),
		newWatchCmd(),
		newDescribeCmd(),
		newDiffCmd(),
	)
	return cmd
}

// filterFlags is the shared filter set for `events list` and `events watch`.
type filterFlags struct {
	since      string
	before     string
	action     string
	actor      string
	actorType  string
	targetType string
	tenant     string
}

// addFilterFlags attaches the shared filter flags to cmd. Each individual
// command also defines its own command-specific flags (limit, interval, etc.).
func addFilterFlags(cmd *cobra.Command, f *filterFlags) {
	cmd.Flags().StringVar(&f.since, "since", "", "include only events at or after this RFC3339 timestamp")
	cmd.Flags().StringVar(&f.before, "before", "", "include only events strictly before this RFC3339 timestamp")
	cmd.Flags().StringVar(&f.action, "action", "", "filter by action string (e.g. user.login)")
	cmd.Flags().StringVar(&f.actor, "actor", "", "filter by actor ID")
	cmd.Flags().StringVar(&f.actorType, "actor-type", "", "filter by actor type (e.g. user, service)")
	cmd.Flags().StringVar(&f.targetType, "target-type", "", "filter by target type")
	cmd.Flags().StringVar(&f.tenant, "tenant", "", "filter by tenant ID")
}
