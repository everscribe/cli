package projects

import (
	"context"
	"fmt"
	"io"

	"github.com/spf13/cobra"

	"github.com/everscribe/cli/internal/client"
	"github.com/everscribe/cli/internal/cobrax"
	"github.com/everscribe/cli/internal/config"
)

func newDeleteCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:     "delete <id>",
		Aliases: []string{"rm"},
		Short:   "Delete a project (server-side soft delete)",
		Args:    cobrax.RequireArgs("id"),
		RunE: func(cmd *cobra.Command, args []string) error {
			return runDelete(cmd.Context(), cmd.OutOrStdout(), args[0])
		},
	}
	return cmd
}

func runDelete(ctx context.Context, stdout io.Writer, id string) error {
	pat, err := config.Load()
	if err != nil {
		return err
	}
	c := client.New(pat.Token)
	if err := c.DeleteProject(ctx, id); err != nil {
		return err
	}
	fmt.Fprintf(stdout, "Project %s deleted.\n", id)
	return nil
}
