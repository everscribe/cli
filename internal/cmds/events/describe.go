package events

import (
	"errors"

	"github.com/spf13/cobra"
)

func newDescribeCmd() *cobra.Command {
	var (
		project string
		format  string
	)
	cmd := &cobra.Command{
		Use:   "describe <event-id>",
		Short: "Print the full event with all nested fields decoded",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			_ = project
			_ = format
			return errors.New("not implemented (step 6)")
		},
	}
	cmd.Flags().StringVar(&project, "project", "", "project ID (required)")
	cmd.Flags().StringVar(&format, "format", "yaml", "output format: yaml | json")
	_ = cmd.MarkFlagRequired("project")
	return cmd
}
