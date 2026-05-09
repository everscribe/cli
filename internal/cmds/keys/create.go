package keys

import (
	"context"
	"io"

	"github.com/spf13/cobra"

	"github.com/everscribe/cli/internal/client"
	"github.com/everscribe/cli/internal/config"
)

func newCreateCmd() *cobra.Command {
	var (
		project string
		name    string
		format  string
	)
	cmd := &cobra.Command{
		Use:   "create",
		Short: "Create a new ingest API key in a project",
		RunE: func(cmd *cobra.Command, args []string) error {
			projectID, err := config.ResolveProjectID(project)
			if err != nil {
				return err
			}
			return runCreate(cmd.Context(), cmd.OutOrStdout(), projectID, name, format)
		},
	}
	cmd.Flags().StringVar(&project, "project", "", "project ID (defaults to the project saved by 'es projects use')")
	cmd.Flags().StringVar(&name, "name", "", "key name (required)")
	cmd.Flags().StringVar(&format, "format", "table", "output format: table | json | yaml")
	_ = cmd.MarkFlagRequired("name")
	return cmd
}

func runCreate(ctx context.Context, stdout io.Writer, projectID, name, format string) error {
	pat, err := config.Load()
	if err != nil {
		return err
	}
	c := client.New(pat.Token)
	resp, err := c.CreateAPIKey(ctx, projectID, name)
	if err != nil {
		return err
	}
	return renderCreatedKey(stdout, format, resp)
}
