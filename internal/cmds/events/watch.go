package events

import (
	"errors"
	"time"

	"github.com/spf13/cobra"
)

func newWatchCmd() *cobra.Command {
	var (
		project  string
		interval time.Duration
		format   string
		filters  filterFlags
	)
	cmd := &cobra.Command{
		Use:   "watch",
		Short: "Stream new events by polling, dedupe by event ID",
		RunE: func(cmd *cobra.Command, args []string) error {
			_ = project
			_ = interval
			_ = format
			_ = filters
			return errors.New("not implemented (step 6)")
		},
	}
	cmd.Flags().StringVar(&project, "project", "", "project ID (required)")
	cmd.Flags().DurationVar(&interval, "interval", 3*time.Second, "polling interval")
	cmd.Flags().StringVar(&format, "format", "table", "output format: table | json | yaml")
	addFilterFlags(cmd, &filters)
	_ = cmd.MarkFlagRequired("project")
	return cmd
}
