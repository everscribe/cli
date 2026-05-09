package keys

import (
	"errors"

	"github.com/spf13/cobra"
)

func newListCmd() *cobra.Command {
	var (
		project string
		format  string
	)
	cmd := &cobra.Command{
		Use:   "list",
		Short: "List active API keys for a project",
		RunE: func(cmd *cobra.Command, args []string) error {
			_ = project
			_ = format
			return errors.New("not implemented (step 5)")
		},
	}
	cmd.Flags().StringVar(&project, "project", "", "project ID (required)")
	cmd.Flags().StringVar(&format, "format", "table", "output format: table | json | yaml")
	_ = cmd.MarkFlagRequired("project")
	return cmd
}
