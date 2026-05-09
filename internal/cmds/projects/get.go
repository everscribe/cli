package projects

import (
	"errors"

	"github.com/spf13/cobra"
)

func newGetCmd() *cobra.Command {
	var format string
	cmd := &cobra.Command{
		Use:   "get <id>",
		Short: "Fetch a single project by ID",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			_ = format
			return errors.New("not implemented (step 4)")
		},
	}
	cmd.Flags().StringVar(&format, "format", "table", "output format: table | json | yaml")
	return cmd
}
