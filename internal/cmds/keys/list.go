package keys

import (
	"context"
	"io"

	"github.com/spf13/cobra"

	"github.com/everscribe/cli/internal/client"
	"github.com/everscribe/cli/internal/config"
)

func newListCmd() *cobra.Command {
	var (
		project string
		format  string
	)
	cmd := &cobra.Command{
		Use:   "list",
		Short: "List active API keys for a project",
		RunE: func(cmd *cobra.Command, args []string) error {
			return runList(cmd.Context(), cmd.OutOrStdout(), project, format)
		},
	}
	cmd.Flags().StringVar(&project, "project", "", "project ID (required)")
	cmd.Flags().StringVar(&format, "format", "table", "output format: table | json | yaml")
	_ = cmd.MarkFlagRequired("project")
	return cmd
}

func runList(ctx context.Context, stdout io.Writer, projectID, format string) error {
	pat, err := config.Load()
	if err != nil {
		return err
	}
	c := client.New(pat.Token)
	ks, err := c.ListAPIKeys(ctx, projectID)
	if err != nil {
		return err
	}
	return renderKeys(stdout, format, ks)
}
