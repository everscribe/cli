package auth

import (
	"errors"

	"github.com/spf13/cobra"
)

func newWhoamiCmd() *cobra.Command {
	var format string
	cmd := &cobra.Command{
		Use:   "whoami",
		Short: "Print the user identity associated with the saved PAT",
		RunE: func(cmd *cobra.Command, args []string) error {
			_ = format
			return errors.New("not implemented (step 3)")
		},
	}
	cmd.Flags().StringVar(&format, "format", "table", "output format: table | json | yaml")
	return cmd
}
