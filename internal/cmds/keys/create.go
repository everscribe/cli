package keys

import (
	"errors"

	"github.com/spf13/cobra"
)

func newCreateCmd() *cobra.Command {
	var (
		project string
		name    string
		format  string
	)
	cmd := &cobra.Command{
		Use:   "create",
		Short: "Create a new ingest API key in a project",
		RunE: func(cmd *cobra.Command, args []string) error {
			_ = project
			_ = name
			_ = format
			return errors.New("not implemented (step 5)")
		},
	}
	cmd.Flags().StringVar(&project, "project", "", "project ID (required)")
	cmd.Flags().StringVar(&name, "name", "", "key name (required)")
	cmd.Flags().StringVar(&format, "format", "table", "output format: table | json | yaml")
	_ = cmd.MarkFlagRequired("project")
	_ = cmd.MarkFlagRequired("name")
	return cmd
}
