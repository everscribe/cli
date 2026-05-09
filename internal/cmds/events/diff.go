package events

import (
	"context"
	"errors"
	"fmt"
	"io"

	"github.com/spf13/cobra"

	"github.com/everscribe/cli/internal/client"
	"github.com/everscribe/cli/internal/config"
	"github.com/everscribe/cli/internal/output"
)

func newDiffCmd() *cobra.Command {
	var project string
	cmd := &cobra.Command{
		Use:   "diff <event-id>",
		Short: "Show the before/after diff for an event with a change record",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			projectID, err := config.ResolveProjectID(project)
			if err != nil {
				return err
			}
			return runDiff(cmd.Context(), cmd.OutOrStdout(), projectID, args[0])
		},
	}
	cmd.Flags().StringVar(&project, "project", "", "project ID (defaults to the project saved by 'es projects use')")
	return cmd
}

func runDiff(ctx context.Context, stdout io.Writer, projectID, eventID string) error {
	pat, err := config.Load()
	if err != nil {
		return err
	}
	c := client.New(pat.Token)
	e, err := c.GetEvent(ctx, projectID, eventID)
	if err != nil {
		return err
	}

	sty := output.NewStylist(stdout)
	if err := renderDiff(stdout, e.Change, sty); err != nil {
		if errors.Is(err, ErrNoChange) {
			return fmt.Errorf("event %s has no change record (only mutation events have before/after data)", eventID)
		}
		return err
	}
	return nil
}
