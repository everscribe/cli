package events

import (
	"errors"

	"github.com/spf13/cobra"
)

func newDiffCmd() *cobra.Command {
	var project string
	cmd := &cobra.Command{
		Use:   "diff <event-id>",
		Short: "Show the before/after diff for an event with a change record",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			_ = project
			return errors.New("not implemented (step 6)")
		},
	}
	cmd.Flags().StringVar(&project, "project", "", "project ID (required)")
	_ = cmd.MarkFlagRequired("project")
	return cmd
}
