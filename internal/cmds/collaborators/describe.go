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
	"github.com/everscribe/cli/internal/output"
	"github.com/everscribe/cli/internal/types"
)

func newDescribeCmd() *cobra.Command {
	var format string
	cmd := &cobra.Command{
		Use:   "describe <email>",
		Short: "Show a collaborator and the projects they can access",
		Long: "Prints a collaborator's details plus every project you've shared with\n" +
			"them and the role they hold on each.",
		Args: cobrax.RequireArgs("email"),
		RunE: func(cmd *cobra.Command, args []string) error {
			return runDescribe(cmd.Context(), cmd.OutOrStdout(), args[0], format)
		},
	}
	cmd.Flags().StringVar(&format, "format", "yaml", "output format: yaml | json | table")
	return cmd
}

func runDescribe(ctx context.Context, stdout io.Writer, email, format string) error {
	pat, err := config.Load()
	if err != nil {
		return err
	}
	c := client.New(pat.Token)

	// The list endpoint already returns each collaborator's shared projects, so
	// describe is a client-side lookup by email — no separate call needed.
	cols, err := c.ListCollaborators(ctx)
	if err != nil {
		return err
	}
	var found *types.Collaborator
	for i := range cols {
		if strings.EqualFold(cols[i].Email, email) {
			found = &cols[i]
			break
		}
	}
	if found == nil {
		return fmt.Errorf("%s is not one of your collaborators", email)
	}
	return renderCollaboratorDetail(stdout, format, *found)
}

func renderCollaboratorDetail(w io.Writer, format string, col types.Collaborator) error {
	f, err := output.ParseFormat(format, true)
	if err != nil {
		return err
	}
	switch f {
	case output.FormatJSON:
		return output.JSON(w, col)
	case output.FormatYAML:
		return output.YAML(w, col)
	default:
		fmt.Fprintf(w, "Email:    %s\n", col.Email)
		if col.Username != "" {
			fmt.Fprintf(w, "Username: %s\n", col.Username)
		}
		fmt.Fprintf(w, "Status:   %s\n", col.Status)
		fmt.Fprintf(w, "Added:    %s\n\n", output.Age(col.CreatedAt))

		if len(col.Projects) == 0 {
			fmt.Fprintln(w, "No projects shared with this collaborator.")
			return nil
		}
		fmt.Fprintln(w, "Projects:")
		tbl := output.NewTable(w)
		tbl.Header("NAME", "ROLE", "PROJECT-ID")
		for _, p := range col.Projects {
			tbl.Row(p.Name, p.Role, p.ID)
		}
		return tbl.Flush()
	}
}
