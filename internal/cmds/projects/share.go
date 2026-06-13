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

func newShareCmd() *cobra.Command {
	var (
		project string
		role    string
	)
	cmd := &cobra.Command{
		Use:   "share <email>",
		Short: "Share a project with a collaborator",
		Long: "Grants an active collaborator access to a project at the given role.\n" +
			"Add them with `es collaborators add` first.",
		Args: cobrax.RequireArgs("email"),
		RunE: func(cmd *cobra.Command, args []string) error {
			projectID, err := config.ResolveProjectID(project)
			if err != nil {
				return err
			}
			return runShare(cmd.Context(), cmd.OutOrStdout(), projectID, args[0], role)
		},
	}
	cmd.Flags().StringVar(&project, "project", "", "project ID (defaults to the project saved by 'es projects use')")
	cmd.Flags().StringVar(&role, "role", "viewer", "access role: viewer | editor | admin")
	return cmd
}

func runShare(ctx context.Context, stdout io.Writer, projectID, email, role string) error {
	if !validProjectRole(role) {
		return fmt.Errorf("role must be viewer, editor, or admin")
	}
	pat, err := config.Load()
	if err != nil {
		return err
	}
	c := client.New(pat.Token)

	userID, err := activeCollaboratorUserID(ctx, c, email)
	if err != nil {
		return err
	}
	if err := c.SetProjectRole(ctx, projectID, userID, role); err != nil {
		return err
	}
	fmt.Fprintf(stdout, "Shared project %s with %s as %s.\n", projectID, email, role)
	return nil
}
