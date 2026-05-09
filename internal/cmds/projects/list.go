package projects

import (
	"errors"

	"github.com/spf13/cobra"
)

func newListCmd() *cobra.Command {
	var format string
	cmd := &cobra.Command{
		Use:   "list",
		Short: "List your projects",
		RunE: func(cmd *cobra.Command, args []string) error {
			_ = format
			return errors.New("not implemented (step 4)")
		},
	}
	cmd.Flags().StringVar(&format, "format", "table", "output format: table | json | yaml")
	return cmd
}
