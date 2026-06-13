package collaborators

import (
	"context"
	"fmt"
	"io"
	"strings"

	"github.com/spf13/cobra"

	"github.com/everscribe/cli/internal/client"
	"github.com/everscribe/cli/internal/cobrax"
	"github.com/everscribe/cli/internal/config"
)

func newRemoveCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:     "remove <email>",
		Aliases: []string{"rm"},
		Short:   "Remove a collaborator (revokes their access to your projects)",
		Args:    cobrax.RequireArgs("email"),
		RunE: func(cmd *cobra.Command, args []string) error {
			return runRemove(cmd.Context(), cmd.OutOrStdout(), args[0])
		},
	}
	return cmd
}

func runRemove(ctx context.Context, stdout io.Writer, email string) error {
	pat, err := config.Load()
	if err != nil {
		return err
	}
	c := client.New(pat.Token)

	// The API removes by collaborator id; resolve it from the email.
	cols, err := c.ListCollaborators(ctx)
	if err != nil {
		return err
	}
	var id string
	for _, col := range cols {
		if strings.EqualFold(col.Email, email) {
			id = col.ID
			break
		}
	}
	if id == "" {
		return fmt.Errorf("%s is not one of your collaborators", email)
	}

	if err := c.RemoveCollaborator(ctx, id); err != nil {
		return err
	}
	fmt.Fprintf(stdout, "Removed %s.\n", email)
	return nil
}
