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
			return runCreate(cmd.Context(), cmd.OutOrStdout(), project, name, format)
		},
	}
	cmd.Flags().StringVar(&project, "project", "", "project ID (required)")
	cmd.Flags().StringVar(&name, "name", "", "key name (required)")
	cmd.Flags().StringVar(&format, "format", "table", "output format: table | json | yaml")
	_ = cmd.MarkFlagRequired("project")
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
