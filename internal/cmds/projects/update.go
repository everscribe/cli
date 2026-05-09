package projects

import (
	"context"
	"io"

	"github.com/spf13/cobra"

	"github.com/everscribe/cli/internal/client"
	"github.com/everscribe/cli/internal/config"
)

func newUpdateCmd() *cobra.Command {
	var (
		name   string
		format string
	)
	cmd := &cobra.Command{
		Use:   "update <id>",
		Short: "Update an existing project",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			return runUpdate(cmd.Context(), cmd.OutOrStdout(), args[0], name, format)
		},
	}
	cmd.Flags().StringVar(&name, "name", "", "new project name (required)")
	cmd.Flags().StringVar(&format, "format", "table", "output format: table | json | yaml")
	_ = cmd.MarkFlagRequired("name")
	return cmd
}

func runUpdate(ctx context.Context, stdout io.Writer, id, name, format string) error {
	pat, err := config.Load()
	if err != nil {
		return err
	}
	c := client.New(pat.Token)
	p, err := c.UpdateProject(ctx, id, name)
	if err != nil {
		return err
	}
	return renderProject(stdout, format, *p)
}
