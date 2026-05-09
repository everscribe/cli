package events

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

func newListCmd() *cobra.Command {
	var (
		project string
		limit   int
		all     bool
		format  string
		filters filterFlags
	)
	cmd := &cobra.Command{
		Use:   "list",
		Short: "List events in a project (most recent first)",
		RunE: func(cmd *cobra.Command, args []string) error {
			projectID, err := config.ResolveProjectID(project)
			if err != nil {
				return err
			}
			return runList(cmd.Context(), cmd.OutOrStdout(), projectID, filters, limit, all, format)
		},
	}
	cmd.Flags().StringVar(&project, "project", "", "project ID (defaults to the project saved by 'es projects use')")
	cmd.Flags().IntVar(&limit, "limit", 50, "max events per page")
	cmd.Flags().BoolVar(&all, "all", false, "fetch all pages instead of one")
	cmd.Flags().StringVar(&format, "format", "table", "output format: table | json | yaml")
	addFilterFlags(cmd, &filters)
	return cmd
}

func runList(ctx context.Context, stdout io.Writer, projectID string, ff filterFlags, limit int, all bool, format string) error {
	pat, err := config.Load()
	if err != nil {
		return err
	}
	cf, err := ff.toClientFilter()
	if err != nil {
		return err
	}
	cf.Limit = limit

	c := client.New(pat.Token)

	var allEvents []types.Event
	resp, err := c.ListEvents(ctx, projectID, cf)
	if err != nil {
		return err
	}
	allEvents = resp.Events

	if all {
		for resp.NextCursor != "" {
			cf.Cursor = resp.NextCursor
			resp, err = c.ListEvents(ctx, projectID, cf)
			if err != nil {
				return err
			}
			allEvents = append(allEvents, resp.Events...)
		}
	}

	if err := renderEventsList(stdout, format, allEvents, true); err != nil {
		return err
	}

	// Pagination hint is only useful in human (table) output. Skip
	// for json/yaml so machine consumers don't get a stray comment
	// mid-stream.
	if !all && resp.NextCursor != "" && parsedFormat(format) == output.FormatTable {
		fmt.Fprintln(stdout)
		fmt.Fprintln(stdout, "More events available. Pass --all to fetch every page.")
	}
	return nil
}

// parsedFormat is a best-effort re-parse of --format that returns the
// zero value for invalid input — used only by display-only branches
// where the renderer has already validated the flag.
func parsedFormat(s string) output.Format {
	f, _ := output.ParseFormat(s, true)
	return f
}
