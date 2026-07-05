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

// projectDescribeView is the full project detail: the project plus everyone with
// access. Members embed in JSON/YAML so `es projects describe <id> -o json | jq`
// can reach them.
type projectDescribeView struct {
	types.Project
	Members []types.ProjectMemberView `json:"members" yaml:"members"`
}

func newDescribeCmd() *cobra.Command {
	var (
		project string
		format  string
	)
	cmd := &cobra.Command{
		Use:   "describe [id]",
		Short: "Show a project's details and everyone with access",
		Long: "Prints a project's metadata plus its members and the role each holds\n" +
			"(the people you've shared it with, plus you as owner).\n\n" +
			"Target the project by positional id, --project, or - with neither - the\n" +
			"project saved by `es projects use`.",
		Args: cobra.MaximumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			// Positional id wins over --project, which falls back to the saved default.
			id := project
			if len(args) > 0 {
				id = args[0]
			}
			projectID, err := config.ResolveProjectID(id)
			if err != nil {
				return err
			}
			return runDescribe(cmd.Context(), cmd.OutOrStdout(), projectID, format)
		},
	}
	cmd.Flags().StringVar(&project, "project", "", "project ID (defaults to the project saved by 'es projects use')")
	cmd.Flags().StringVar(&format, "format", "yaml", "output format: yaml | json | table")
	return cmd
}

func runDescribe(ctx context.Context, stdout io.Writer, id, format string) error {
	pat, err := config.Load()
	if err != nil {
		return err
	}
	c := client.New(pat.Token)
	p, err := c.GetProject(ctx, id)
	if err != nil {
		return err
	}
	members, err := c.ListProjectMembers(ctx, id)
	if err != nil {
		return err
	}
	return renderProjectDescribe(stdout, format, projectDescribeView{Project: *p, Members: members.Members})
}

func renderProjectDescribe(w io.Writer, format string, v projectDescribeView) error {
	f, err := output.ParseFormat(format, true)
	if err != nil {
		return err
	}
	switch f {
	case output.FormatJSON:
		return output.JSON(w, v)
	case output.FormatYAML:
		return output.YAML(w, v)
	default:
		fmt.Fprintf(w, "ID:      %s\n", v.ID)
		fmt.Fprintf(w, "Name:    %s\n", v.Name)
		fmt.Fprintf(w, "Created: %s\n", output.Age(v.CreatedAt))
		fmt.Fprintf(w, "Updated: %s\n\n", output.Age(v.UpdatedAt))

		if len(v.Members) == 0 {
			fmt.Fprintln(w, "No members.")
			return nil
		}
		fmt.Fprintln(w, "Members:")
		tbl := output.NewTable(w)
		tbl.Header("USERNAME", "EMAIL", "ROLE")
		for _, m := range v.Members {
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
		return "-"
	}
	return m.ProjectRole
}
