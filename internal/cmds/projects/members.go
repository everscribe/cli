package projects

import (
	"context"
	"fmt"
	"io"

	"github.com/spf13/cobra"

	"github.com/everscribe/cli/internal/client"
	"github.com/everscribe/cli/internal/config"
	"github.com/everscribe/cli/internal/output"
	"github.com/everscribe/cli/internal/types"
)

func newMembersCmd() *cobra.Command {
	var (
		project string
		format  string
	)
	cmd := &cobra.Command{
		Use:   "members",
		Short: "List who has access to a project",
		RunE: func(cmd *cobra.Command, args []string) error {
			projectID, err := config.ResolveProjectID(project)
			if err != nil {
				return err
			}
			return runMembers(cmd.Context(), cmd.OutOrStdout(), projectID, format)
		},
	}
	cmd.Flags().StringVar(&project, "project", "", "project ID (defaults to the project saved by 'es projects use')")
	cmd.Flags().StringVar(&format, "format", "table", "output format: table | json | yaml")
	return cmd
}

func runMembers(ctx context.Context, stdout io.Writer, projectID, format string) error {
	pat, err := config.Load()
	if err != nil {
		return err
	}
	c := client.New(pat.Token)
	resp, err := c.ListProjectMembers(ctx, projectID)
	if err != nil {
		return err
	}
	return renderMembers(stdout, format, resp.Members)
}

func renderMembers(w io.Writer, format string, members []types.ProjectMemberView) error {
	f, err := output.ParseFormat(format, true)
	if err != nil {
		return err
	}
	switch f {
	case output.FormatJSON:
		return output.JSON(w, members)
	case output.FormatYAML:
		return output.YAML(w, members)
	default:
		if len(members) == 0 {
			fmt.Fprintln(w, "No members.")
			return nil
		}
		tbl := output.NewTable(w)
		tbl.Header("USERNAME", "EMAIL", "ROLE")
		for _, m := range members {
			tbl.Row(m.Username, m.Email, memberRole(m))
		}
		return tbl.Flush()
	}
}

// memberRole is the effective role to display: account owner/admin show as such,
// otherwise the explicit project role.
func memberRole(m types.ProjectMemberView) string {
	switch m.AccountRole {
	case "owner":
		return "owner"
	case "admin":
		return "admin"
	}
	if m.ProjectRole == "" {
		return "—"
	}
	return m.ProjectRole
}
