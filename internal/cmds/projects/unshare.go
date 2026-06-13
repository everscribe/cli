package projects

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

func newUnshareCmd() *cobra.Command {
	var project string
	cmd := &cobra.Command{
		Use:   "unshare <email>",
		Short: "Revoke a collaborator's access to a project",
		Args:  cobrax.RequireArgs("email"),
		RunE: func(cmd *cobra.Command, args []string) error {
			projectID, err := config.ResolveProjectID(project)
			if err != nil {
				return err
			}
			return runUnshare(cmd.Context(), cmd.OutOrStdout(), projectID, args[0])
		},
	}
	cmd.Flags().StringVar(&project, "project", "", "project ID (defaults to the project saved by 'es projects use')")
	return cmd
}

func runUnshare(ctx context.Context, stdout io.Writer, projectID, email string) error {
	pat, err := config.Load()
	if err != nil {
		return err
	}
	c := client.New(pat.Token)

	// Resolve the email to a user id from the project's current members.
	resp, err := c.ListProjectMembers(ctx, projectID)
	if err != nil {
		return err
	}
	var userID string
	for _, m := range resp.Members {
		if strings.EqualFold(m.Email, email) {
			userID = m.UserID
			break
		}
	}
	if userID == "" {
		return fmt.Errorf("%s does not have access to project %s", email, projectID)
	}
	if err := c.RemoveProjectMember(ctx, projectID, userID); err != nil {
		return err
	}
	fmt.Fprintf(stdout, "Removed %s from project %s.\n", email, projectID)
	return nil
}
