package projects

import (
	"errors"

	"github.com/spf13/cobra"
)

func newCreateCmd() *cobra.Command {
	var (
		name   string
		format string
	)
	cmd := &cobra.Command{
		Use:   "create",
		Short: "Create a new project",
		RunE: func(cmd *cobra.Command, args []string) error {
			_ = name
			_ = format
			return errors.New("not implemented (step 4)")
		},
	}
	cmd.Flags().StringVar(&name, "name", "", "project name (required)")
	cmd.Flags().StringVar(&format, "format", "table", "output format: table | json | yaml")
	_ = cmd.MarkFlagRequired("name")
	return cmd
}
