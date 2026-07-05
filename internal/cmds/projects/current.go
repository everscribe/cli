package projects

import (
	"context"
	"fmt"
	"io"

	"github.com/spf13/cobra"

	"github.com/everscribe/cli/internal/client"
	"github.com/everscribe/cli/internal/config"
)

func newCurrentCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "current",
		Short: "Print the default project ID and name, if one is set",
		RunE: func(cmd *cobra.Command, args []string) error {
			return runCurrent(cmd.Context(), cmd.OutOrStdout())
		},
	}
	return cmd
}

func runCurrent(ctx context.Context, stdout io.Writer) error {
	cfg, err := config.LoadConfig()
	if err != nil {
		return err
	}
	if cfg.DefaultProjectID == "" {
		fmt.Fprintln(stdout, "No default project set. Run `es projects use <id>` to pick one.")
		return nil
	}

	// Try to enrich the saved ID with the project's name. If the user
	// is logged out or the project is unreachable (404 / network),
	// fall back to printing the bare ID rather than failing - the
	// command's primary value is "tell me what's saved", and the saved
	// value is always the ID.
	pat, err := config.Load()
	if err != nil {
		fmt.Fprintln(stdout, cfg.DefaultProjectID)
		return nil
	}
	p, err := client.New(pat.Token).GetProject(ctx, cfg.DefaultProjectID)
	if err != nil {
		fmt.Fprintln(stdout, cfg.DefaultProjectID)
		return nil
	}
	fmt.Fprintf(stdout, "%s (%s)\n", p.ID, p.Name)
	return nil
}
