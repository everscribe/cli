package events

import (
	"context"
	"fmt"
	"io"

	"github.com/spf13/cobra"

	"github.com/everscribe/cli/internal/client"
	"github.com/everscribe/cli/internal/cobrax"
	"github.com/everscribe/cli/internal/config"
	"github.com/everscribe/cli/internal/output"
)

func newDescribeCmd() *cobra.Command {
	var (
		project string
		format  string
	)
	cmd := &cobra.Command{
		Use:   "describe <event-id>",
		Short: "Print the full event with all nested fields decoded",
		Args:  cobrax.RequireArgs("event-id"),
		RunE: func(cmd *cobra.Command, args []string) error {
			projectID, err := config.ResolveProjectID(project)
			if err != nil {
				return err
			}
			return runDescribe(cmd.Context(), cmd.OutOrStdout(), projectID, args[0], format)
		},
	}
	cmd.Flags().StringVar(&project, "project", "", "project ID (defaults to the project saved by 'es projects use')")
	cmd.Flags().StringVar(&format, "format", "yaml", "output format: yaml | json")
	return cmd
}

func runDescribe(ctx context.Context, stdout io.Writer, projectID, eventID, format string) error {
	f, err := output.ParseFormat(format, false)
	if err != nil {
		return err
	}
	pat, err := config.Load()
	if err != nil {
		return err
	}
	c := client.New(pat.Token)
	e, err := c.GetEvent(ctx, projectID, eventID)
	if err != nil {
		return err
	}

	view := buildDescribeView(*e)
	switch f {
	case output.FormatJSON:
		return output.JSON(stdout, view)
	case output.FormatYAML:
		return output.YAML(stdout, view)
	default:
		return fmt.Errorf("unexpected format %q", f)
	}
}
