package collaborators

import (
	"context"
	"io"

	"github.com/spf13/cobra"

	"github.com/everscribe/cli/internal/client"
	"github.com/everscribe/cli/internal/config"
)

func newListCmd() *cobra.Command {
	var format string
	cmd := &cobra.Command{
		Use:   "list",
		Short: "List your collaborators",
		RunE: func(cmd *cobra.Command, args []string) error {
			return runList(cmd.Context(), cmd.OutOrStdout(), format)
		},
	}
	cmd.Flags().StringVar(&format, "format", "table", "output format: table | json | yaml")
	return cmd
}

func runList(ctx context.Context, stdout io.Writer, format string) error {
	pat, err := config.Load()
	if err != nil {
		return err
	}
	c := client.New(pat.Token)
	cols, err := c.ListCollaborators(ctx)
	if err != nil {
		return err
	}
	return renderCollaborators(stdout, format, cols)
}
