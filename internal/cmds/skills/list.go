package skills

import (
	"context"
	"io"

	"github.com/spf13/cobra"
)

func newListCmd() *cobra.Command {
	var format string
	cmd := &cobra.Command{
		Use:   "list",
		Short: "List skills available on everscribe.io",
		RunE: func(cmd *cobra.Command, args []string) error {
			return runList(cmd.Context(), cmd.OutOrStdout(), format)
		},
	}
	cmd.Flags().StringVar(&format, "format", "table", "output format: table | json | yaml")
	return cmd
}

func runList(ctx context.Context, stdout io.Writer, format string) error {
	cat, err := NewFetcher().Catalog(ctx)
	if err != nil {
		return err
	}
	return renderSkills(stdout, format, cat.Skills)
}
