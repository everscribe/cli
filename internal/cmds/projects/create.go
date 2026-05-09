package projects

import (
	"context"
	"io"

	"github.com/spf13/cobra"

	"github.com/everscribe/cli/internal/client"
	"github.com/everscribe/cli/internal/config"
)

func newCreateCmd() *cobra.Command {
	var (
		name   string
		format string
	)
	cmd := &cobra.Command{
		Use:   "create",
		Short: "Create a new project",
		RunE: func(cmd *cobra.Command, args []string) error {
			return runCreate(cmd.Context(), cmd.OutOrStdout(), name, format)
		},
	}
	cmd.Flags().StringVar(&name, "name", "", "project name (required)")
	cmd.Flags().StringVar(&format, "format", "table", "output format: table | json | yaml")
	_ = cmd.MarkFlagRequired("name")
	return cmd
}

func runCreate(ctx context.Context, stdout io.Writer, name, format string) error {
	pat, err := config.Load()
	if err != nil {
		return err
	}
	c := client.New(pat.Token)
	p, err := c.CreateProject(ctx, name)
	if err != nil {
		return err
	}
	return renderProject(stdout, format, *p)
}
