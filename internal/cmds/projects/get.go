package projects

import (
	"context"
	"io"

	"github.com/spf13/cobra"

	"github.com/everscribe/cli/internal/client"
	"github.com/everscribe/cli/internal/cobrax"
	"github.com/everscribe/cli/internal/config"
)

func newGetCmd() *cobra.Command {
	var format string
	cmd := &cobra.Command{
		Use:   "get <id>",
		Short: "Fetch a single project by ID",
		Args:  cobrax.RequireArgs("id"),
		RunE: func(cmd *cobra.Command, args []string) error {
			return runGet(cmd.Context(), cmd.OutOrStdout(), args[0], format)
		},
	}
	cmd.Flags().StringVar(&format, "format", "table", "output format: table | json | yaml")
	return cmd
}

func runGet(ctx context.Context, stdout io.Writer, id, format string) error {
	pat, err := config.Load()
	if err != nil {
		return err
	}
	c := client.New(pat.Token)
	p, err := c.GetProject(ctx, id)
	if err != nil {
		return err
	}
	return renderProject(stdout, format, *p)
}
