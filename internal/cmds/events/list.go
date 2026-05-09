package events

import (
	"errors"

	"github.com/spf13/cobra"
)

func newListCmd() *cobra.Command {
	var (
		project string
		limit   int
		all     bool
		format  string
		filters filterFlags
	)
	cmd := &cobra.Command{
		Use:   "list",
		Short: "List events in a project (most recent first)",
		RunE: func(cmd *cobra.Command, args []string) error {
			_ = project
			_ = limit
			_ = all
			_ = format
			_ = filters
			return errors.New("not implemented (step 6)")
		},
	}
	cmd.Flags().StringVar(&project, "project", "", "project ID (required)")
	cmd.Flags().IntVar(&limit, "limit", 50, "max events per page")
	cmd.Flags().BoolVar(&all, "all", false, "fetch all pages instead of one")
	cmd.Flags().StringVar(&format, "format", "table", "output format: table | json | yaml")
	addFilterFlags(cmd, &filters)
	_ = cmd.MarkFlagRequired("project")
	return cmd
}
