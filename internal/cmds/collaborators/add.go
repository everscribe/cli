package collaborators

import (
	"context"
	"io"

	"github.com/spf13/cobra"

	"github.com/everscribe/cli/internal/client"
	"github.com/everscribe/cli/internal/cobrax"
	"github.com/everscribe/cli/internal/config"
)

func newAddCmd() *cobra.Command {
	var format string
	cmd := &cobra.Command{
		Use:   "add <email>",
		Short: "Add a collaborator by email",
		Long: "Adds someone to your collaborator list. If they already have an " +
			"Everscribe account they're added right away (active); otherwise they " +
			"get an invite to create one and show up as pending until they do.",
		Args: cobrax.RequireArgs("email"),
		RunE: func(cmd *cobra.Command, args []string) error {
			return runAdd(cmd.Context(), cmd.OutOrStdout(), args[0], format)
		},
	}
	cmd.Flags().StringVar(&format, "format", "table", "output format: table | json | yaml")
	return cmd
}

func runAdd(ctx context.Context, stdout io.Writer, email, format string) error {
	pat, err := config.Load()
	if err != nil {
		return err
	}
	c := client.New(pat.Token)
	col, err := c.AddCollaborator(ctx, email)
	if err != nil {
		return err
	}
	return renderCollaborator(stdout, format, *col)
}
